package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/somaz94/bash-pilot/internal/ssh"
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

	if len(gitConfigCalls) != 2 {
		t.Fatalf("expected 2 git config calls, got %d", len(gitConfigCalls))
	}
	if !strings.Contains(gitConfigCalls[0], "user.name") {
		t.Error("expected user.name config call")
	}

	if len(result.DirsCreated) != 1 {
		t.Fatalf("expected 1 dir created, got %d", len(result.DirsCreated))
	}

	if len(result.ProfilesWritten) != 1 {
		t.Fatalf("expected 1 profile written, got %d", len(result.ProfilesWritten))
	}

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

	block := buildHostBlock(h, "/home/user", true)

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

func TestBuildHostBlock_ProxyCommand(t *testing.T) {
	tests := []struct {
		name  string
		entry SSHHostEntry
		want  string
	}{
		{"proxy_command field", SSHHostEntry{Name: "a", ProxyCommand: "ssh -W %h:%p bastion"}, "  ProxyCommand ssh -W %h:%p bastion\n"},
		{"legacy command in proxy_jump", SSHHostEntry{Name: "a", ProxyJump: "ssh -W %h:%p bastion"}, "  ProxyCommand ssh -W %h:%p bastion\n"},
		{"jump spec stays ProxyJump", SSHHostEntry{Name: "a", ProxyJump: "admin@bastion:2222"}, "  ProxyJump admin@bastion:2222\n"},
		{"jump spec with trailing comment stays ProxyJump", SSHHostEntry{Name: "a", ProxyJump: "bastion # office"}, "  ProxyJump bastion # office\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := buildHostBlock(tt.entry, "/home/user", false)
			if !strings.Contains(block, tt.want) {
				t.Errorf("block =\n%s\nwant line %q", block, tt.want)
			}
			if strings.Count(block, "Proxy") != 1 {
				t.Errorf("block should carry exactly one proxy directive:\n%s", block)
			}
		})
	}
}

func TestImport_OnlySSH(t *testing.T) {
	origUserHomeDir := userHomeDir
	origWriteFile := writeFile
	origReadFile := readFile
	origMkdirAll := mkdirAll
	origStatFile := statFile
	origRunGitConfig := runGitConfig
	defer func() {
		userHomeDir = origUserHomeDir
		writeFile = origWriteFile
		readFile = origReadFile
		mkdirAll = origMkdirAll
		statFile = origStatFile
		runGitConfig = origRunGitConfig
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	mkdirAll = func(path string, perm os.FileMode) error { return nil }
	statFile = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
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
	origRunGitConfig := runGitConfig
	defer func() {
		userHomeDir = origUserHomeDir
		writeFile = origWriteFile
		readFile = origReadFile
		mkdirAll = origMkdirAll
		statFile = origStatFile
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

func TestAddIncludeIf_ParentOfExisting(t *testing.T) {
	home := t.TempDir()
	existing := "[includeIf \"gitdir:~/work/sub/\"]\n\tpath = ~/.gitconfig-sub\n"
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := addIncludeIf(home, GitProfileExport{Name: "work", Directory: "~/work"}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(home, ".gitconfig"))
	if !strings.Contains(string(got), "[includeIf \"gitdir:~/work/\"]") {
		t.Errorf("includeIf for ~/work/ was not written:\n%s", got)
	}
}

func TestParseExistingHosts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	content := "Host plain\n" +
		"Host=equals\n" +
		"Host\ttabbed\n" +
		"Host multi-a multi-b\n" +
		"Host bastion *\n" +
		"Host *\n" +
		"  HostName not-a-host\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	got := parseExistingHosts(path)
	for _, name := range []string{"plain", "equals", "tabbed", "multi-a", "multi-b", "bastion"} {
		if !got[name] {
			t.Errorf("%q not found in %v", name, got)
		}
	}
	for _, name := range []string{"*", "not-a-host", "multi-a multi-b"} {
		if got[name] {
			t.Errorf("%q should not be an existing host: %v", name, got)
		}
	}

	if missing := parseExistingHosts(filepath.Join(t.TempDir(), "none")); len(missing) != 0 {
		t.Errorf("missing config = %v, want empty", missing)
	}
}

func TestImport_SSHHostSharesExistingPattern(t *testing.T) {
	origUserHomeDir := userHomeDir
	defer func() { userHomeDir = origUserHomeDir }()
	home := t.TempDir()
	userHomeDir = func() (string, error) { return home, nil }

	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "config"), []byte("Host=web1 web2\n  HostName 192.0.2.10\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &MigrateConfig{SSH: SSHExport{Hosts: []SSHHostEntry{
		{Name: "web2", Hostname: "192.0.2.20"},
		{Name: "web9 web1", Hostname: "192.0.2.90"},
		{Name: "web3", Hostname: "192.0.2.30"},
	}}}
	result, err := Import(cfg, true, map[string]bool{"ssh": true})
	if err != nil {
		t.Fatal(err)
	}
	if result.SSHHostsSkipped != 2 || result.SSHHostsAdded != 1 {
		t.Errorf("skipped = %d, added = %d; want web2 and 'web9 web1' skipped, web3 added", result.SSHHostsSkipped, result.SSHHostsAdded)
	}
	want := []string{
		"Host 'web2' already exists in SSH config, skipping",
		"Host 'web9 web1' shares 'web1' with an existing SSH config entry, skipping",
	}
	if strings.Join(result.Warnings, "\n") != strings.Join(want, "\n") {
		t.Errorf("warnings = %q, want %q", result.Warnings, want)
	}
}

