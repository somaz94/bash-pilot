package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/somaz94/bash-pilot/internal/ssh"
)

func TestAuditSSHConfig(t *testing.T) {
	t.Run("unreadable config is a fail finding", func(t *testing.T) {
		result := auditSSHConfig(filepath.Join(t.TempDir(), "missing"))
		if len(result.Findings) != 1 {
			t.Fatalf("findings = %+v, want exactly one", result.Findings)
		}
		f := result.Findings[0]
		if f.Severity != ssh.SeverityFail || !strings.HasPrefix(f.Message, "Cannot read SSH config: ") {
			t.Errorf("finding = %+v", f)
		}
	})
	t.Run("readable config is audited", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config")
		if err := os.WriteFile(path, []byte("Host a\n  Hostname 192.0.2.1\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		result := auditSSHConfig(path)
		for _, f := range result.Findings {
			if strings.HasPrefix(f.Message, "Cannot read") {
				t.Errorf("unexpected read failure: %+v", f)
			}
		}
		if len(result.Findings) == 0 {
			t.Error("expected the no-IdentityFile finding for host a")
		}
	})
}
