package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCapture_Basic(t *testing.T) {
	snap := Capture()

	if snap.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
	if snap.OS != runtime.GOOS {
		t.Errorf("expected OS %s, got %s", runtime.GOOS, snap.OS)
	}
	if snap.Arch != runtime.GOARCH {
		t.Errorf("expected Arch %s, got %s", runtime.GOARCH, snap.Arch)
	}
}

func TestCaptureShell(t *testing.T) {
	origGetenv := getenv
	defer func() { getenv = origGetenv }()

	envMap := map[string]string{
		"SHELL":  "/bin/zsh",
		"EDITOR": "vim",
		"TERM":   "xterm-256color",
	}
	getenv = func(key string) string { return envMap[key] }

	origRunCommand := runCommand
	defer func() { runCommand = origRunCommand }()
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not bash")
	}

	snap := &Snapshot{}
	captureShell(snap)

	if snap.Shell.Shell != "/bin/zsh" {
		t.Errorf("expected shell /bin/zsh, got %s", snap.Shell.Shell)
	}
	if snap.Shell.Editor != "vim" {
		t.Errorf("expected editor vim, got %s", snap.Shell.Editor)
	}
	if snap.Shell.Term != "xterm-256color" {
		t.Errorf("expected term xterm-256color, got %s", snap.Shell.Term)
	}
}

func TestCaptureShell_BashVersion(t *testing.T) {
	origGetenv := getenv
	origRunCommand := runCommand
	defer func() {
		getenv = origGetenv
		runCommand = origRunCommand
	}()

	getenv = func(key string) string {
		if key == "SHELL" {
			return "/bin/bash"
		}
		return ""
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if strings.Contains(name, "bash") {
			return []byte("GNU bash, version 5.2.15\nsome other info\n"), nil
		}
		return nil, fmt.Errorf("not found")
	}

	snap := &Snapshot{}
	captureShell(snap)

	if snap.Shell.Version != "GNU bash, version 5.2.15" {
		t.Errorf("expected bash version string, got %q", snap.Shell.Version)
	}
}

func TestCaptureShell_EditorFallbackToVisual(t *testing.T) {
	origGetenv := getenv
	defer func() { getenv = origGetenv }()

	getenv = func(key string) string {
		if key == "VISUAL" {
			return "code"
		}
		return ""
	}

	snap := &Snapshot{}
	captureShell(snap)

	if snap.Shell.Editor != "code" {
		t.Errorf("expected editor 'code' from VISUAL, got %q", snap.Shell.Editor)
	}
}

func TestCaptureTools(t *testing.T) {
	origLookPath := lookPath
	origRunCommand := runCommand
	defer func() {
		lookPath = origLookPath
		runCommand = origRunCommand
	}()

	lookPath = func(file string) (string, error) {
		switch file {
		case "git":
			return "/usr/bin/git", nil
		case "node":
			return "/usr/local/bin/node", nil
		default:
			return "", fmt.Errorf("not found")
		}
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		switch name {
		case "git":
			return []byte("git version 2.42.0\n"), nil
		case "node":
			return []byte("v20.10.0\n"), nil
		}
		return nil, fmt.Errorf("unknown")
	}

	snap := &Snapshot{}
	captureTools(snap)

	if len(snap.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(snap.Tools))
	}

	toolMap := make(map[string]ToolInfo)
	for _, tool := range snap.Tools {
		toolMap[tool.Name] = tool
	}

	git, ok := toolMap["git"]
	if !ok {
		t.Fatal("git not found in tools")
	}
	if git.Version != "git version 2.42.0" {
		t.Errorf("expected git version, got %q", git.Version)
	}
	if git.Path != "/usr/bin/git" {
		t.Errorf("expected git path /usr/bin/git, got %s", git.Path)
	}
}

func TestCaptureTools_VersionError(t *testing.T) {
	origLookPath := lookPath
	origRunCommand := runCommand
	defer func() {
		lookPath = origLookPath
		runCommand = origRunCommand
	}()

	lookPath = func(file string) (string, error) {
		if file == "git" {
			return "/usr/bin/git", nil
		}
		return "", fmt.Errorf("not found")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("version error")
	}

	snap := &Snapshot{}
	captureTools(snap)

	if len(snap.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(snap.Tools))
	}
	if snap.Tools[0].Version != "" {
		t.Errorf("expected empty version on error, got %q", snap.Tools[0].Version)
	}
}

func TestCaptureGit(t *testing.T) {
	origRunCommand := runCommand
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		userHomeDir = origUserHomeDir
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }

	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) >= 3 && args[2] == "user.email" {
			return []byte("test@example.com\n"), nil
		}
		if name == "git" && len(args) >= 3 && args[2] == "user.name" {
			return []byte("Test User\n"), nil
		}
		return nil, fmt.Errorf("unknown")
	}

	snap := &Snapshot{}
	captureGit(snap)

	if snap.Git.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", snap.Git.Email)
	}
	if snap.Git.Name != "Test User" {
		t.Errorf("expected name Test User, got %s", snap.Git.Name)
	}
}

