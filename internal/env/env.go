package env

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// CheckResult holds all findings from env check.
type CheckResult struct {
	Findings []Finding `json:"findings"`
}

// Finding represents a single check result.
type Finding struct {
	Severity string `json:"severity"` // "ok", "warn", "error"
	Category string `json:"category"`
	Message  string `json:"message"`
}

// PathResult holds PATH analysis results.
type PathResult struct {
	Entries    []PathEntry `json:"entries"`
	Duplicates []string    `json:"duplicates,omitempty"`
	Missing    []string    `json:"missing,omitempty"`
	Total      int         `json:"total"`
}

// PathEntry represents a single PATH directory.
type PathEntry struct {
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	Index  int    `json:"index"`
}

// Check performs a shell environment health scan.
func Check() *CheckResult {
	result := &CheckResult{}

	checkShell(result)
	checkCommonTools(result)
	checkSSHAgent(result)
	checkGitConfig(result)
	checkHomeDir(result)
	checkEditorSet(result)

	return result
}

// AnalyzePath analyzes the PATH environment variable.
func AnalyzePath() *PathResult {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return &PathResult{}
	}

	dirs := strings.Split(pathEnv, ":")
	result := &PathResult{
		Total: len(dirs),
	}

	seen := make(map[string]int) // path -> first occurrence index
	for i, dir := range dirs {
		expanded := expandEnvPath(dir)
		entry := PathEntry{
			Path:  dir,
			Index: i + 1,
		}

		if info, err := os.Stat(expanded); err == nil && info.IsDir() {
			entry.Exists = true
		} else {
			entry.Exists = false
			result.Missing = append(result.Missing, dir)
		}

		result.Entries = append(result.Entries, entry)

		// Track duplicates.
		canonical := filepath.Clean(expanded)
		if firstIdx, exists := seen[canonical]; exists {
			dup := fmt.Sprintf("%s (index %d and %d)", dir, firstIdx, i+1)
			// Only add if not already reported.
			found := false
			for _, d := range result.Duplicates {
				if strings.HasPrefix(d, dir+" ") || strings.HasPrefix(d, canonical+" ") {
					found = true
					break
				}
			}
			if !found {
				result.Duplicates = append(result.Duplicates, dup)
			}
		} else {
			seen[canonical] = i + 1
		}
	}

	return result
}

// Helper functions.

func checkShell(result *CheckResult) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		result.Findings = append(result.Findings, Finding{
			Severity: "warn",
			Category: "shell",
			Message:  "SHELL environment variable not set",
		})
		return
	}

	result.Findings = append(result.Findings, Finding{
		Severity: "ok",
		Category: "shell",
		Message:  fmt.Sprintf("Shell: %s", shell),
	})

	// Check bash version if using bash.
	if strings.Contains(shell, "bash") {
		out, err := exec.Command(shell, "--version").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				version := lines[0]
				if strings.Contains(version, "version 3.") {
					result.Findings = append(result.Findings, Finding{
						Severity: "warn",
						Category: "shell",
						Message:  fmt.Sprintf("Old bash version detected: %s (recommend 5.x via Homebrew)", version),
					})
				} else {
					result.Findings = append(result.Findings, Finding{
						Severity: "ok",
						Category: "shell",
						Message:  fmt.Sprintf("Bash version: %s", version),
					})
				}
			}
		}
	}
}

func checkCommonTools(result *CheckResult) {
	tools := []struct {
		name     string
		required bool
	}{
		{"git", true},
		{"ssh", true},
		{"curl", true},
		{"make", false},
		{"docker", false},
		{"kubectl", false},
		{"helm", false},
		{"terraform", false},
		{"go", false},
		{"node", false},
		{"python3", false},
	}

	for _, tool := range tools {
		path, err := exec.LookPath(tool.name)
		if err != nil {
			sev := "warn"
			if tool.required {
				sev = "error"
			}
			result.Findings = append(result.Findings, Finding{
				Severity: sev,
				Category: "tools",
				Message:  fmt.Sprintf("%s: not found", tool.name),
			})
		} else {
			result.Findings = append(result.Findings, Finding{
				Severity: "ok",
				Category: "tools",
				Message:  fmt.Sprintf("%s: %s", tool.name, path),
			})
		}
	}
}

