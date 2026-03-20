package snapshot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Function variables for testing.
var (
	runCommand = func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).Output()
	}
	lookPath    = exec.LookPath
	userHomeDir = os.UserHomeDir
	getenv      = os.Getenv
)

// Snapshot represents a full environment capture.
type Snapshot struct {
	Timestamp string       `json:"timestamp"`
	Hostname  string       `json:"hostname"`
	OS        string       `json:"os"`
	Arch      string       `json:"arch"`
	Shell     ShellInfo    `json:"shell"`
	Tools     []ToolInfo   `json:"tools"`
	Git       GitInfo      `json:"git"`
	SSHKeys   []SSHKeyInfo `json:"ssh_keys"`
	K8s       []K8sContext `json:"k8s_contexts,omitempty"`
	Path      []string     `json:"path"`
	Brew      []string     `json:"brew_packages,omitempty"`
}

// ShellInfo holds shell environment details.
type ShellInfo struct {
	Shell   string `json:"shell"`
	Version string `json:"version,omitempty"`
	Editor  string `json:"editor,omitempty"`
	Term    string `json:"term,omitempty"`
}

// ToolInfo holds a tool name and version.
type ToolInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path,omitempty"`
	Version string `json:"version,omitempty"`
}

// GitInfo holds git identity information.
type GitInfo struct {
	Email    string       `json:"email,omitempty"`
	Name     string       `json:"name,omitempty"`
	Profiles []GitProfile `json:"profiles,omitempty"`
}

// GitProfile holds an includeIf profile.
type GitProfile struct {
	Name      string `json:"name"`
	Email     string `json:"email,omitempty"`
	Directory string `json:"directory"`
}

// SSHKeyInfo holds SSH key metadata.
type SSHKeyInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// K8sContext holds a kubernetes context.
type K8sContext struct {
	Name    string `json:"name"`
	Cluster string `json:"cluster,omitempty"`
	Current bool   `json:"current,omitempty"`
}

// Capture takes a snapshot of the current environment.
func Capture() *Snapshot {
	hostname, _ := os.Hostname()

	snap := &Snapshot{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Hostname:  hostname,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	captureShell(snap)
	captureTools(snap)
	captureGit(snap)
	captureSSHKeys(snap)
	captureK8s(snap)
	capturePath(snap)
	captureBrew(snap)

	return snap
}

func captureShell(snap *Snapshot) {
	snap.Shell.Shell = getenv("SHELL")
	snap.Shell.Editor = getenv("EDITOR")
	if snap.Shell.Editor == "" {
		snap.Shell.Editor = getenv("VISUAL")
	}
	snap.Shell.Term = getenv("TERM")

	if snap.Shell.Shell != "" && strings.Contains(snap.Shell.Shell, "bash") {
		out, err := runCommand(snap.Shell.Shell, "--version")
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				snap.Shell.Version = strings.TrimSpace(lines[0])
			}
		}
	}
}

func captureTools(snap *Snapshot) {
	tools := []struct {
		name       string
		versionArg string
	}{
		{"git", "--version"},
		{"ssh", "-V"},
		{"curl", "--version"},
		{"make", "--version"},
		{"docker", "--version"},
		{"kubectl", "version --client --short"},
		{"helm", "version --short"},
		{"terraform", "--version"},
		{"go", "version"},
		{"node", "--version"},
		{"python3", "--version"},
	}

	for _, tool := range tools {
		path, err := lookPath(tool.name)
		if err != nil {
			continue
		}

		info := ToolInfo{
			Name: tool.name,
			Path: path,
		}

		args := strings.Fields(tool.versionArg)
		out, err := runCommand(tool.name, args...)
		if err == nil {
			version := strings.TrimSpace(string(out))
			// Take first line only.
			if idx := strings.IndexByte(version, '\n'); idx != -1 {
				version = version[:idx]
			}
			info.Version = version
		}

		snap.Tools = append(snap.Tools, info)
	}
}