func TestSSHArg(t *testing.T) {
	tests := []struct{ in, want string }{
		{"192.0.2.5", "192.0.2.5"},
		{"/keys/my key", `"/keys/my key"`},
		{`o'brien`, `"o'brien"`},
		{"web#1", `"web#1"`},
		{`a"b\c`, `"a\"b\\c"`},
		{"=foo", `"=foo"`},
	}
	for _, tt := range tests {
		if got := sshArg(tt.in); got != tt.want {
			t.Errorf("sshArg(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

// A source config with quoting must survive export, import and a re-parse unchanged.
func TestExportImport_QuotedValuesRoundTrip(t *testing.T) {
	oldHome, newHome := t.TempDir(), t.TempDir()
	t.Setenv("HOME", oldHome)
	src := filepath.Join(oldHome, "config")
	content := "Host \"a'b\" odd\n  HostName 192.0.2.5\n  User o\\'brien\n  IdentityFile \"~/.ssh/my key #1\"\n  ProxyJump jump\n"
	if err := os.WriteFile(src, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := &MigrateConfig{}
	exportSSH(cfg, src, oldHome)
	if len(cfg.SSH.Hosts) != 1 {
		t.Fatalf("exported hosts = %+v", cfg.SSH.Hosts)
	}

	t.Setenv("HOME", newHome)
	dst := filepath.Join(newHome, "config")
	if err := os.WriteFile(dst, []byte(buildHostBlock(cfg.SSH.Hosts[0], newHome, true)), 0o600); err != nil {
		t.Fatal(err)
	}
	hosts, err := ssh.ParseConfig(dst)
	if err != nil || len(hosts) != 1 {
		t.Fatalf("re-parse: hosts = %+v, err = %v", hosts, err)
	}
	h := hosts[0]
	if want := filepath.Join(newHome, ".ssh", "my key #1"); h.IdentityFile != want {
		t.Errorf("IdentityFile = %q, want %q", h.IdentityFile, want)
	}
	if h.Name != "a'b odd" || h.User != "o'brien" || h.Hostname != "192.0.2.5" || h.ProxyJump != "jump" {
		t.Errorf("host = %+v", h)
	}
}

// Version 1 exports hold raw ssh_config text, so quoting it would turn a trailing
// comment into part of the value; version 2 values are words that need quoting.
func TestImport_QuotesOnlyFormat2Values(t *testing.T) {
	origUserHomeDir := userHomeDir
	defer func() { userHomeDir = origUserHomeDir }()
	for _, tt := range []struct {
		version, hostname, want string
		wantWarning             bool
	}{
		{version: "1", hostname: "192.0.2.5 # primary", want: "  Hostname 192.0.2.5 # primary\n"},
		{version: "", hostname: "192.0.2.5 # primary", want: "  Hostname 192.0.2.5 # primary\n"},
		{version: FormatVersion, hostname: "my host", want: "  Hostname \"my host\"\n"},
		{version: "3", hostname: "my host", want: "  Hostname \"my host\"\n", wantWarning: true},
	} {
		t.Run("version "+tt.version, func(t *testing.T) {
			home := t.TempDir()
			userHomeDir = func() (string, error) { return home, nil }
			cfg := &MigrateConfig{Version: tt.version, SSH: SSHExport{Hosts: []SSHHostEntry{{Name: "h", Hostname: tt.hostname}}}}
			result, err := Import(cfg, false, map[string]bool{"ssh": true})
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(filepath.Join(home, ".ssh", "config"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(got), tt.want) {
				t.Errorf("config =\n%s\nwant line %q", got, tt.want)
			}
			warned := strings.Contains(strings.Join(result.Warnings, "\n"), "unknown export format version")
			if warned != tt.wantWarning {
				t.Errorf("warnings = %q, want unknown-version warning: %v", result.Warnings, tt.wantWarning)
			}
		})
	}
}
