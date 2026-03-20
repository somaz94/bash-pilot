package migrate

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ImportResult holds the result of an import operation.
type ImportResult struct {
	SSHConfigWritten bool        `json:"ssh_config_written"`
	SSHHostsAdded    int         `json:"ssh_hosts_added"`
	SSHHostsSkipped  int         `json:"ssh_hosts_skipped"`
	SSHKeysNeeded    []KeyAction `json:"ssh_keys_needed,omitempty"`
	GitConfigWritten bool        `json:"git_config_written"`
	DirsCreated      []string    `json:"dirs_created,omitempty"`
	ProfilesWritten  []string    `json:"profiles_written,omitempty"`
	Warnings         []string    `json:"warnings,omitempty"`
}

// KeyAction describes an SSH key that needs to be generated.
type KeyAction struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Command string `json:"command"`
}

// Function variables for testing.
var (
	writeFile    = os.WriteFile
	readFile     = os.ReadFile
	mkdirAll     = os.MkdirAll
	statFile     = os.Stat
	openFile     = os.Open
	runGitConfig = func(args ...string) error {
		allArgs := append([]string{"config", "--global"}, args...)
		_, err := runCommand("git", allArgs...)
		return err
	}
)

// Import applies the migrate config to the current machine.
func Import(cfg *MigrateConfig, dryRun bool) (*ImportResult, error) {
	home, err := userHomeDir()
	if err != nil {
		return nil, err
	}

	result := &ImportResult{}

	importSSH(cfg, home, dryRun, result)
	importGit(cfg, home, dryRun, result)

	return result, nil
}

func importSSH(cfg *MigrateConfig, home string, dryRun bool, result *ImportResult) {
	sshDir := filepath.Join(home, ".ssh")

	// Ensure ~/.ssh exists.
	if !dryRun {
		mkdirAll(sshDir, 0700)
	}

	// Read existing SSH config to detect duplicates.
	sshConfigPath := filepath.Join(sshDir, "config")
	existingHosts := parseExistingHosts(sshConfigPath)

	// Build new host blocks.
	var newBlocks []string
	for _, h := range cfg.SSH.Hosts {
		if _, exists := existingHosts[h.Name]; exists {
			result.SSHHostsSkipped++
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Host '%s' already exists in SSH config, skipping", h.Name))
			continue
		}

		block := buildHostBlock(h, home)
		newBlocks = append(newBlocks, block)
		result.SSHHostsAdded++
	}

	if len(newBlocks) > 0 && !dryRun {
		// Append to existing config.
		content := "\n# Imported by bash-pilot migrate\n" + strings.Join(newBlocks, "\n")

		f, err := os.OpenFile(sshConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			f.WriteString(content)
			f.Close()
			result.SSHConfigWritten = true
		}
	} else if len(newBlocks) > 0 {
		result.SSHConfigWritten = true // would be written
	}

	// Check which SSH keys need to be generated.
	for _, key := range cfg.SSH.Keys {
		keyPath := expandHome(key.Path, home)
		if _, err := statFile(keyPath); err == nil {
			continue // key already exists
		}

		keyType := strings.ToLower(key.Type)
		if keyType == "" {
			keyType = "ed25519"
		}

		result.SSHKeysNeeded = append(result.SSHKeysNeeded, KeyAction{
			Name:    key.Name,
			Type:    key.Type,
			Command: fmt.Sprintf("ssh-keygen -t %s -f %s", keyType, keyPath),
		})
	}
}

func importGit(cfg *MigrateConfig, home string, dryRun bool, result *ImportResult) {
	// Set global user.name and user.email.
	if cfg.Git.UserName != "" && !dryRun {
		if err := runGitConfig("user.name", cfg.Git.UserName); err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Failed to set git user.name: %s", err))
		} else {
			result.GitConfigWritten = true
		}
	} else if cfg.Git.UserName != "" {
		result.GitConfigWritten = true
	}

	if cfg.Git.UserEmail != "" && !dryRun {
		if err := runGitConfig("user.email", cfg.Git.UserEmail); err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Failed to set git user.email: %s", err))
		} else {
			result.GitConfigWritten = true
		}
	} else if cfg.Git.UserEmail != "" {
		result.GitConfigWritten = true
	}

	// Create includeIf profiles.
	for _, p := range cfg.Git.Profiles {
		// Create the directory.
		dir := expandHome(p.Directory, home)
		if !dryRun {
			if err := mkdirAll(dir, 0755); err != nil {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Failed to create directory %s: %s", dir, err))
				continue
			}
		}
		result.DirsCreated = append(result.DirsCreated, p.Directory)

		// Write profile config file (e.g., ~/.gitconfig-work).
		profileConfigPath := filepath.Join(home, ".gitconfig-"+p.Name)
		if _, err := statFile(profileConfigPath); err == nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("%s already exists, skipping", profileConfigPath))
			continue
		}

		var content strings.Builder
		content.WriteString("[user]\n")
		if p.Email != "" {
			content.WriteString(fmt.Sprintf("\temail = %s\n", p.Email))
		}
		if p.SignKey != "" {
			content.WriteString(fmt.Sprintf("\tsigningkey = %s\n", p.SignKey))
		}

		if !dryRun {
			if err := writeFile(profileConfigPath, []byte(content.String()), 0600); err != nil {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Failed to write %s: %s", profileConfigPath, err))
				continue
			}
		}
		result.ProfilesWritten = append(result.ProfilesWritten, "~/.gitconfig-"+p.Name)

		// Add includeIf to ~/.gitconfig if not already present.
		if !dryRun {
			addIncludeIf(home, p)
		}
	}
}

