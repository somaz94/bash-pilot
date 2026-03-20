package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/somaz94/bash-pilot/internal/ssh"
)

func TestExport_Basic(t *testing.T) {
	origUserHomeDir := userHomeDir
	origParseSSHConf := parseSSHConf
	origRunCommand := runCommand
	origReadDir := readDir
	defer func() {
		userHomeDir = origUserHomeDir
		parseSSHConf = origParseSSHConf
		runCommand = origRunCommand
		readDir = origReadDir
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }

	// Create gitconfig.
	gitconfig := fmt.Sprintf(`[user]
	name = Test User
	email = test@example.com
[includeIf "gitdir:%s/work/"]
	path = %s/.gitconfig-work
`, tmpDir, tmpDir)
	os.WriteFile(filepath.Join(tmpDir, ".gitconfig"), []byte(gitconfig), 0644)

	workConfig := "[user]\n\temail = work@company.com\n\tsigningkey = ABC123\n"
	os.WriteFile(filepath.Join(tmpDir, ".gitconfig-work"), []byte(workConfig), 0644)

	parseSSHConf = func(path string) ([]ssh.Host, error) {
		return []ssh.Host{
			{Name: "github.com-personal", Hostname: "github.com", User: "git",
				IdentityFile: tmpDir + "/.ssh/id_ed25519"},
			{Name: "server1", Hostname: "10.10.10.10", User: "deploy",
				Port: "2222", IdentityFile: tmpDir + "/.ssh/id_rsa_work"},
		}, nil
	}

	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) >= 3 {
			if args[2] == "user.name" {
				return []byte("Test User\n"), nil
			}
			if args[2] == "user.email" {
				return []byte("test@example.com\n"), nil
			}
		}
		if name == "ssh-keygen" {
			return []byte("256 SHA256:abc user@host (ED25519)\n"), nil
		}
		return nil, fmt.Errorf("unknown")
	}

	// Create SSH key files.
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("key"), 0600)
	os.WriteFile(filepath.Join(sshDir, "id_ed25519.pub"), []byte("pub"), 0644)
	os.WriteFile(filepath.Join(sshDir, "config"), []byte("cfg"), 0644)
	readDir = os.ReadDir

	cfg, err := Export("")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Version != "1" {
		t.Errorf("expected version 1, got %s", cfg.Version)
	}
	if cfg.SourceHome != tmpDir {
		t.Errorf("expected source home %s, got %s", tmpDir, cfg.SourceHome)
	}

	// SSH hosts.
	if len(cfg.SSH.Hosts) != 2 {
		t.Fatalf("expected 2 SSH hosts, got %d", len(cfg.SSH.Hosts))
	}
	if cfg.SSH.Hosts[0].IdentityFile != "~/.ssh/id_ed25519" {
		t.Errorf("expected normalized path, got %s", cfg.SSH.Hosts[0].IdentityFile)
	}
	if cfg.SSH.Hosts[1].Port != "2222" {
		t.Errorf("expected port 2222, got %s", cfg.SSH.Hosts[1].Port)
	}

	// SSH keys.
	if len(cfg.SSH.Keys) != 1 {
		t.Fatalf("expected 1 SSH key, got %d", len(cfg.SSH.Keys))
	}
	if cfg.SSH.Keys[0].Name != "id_ed25519" {
		t.Errorf("expected key name id_ed25519, got %s", cfg.SSH.Keys[0].Name)
	}
	if cfg.SSH.Keys[0].Type != "ED25519" {
		t.Errorf("expected key type ED25519, got %s", cfg.SSH.Keys[0].Type)
	}

	// Git.
	if cfg.Git.UserName != "Test User" {
		t.Errorf("expected git user name, got %s", cfg.Git.UserName)
	}
	if cfg.Git.UserEmail != "test@example.com" {
		t.Errorf("expected git user email, got %s", cfg.Git.UserEmail)
	}
	if len(cfg.Git.Profiles) != 1 {
		t.Fatalf("expected 1 git profile, got %d", len(cfg.Git.Profiles))
	}
	if cfg.Git.Profiles[0].Email != "work@company.com" {
		t.Errorf("expected profile email, got %s", cfg.Git.Profiles[0].Email)
	}
	if cfg.Git.Profiles[0].SignKey != "ABC123" {
		t.Errorf("expected profile sign key, got %s", cfg.Git.Profiles[0].SignKey)
	}
}

