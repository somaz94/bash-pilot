package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	result := Check()

	if len(result.Findings) == 0 {
		t.Error("expected at least some findings")
	}

	// Should have shell finding.
	found := false
	for _, f := range result.Findings {
		if f.Category == "shell" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected shell category in findings")
	}
}

func TestCheck_HasToolFindings(t *testing.T) {
	result := Check()

	toolCount := 0
	for _, f := range result.Findings {
		if f.Category == "tools" {
			toolCount++
		}
	}
	if toolCount == 0 {
		t.Error("expected tool findings")
	}
}

func TestCheck_HasHomeFindings(t *testing.T) {
	result := Check()

	found := false
	for _, f := range result.Findings {
		if f.Category == "home" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected home category in findings")
	}
}

func TestAnalyzePath(t *testing.T) {
	result := AnalyzePath()

	if result.Total == 0 {
		t.Error("expected at least one PATH entry")
	}

	if len(result.Entries) != result.Total {
		t.Errorf("entries count %d != total %d", len(result.Entries), result.Total)
	}
}

func TestAnalyzePath_WithDuplicates(t *testing.T) {
	original := os.Getenv("PATH")
	defer os.Setenv("PATH", original)

	os.Setenv("PATH", "/usr/bin:/usr/local/bin:/usr/bin")

	result := AnalyzePath()

	if len(result.Duplicates) == 0 {
		t.Error("expected duplicates to be detected")
	}
}

func TestAnalyzePath_WithMissing(t *testing.T) {
	original := os.Getenv("PATH")
	defer os.Setenv("PATH", original)

	os.Setenv("PATH", "/usr/bin:/nonexistent/path/that/does/not/exist")

	result := AnalyzePath()

	if len(result.Missing) == 0 {
		t.Error("expected missing directories to be detected")
	}
}

func TestAnalyzePath_EmptyPath(t *testing.T) {
	original := os.Getenv("PATH")
	defer os.Setenv("PATH", original)

	os.Setenv("PATH", "")

	result := AnalyzePath()

	if result.Total != 0 {
		t.Errorf("expected 0 entries for empty PATH, got %d", result.Total)
	}
}

func TestExpandEnvPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/bin", filepath.Join(home, "bin")},
		{"/usr/bin", "/usr/bin"},
		{"relative", "relative"},
	}

	for _, tt := range tests {
		got := expandEnvPath(tt.input)
		if got != tt.expected {
			t.Errorf("expandEnvPath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestSummarizeFindings(t *testing.T) {
	findings := []Finding{
		{Severity: "ok"},
		{Severity: "ok"},
		{Severity: "warn"},
		{Severity: "error"},
		{Severity: "ok"},
	}

	ok, warn, errCount := SummarizeFindings(findings)
	if ok != 3 {
		t.Errorf("expected 3 ok, got %d", ok)
	}
	if warn != 1 {
		t.Errorf("expected 1 warn, got %d", warn)
	}
	if errCount != 1 {
		t.Errorf("expected 1 error, got %d", errCount)
	}
}

func TestSummarizeFindings_Empty(t *testing.T) {
	ok, warn, errCount := SummarizeFindings(nil)
	if ok != 0 || warn != 0 || errCount != 0 {
		t.Error("expected all zeros for empty findings")
	}
}

func TestGroupFindingsByCategory(t *testing.T) {
	findings := []Finding{
		{Severity: "ok", Category: "shell", Message: "test1"},
		{Severity: "ok", Category: "tools", Message: "test2"},
		{Severity: "warn", Category: "shell", Message: "test3"},
		{Severity: "ok", Category: "git", Message: "test4"},
	}

	groups, keys := GroupFindingsByCategory(findings)

	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}

	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	}

	// Keys should be sorted.
	for i := 1; i < len(keys); i++ {
		if keys[i] < keys[i-1] {
			t.Error("keys should be sorted")
		}
	}

	if len(groups["shell"]) != 2 {
		t.Errorf("expected 2 shell findings, got %d", len(groups["shell"]))
	}
}

func TestCheckShell(t *testing.T) {
	result := &CheckResult{}
	checkShell(result)

	if len(result.Findings) == 0 {
		t.Error("expected at least one shell finding")
	}
}

func TestCheckShell_NoShell(t *testing.T) {
	original := os.Getenv("SHELL")
	defer os.Setenv("SHELL", original)

	os.Setenv("SHELL", "")

	result := &CheckResult{}
	checkShell(result)

	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	if result.Findings[0].Severity != "warn" {
		t.Error("expected warning when SHELL not set")
	}
}

