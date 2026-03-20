package env

import (
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