func buildHostBlock(h SSHHostEntry, home string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Host %s\n", h.Name))
	if h.Hostname != "" {
		b.WriteString(fmt.Sprintf("  Hostname %s\n", h.Hostname))
	}
	if h.User != "" {
		b.WriteString(fmt.Sprintf("  User %s\n", h.User))
	}
	if h.Port != "" {
		b.WriteString(fmt.Sprintf("  Port %s\n", h.Port))
	}
	if h.IdentityFile != "" {
		// Expand ~/ to local home.
		path := expandHome(h.IdentityFile, home)
		b.WriteString(fmt.Sprintf("  IdentityFile %s\n", path))
	}
	if h.ProxyJump != "" {
		b.WriteString(fmt.Sprintf("  ProxyJump %s\n", h.ProxyJump))
	}
	if h.ForwardAgent {
		b.WriteString("  ForwardAgent yes\n")
	}
	return b.String()
}

func parseExistingHosts(path string) map[string]bool {
	hosts := make(map[string]bool)
	f, err := openFile(path)
	if err != nil {
		return hosts
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "host ") {
			name := strings.TrimSpace(line[5:])
			if name != "*" {
				hosts[name] = true
			}
		}
	}
	return hosts
}

func addIncludeIf(home string, p GitProfileExport) {
	gitconfigPath := filepath.Join(home, ".gitconfig")
	data, err := readFile(gitconfigPath)
	if err != nil {
		// Create new gitconfig.
		data = []byte{}
	}

	content := string(data)
	dir := p.Directory
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	// Check if already present.
	checkStr := fmt.Sprintf("gitdir:%s", dir)
	if strings.Contains(content, checkStr) {
		return
	}

	// Append includeIf block.
	block := fmt.Sprintf("\n[includeIf \"gitdir:%s\"]\n\tpath = ~/.gitconfig-%s\n", dir, p.Name)
	content += block

	writeFile(gitconfigPath, []byte(content), 0600)
}

// FormatImportResult returns a human-readable summary.
func FormatImportResult(result *ImportResult) string {
	var b strings.Builder

	// SSH section.
	b.WriteString("  SSH:\n")
	if result.SSHHostsAdded > 0 || result.SSHHostsSkipped > 0 {
		b.WriteString(fmt.Sprintf("    %d host(s) added, %d skipped\n", result.SSHHostsAdded, result.SSHHostsSkipped))
	} else {
		b.WriteString("    no hosts to add\n")
	}

	if len(result.SSHKeysNeeded) > 0 {
		b.WriteString(fmt.Sprintf("    %d key(s) to generate:\n", len(result.SSHKeysNeeded)))
		for _, k := range result.SSHKeysNeeded {
			b.WriteString(fmt.Sprintf("      %s\n", k.Command))
		}
	}

	// Git section.
	b.WriteString("  Git:\n")
	if result.GitConfigWritten {
		b.WriteString("    global user.name/email configured\n")
	}
	if len(result.DirsCreated) > 0 {
		b.WriteString(fmt.Sprintf("    %d profile directory(s) created\n", len(result.DirsCreated)))
	}
	if len(result.ProfilesWritten) > 0 {
		for _, p := range result.ProfilesWritten {
			b.WriteString(fmt.Sprintf("    wrote %s\n", p))
		}
	}

	// Warnings.
	if len(result.Warnings) > 0 {
		b.WriteString("  Warnings:\n")
		for _, w := range result.Warnings {
			b.WriteString(fmt.Sprintf("    ! %s\n", w))
		}
	}

	return b.String()
}