func TestCaptureGit_WithProfiles(t *testing.T) {
	origRunCommand := runCommand
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		userHomeDir = origUserHomeDir
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }

	// Create a gitconfig with includeIf.
	gitconfig := `[user]
	email = default@example.com
	name = Default User
[includeIf "gitdir:~/work/"]
	path = ~/.gitconfig-work
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitconfig"), []byte(gitconfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Create the included config.
	workConfig := `[user]
	email = work@company.com
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitconfig-work"), []byte(workConfig), 0644); err != nil {
		t.Fatal(err)
	}

	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) >= 3 && args[2] == "user.email" {
			return []byte("default@example.com\n"), nil
		}
		if name == "git" && len(args) >= 3 && args[2] == "user.name" {
			return []byte("Default User\n"), nil
		}
		return nil, fmt.Errorf("unknown")
	}

	snap := &Snapshot{}
	captureGit(snap)

	if len(snap.Git.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(snap.Git.Profiles))
	}
	if snap.Git.Profiles[0].Name != "work" {
		t.Errorf("expected profile name 'work', got %s", snap.Git.Profiles[0].Name)
	}
	if snap.Git.Profiles[0].Email != "work@company.com" {
		t.Errorf("expected profile email work@company.com, got %s", snap.Git.Profiles[0].Email)
	}
}

func TestCaptureGit_HomeDirError(t *testing.T) {
	origRunCommand := runCommand
	origUserHomeDir := userHomeDir
	defer func() {
		runCommand = origRunCommand
		userHomeDir = origUserHomeDir
	}()

	userHomeDir = func() (string, error) { return "", fmt.Errorf("no home") }
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("error")
	}

	snap := &Snapshot{}
	captureGit(snap) // should not panic
}

func TestCaptureSSHKeys(t *testing.T) {
	origUserHomeDir := userHomeDir
	origRunCommand := runCommand
	defer func() {
		userHomeDir = origUserHomeDir
		runCommand = origRunCommand
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }

	sshDir := filepath.Join(tmpDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create key files.
	os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("key"), 0600)
	os.WriteFile(filepath.Join(sshDir, "id_ed25519.pub"), []byte("pub"), 0644)
	os.WriteFile(filepath.Join(sshDir, "config"), []byte("config"), 0644)
	os.WriteFile(filepath.Join(sshDir, "known_hosts"), []byte("hosts"), 0644)

	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "ssh-keygen" {
			return []byte("256 SHA256:abc123 user@host (ED25519)\n"), nil
		}
		return nil, fmt.Errorf("unknown")
	}

	snap := &Snapshot{}
	captureSSHKeys(snap)

	if len(snap.SSHKeys) != 1 {
		t.Fatalf("expected 1 SSH key, got %d", len(snap.SSHKeys))
	}
	if snap.SSHKeys[0].Name != "id_ed25519" {
		t.Errorf("expected key name id_ed25519, got %s", snap.SSHKeys[0].Name)
	}
	if snap.SSHKeys[0].Fingerprint != "SHA256:abc123" {
		t.Errorf("expected fingerprint SHA256:abc123, got %s", snap.SSHKeys[0].Fingerprint)
	}
	if snap.SSHKeys[0].Type != "ED25519" {
		t.Errorf("expected type ED25519, got %s", snap.SSHKeys[0].Type)
	}
}

func TestCaptureSSHKeys_HomeDirError(t *testing.T) {
	origUserHomeDir := userHomeDir
	defer func() { userHomeDir = origUserHomeDir }()

	userHomeDir = func() (string, error) { return "", fmt.Errorf("no home") }

	snap := &Snapshot{}
	captureSSHKeys(snap) // should not panic
	if len(snap.SSHKeys) != 0 {
		t.Error("expected no SSH keys on home dir error")
	}
}

func TestCaptureK8s(t *testing.T) {
	origLookPath := lookPath
	origRunCommand := runCommand
	defer func() {
		lookPath = origLookPath
		runCommand = origRunCommand
	}()

	lookPath = func(file string) (string, error) {
		if file == "kubectl" {
			return "/usr/local/bin/kubectl", nil
		}
		return "", fmt.Errorf("not found")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "kubectl" {
			joined := strings.Join(args, " ")
			if strings.Contains(joined, "current-context") {
				return []byte("my-cluster\n"), nil
			}
			if strings.Contains(joined, "get-contexts") {
				return []byte("my-cluster\nother-cluster\n"), nil
			}
		}
		return nil, fmt.Errorf("unknown")
	}

	snap := &Snapshot{}
	captureK8s(snap)

	if len(snap.K8s) != 2 {
		t.Fatalf("expected 2 k8s contexts, got %d", len(snap.K8s))
	}

	found := false
	for _, ctx := range snap.K8s {
		if ctx.Name == "my-cluster" && ctx.Current {
			found = true
		}
	}
	if !found {
		t.Error("expected my-cluster to be marked as current")
	}
}