func captureGit(snap *Snapshot) {
	out, err := runCommand("git", "config", "--global", "user.email")
	if err == nil {
		snap.Git.Email = strings.TrimSpace(string(out))
	}

	out, err = runCommand("git", "config", "--global", "user.name")
	if err == nil {
		snap.Git.Name = strings.TrimSpace(string(out))
	}

	// Parse includeIf profiles from gitconfig.
	home, err := userHomeDir()
	if err != nil {
		return
	}
	gitconfigPath := filepath.Join(home, ".gitconfig")
	data, err := os.ReadFile(gitconfigPath)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "[includeif \"gitdir:") {
			// Extract directory.
			start := strings.Index(trimmed, "gitdir:")
			if start == -1 {
				continue
			}
			dir := trimmed[start+7:]
			dir = strings.TrimSuffix(dir, "\"]")
			dir = strings.TrimSuffix(dir, "/")

			// Extract profile name from directory.
			name := filepath.Base(dir)

			// Try to read email from the included config.
			profile := GitProfile{
				Name:      name,
				Directory: dir,
			}

			// Find the path line after this includeIf.
			for _, l2 := range lines {
				t2 := strings.TrimSpace(l2)
				if strings.HasPrefix(t2, "path = ") && strings.Contains(t2, name) {
					includePath := strings.TrimPrefix(t2, "path = ")
					includePath = expandHome(includePath, home)
					incData, err := os.ReadFile(includePath)
					if err == nil {
						for _, il := range strings.Split(string(incData), "\n") {
							it := strings.TrimSpace(il)
							if strings.HasPrefix(it, "email = ") {
								profile.Email = strings.TrimPrefix(it, "email = ")
							}
						}
					}
				}
			}

			snap.Git.Profiles = append(snap.Git.Profiles, profile)
		}
	}
}

func captureSSHKeys(snap *Snapshot) {
	home, err := userHomeDir()
	if err != nil {
		return
	}

	sshDir := filepath.Join(home, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip public keys, config, known_hosts, etc.
		if strings.HasSuffix(name, ".pub") ||
			name == "config" ||
			name == "known_hosts" ||
			name == "known_hosts.old" ||
			name == "authorized_keys" {
			continue
		}

		keyPath := filepath.Join(sshDir, name)
		info := SSHKeyInfo{Name: name}

		// Get fingerprint and type.
		out, err := runCommand("ssh-keygen", "-l", "-f", keyPath)
		if err == nil {
			parts := strings.Fields(strings.TrimSpace(string(out)))
			if len(parts) >= 2 {
				info.Fingerprint = parts[1]
			}
			if len(parts) >= 4 {
				info.Type = strings.Trim(parts[len(parts)-1], "()")
			}
		}

		snap.SSHKeys = append(snap.SSHKeys, info)
	}
}

func captureK8s(snap *Snapshot) {
	if _, err := lookPath("kubectl"); err != nil {
		return
	}

	// Get current context.
	currentCtx := ""
	out, err := runCommand("kubectl", "config", "current-context")
	if err == nil {
		currentCtx = strings.TrimSpace(string(out))
	}

	// Get all contexts.
	out, err = runCommand("kubectl", "config", "get-contexts", "-o", "name")
	if err != nil {
		return
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		ctx := strings.TrimSpace(line)
		if ctx == "" {
			continue
		}
		snap.K8s = append(snap.K8s, K8sContext{
			Name:    ctx,
			Current: ctx == currentCtx,
		})
	}
}

func capturePath(snap *Snapshot) {
	pathEnv := getenv("PATH")
	if pathEnv == "" {
		return
	}
	snap.Path = strings.Split(pathEnv, ":")
}

func captureBrew(snap *Snapshot) {
	if runtime.GOOS != "darwin" {
		return
	}
	if _, err := lookPath("brew"); err != nil {
		return
	}

	out, err := runCommand("brew", "list", "--formula", "-1")
	if err != nil {
		return
	}

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		pkg := strings.TrimSpace(line)
		if pkg != "" {
			snap.Brew = append(snap.Brew, pkg)
		}
	}
}

func expandHome(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

// FormatSummary returns a human-readable summary of the snapshot.
func FormatSummary(snap *Snapshot) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Hostname:  %s\n", snap.Hostname))
	b.WriteString(fmt.Sprintf("OS/Arch:   %s/%s\n", snap.OS, snap.Arch))
	b.WriteString(fmt.Sprintf("Shell:     %s\n", snap.Shell.Shell))
	b.WriteString(fmt.Sprintf("Tools:     %d installed\n", len(snap.Tools)))
	b.WriteString(fmt.Sprintf("SSH Keys:  %d\n", len(snap.SSHKeys)))
	b.WriteString(fmt.Sprintf("K8s:       %d context(s)\n", len(snap.K8s)))
	b.WriteString(fmt.Sprintf("PATH:      %d entries\n", len(snap.Path)))
	if len(snap.Brew) > 0 {
		b.WriteString(fmt.Sprintf("Brew:      %d packages\n", len(snap.Brew)))
	}
	b.WriteString(fmt.Sprintf("Captured:  %s\n", snap.Timestamp))

	return b.String()
}
