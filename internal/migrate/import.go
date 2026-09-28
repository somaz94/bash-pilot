package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/somaz94/bash-pilot/internal/config"
	"github.com/somaz94/bash-pilot/internal/ssh"
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
	runGitConfig = func(args ...string) error {
		allArgs := append([]string{"config", "--global"}, args...)
		_, err := runCommand("git", allArgs...)
		return err
	}
)

// Import applies the migrate config to the current machine.
func Import(cfg *MigrateConfig, dryRun bool, onlyOpts ...map[string]bool) (*ImportResult, error) {
	home, err := userHomeDir()
	if err != nil {
		return nil, err
	}

	var only map[string]bool
	if len(onlyOpts) > 0 {
		only = onlyOpts[0]
	}

	result := &ImportResult{}

	if only == nil || only["ssh"] {
		importSSH(cfg, home, dryRun, result)
	}
	if only == nil || only["git"] {
		importGit(cfg, home, dryRun, result)
	}

	return result, nil
}

func importSSH(cfg *MigrateConfig, home string, dryRun bool, result *ImportResult) {
	sshDir := filepath.Join(home, ".ssh")

	if !dryRun {
		// A failure here surfaces when the config file below cannot be opened.
		_ = mkdirAll(sshDir, config.PermSSHDir)
	}

	sshConfigPath := filepath.Join(sshDir, "config")
	existingHosts := parseExistingHosts(sshConfigPath)
	parsedValues := cfg.Version != "" && cfg.Version != "1"
	if parsedValues && cfg.Version != FormatVersion {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"unknown export format version %q (this release writes %q); SSH values may be written incorrectly", cfg.Version, FormatVersion))
	}

	var newBlocks []string
	for _, h := range cfg.SSH.Hosts {
		if pattern, ok := existingPattern(h.Name, existingHosts); ok {
			result.SSHHostsSkipped++
			msg := fmt.Sprintf("Host '%s' already exists in SSH config, skipping", h.Name)
			if pattern != h.Name {
				msg = fmt.Sprintf("Host '%s' shares '%s' with an existing SSH config entry, skipping", h.Name, pattern)
			}
			result.Warnings = append(result.Warnings, msg)
			continue
		}

		block := buildHostBlock(h, home, parsedValues)
		newBlocks = append(newBlocks, block)
		result.SSHHostsAdded++
	}

	if len(newBlocks) > 0 && !dryRun {
		content := "\n# Imported by bash-pilot migrate\n" + strings.Join(newBlocks, "\n")

		// Mark the config written only after the write and close succeed, never on a failed open.
		f, err := os.OpenFile(sshConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, config.PermSSHConfigFile)
		if err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("could not open SSH config %s: %v", sshConfigPath, err))
		} else {
			_, writeErr := f.WriteString(content)
			closeErr := f.Close()
			if writeErr == nil {
				writeErr = closeErr
			}
			if writeErr != nil {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("could not write SSH config %s: %v", sshConfigPath, writeErr))
			} else {
				result.SSHConfigWritten = true
			}
		}
	} else if len(newBlocks) > 0 {
		result.SSHConfigWritten = true // would be written
	}

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

	for _, p := range cfg.Git.Profiles {
		dir := expandHome(p.Directory, home)
		if !dryRun {
			if err := mkdirAll(dir, config.PermConfigDir); err != nil {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Failed to create directory %s: %s", dir, err))
				continue
			}
		}
		result.DirsCreated = append(result.DirsCreated, p.Directory)

		profileConfigPath := filepath.Join(home, ".gitconfig-"+p.Name)
		if _, err := statFile(profileConfigPath); err == nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("%s already exists, skipping", profileConfigPath))
			continue
		}

		var content strings.Builder
		content.WriteString("[user]\n")
		if p.UserName != "" {
			content.WriteString(fmt.Sprintf("\tname = %s\n", p.UserName))
		}
		if p.Email != "" {
			content.WriteString(fmt.Sprintf("\temail = %s\n", p.Email))
		}
		if p.SignKey != "" {
			content.WriteString(fmt.Sprintf("\tsigningkey = %s\n", p.SignKey))
		}

		if !dryRun {
			if err := writeFile(profileConfigPath, []byte(content.String()), config.PermGitConfigFile); err != nil {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Failed to write %s: %s", profileConfigPath, err))
				continue
			}
		}
		result.ProfilesWritten = append(result.ProfilesWritten, "~/.gitconfig-"+p.Name)

		if !dryRun {
			if err := addIncludeIf(home, p); err != nil {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Failed to add includeIf for profile %s: %s", p.Name, err))
			}
		}
	}
}