func checkSSHAgent(result *CheckResult) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		result.Findings = append(result.Findings, Finding{
			Severity: "warn",
			Category: "ssh-agent",
			Message:  "SSH_AUTH_SOCK not set (ssh-agent not running)",
		})
		return
	}

	if _, err := os.Stat(sock); os.IsNotExist(err) {
		result.Findings = append(result.Findings, Finding{
			Severity: "warn",
			Category: "ssh-agent",
			Message:  fmt.Sprintf("SSH_AUTH_SOCK points to missing socket: %s", sock),
		})
		return
	}

	// Check loaded keys.
	out, err := exec.Command("ssh-add", "-l").Output()
	if err != nil {
		result.Findings = append(result.Findings, Finding{
			Severity: "warn",
			Category: "ssh-agent",
			Message:  "ssh-agent running but no keys loaded",
		})
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	result.Findings = append(result.Findings, Finding{
		Severity: "ok",
		Category: "ssh-agent",
		Message:  fmt.Sprintf("ssh-agent: %d key(s) loaded", len(lines)),
	})
}

func checkGitConfig(result *CheckResult) {
	out, err := exec.Command("git", "config", "--global", "user.email").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		result.Findings = append(result.Findings, Finding{
			Severity: "warn",
			Category: "git",
			Message:  "No global git user.email configured",
		})
	} else {
		result.Findings = append(result.Findings, Finding{
			Severity: "ok",
			Category: "git",
			Message:  fmt.Sprintf("git user.email: %s", strings.TrimSpace(string(out))),
		})
	}

	out, err = exec.Command("git", "config", "--global", "user.name").Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		result.Findings = append(result.Findings, Finding{
			Severity: "warn",
			Category: "git",
			Message:  "No global git user.name configured",
		})
	} else {
		result.Findings = append(result.Findings, Finding{
			Severity: "ok",
			Category: "git",
			Message:  fmt.Sprintf("git user.name: %s", strings.TrimSpace(string(out))),
		})
	}
}

func checkHomeDir(result *CheckResult) {
	home, err := os.UserHomeDir()
	if err != nil {
		result.Findings = append(result.Findings, Finding{
			Severity: "error",
			Category: "home",
			Message:  "Cannot determine home directory",
		})
		return
	}

	// Check common config dirs.
	dirs := []string{
		filepath.Join(home, ".ssh"),
		filepath.Join(home, ".config"),
	}

	for _, dir := range dirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			perm := info.Mode().Perm()
			if dir == filepath.Join(home, ".ssh") && perm&0077 != 0 {
				result.Findings = append(result.Findings, Finding{
					Severity: "warn",
					Category: "home",
					Message:  fmt.Sprintf("%s: permissions too open (%04o, recommend 0700)", dir, perm),
				})
			} else {
				result.Findings = append(result.Findings, Finding{
					Severity: "ok",
					Category: "home",
					Message:  fmt.Sprintf("%s: OK", dir),
				})
			}
		} else {
			result.Findings = append(result.Findings, Finding{
				Severity: "warn",
				Category: "home",
				Message:  fmt.Sprintf("%s: not found", dir),
			})
		}
	}

	// Check OS-specific profile.
	var profileFile string
	if runtime.GOOS == "darwin" {
		profileFile = filepath.Join(home, ".bash_profile")
	} else {
		profileFile = filepath.Join(home, ".bashrc")
	}

	if _, err := os.Stat(profileFile); err == nil {
		result.Findings = append(result.Findings, Finding{
			Severity: "ok",
			Category: "home",
			Message:  fmt.Sprintf("%s: exists", profileFile),
		})
	} else {
		result.Findings = append(result.Findings, Finding{
			Severity: "warn",
			Category: "home",
			Message:  fmt.Sprintf("%s: not found", profileFile),
		})
	}
}

func checkEditorSet(result *CheckResult) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		result.Findings = append(result.Findings, Finding{
			Severity: "warn",
			Category: "editor",
			Message:  "EDITOR/VISUAL not set (default editor undefined)",
		})
		return
	}

	result.Findings = append(result.Findings, Finding{
		Severity: "ok",
		Category: "editor",
		Message:  fmt.Sprintf("Editor: %s", editor),
	})
}

func expandEnvPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return os.ExpandEnv(path)
}

// SummarizeFindings returns counts by severity.
func SummarizeFindings(findings []Finding) (ok, warn, errCount int) {
	for _, f := range findings {
		switch f.Severity {
		case "ok":
			ok++
		case "warn":
			warn++
		case "error":
			errCount++
		}
	}
	return
}

// GroupFindingsByCategory groups findings by category and returns sorted keys.
func GroupFindingsByCategory(findings []Finding) (map[string][]Finding, []string) {
	groups := make(map[string][]Finding)
	for _, f := range findings {
		groups[f.Category] = append(groups[f.Category], f)
	}

	var keys []string
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return groups, keys
}
