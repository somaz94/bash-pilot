package ssh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAudit_SharedKeys(t *testing.T) {
	hosts := []Host{
		{Name: "server1", IdentityFile: "/tmp/test-key"},
		{Name: "server2", IdentityFile: "/tmp/test-key"},
		{Name: "server3", IdentityFile: "/tmp/test-key"},
		{Name: "server4", IdentityFile: "/tmp/test-key"},
	}

	result := Audit(hosts)

	// Should warn about shared key (4 hosts > 3 threshold).
	found := false
	for _, f := range result.Findings {
		if f.Severity == SeverityWarn && f.Key == "test-key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected warning about shared key 'test-key'")
	}
}

func TestAudit_KeyPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key")

	if err := os.WriteFile(keyPath, []byte("fake-key"), 0600); err != nil {
		t.Fatal(err)
	}
	// Chmod, not WriteFile's mode, so the result does not depend on umask.
	if err := os.Chmod(keyPath, 0644); err != nil {
		t.Fatal(err)
	}

	hosts := []Host{
		{Name: "server1", IdentityFile: keyPath},
	}

	result := Audit(hosts)

	foundPerm := false
	for _, f := range result.Findings {
		if f.Severity == SeverityWarn && f.Key == "test_key" {
			foundPerm = true
			break
		}
	}
	if !foundPerm {
		t.Error("expected warning about key permissions")
	}
}

func TestAudit_MissingKey(t *testing.T) {
	hosts := []Host{
		{Name: "server1", IdentityFile: "/nonexistent/key"},
	}

	result := Audit(hosts)

	foundMissing := false
	for _, f := range result.Findings {
		if f.Severity == SeverityFail && strings.Contains(f.Message, "key file not found") {
			foundMissing = true
			break
		}
	}
	if !foundMissing {
		t.Error("expected fail finding for missing key")
	}
}

func TestAudit_NoIdentityFile(t *testing.T) {
	hosts := []Host{
		{Name: "server1"},
	}

	result := Audit(hosts)

	found := false
	for _, f := range result.Findings {
		if f.Severity == SeverityWarn && f.Key == "server1" {
			found = true
			if want := "server1: no IdentityFile specified (will use default keys)"; f.Message != want {
				t.Errorf("Message = %q, want %q", f.Message, want)
			}
			break
		}
	}
	if !found {
		t.Error("expected warning about missing IdentityFile")
	}
}

func TestAudit_KeyFindingsFollowConfigOrder(t *testing.T) {
	hosts := []Host{
		{Name: "h1", IdentityFile: "/nonexistent/k1"},
		{Name: "h2", IdentityFile: "/nonexistent/k2"},
		{Name: "h3", IdentityFile: "/nonexistent/k3"},
		{Name: "h4", IdentityFile: "/nonexistent/k4"},
	}
	var got []string
	for _, f := range Audit(hosts).Findings {
		if f.Severity == SeverityOK {
			got = append(got, f.Key)
		}
	}
	want := []string{"k1", "k2", "k3", "k4"}
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Fatalf("usage findings order = %v, want %v", got, want)
		}
	}
}