func TestCheckSSHAgent_NoSock(t *testing.T) {
	original := os.Getenv("SSH_AUTH_SOCK")
	defer os.Setenv("SSH_AUTH_SOCK", original)

	os.Setenv("SSH_AUTH_SOCK", "")

	result := &CheckResult{}
	checkSSHAgent(result)

	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	if result.Findings[0].Severity != "warn" {
		t.Error("expected warning when SSH_AUTH_SOCK not set")
	}
}

func TestCheckSSHAgent_MissingSock(t *testing.T) {
	original := os.Getenv("SSH_AUTH_SOCK")
	defer os.Setenv("SSH_AUTH_SOCK", original)

	os.Setenv("SSH_AUTH_SOCK", "/nonexistent/socket")

	result := &CheckResult{}
	checkSSHAgent(result)

	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	if !strings.Contains(result.Findings[0].Message, "missing socket") {
		t.Error("expected message about missing socket")
	}
}

func TestCheckEditorSet(t *testing.T) {
	original := os.Getenv("EDITOR")
	originalVisual := os.Getenv("VISUAL")
	defer func() {
		os.Setenv("EDITOR", original)
		os.Setenv("VISUAL", originalVisual)
	}()

	os.Setenv("EDITOR", "")
	os.Setenv("VISUAL", "")

	result := &CheckResult{}
	checkEditorSet(result)

	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	if result.Findings[0].Severity != "warn" {
		t.Error("expected warning when no editor set")
	}
}

func TestCheckEditorSet_WithEditor(t *testing.T) {
	original := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", original)

	os.Setenv("EDITOR", "vim")

	result := &CheckResult{}
	checkEditorSet(result)

	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	if result.Findings[0].Severity != "ok" {
		t.Error("expected ok when EDITOR is set")
	}
}

func TestCheckEditorSet_WithVisual(t *testing.T) {
	originalEditor := os.Getenv("EDITOR")
	originalVisual := os.Getenv("VISUAL")
	defer func() {
		os.Setenv("EDITOR", originalEditor)
		os.Setenv("VISUAL", originalVisual)
	}()

	os.Setenv("EDITOR", "")
	os.Setenv("VISUAL", "code")

	result := &CheckResult{}
	checkEditorSet(result)

	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	if result.Findings[0].Severity != "ok" {
		t.Error("expected ok when VISUAL is set")
	}
}

func TestAnalyzePath_AllExist(t *testing.T) {
	original := os.Getenv("PATH")
	defer os.Setenv("PATH", original)

	os.Setenv("PATH", "/usr/bin:/tmp")

	result := AnalyzePath()

	if len(result.Missing) != 0 {
		t.Errorf("expected no missing dirs, got %d", len(result.Missing))
	}
}

// --- Tests using function variable overrides ---

func TestCheckCommonTools_AllFound(t *testing.T) {
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	lookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}

	result := &CheckResult{}
	checkCommonTools(result)

	for _, f := range result.Findings {
		if f.Severity != "ok" {
			t.Errorf("expected all ok, got %s for %s", f.Severity, f.Message)
		}
	}
	if len(result.Findings) != 11 {
		t.Errorf("expected 11 findings, got %d", len(result.Findings))
	}
}

func TestCheckCommonTools_RequiredMissing(t *testing.T) {
	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()

	lookPath = func(file string) (string, error) {
		return "", fmt.Errorf("not found")
	}

	result := &CheckResult{}
	checkCommonTools(result)

	errorCount := 0
	warnCount := 0
	for _, f := range result.Findings {
		switch f.Severity {
		case "error":
			errorCount++
		case "warn":
			warnCount++
		}
	}
	// git, ssh, curl are required → error
	if errorCount != 3 {
		t.Errorf("expected 3 errors (required tools), got %d", errorCount)
	}
	// the rest are optional → warn
	if warnCount != 8 {
		t.Errorf("expected 8 warnings (optional tools), got %d", warnCount)
	}
}

