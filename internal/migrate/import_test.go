package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImport_SSHHosts(t *testing.T) {
	origUserHomeDir := userHomeDir
	origStatFile := statFile
	origRunCommand := runCommand
	defer func() {
		userHomeDir = origUserHomeDir
		statFile = origStatFile
		runCommand = origRunCommand
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	statFile = os.Stat
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not needed")
	}

	// Create empty SSH dir.
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)

	cfg := &MigrateConfig{
		SSH: SSHExport{
			Hosts: []SSHHostEntry{
				{Name: "server1", Hostname: "10.0.0.1", User: "deploy", IdentityFile: "~/.ssh/id_rsa"},
				{Name: "server2", Hostname: "10.0.0.2", User: "admin"},
			},
			Keys: []SSHKeyRef{
				{Name: "id_rsa", Type: "RSA", Path: "~/.ssh/id_rsa"},
			},
		},
	}

	result, err := Import(cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	if result.SSHHostsAdded != 2 {
		t.Errorf("expected 2 hosts added, got %d", result.SSHHostsAdded)
	}
	if result.SSHHostsSkipped != 0 {
		t.Errorf("expected 0 hosts skipped, got %d", result.SSHHostsSkipped)
	}
	if !result.SSHConfigWritten {
		t.Error("expected SSH config written")
	}

	// Verify config file content.
	data, err := os.ReadFile(filepath.Join(sshDir, "config"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "Host server1") {
		t.Error("expected server1 in SSH config")
	}
	if !strings.Contains(content, "Host server2") {
		t.Error("expected server2 in SSH config")
	}
	if !strings.Contains(content, filepath.Join(tmpDir, ".ssh", "id_rsa")) {
		t.Error("expected expanded IdentityFile path in SSH config")
	}

	// Key should be needed since it doesn't exist.
	if len(result.SSHKeysNeeded) != 1 {
		t.Fatalf("expected 1 key needed, got %d", len(result.SSHKeysNeeded))
	}
	if !strings.Contains(result.SSHKeysNeeded[0].Command, "ssh-keygen") {
		t.Error("expected ssh-keygen command")
	}
}

func TestImport_SSHHostDuplicate(t *testing.T) {
	origUserHomeDir := userHomeDir
	origStatFile := statFile
	origRunCommand := runCommand
	defer func() {
		userHomeDir = origUserHomeDir
		statFile = origStatFile
		runCommand = origRunCommand
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	statFile = os.Stat
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not needed")
	}

	// Create SSH config with existing host.
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	os.WriteFile(filepath.Join(sshDir, "config"), []byte("Host server1\n  Hostname old.host\n"), 0600)

	cfg := &MigrateConfig{
		SSH: SSHExport{
			Hosts: []SSHHostEntry{
				{Name: "server1", Hostname: "10.0.0.1"},
				{Name: "server2", Hostname: "10.0.0.2"},
			},
		},
	}

	result, err := Import(cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	if result.SSHHostsAdded != 1 {
		t.Errorf("expected 1 host added, got %d", result.SSHHostsAdded)
	}
	if result.SSHHostsSkipped != 1 {
		t.Errorf("expected 1 host skipped, got %d", result.SSHHostsSkipped)
	}
	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}
}

func TestImport_SSHKeyExists(t *testing.T) {
	origUserHomeDir := userHomeDir
	origStatFile := statFile
	origRunCommand := runCommand
	defer func() {
		userHomeDir = origUserHomeDir
		statFile = origStatFile
		runCommand = origRunCommand
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	statFile = os.Stat
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not needed")
	}

	// Create SSH dir and key.
	sshDir := filepath.Join(tmpDir, ".ssh")
	os.MkdirAll(sshDir, 0700)
	os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("key"), 0600)

	cfg := &MigrateConfig{
		SSH: SSHExport{
			Keys: []SSHKeyRef{
				{Name: "id_ed25519", Type: "ED25519", Path: "~/.ssh/id_ed25519"},
			},
		},
	}

	result, err := Import(cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.SSHKeysNeeded) != 0 {
		t.Errorf("expected 0 keys needed (key exists), got %d", len(result.SSHKeysNeeded))
	}
}

