package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Profile represents a git identity profile.
type Profile struct {
	Name      string `json:"name"`
	Directory string `json:"directory,omitempty"`
	Email     string `json:"email"`
	SignKey   string `json:"sign_key,omitempty"`
	Active    bool   `json:"active"`
}

// Issue represents a problem found in gitconfig.
type Issue struct {
	Severity string `json:"severity"` // "ok", "warn", "error"
	Category string `json:"category"`
	Message  string `json:"message"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

// DoctorResult holds all findings from git doctor.
type DoctorResult struct {
	Issues []Issue `json:"issues"`
}

// CleanResult holds the results of a clean operation.
type CleanResult struct {
	Removed   []string `json:"removed"`
	File      string   `json:"file"`
	DryRun    bool     `json:"dry_run"`
	BackupDir string   `json:"backup_dir,omitempty"`
}

// IncludeIfEntry represents a conditional include in gitconfig.
type IncludeIfEntry struct {
	Condition string
	Path      string
	Line      int
}

// SafeDirEntry represents a safe.directory entry.
type SafeDirEntry struct {
	Directory string
	File      string
	Line      int
}

// ParseGitConfigFile parses a gitconfig file and returns its sections and key-value pairs.
func ParseGitConfigFile(path string) ([]Section, error) {
	path = expandPath(path)
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %w", path, err)
	}
	defer f.Close()

	var sections []Section
	var current *Section
	scanner := bufio.NewScanner(f)
	lineNum := 0

	sectionRe := regexp.MustCompile(`^\s*\[([^\]]+)\]\s*$`)
	kvRe := regexp.MustCompile(`^\s*(\S+)\s*=\s*(.*)$`)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if m := sectionRe.FindStringSubmatch(line); m != nil {
			sections = append(sections, Section{
				Header: m[1],
				Line:   lineNum,
			})
			current = &sections[len(sections)-1]
			continue
		}

		if m := kvRe.FindStringSubmatch(line); m != nil && current != nil {
			current.Entries = append(current.Entries, Entry{
				Key:   m[1],
				Value: strings.TrimSpace(m[2]),
				Line:  lineNum,
			})
		}
	}

	return sections, scanner.Err()
}

// Section represents a gitconfig section.
type Section struct {
	Header  string
	Line    int
	Entries []Entry
}

// Entry represents a key-value pair in a section.
type Entry struct {
	Key   string
	Value string
	Line  int
}

// GetProfiles reads git profiles from includeIf directives in gitconfig.
func GetProfiles(gitconfigPath string) ([]Profile, error) {
	path := expandPath(gitconfigPath)
	sections, err := ParseGitConfigFile(path)
	if err != nil {
		return nil, err
	}

	// Get current working directory for active detection.
	cwd, _ := os.Getwd()

	var profiles []Profile

	for _, sec := range sections {
		// Check for includeIf sections (case-insensitive, git allows includeIF, includeIf, etc.).
		headerLower := strings.ToLower(sec.Header)
		if !strings.HasPrefix(headerLower, "includeif ") {
			continue
		}

		// Parse condition: includeIf "gitdir:~/work/"
		condition := sec.Header[len("includeIf "):]
		condition = strings.Trim(condition, "\"")

		if !strings.HasPrefix(strings.ToLower(condition), "gitdir:") {
			continue
		}

		dir := condition[len("gitdir:"):]
		dir = strings.TrimSuffix(dir, "/")

		var includePath string
		for _, e := range sec.Entries {
			if e.Key == "path" {
				includePath = expandPath(e.Value)
			}
		}

		if includePath == "" {
			continue
		}

		// Parse the included config for email and signing key.
		email, signKey := parseIncludedConfig(includePath)

		// Determine profile name from directory.
		name := filepath.Base(expandPath(dir))

		active := false
		if cwd != "" {
			expandedDir := expandPath(dir)
			cleanDir := filepath.Clean(expandedDir)
			if cwd == cleanDir || strings.HasPrefix(cwd, cleanDir+string(filepath.Separator)) {
				active = true
			}
		}

		profiles = append(profiles, Profile{
			Name:      name,
			Directory: dir,
			Email:     email,
			SignKey:   signKey,
			Active:    active,
		})
	}

	// Also check global user config.
	globalEmail := getGitConfigValue(sections, "user", "email")
	globalName := getGitConfigValue(sections, "user", "name")
	if globalEmail != "" {
		p := Profile{
			Name:  "global",
			Email: globalEmail,
		}
		if globalName != "" {
			p.Name = "global (" + globalName + ")"
		}
		// Global is active if no other profile is active.
		hasActive := false
		for _, pp := range profiles {
			if pp.Active {
				hasActive = true
				break
			}
		}
		if !hasActive {
			p.Active = true
		}
		// Prepend global profile.
		profiles = append([]Profile{p}, profiles...)
	}

	return profiles, nil
}

// Doctor diagnoses common gitconfig issues.
func Doctor(gitconfigPath string) (*DoctorResult, error) {
	path := expandPath(gitconfigPath)
	result := &DoctorResult{}

	sections, err := ParseGitConfigFile(path)
	if err != nil {
		result.Issues = append(result.Issues, Issue{
			Severity: "error",
			Category: "config",
			Message:  fmt.Sprintf("Cannot read %s: %s", path, err),
		})
		return result, nil
	}

	// Check 1: Duplicate safe.directory entries.
	checkDuplicateSafeDirs(sections, path, result)

	// Check 2: Missing or invalid includeIf targets.
	checkIncludeIfs(sections, path, result)

	// Check 3: No user.email configured.
	checkUserIdentity(sections, path, result)

	// Check 4: Duplicate remote URLs.
	checkDuplicateRemotes(sections, path, result)

	// Check 5: File permissions.
	checkFilePermissions(path, result)

	if len(result.Issues) == 0 {
		result.Issues = append(result.Issues, Issue{
			Severity: "ok",
			Category: "all",
			Message:  "No issues found in gitconfig",
			File:     path,
		})
	}

	return result, nil
}

// FindDuplicateSafeDirs returns duplicate safe.directory entries.
func FindDuplicateSafeDirs(sections []Section) []SafeDirEntry {
	seen := make(map[string]SafeDirEntry)
	var duplicates []SafeDirEntry

	for _, sec := range sections {
		if sec.Header != "safe" {
			continue
		}
		for _, e := range sec.Entries {
			if e.Key != "directory" {
				continue
			}
			dir := e.Value
			if _, exists := seen[dir]; exists {
				duplicates = append(duplicates, SafeDirEntry{
					Directory: dir,
					Line:      e.Line,
				})
			} else {
				seen[dir] = SafeDirEntry{
					Directory: dir,
					Line:      e.Line,
				}
			}
		}
	}

	return duplicates
}

// Clean removes duplicate safe.directory entries from gitconfig.
func Clean(gitconfigPath string, dryRun bool) (*CleanResult, error) {
	path := expandPath(gitconfigPath)
	result := &CleanResult{
		File:   path,
		DryRun: dryRun,
	}

	sections, err := ParseGitConfigFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s: %w", path, err)
	}

	// Find duplicate safe.directory lines.
	removeLines := make(map[int]bool)
	duplicates := FindDuplicateSafeDirs(sections)
	for _, d := range duplicates {
		removeLines[d.Line] = true
		result.Removed = append(result.Removed, fmt.Sprintf("safe.directory=%s (line %d)", d.Directory, d.Line))
	}

	// Also find and remove stale safe.directory entries (dirs that don't exist).
	for _, sec := range sections {
		if sec.Header != "safe" {
			continue
		}
		for _, e := range sec.Entries {
			if e.Key != "directory" || removeLines[e.Line] {
				continue
			}
			dir := expandPath(e.Value)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				removeLines[e.Line] = true
				result.Removed = append(result.Removed, fmt.Sprintf("safe.directory=%s (line %d, directory not found)", e.Value, e.Line))
			}
		}
	}

	if len(result.Removed) == 0 {
		return result, nil
	}

	if dryRun {
		return result, nil
	}

	// Create backup.
	backupPath := path + ".bak"
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return nil, fmt.Errorf("cannot create backup: %w", err)
	}
	result.BackupDir = backupPath

	// Rewrite file without duplicate/stale lines.
	lines := strings.Split(string(data), "\n")
	var newLines []string
	for i, line := range lines {
		lineNum := i + 1
		if removeLines[lineNum] {
			continue
		}
		newLines = append(newLines, line)
	}

	// Remove empty [safe] sections.
	output := removeEmptySections(strings.Join(newLines, "\n"), "safe")

	if err := os.WriteFile(path, []byte(output), 0600); err != nil {
		return nil, fmt.Errorf("cannot write %s: %w", path, err)
	}

	return result, nil
}

// Helper functions.

func checkDuplicateSafeDirs(sections []Section, file string, result *DoctorResult) {
	seen := make(map[string]int)

	for _, sec := range sections {
		if sec.Header != "safe" {
			continue
		}
		for _, e := range sec.Entries {
			if e.Key == "directory" {
				seen[e.Value]++
			}
		}
	}

	for dir, count := range seen {
		if count > 1 {
			result.Issues = append(result.Issues, Issue{
				Severity: "warn",
				Category: "safe.directory",
				Message:  fmt.Sprintf("Duplicate safe.directory: %s (%d times)", dir, count),
				File:     file,
			})
		}
	}
}

func checkIncludeIfs(sections []Section, file string, result *DoctorResult) {
	for _, sec := range sections {
		if !strings.HasPrefix(strings.ToLower(sec.Header), "includeif ") {
			continue
		}

		for _, e := range sec.Entries {
			if e.Key != "path" {
				continue
			}
			target := expandPath(e.Value)
			if _, err := os.Stat(target); os.IsNotExist(err) {
				result.Issues = append(result.Issues, Issue{
					Severity: "warn",
					Category: "includeIf",
					Message:  fmt.Sprintf("Include target not found: %s", e.Value),
					File:     file,
					Line:     e.Line,
				})
			}
		}
	}
}

func checkUserIdentity(sections []Section, file string, result *DoctorResult) {
	hasEmail := false
	for _, sec := range sections {
		if sec.Header != "user" {
			continue
		}
		for _, e := range sec.Entries {
			if e.Key == "email" {
				hasEmail = true
			}
		}
	}

	if !hasEmail {
		// Check if there are includeIf profiles that provide email.
		hasIncludeIf := false
		for _, sec := range sections {
			if strings.HasPrefix(strings.ToLower(sec.Header), "includeif ") {
				hasIncludeIf = true
				break
			}
		}
		if hasIncludeIf {
			result.Issues = append(result.Issues, Issue{
				Severity: "ok",
				Category: "user.identity",
				Message:  "No global user.email (using includeIf profiles)",
				File:     file,
			})
		} else {
			result.Issues = append(result.Issues, Issue{
				Severity: "warn",
				Category: "user.identity",
				Message:  "No user.email configured",
				File:     file,
			})
		}
	}
}

func checkDuplicateRemotes(sections []Section, file string, result *DoctorResult) {
	remoteURLs := make(map[string][]string)
	for _, sec := range sections {
		if !strings.HasPrefix(sec.Header, "remote ") {
			continue
		}
		remoteName := strings.Trim(strings.TrimPrefix(sec.Header, "remote "), "\"")
		for _, e := range sec.Entries {
			if e.Key == "url" {
				remoteURLs[e.Value] = append(remoteURLs[e.Value], remoteName)
			}
		}
	}

	for url, remotes := range remoteURLs {
		if len(remotes) > 1 {
			result.Issues = append(result.Issues, Issue{
				Severity: "warn",
				Category: "remote",
				Message:  fmt.Sprintf("Duplicate remote URL: %s (in: %s)", url, strings.Join(remotes, ", ")),
				File:     file,
			})
		}
	}
}

func checkFilePermissions(path string, result *DoctorResult) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	perm := info.Mode().Perm()
	if perm&0077 != 0 {
		result.Issues = append(result.Issues, Issue{
			Severity: "warn",
			Category: "permissions",
			Message:  fmt.Sprintf("Gitconfig too open: %s (mode %04o, recommend 0600)", path, perm),
			File:     path,
		})
	}
}

func parseIncludedConfig(path string) (email, signKey string) {
	sections, err := ParseGitConfigFile(path)
	if err != nil {
		return "", ""
	}
	for _, sec := range sections {
		if sec.Header != "user" {
			continue
		}
		for _, e := range sec.Entries {
			if e.Key == "email" {
				email = e.Value
			}
			if e.Key == "signingkey" {
				signKey = e.Value
			}
		}
	}
	return
}

func getGitConfigValue(sections []Section, section, key string) string {
	for _, sec := range sections {
		if sec.Header != section {
			continue
		}
		for _, e := range sec.Entries {
			if e.Key == key {
				return e.Value
			}
		}
	}
	return ""
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

func removeEmptySections(content, sectionName string) string {
	lines := strings.Split(content, "\n")
	var result []string
	sectionRe := regexp.MustCompile(`^\s*\[` + regexp.QuoteMeta(sectionName) + `\]\s*$`)

	i := 0
	for i < len(lines) {
		if sectionRe.MatchString(lines[i]) {
			// Check if next non-empty lines are another section or EOF.
			j := i + 1
			for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
				j++
			}
			if j >= len(lines) || (len(strings.TrimSpace(lines[j])) > 0 && strings.TrimSpace(lines[j])[0] == '[') {
				// Empty section — skip it and trailing blank lines.
				i = j
				continue
			}
		}
		result = append(result, lines[i])
		i++
	}

	return strings.Join(result, "\n")
}

// DefaultGitConfigPath returns the default gitconfig path.
func DefaultGitConfigPath() string {
	// Try git config --list --show-origin to find the global config.
	out, err := exec.Command("git", "config", "--global", "--list", "--show-origin").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "file:") {
				parts := strings.SplitN(line, "\t", 2)
				if len(parts) >= 1 {
					return strings.TrimPrefix(parts[0], "file:")
				}
			}
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "~/.gitconfig"
	}
	return filepath.Join(home, ".gitconfig")
}