func TestCheckSSHAgent_WithKeys(t *testing.T) {
	origSock := os.Getenv("SSH_AUTH_SOCK")
	origStat := statFunc
	origRun := runCommand
	defer func() {
		os.Setenv("SSH_AUTH_SOCK", origSock)
		statFunc = origStat
		runCommand = origRun
	}()

	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "agent.sock")
	os.WriteFile(sockPath, []byte{}, 0600)
	os.Setenv("SSH_AUTH_SOCK", sockPath)

	statFunc = os.Stat
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "ssh-add" {
			return []byte("2048 SHA256:abc key1 (RSA)\n4096 SHA256:def key2 (ED25519)"), nil
		}
		return nil, fmt.Errorf("unexpected command")
	}

	result := &CheckResult{}
	checkSSHAgent(result)

	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	if result.Findings[0].Severity != "ok" {
		t.Errorf("expected ok, got %s", result.Findings[0].Severity)
	}
	if !strings.Contains(result.Findings[0].Message, "2 key(s) loaded") {
		t.Errorf("expected 2 keys message, got %s", result.Findings[0].Message)
	}
}

func TestCheckSSHAgent_NoKeys(t *testing.T) {
	origSock := os.Getenv("SSH_AUTH_SOCK")
	origStat := statFunc
	origRun := runCommand
	defer func() {
		os.Setenv("SSH_AUTH_SOCK", origSock)
		statFunc = origStat
		runCommand = origRun
	}()

	tmpDir := t.TempDir()
	sockPath := filepath.Join(tmpDir, "agent.sock")
	os.WriteFile(sockPath, []byte{}, 0600)
	os.Setenv("SSH_AUTH_SOCK", sockPath)

	statFunc = os.Stat
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("exit status 1")
	}

	result := &CheckResult{}
	checkSSHAgent(result)

	if len(result.Findings) == 0 {
		t.Fatal("expected findings")
	}
	if !strings.Contains(result.Findings[0].Message, "no keys loaded") {
		t.Errorf("expected no keys message, got %s", result.Findings[0].Message)
	}
}

func TestCheckGitConfig_BothSet(t *testing.T) {
	origRun := runCommand
	defer func() { runCommand = origRun }()

	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) >= 3 {
			switch args[2] {
			case "user.email":
				return []byte("user@example.com\n"), nil
			case "user.name":
				return []byte("Test User\n"), nil
			}
		}
		return nil, fmt.Errorf("unexpected")
	}

	result := &CheckResult{}
	checkGitConfig(result)

	if len(result.Findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(result.Findings))
	}
	for _, f := range result.Findings {
		if f.Severity != "ok" {
			t.Errorf("expected ok, got %s: %s", f.Severity, f.Message)
		}
	}
}

func TestCheckGitConfig_NoneSet(t *testing.T) {
	origRun := runCommand
	defer func() { runCommand = origRun }()

	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("exit status 1")
	}

	result := &CheckResult{}
	checkGitConfig(result)

	if len(result.Findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(result.Findings))
	}
	for _, f := range result.Findings {
		if f.Severity != "warn" {
			t.Errorf("expected warn, got %s: %s", f.Severity, f.Message)
		}
	}
}

func TestCheckHomeDir_WithTempDir(t *testing.T) {
	origHome := userHomeDir
	origStat := statFunc
	defer func() {
		userHomeDir = origHome
		statFunc = origStat
	}()

	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".ssh"), 0700)
	os.MkdirAll(filepath.Join(tmpDir, ".config"), 0755)

	userHomeDir = func() (string, error) { return tmpDir, nil }
	statFunc = os.Stat

	result := &CheckResult{}
	checkHomeDir(result)

	if len(result.Findings) < 2 {
		t.Fatalf("expected at least 2 findings, got %d", len(result.Findings))
	}
	// .ssh should be ok (0700)
	if result.Findings[0].Severity != "ok" {
		t.Errorf("expected ok for .ssh, got %s: %s", result.Findings[0].Severity, result.Findings[0].Message)
	}
	// .config should be ok
	if result.Findings[1].Severity != "ok" {
		t.Errorf("expected ok for .config, got %s: %s", result.Findings[1].Severity, result.Findings[1].Message)
	}
}

func TestCheckHomeDir_SSHPermsTooOpen(t *testing.T) {
	origHome := userHomeDir
	origStat := statFunc
	defer func() {
		userHomeDir = origHome
		statFunc = origStat
	}()

	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".ssh"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, ".config"), 0755)

	userHomeDir = func() (string, error) { return tmpDir, nil }
	statFunc = os.Stat

	result := &CheckResult{}
	checkHomeDir(result)

	if len(result.Findings) < 1 {
		t.Fatal("expected findings")
	}
	if result.Findings[0].Severity != "warn" {
		t.Errorf("expected warn for .ssh perms, got %s: %s", result.Findings[0].Severity, result.Findings[0].Message)
	}
	if !strings.Contains(result.Findings[0].Message, "permissions too open") {
		t.Errorf("expected permissions warning, got %s", result.Findings[0].Message)
	}
}