func TestImport_GitConfig(t *testing.T) {
	origUserHomeDir := userHomeDir
	origStatFile := statFile
	origRunCommand := runCommand
	origWriteFile := writeFile
	origReadFile := readFile
	origMkdirAll := mkdirAll
	origRunGitConfig := runGitConfig
	defer func() {
		userHomeDir = origUserHomeDir
		statFile = origStatFile
		runCommand = origRunCommand
		writeFile = origWriteFile
		readFile = origReadFile
		mkdirAll = origMkdirAll
		runGitConfig = origRunGitConfig
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	statFile = os.Stat
	writeFile = os.WriteFile
	readFile = os.ReadFile
	mkdirAll = os.MkdirAll
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not needed")
	}

	gitConfigCalls := []string{}
	runGitConfig = func(args ...string) error {
		gitConfigCalls = append(gitConfigCalls, strings.Join(args, " "))
		return nil
	}

	// Create empty gitconfig.
	os.WriteFile(filepath.Join(tmpDir, ".gitconfig"), []byte(""), 0600)
	os.MkdirAll(filepath.Join(tmpDir, ".ssh"), 0700)

	cfg := &MigrateConfig{
		Git: GitExport{
			UserName:  "New User",
			UserEmail: "new@example.com",
			Profiles: []GitProfileExport{
				{Name: "work", Directory: "~/work", Email: "work@company.com", SignKey: "KEY123"},
			},
		},
	}

	result, err := Import(cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	if !result.GitConfigWritten {
		t.Error("expected git config written")
	}

	// Check git config calls.
	if len(gitConfigCalls) != 2 {
		t.Fatalf("expected 2 git config calls, got %d", len(gitConfigCalls))
	}
	if !strings.Contains(gitConfigCalls[0], "user.name") {
		t.Error("expected user.name config call")
	}

	// Check profile directory created.
	if len(result.DirsCreated) != 1 {
		t.Fatalf("expected 1 dir created, got %d", len(result.DirsCreated))
	}

	// Check profile config written.
	if len(result.ProfilesWritten) != 1 {
		t.Fatalf("expected 1 profile written, got %d", len(result.ProfilesWritten))
	}

	// Verify profile config file.
	profileData, err := os.ReadFile(filepath.Join(tmpDir, ".gitconfig-work"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(profileData), "work@company.com") {
		t.Error("expected email in profile config")
	}
	if !strings.Contains(string(profileData), "KEY123") {
		t.Error("expected signing key in profile config")
	}

	// Verify includeIf added to gitconfig.
	gitconfigData, err := os.ReadFile(filepath.Join(tmpDir, ".gitconfig"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gitconfigData), "includeIf") {
		t.Error("expected includeIf in gitconfig")
	}
}

func TestImport_GitProfileExists(t *testing.T) {
	origUserHomeDir := userHomeDir
	origStatFile := statFile
	origRunCommand := runCommand
	origWriteFile := writeFile
	origReadFile := readFile
	origMkdirAll := mkdirAll
	origRunGitConfig := runGitConfig
	defer func() {
		userHomeDir = origUserHomeDir
		statFile = origStatFile
		runCommand = origRunCommand
		writeFile = origWriteFile
		readFile = origReadFile
		mkdirAll = origMkdirAll
		runGitConfig = origRunGitConfig
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	statFile = os.Stat
	writeFile = os.WriteFile
	readFile = os.ReadFile
	mkdirAll = os.MkdirAll
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not needed")
	}
	runGitConfig = func(args ...string) error { return nil }

	os.MkdirAll(filepath.Join(tmpDir, ".ssh"), 0700)
	// Profile config already exists.
	os.WriteFile(filepath.Join(tmpDir, ".gitconfig-work"), []byte("existing"), 0600)
	os.WriteFile(filepath.Join(tmpDir, ".gitconfig"), []byte(""), 0600)

	cfg := &MigrateConfig{
		Git: GitExport{
			Profiles: []GitProfileExport{
				{Name: "work", Directory: "~/work", Email: "work@company.com"},
			},
		},
	}

	result, err := Import(cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	// Should warn and skip.
	if len(result.ProfilesWritten) != 0 {
		t.Error("expected no profiles written when file exists")
	}
	foundWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "already exists") {
			foundWarning = true
		}
	}
	if !foundWarning {
		t.Error("expected warning about existing profile config")
	}
}

func TestImport_DryRun(t *testing.T) {
	origUserHomeDir := userHomeDir
	origStatFile := statFile
	origRunCommand := runCommand
	defer func() {
		userHomeDir = origUserHomeDir
		statFile = origStatFile
		runCommand = origRunCommand
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	statFile = os.Stat
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not needed")
	}

	cfg := &MigrateConfig{
		SSH: SSHExport{
			Hosts: []SSHHostEntry{
				{Name: "server1", Hostname: "10.0.0.1"},
			},
		},
		Git: GitExport{
			UserName: "Test",
		},
	}

	result, err := Import(cfg, true)
	if err != nil {
		t.Fatal(err)
	}

	if result.SSHHostsAdded != 1 {
		t.Errorf("expected 1 host added in dry-run, got %d", result.SSHHostsAdded)
	}

	// Verify no files were actually created.
	sshConfigPath := filepath.Join(tmpDir, ".ssh", "config")
	if _, err := os.Stat(sshConfigPath); err == nil {
		t.Error("expected no SSH config file in dry-run")
	}
}

func TestImport_HomeDirError(t *testing.T) {
	origUserHomeDir := userHomeDir
	defer func() { userHomeDir = origUserHomeDir }()

	userHomeDir = func() (string, error) { return "", fmt.Errorf("no home") }

	_, err := Import(&MigrateConfig{}, false)
	if err == nil {
		t.Error("expected error on home dir failure")
	}
}

func TestImport_SSHKeyNoType(t *testing.T) {
	origUserHomeDir := userHomeDir
	origStatFile := statFile
	origRunCommand := runCommand
	defer func() {
		userHomeDir = origUserHomeDir
		statFile = origStatFile
		runCommand = origRunCommand
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	statFile = os.Stat
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("not needed")
	}

	os.MkdirAll(filepath.Join(tmpDir, ".ssh"), 0700)

	cfg := &MigrateConfig{
		SSH: SSHExport{
			Keys: []SSHKeyRef{
				{Name: "id_custom", Path: "~/.ssh/id_custom"}, // no type
			},
		},
	}

	result, err := Import(cfg, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.SSHKeysNeeded) != 1 {
		t.Fatalf("expected 1 key needed, got %d", len(result.SSHKeysNeeded))
	}
	// Should default to ed25519.
	if !strings.Contains(result.SSHKeysNeeded[0].Command, "ed25519") {
		t.Errorf("expected ed25519 default, got %s", result.SSHKeysNeeded[0].Command)
	}
}

func TestFormatImportResult(t *testing.T) {
	result := &ImportResult{
		SSHHostsAdded:   3,
		SSHHostsSkipped: 1,
		SSHKeysNeeded: []KeyAction{
			{Name: "id_rsa", Command: "ssh-keygen -t rsa -f ~/.ssh/id_rsa"},
		},
		GitConfigWritten: true,
		DirsCreated:      []string{"~/work"},
		ProfilesWritten:  []string{"~/.gitconfig-work"},
		Warnings:         []string{"Host 'old' already exists"},
	}

	output := FormatImportResult(result)

	checks := []string{
		"3 host(s) added",
		"1 skipped",
		"1 key(s) to generate",
		"ssh-keygen",
		"user.name/email configured",
		"1 profile directory",
		"~/.gitconfig-work",
		"Host 'old' already exists",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q", check)
		}
	}
}

