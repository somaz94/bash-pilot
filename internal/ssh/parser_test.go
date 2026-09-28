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
  User git
  IdentityFile ~/.ssh/id_rsa_somaz94

Host test-server
  Hostname 198.51.100.67
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

	h := hosts[0]
	if h.Name != "github.com-somaz94" {
		t.Errorf("host[0].Name = %q, want %q", h.Name, "github.com-somaz94")
	}
	if h.Hostname != "github.com" {
		t.Errorf("host[0].Hostname = %q, want %q", h.Hostname, "github.com")
	}
	if h.User != "git" {
		t.Errorf("host[0].User = %q, want %q", h.User, "git")
	}

	h = hosts[1]
	if h.Name != "test-server" {
		t.Errorf("host[1].Name = %q, want %q", h.Name, "test-server")
	}
	if h.Hostname != "198.51.100.67" {
		t.Errorf("host[1].Hostname = %q, want %q", h.Hostname, "198.51.100.67")
	}
}

func TestParseConfig_AllDirectives(t *testing.T) {
	content := `
Include /some/other/config

Host jump-server
  Hostname 203.0.113.50
  User admin
  Port 2222
  IdentityFile /absolute/path/key
  ProxyJump bastion
  ForwardAgent yes

Host proxy-host
  Hostname 10.0.0.1
  User deploy
  ProxyCommand ssh -W %h:%p bastion
  ForwardAgent no
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

	h := hosts[0]
	if h.Port != "2222" {
		t.Errorf("Port = %q, want %q", h.Port, "2222")
	}
	if h.ProxyJump != "bastion" {
		t.Errorf("ProxyJump = %q, want %q", h.ProxyJump, "bastion")
	}
	if !h.ForwardAgent {
		t.Error("ForwardAgent should be true")
	}
	if h.IdentityFile != "/absolute/path/key" {
		t.Errorf("IdentityFile = %q, want absolute path", h.IdentityFile)
	}

	h2 := hosts[1]
	if h2.ProxyCommand != "ssh -W %h:%p bastion" || h2.ProxyJump != "" {
		t.Errorf("ProxyCommand = %q, ProxyJump = %q; want the command kept out of ProxyJump", h2.ProxyCommand, h2.ProxyJump)
	}
	if h2.ForwardAgent {
		t.Error("ForwardAgent should be false for 'no'")
	}
}

func TestParseConfig_FieldsBeforeHost(t *testing.T) {
	// Fields before any Host block should be ignored.
	content := `
Hostname orphan.com
User nobody

Host real-host
  Hostname 10.0.0.1
  User deploy
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

	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].Name != "real-host" {
		t.Errorf("Name = %q, want %q", hosts[0].Name, "real-host")
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
		{"Host=myserver", "Host", "myserver"},
		{"User=deploy", "User", "deploy"},
		{"OnlyKeyword", "OnlyKeyword", ""},
		{"Host\tmyhost", "Host", "myhost"},
		{"Host = myserver", "Host", "myserver"},
		{"Host \t= \tmyserver", "Host", "myserver"},
		{"ProxyCommand ssh -o StrictHostKeyChecking=no -W %h:%p bastion", "ProxyCommand", "ssh -o StrictHostKeyChecking=no -W %h:%p bastion"},
		{"ProxyCommand=ssh -o StrictHostKeyChecking=no -W %h:%p bastion", "ProxyCommand", "ssh -o StrictHostKeyChecking=no -W %h:%p bastion"},
		{"OnlyKeyword ", "OnlyKeyword", ""},
		{"=orphan", "", "orphan"},
	}

	for _, tt := range tests {
		key, value := parseKeyValue(tt.line)
		if key != tt.wantKey || value != tt.wantValue {
			t.Errorf("parseKeyValue(%q) = (%q, %q), want (%q, %q)",
				tt.line, key, value, tt.wantKey, tt.wantValue)
		}
	}
}

func TestExpandPath(t *testing.T) {
	expanded := expandPath("~/test/key")
	if expanded == "~/test/key" {
		t.Error("expandPath should expand ~ prefix")
	}

	abs := expandPath("/absolute/path/key")
	if abs != "/absolute/path/key" {
		t.Errorf("expandPath(%q) = %q, should be unchanged", "/absolute/path/key", abs)
	}

	rel := expandPath("relative/path")
	if rel != "relative/path" {
		t.Errorf("expandPath(%q) = %q, should be unchanged", "relative/path", rel)
	}
}

func TestParseConfig_ProxyCommandWithEquals(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config")
	content := "=orphan\nHost proxy-host\n  ProxyCommand ssh -o StrictHostKeyChecking=no -W %h:%p bastion\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	hosts, err := ParseConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 1 || hosts[0].ProxyCommand != "ssh -o StrictHostKeyChecking=no -W %h:%p bastion" {
		t.Fatalf("hosts = %+v", hosts)
	}
}

func TestParseConfig_TildePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, "ssh_config"), []byte("Host a\n  Hostname 192.0.2.1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	hosts, err := ParseConfig("~/ssh_config")
	if err != nil {
		t.Fatalf("ParseConfig(~/ssh_config): %v", err)
	}
	if len(hosts) != 1 || hosts[0].Name != "a" {
		t.Errorf("hosts = %+v", hosts)
	}
}

func TestParseConfig_FirstProxyDirectiveWins(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config")
	content := "Host cmd-first\n  ProxyCommand ssh -W %h:%p b1\n  ProxyJump b2\n" +
		"Host jump-first\n  ProxyJump b3\n  ProxyCommand ssh -W %h:%p b4\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	hosts, err := ParseConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 2 {
		t.Fatalf("hosts = %+v", hosts)
	}
	if h := hosts[0]; h.ProxyCommand != "ssh -W %h:%p b1" || h.ProxyJump != "" {
		t.Errorf("cmd-first: ProxyCommand = %q, ProxyJump = %q", h.ProxyCommand, h.ProxyJump)
	}
	if h := hosts[1]; h.ProxyJump != "b3" || h.ProxyCommand != "" {
		t.Errorf("jump-first: ProxyJump = %q, ProxyCommand = %q", h.ProxyJump, h.ProxyCommand)
	}
}