func TestCheckHomeDir_MissingDirs(t *testing.T) {
	origHome := userHomeDir
	origStat := statFunc
	defer func() {
		userHomeDir = origHome
		statFunc = origStat
	}()

	tmpDir := t.TempDir()
	// Don't create .ssh or .config

	userHomeDir = func() (string, error) { return tmpDir, nil }
	statFunc = os.Stat

	result := &CheckResult{}
	checkHomeDir(result)

	warnCount := 0
	for _, f := range result.Findings {
		if f.Severity == "warn" {
			warnCount++
		}
	}
	// .ssh not found, .config not found, profile not found = 3 warnings
	if warnCount < 2 {
		t.Errorf("expected at least 2 warnings for missing dirs, got %d", warnCount)
	}
}

func TestCheckHomeDir_HomeDirError(t *testing.T) {
	origHome := userHomeDir
	defer func() { userHomeDir = origHome }()

	userHomeDir = func() (string, error) { return "", errors.New("no home") }

	result := &CheckResult{}
	checkHomeDir(result)

	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Findings))
	}
	if result.Findings[0].Severity != "error" {
		t.Errorf("expected error, got %s", result.Findings[0].Severity)
	}
}

func TestCheckShell_WithBash(t *testing.T) {
	origShell := os.Getenv("SHELL")
	origRun := runCommand
	defer func() {
		os.Setenv("SHELL", origShell)
		runCommand = origRun
	}()

	os.Setenv("SHELL", "/bin/bash")
	runCommand = func(name string, args ...string) ([]byte, error) {
		return []byte("GNU bash, version 5.2.15(1)-release (aarch64-apple-darwin)\n"), nil
	}

	result := &CheckResult{}
	checkShell(result)

	if len(result.Findings) < 2 {
		t.Fatalf("expected at least 2 findings, got %d", len(result.Findings))
	}
	// First: shell name, second: bash version
	if result.Findings[1].Severity != "ok" {
		t.Errorf("expected ok for bash 5.x, got %s: %s", result.Findings[1].Severity, result.Findings[1].Message)
	}
}

func TestCheckShell_OldBash(t *testing.T) {
	origShell := os.Getenv("SHELL")
	origRun := runCommand
	defer func() {
		os.Setenv("SHELL", origShell)
		runCommand = origRun
	}()

	os.Setenv("SHELL", "/bin/bash")
	runCommand = func(name string, args ...string) ([]byte, error) {
		return []byte("GNU bash, version 3.2.57(1)-release (x86_64-apple-darwin)\n"), nil
	}

	result := &CheckResult{}
	checkShell(result)

	found := false
	for _, f := range result.Findings {
		if f.Severity == "warn" && strings.Contains(f.Message, "Old bash version") {
			found = true
		}
	}
	if !found {
		t.Error("expected warning for old bash version 3.x")
	}
}

func TestCheckShell_VersionError(t *testing.T) {
	origShell := os.Getenv("SHELL")
	origRun := runCommand
	defer func() {
		os.Setenv("SHELL", origShell)
		runCommand = origRun
	}()

	os.Setenv("SHELL", "/bin/bash")
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("command failed")
	}

	result := &CheckResult{}
	checkShell(result)

	// Should still have the shell finding, just no version
	if len(result.Findings) != 1 {
		t.Errorf("expected 1 finding (shell only, no version), got %d", len(result.Findings))
	}
}

func TestExpandEnvPath_HomeDirError(t *testing.T) {
	origHome := userHomeDir
	defer func() { userHomeDir = origHome }()

	userHomeDir = func() (string, error) { return "", errors.New("no home") }

	got := expandEnvPath("~/bin")
	if got != "~/bin" {
		t.Errorf("expected ~/bin unchanged on error, got %s", got)
	}
}

func TestAnalyzePath_DuplicateAlreadyReported(t *testing.T) {
	original := os.Getenv("PATH")
	defer os.Setenv("PATH", original)

	// Three occurrences: second triggers dup, third should not re-add
	os.Setenv("PATH", "/usr/bin:/tmp:/usr/bin:/usr/bin")

	result := AnalyzePath()

	if len(result.Duplicates) != 1 {
		t.Errorf("expected 1 duplicate entry, got %d: %v", len(result.Duplicates), result.Duplicates)
	}
}