func TestFormatImportResult_Empty(t *testing.T) {
	result := &ImportResult{}
	output := FormatImportResult(result)

	if !strings.Contains(output, "no hosts to add") {
		t.Error("expected 'no hosts to add' for empty result")
	}
}

func TestBuildHostBlock(t *testing.T) {
	h := SSHHostEntry{
		Name:         "server1",
		Hostname:     "10.0.0.1",
		User:         "deploy",
		Port:         "2222",
		IdentityFile: "~/.ssh/id_rsa",
		ProxyJump:    "bastion",
		ForwardAgent: true,
	}

	block := buildHostBlock(h, "/home/user")

	checks := []string{
		"Host server1",
		"Hostname 10.0.0.1",
		"User deploy",
		"Port 2222",
		"IdentityFile /home/user/.ssh/id_rsa",
		"ProxyJump bastion",
		"ForwardAgent yes",
	}
	for _, check := range checks {
		if !strings.Contains(block, check) {
			t.Errorf("expected block to contain %q", check)
		}
	}
}

func TestImport_OnlySSH(t *testing.T) {
	origUserHomeDir := userHomeDir
	origWriteFile := writeFile
	origReadFile := readFile
	origMkdirAll := mkdirAll
	origStatFile := statFile
	origOpenFile := openFile
	origRunGitConfig := runGitConfig
	defer func() {
		userHomeDir = origUserHomeDir
		writeFile = origWriteFile
		readFile = origReadFile
		mkdirAll = origMkdirAll
		statFile = origStatFile
		openFile = origOpenFile
		runGitConfig = origRunGitConfig
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	mkdirAll = func(path string, perm os.FileMode) error { return nil }
	statFile = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	openFile = func(name string) (*os.File, error) { return nil, os.ErrNotExist }
	writeFile = func(name string, data []byte, perm os.FileMode) error { return nil }
	readFile = func(name string) ([]byte, error) { return nil, os.ErrNotExist }
	runGitConfig = func(args ...string) error {
		t.Error("git config should not be called with --only ssh")
		return nil
	}

	cfg := &MigrateConfig{
		SSH: SSHExport{
			Hosts: []SSHHostEntry{{Name: "server1", Hostname: "10.0.0.1"}},
		},
		Git: GitExport{
			UserName:  "test",
			UserEmail: "test@example.com",
		},
	}

	only := map[string]bool{"ssh": true}
	result, err := Import(cfg, true, only)
	if err != nil {
		t.Fatal(err)
	}

	if result.SSHHostsAdded != 1 {
		t.Errorf("expected 1 SSH host added, got %d", result.SSHHostsAdded)
	}
	if result.GitConfigWritten {
		t.Error("expected git config NOT written with --only ssh")
	}
}

func TestImport_OnlyGit(t *testing.T) {
	origUserHomeDir := userHomeDir
	origWriteFile := writeFile
	origReadFile := readFile
	origMkdirAll := mkdirAll
	origStatFile := statFile
	origOpenFile := openFile
	origRunGitConfig := runGitConfig
	defer func() {
		userHomeDir = origUserHomeDir
		writeFile = origWriteFile
		readFile = origReadFile
		mkdirAll = origMkdirAll
		statFile = origStatFile
		openFile = origOpenFile
		runGitConfig = origRunGitConfig
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	mkdirAll = func(path string, perm os.FileMode) error { return nil }
	statFile = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	writeFile = func(name string, data []byte, perm os.FileMode) error { return nil }
	readFile = func(name string) ([]byte, error) { return nil, os.ErrNotExist }
	runGitConfig = func(args ...string) error { return nil }

	cfg := &MigrateConfig{
		SSH: SSHExport{
			Hosts: []SSHHostEntry{{Name: "server1", Hostname: "10.0.0.1"}},
		},
		Git: GitExport{
			UserName:  "test",
			UserEmail: "test@example.com",
		},
	}

	only := map[string]bool{"git": true}
	result, err := Import(cfg, false, only)
	if err != nil {
		t.Fatal(err)
	}

	if result.SSHHostsAdded != 0 {
		t.Errorf("expected 0 SSH hosts with --only git, got %d", result.SSHHostsAdded)
	}
	if !result.GitConfigWritten {
		t.Error("expected git config written with --only git")
	}
}
