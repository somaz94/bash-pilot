package ssh

import (
	"os"
	"path/filepath"
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

	// Create a key file with bad permissions.
	if err := os.WriteFile(keyPath, []byte("fake-key"), 0644); err != nil {
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
		if f.Severity == SeverityFail && f.Message == "key file not found" {
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
			break
		}
	}
	if !found {
		t.Error("expected warning about missing IdentityFile")
	}
}