func TestCaptureK8s_NoKubectl(t *testing.T) {
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	lookPath = func(file string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	snap := &Snapshot{}
	captureK8s(snap)

	if len(snap.K8s) != 0 {
		t.Error("expected no k8s contexts when kubectl not found")
	}
}

func TestCapturePath(t *testing.T) {
	origGetenv := getenv
	defer func() { getenv = origGetenv }()

	getenv = func(key string) string {
		if key == "PATH" {
			return "/usr/bin:/usr/local/bin:/home/user/bin"
		}
		return ""
	}

	snap := &Snapshot{}
	capturePath(snap)

	if len(snap.Path) != 3 {
		t.Fatalf("expected 3 path entries, got %d", len(snap.Path))
	}
	if snap.Path[0] != "/usr/bin" {
		t.Errorf("expected first path /usr/bin, got %s", snap.Path[0])
	}
}

func TestCapturePath_Empty(t *testing.T) {
	origGetenv := getenv
	defer func() { getenv = origGetenv }()

	getenv = func(key string) string { return "" }

	snap := &Snapshot{}
	capturePath(snap)

	if snap.Path != nil {
		t.Error("expected nil path on empty PATH")
	}
}

func TestCaptureBrew_NonDarwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping on darwin")
	}

	snap := &Snapshot{}
	captureBrew(snap)

	if len(snap.Brew) != 0 {
		t.Error("expected no brew packages on non-darwin")
	}
}

func TestCaptureBrew_NoBrew(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-darwin")
	}

	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	lookPath = func(file string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	snap := &Snapshot{}
	captureBrew(snap)

	if len(snap.Brew) != 0 {
		t.Error("expected no brew packages when brew not found")
	}
}

func TestCaptureBrew_WithPackages(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-darwin")
	}

	origLookPath := lookPath
	origRunCommand := runCommand
	defer func() {
		lookPath = origLookPath
		runCommand = origRunCommand
	}()

	lookPath = func(file string) (string, error) {
		if file == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", fmt.Errorf("not found")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "brew" {
			return []byte("git\nwget\njq\n"), nil
		}
		return nil, fmt.Errorf("unknown")
	}

	snap := &Snapshot{}
	captureBrew(snap)

	if len(snap.Brew) != 3 {
		t.Fatalf("expected 3 brew packages, got %d", len(snap.Brew))
	}
	if snap.Brew[0] != "git" || snap.Brew[1] != "wget" || snap.Brew[2] != "jq" {
		t.Errorf("unexpected brew packages: %v", snap.Brew)
	}
}

func TestCaptureBrew_ListError(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("skipping on non-darwin")
	}

	origLookPath := lookPath
	origRunCommand := runCommand
	defer func() {
		lookPath = origLookPath
		runCommand = origRunCommand
	}()

	lookPath = func(file string) (string, error) {
		if file == "brew" {
			return "/opt/homebrew/bin/brew", nil
		}
		return "", fmt.Errorf("not found")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("brew error")
	}

	snap := &Snapshot{}
	captureBrew(snap)

	if len(snap.Brew) != 0 {
		t.Error("expected no brew packages on error")
	}
}

func TestCaptureK8s_GetContextsError(t *testing.T) {
	origLookPath := lookPath
	origRunCommand := runCommand
	defer func() {
		lookPath = origLookPath
		runCommand = origRunCommand
	}()

	lookPath = func(file string) (string, error) {
		if file == "kubectl" {
			return "/usr/local/bin/kubectl", nil
		}
		return "", fmt.Errorf("not found")
	}
	callCount := 0
	runCommand = func(name string, args ...string) ([]byte, error) {
		callCount++
		if callCount == 1 {
			// current-context succeeds
			return []byte("my-cluster\n"), nil
		}
		// get-contexts fails
		return nil, fmt.Errorf("error")
	}

	snap := &Snapshot{}
	captureK8s(snap)

	if len(snap.K8s) != 0 {
		t.Error("expected no k8s contexts when get-contexts fails")
	}
}

func TestExpandHome(t *testing.T) {
	result := expandHome("~/config", "/home/user")
	if result != "/home/user/config" {
		t.Errorf("expected /home/user/config, got %s", result)
	}

	result = expandHome("/absolute/path", "/home/user")
	if result != "/absolute/path" {
		t.Errorf("expected /absolute/path, got %s", result)
	}
}

func TestFormatSummary(t *testing.T) {
	snap := &Snapshot{
		Hostname:  "myhost",
		OS:        "darwin",
		Arch:      "arm64",
		Timestamp: "2024-01-01T00:00:00Z",
		Shell:     ShellInfo{Shell: "/bin/zsh"},
		Tools:     []ToolInfo{{Name: "git"}, {Name: "node"}},
		SSHKeys:   []SSHKeyInfo{{Name: "id_ed25519"}},
		K8s:       []K8sContext{{Name: "ctx1"}},
		Path:      []string{"/usr/bin", "/usr/local/bin"},
		Brew:      []string{"pkg1", "pkg2"},
	}

	output := FormatSummary(snap)

	checks := []string{
		"myhost",
		"darwin/arm64",
		"/bin/zsh",
		"2 installed",
		"1",
		"1 context",
		"2 entries",
		"2 packages",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected summary to contain %q", check)
		}
	}
}
