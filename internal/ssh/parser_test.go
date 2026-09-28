package ssh

import (
	"os"
	"path/filepath"
	"slices"
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

// Expected words match what `ssh -G` reports for the same argument.
func TestSplitArgs(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"192.0.2.10", []string{"192.0.2.10"}},
		{"a  b\tc", []string{"a", "b", "c"}},
		{"192.0.2.10 # primary", []string{"192.0.2.10"}},
		{"web#1", []string{"web#1"}},
		{"# only a comment", nil},
		{`"/tmp/my key"`, []string{"/tmp/my key"}},
		{`"~/.ssh/key #1"`, []string{"~/.ssh/key #1"}},
		{`'a # b' # c`, []string{"a # b"}},
		{`"a"b'c'`, []string{"abc"}},
		{`'it''s'`, []string{"its"}},
		{`o\'brien # c`, []string{"o'brien"}},
		{`/tmp/x\ #2`, []string{"/tmp/x #2"}},
		{`"a\" # b" # c`, []string{`a" # b`}},
		{`/tmp/back\\slash`, []string{`/tmp/back\slash`}},
		{`/tmp/lone\x`, []string{`/tmp/lone\x`}},
		{`"in quotes\ stays"`, []string{`in quotes\ stays`}},
	}
	for _, tt := range tests {
		if got := splitArgs(tt.in); !slices.Equal(got, tt.want) {
			t.Errorf("splitArgs(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseConfig_TrailingComments(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config")
	content := "Host * # defaults\n  User nobody\n" +
		"Host # retired\n  User nobody\n" +
		"Host prod # production box\n  Hostname 192.0.2.10 # primary\n  User deploy #ops\n" +
		"Host via-bastion\n  ProxyCommand ssh -W %h:%p bastion # shell comment\n" +
		"Host=eq # c\n  ForwardAgent yes # c\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	hosts, err := ParseConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 3 {
		t.Fatalf("hosts = %+v, want prod, via-bastion and eq (Host * and a comment-only Host are not hosts)", hosts)
	}
	if h := hosts[0]; h.Name != "prod" || h.Hostname != "192.0.2.10" || h.User != "deploy" {
		t.Errorf("prod = %+v", h)
	}
	if h := hosts[1]; h.ProxyCommand != "ssh -W %h:%p bastion # shell comment" {
		t.Errorf("ProxyCommand = %q, want it verbatim", h.ProxyCommand)
	}
	if h := hosts[2]; h.Name != "eq" || !h.ForwardAgent {
		t.Errorf("eq = %+v, want ForwardAgent yes despite the trailing comment", h)
	}
}

func TestParseConfig_MatchEndsHostBlock(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config")
	content := "Host a\n  Hostname 192.0.2.1\n" +
		"Match host b\n  User nobody\n  IdentityFile ~/.ssh/other\n" +
		"Host c\n  User carol\n"
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
	if h := hosts[0]; h.Hostname != "192.0.2.1" || h.User != "" || h.IdentityFile != "" {
		t.Errorf("a picked up Match settings: %+v", h)
	}
	if hosts[1].User != "carol" {
		t.Errorf("c.User = %q, want carol", hosts[1].User)
	}
}

// Expected values match what `ssh -G` reports for the same block.
func TestParseConfig_FirstValueWins(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config")
	content := "Host dup\n" +
		"  HostName 192.0.2.1\n  HostName 192.0.2.2\n" +
		"  User first\n  User second\n" +
		"  Port 2201\n  Port 2202\n" +
		"  IdentityFile /keys/a\n  IdentityFile /keys/b\n" +
		"  ForwardAgent no\n  ForwardAgent yes\n" +
		"  ProxyJump j1\n  ProxyJump j2\n  ProxyCommand nc %h %p\n" +
		"Host next\n  User again\n"
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
	want := Host{Name: "dup", Hostname: "192.0.2.1", User: "first", Port: "2201", IdentityFile: "/keys/a", ProxyJump: "j1"}
	if hosts[0] != want {
		t.Errorf("dup = %+v\nwant  %+v", hosts[0], want)
	}
	if hosts[1].User != "again" {
		t.Errorf("next.User = %q; a new Host must start with no parameters set", hosts[1].User)
	}
}

func TestParseConfig_QuotedValues(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfgPath := filepath.Join(t.TempDir(), "config")
	content := "Host 'multi a' b # c\n" +
		"  IdentityFile \"~/.ssh/my key\"\n" +
		"  User o\\'brien\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	hosts, err := ParseConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 1 {
		t.Fatalf("hosts = %+v", hosts)
	}
	h := hosts[0]
	if want := filepath.Join(home, ".ssh", "my key"); h.IdentityFile != want {
		t.Errorf("IdentityFile = %q, want %q (unquoted, then ~ expanded)", h.IdentityFile, want)
	}
	if h.User != "o'brien" || h.Name != "multi a b" {
		t.Errorf("User = %q, Name = %q", h.User, h.Name)
	}
}

// Expected values match `ssh -G`: ProxyJump is read raw and ends at any '#'.
func TestParseConfig_ProxyJumpIsRaw(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config")
	content := "Host q\n  ProxyJump \"jump\"\n" +
		"Host h\n  ProxyJump j#x\n" +
		"Host c\n  ProxyJump jump # office\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	hosts, err := ParseConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, h := range hosts {
		got = append(got, h.ProxyJump)
	}
	if want := []string{`"jump"`, "j", "jump"}; !slices.Equal(got, want) {
		t.Errorf("ProxyJump values = %q, want %q", got, want)
	}
}

func TestParseConfig_EmptyArgumentsClaimNothing(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config")
	content := "Host \"\"\n  User ghost\n" +
		"Host a\n  User \"\" bob\n  User carol\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	hosts, err := ParseConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 1 || hosts[0].Name != "a" || hosts[0].User != "carol" {
		t.Errorf("hosts = %+v, want only a with User carol", hosts)
	}
}