// buildHostBlock renders h as an ssh_config Host block. parsedValues marks a
// format 2 export, whose values must be quoted again where ssh would split them;
// older exports hold raw ssh_config text and are written as-is.
func buildHostBlock(h SSHHostEntry, home string, parsedValues bool) string {
	arg := func(v string) string {
		if parsedValues {
			return sshArg(v)
		}
		return v
	}
	var b strings.Builder
	name := h.Name
	if parsedValues {
		patterns := strings.Fields(h.Name)
		for i, p := range patterns {
			patterns[i] = sshArg(p)
		}
		name = strings.Join(patterns, " ")
	}
	b.WriteString(fmt.Sprintf("Host %s\n", name))
	if h.Hostname != "" {
		b.WriteString(fmt.Sprintf("  Hostname %s\n", arg(h.Hostname)))
	}
	if h.User != "" {
		b.WriteString(fmt.Sprintf("  User %s\n", arg(h.User)))
	}
	if h.Port != "" {
		b.WriteString(fmt.Sprintf("  Port %s\n", h.Port))
	}
	if h.IdentityFile != "" {
		path := expandHome(h.IdentityFile, home)
		b.WriteString(fmt.Sprintf("  IdentityFile %s\n", arg(path)))
	}
	proxyJump, proxyCommand := h.ProxyJump, h.ProxyCommand
	// Older exports stored ProxyCommand in proxy_jump; a jump spec is one token plus an optional "# comment".
	if f := strings.Fields(proxyJump); !parsedValues && proxyCommand == "" && len(f) > 1 && !strings.HasPrefix(f[1], "#") {
		proxyJump, proxyCommand = "", proxyJump
	}
	if proxyJump != "" {
		b.WriteString(fmt.Sprintf("  ProxyJump %s\n", proxyJump))
	}
	if proxyCommand != "" {
		b.WriteString(fmt.Sprintf("  ProxyCommand %s\n", proxyCommand))
	}
	switch {
	case h.ForwardAgentSocket != "":
		b.WriteString(fmt.Sprintf("  ForwardAgent %s\n", arg(expandHome(h.ForwardAgentSocket, home))))
	case h.ForwardAgent:
		b.WriteString("  ForwardAgent yes\n")
	}
	return b.String()
}

// sshArg quotes v for ssh_config when ssh would otherwise split, unquote or
// comment it; ssh strips the quotes again when it reads the file.
func sshArg(v string) string {
	// A leading '=' would be read as the keyword separator.
	if !strings.ContainsAny(v, " \t\"'#\\") && !strings.HasPrefix(v, "=") {
		return v
	}
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(v) + `"`
}

// parseExistingHosts returns every pattern named on a Host line of the SSH
// config at path. A missing or unreadable config has no existing hosts.
func parseExistingHosts(path string) map[string]bool {
	existing := make(map[string]bool)
	hosts, _ := ssh.ParseConfig(path)
	for _, h := range hosts {
		for _, pattern := range strings.Fields(h.Name) {
			if pattern != "*" {
				existing[pattern] = true
			}
		}
	}
	return existing
}

// existingPattern returns the first pattern of a "Host a b" name that is already
// defined; ssh uses the first block that sets a value, so an appended duplicate would be ignored.
func existingPattern(name string, existing map[string]bool) (string, bool) {
	for _, pattern := range strings.Fields(name) {
		if existing[pattern] {
			return pattern, true
		}
	}
	return "", false
}

// addIncludeIf appends an includeIf block for the profile to ~/.gitconfig.
// It returns nil when the block is already present or was written; a write
// failure is returned so the caller can surface it rather than leaving the
// profile silently unconfigured.
func addIncludeIf(home string, p GitProfileExport) error {
	gitconfigPath := filepath.Join(home, ".gitconfig")
	data, err := readFile(gitconfigPath)
	if err != nil {
		// No readable gitconfig: start a new one.
		data = []byte{}
	}

	content := string(data)
	dir := p.Directory
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	// The closing quote stops a nested "gitdir:~/work/sub/" from matching "gitdir:~/work/".
	checkStr := fmt.Sprintf("gitdir:%s\"", dir)
	if strings.Contains(content, checkStr) {
		return nil
	}

	block := fmt.Sprintf("\n[includeIf \"gitdir:%s\"]\n\tpath = ~/.gitconfig-%s\n", dir, p.Name)
	content += block

	return writeFile(gitconfigPath, []byte(content), config.PermGitConfigFile)
}

// FormatImportResult returns a human-readable summary.
func FormatImportResult(result *ImportResult) string {
	var b strings.Builder

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

	if len(result.Warnings) > 0 {
		b.WriteString("  Warnings:\n")
		for _, w := range result.Warnings {
			b.WriteString(fmt.Sprintf("    ! %s\n", w))
		}
	}

	return b.String()
}
