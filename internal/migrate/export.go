package migrate

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/somaz94/bash-pilot/internal/ssh"
)

// Function variables for testing.
var (
	runCommand = func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).Output()
	}
	userHomeDir  = os.UserHomeDir
	readDir      = os.ReadDir
	parseSSHConf = func(path string) ([]ssh.Host, error) {
		return ssh.ParseConfig(path)
	}
)

// Export captures SSH and Git configuration in a portable format.
func Export(sshConfigPath string) (*MigrateConfig, error) {
	home, err := userHomeDir()
	if err != nil {
		return nil, err
	}

	cfg := &MigrateConfig{
		Version:    "1",
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		SourceOS:   runtime.GOOS,
		SourceHome: home,
	}

	exportSSH(cfg, sshConfigPath, home)
	exportGit(cfg, home)

	return cfg, nil
}

func exportSSH(cfg *MigrateConfig, sshConfigPath, home string) {
	hosts, err := parseSSHConf(sshConfigPath)
	if err != nil {
		return
	}

	for _, h := range hosts {
		entry := SSHHostEntry{
			Name:         h.Name,
			Hostname:     h.Hostname,
			User:         h.User,
			Port:         h.Port,
			IdentityFile: normalizePath(h.IdentityFile, home),
			ProxyJump:    h.ProxyJump,
			ForwardAgent: h.ForwardAgent,
		}
		cfg.SSH.Hosts = append(cfg.SSH.Hosts, entry)
	}

	// Scan SSH keys.
	sshDir := filepath.Join(home, ".ssh")
	entries, err := readDir(sshDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".pub") ||
			name == "config" ||
			name == "known_hosts" ||
			name == "known_hosts.old" ||
			name == "authorized_keys" {
			continue
		}

		keyPath := filepath.Join(sshDir, name)
		ref := SSHKeyRef{
			Name: name,
			Path: "~/.ssh/" + name,
		}

		// Get key type.
		out, err := runCommand("ssh-keygen", "-l", "-f", keyPath)
		if err == nil {
			parts := strings.Fields(strings.TrimSpace(string(out)))
			if len(parts) >= 4 {
				ref.Type = strings.Trim(parts[len(parts)-1], "()")
			}
		}

		cfg.SSH.Keys = append(cfg.SSH.Keys, ref)
	}
}

func exportGit(cfg *MigrateConfig, home string) {
	// Global user.name and user.email.
	out, err := runCommand("git", "config", "--global", "user.name")
	if err == nil {
		cfg.Git.UserName = strings.TrimSpace(string(out))
	}

	out, err = runCommand("git", "config", "--global", "user.email")
	if err == nil {
		cfg.Git.UserEmail = strings.TrimSpace(string(out))
	}

	// Parse includeIf profiles from gitconfig.
	gitconfigPath := filepath.Join(home, ".gitconfig")
	data, err := os.ReadFile(gitconfigPath)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if !strings.HasPrefix(lower, "[includeif \"gitdir:") {
			continue
		}

		start := strings.Index(trimmed, "gitdir:")
		if start == -1 {
			continue
		}
		dir := trimmed[start+7:]
		dir = strings.TrimSuffix(dir, "\"]")
		dir = strings.TrimSuffix(dir, "/")

		// Normalize to tilde path.
		dir = normalizePath(dir, home)

		name := filepath.Base(dir)

		profile := GitProfileExport{
			Name:      name,
			Directory: dir,
		}

		// Find the path = line after this includeIf and read the included config.
		for j := i + 1; j < len(lines) && j < i+5; j++ {
			t := strings.TrimSpace(lines[j])
			if strings.HasPrefix(t, "path = ") {
				includePath := strings.TrimPrefix(t, "path = ")
				includePath = expandHome(includePath, home)
				incData, err := os.ReadFile(includePath)
				if err != nil {
					break
				}
				for _, il := range strings.Split(string(incData), "\n") {
					it := strings.TrimSpace(il)
					if strings.HasPrefix(it, "email = ") {
						profile.Email = strings.TrimPrefix(it, "email = ")
					}
					if strings.HasPrefix(it, "signingkey = ") {
						profile.SignKey = strings.TrimPrefix(it, "signingkey = ")
					}
				}
				break
			}
			// Stop if we hit another section.
			if strings.HasPrefix(t, "[") {
				break
			}
		}

		cfg.Git.Profiles = append(cfg.Git.Profiles, profile)
	}
}

// normalizePath replaces the home directory prefix with ~/.
func normalizePath(path, home string) string {
	if home == "" {
		return path
	}
	// Handle paths that already use ~/.
	if strings.HasPrefix(path, "~/") {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

// expandHome replaces ~/ with the home directory.
func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}
