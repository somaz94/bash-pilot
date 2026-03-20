package ssh

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfig(t *testing.T) {
	content := `# Test SSH config
Host github.com-somaz94
  Hostname github.com
  User somaz
  IdentityFile ~/.ssh/id_rsa_somaz94

Host test-server
  Hostname 3.65.182.184
  User ec2-user
  IdentityFile ~/.ssh/test.pem

Host *
  ServerAliveInterval 60
`

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	hosts, err := ParseConfig(cfgPath)
	if err != nil {
		t.Fatalf("ParseConfig() error: %v", err)
	}

	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	// Verify first host.
	h := hosts[0]
	if h.Name != "github.com-somaz94" {
		t.Errorf("host[0].Name = %q, want %q", h.Name, "github.com-somaz94")
	}
	if h.Hostname != "github.com" {
		t.Errorf("host[0].Hostname = %q, want %q", h.Hostname, "github.com")
	}
	if h.User != "somaz" {
		t.Errorf("host[0].User = %q, want %q", h.User, "somaz")
	}

	// Verify second host.
	h = hosts[1]
	if h.Name != "test-server" {
		t.Errorf("host[1].Name = %q, want %q", h.Name, "test-server")
	}
	if h.Hostname != "3.65.182.184" {
		t.Errorf("host[1].Hostname = %q, want %q", h.Hostname, "3.65.182.184")
	}
}

func TestParseConfig_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config")
	if err := os.WriteFile(cfgPath, []byte("# empty\n"), 0644); err != nil {
		t.Fatal(err)
	}

	hosts, err := ParseConfig(cfgPath)
	if err != nil {
		t.Fatalf("ParseConfig() error: %v", err)
	}
	if len(hosts) != 0 {
		t.Errorf("expected 0 hosts, got %d", len(hosts))
	}
}

func TestParseConfig_NotFound(t *testing.T) {
	_, err := ParseConfig("/nonexistent/path/config")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParseKeyValue(t *testing.T) {
	tests := []struct {
		line      string
		wantKey   string
		wantValue string
	}{
		{"Hostname github.com", "Hostname", "github.com"},
		{"User ec2-user", "User", "ec2-user"},
		{"IdentityFile ~/.ssh/id_rsa", "IdentityFile", "~/.ssh/id_rsa"},
	}

	for _, tt := range tests {
		key, value := parseKeyValue(tt.line)
		if key != tt.wantKey || value != tt.wantValue {
			t.Errorf("parseKeyValue(%q) = (%q, %q), want (%q, %q)",
				tt.line, key, value, tt.wantKey, tt.wantValue)
		}
	}
}