func TestExport_SSHParseError(t *testing.T) {
	origUserHomeDir := userHomeDir
	origParseSSHConf := parseSSHConf
	origRunCommand := runCommand
	origReadDir := readDir
	defer func() {
		userHomeDir = origUserHomeDir
		parseSSHConf = origParseSSHConf
		runCommand = origRunCommand
		readDir = origReadDir
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	parseSSHConf = func(path string) ([]ssh.Host, error) {
		return nil, fmt.Errorf("no config")
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("error")
	}
	readDir = func(name string) ([]os.DirEntry, error) {
		return nil, fmt.Errorf("no dir")
	}

	cfg, err := Export("")
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.SSH.Hosts) != 0 {
		t.Error("expected no SSH hosts on parse error")
	}
}

func TestExport_HomeDirError(t *testing.T) {
	origUserHomeDir := userHomeDir
	defer func() { userHomeDir = origUserHomeDir }()

	userHomeDir = func() (string, error) { return "", fmt.Errorf("no home") }

	_, err := Export("")
	if err == nil {
		t.Error("expected error on home dir failure")
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		path, home, expected string
	}{
		{"/Users/somaz/.ssh/id_rsa", "/Users/somaz", "~/.ssh/id_rsa"},
		{"~/.ssh/id_rsa", "/Users/somaz", "~/.ssh/id_rsa"},
		{"/etc/ssh/key", "/Users/somaz", "/etc/ssh/key"},
		{"", "/Users/somaz", ""},
		{"/Users/somaz/.ssh/key", "", "/Users/somaz/.ssh/key"},
	}

	for _, tt := range tests {
		result := normalizePath(tt.path, tt.home)
		if result != tt.expected {
			t.Errorf("normalizePath(%q, %q) = %q, want %q", tt.path, tt.home, result, tt.expected)
		}
	}
}

func TestExpandHome(t *testing.T) {
	result := expandHome("~/.ssh/id_rsa", "/home/user")
	if result != "/home/user/.ssh/id_rsa" {
		t.Errorf("expected /home/user/.ssh/id_rsa, got %s", result)
	}

	result = expandHome("/absolute/path", "/home/user")
	if result != "/absolute/path" {
		t.Errorf("expected /absolute/path, got %s", result)
	}
}

func TestExport_TildePath(t *testing.T) {
	origUserHomeDir := userHomeDir
	origParseSSHConf := parseSSHConf
	origRunCommand := runCommand
	origReadDir := readDir
	defer func() {
		userHomeDir = origUserHomeDir
		parseSSHConf = origParseSSHConf
		runCommand = origRunCommand
		readDir = origReadDir
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }

	// SSH host already has tilde path.
	parseSSHConf = func(path string) ([]ssh.Host, error) {
		return []ssh.Host{
			{Name: "test", IdentityFile: "~/.ssh/id_rsa"},
		}, nil
	}
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("error")
	}
	readDir = func(name string) ([]os.DirEntry, error) {
		return nil, fmt.Errorf("no dir")
	}

	cfg, err := Export("")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.SSH.Hosts[0].IdentityFile != "~/.ssh/id_rsa" {
		t.Errorf("expected tilde path preserved, got %s", cfg.SSH.Hosts[0].IdentityFile)
	}
}

func TestExport_GitProfileWithTildePath(t *testing.T) {
	origUserHomeDir := userHomeDir
	origParseSSHConf := parseSSHConf
	origRunCommand := runCommand
	origReadDir := readDir
	defer func() {
		userHomeDir = origUserHomeDir
		parseSSHConf = origParseSSHConf
		runCommand = origRunCommand
		readDir = origReadDir
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }

	// Gitconfig with tilde path in includeIf.
	gitconfig := `[includeIf "gitdir:~/personal/"]
	path = ~/.gitconfig-personal
`
	os.WriteFile(filepath.Join(tmpDir, ".gitconfig"), []byte(gitconfig), 0644)
	personalConfig := "[user]\n\temail = me@gmail.com\n"
	os.WriteFile(filepath.Join(tmpDir, ".gitconfig-personal"), []byte(personalConfig), 0644)

	parseSSHConf = func(path string) ([]ssh.Host, error) { return nil, fmt.Errorf("no config") }
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("error")
	}
	readDir = func(name string) ([]os.DirEntry, error) {
		return nil, fmt.Errorf("no dir")
	}

	cfg, err := Export("")
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Git.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(cfg.Git.Profiles))
	}
	if !strings.HasPrefix(cfg.Git.Profiles[0].Directory, "~/") {
		t.Errorf("expected tilde directory, got %s", cfg.Git.Profiles[0].Directory)
	}
	if cfg.Git.Profiles[0].Email != "me@gmail.com" {
		t.Errorf("expected email, got %s", cfg.Git.Profiles[0].Email)
	}
}
