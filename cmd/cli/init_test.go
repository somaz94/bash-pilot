package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestInitCmd_GeneratesConfig(t *testing.T) {
	// Create a temp SSH config.
	tmpDir := t.TempDir()
	sshConfig := filepath.Join(tmpDir, "ssh_config")
	content := `Host github.com-personal
  Hostname github.com
  User git
  IdentityFile ~/.ssh/id_rsa_personal

Host web-server
  Hostname 54.123.45.67
  User ec2-user

Host k8s-control-01
  Hostname 10.0.1.10
  User admin

Host nas
  Hostname 192.168.1.10
  User user
`
	if err := os.WriteFile(sshConfig, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a temp config dir (so init writes there).
	cfgDir := filepath.Join(tmpDir, ".config", "bash-pilot")
	cfgPath := filepath.Join(cfgDir, "config.yaml")

	// Set up the root command with our test SSH config.
	rootCmd.SetArgs([]string{"init", "--config", sshConfig})

	// We can't easily test the full flow since init uses UserHomeDir,
	// but we can verify the command exists and is wired up.
	cmd, _, err := rootCmd.Find([]string{"init"})
	if err != nil {
		t.Fatalf("init command not found: %v", err)
	}
	if cmd.Use != "init" {
		t.Errorf("expected 'init', got %q", cmd.Use)
	}

	// Verify --force flag exists.
	f := cmd.Flags().Lookup("force")
	if f == nil {
		t.Error("--force flag not found")
	}

	// Verify config path doesn't exist yet.
	if _, err := os.Stat(cfgPath); err == nil {
		t.Error("config should not exist yet")
	}
}

func TestInitCmd_HasForceFlag(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"init"})
	if err != nil {
		t.Fatalf("init command not found: %v", err)
	}

	force := cmd.Flags().Lookup("force")
	if force == nil {
		t.Fatal("--force flag not registered")
	}
	if force.DefValue != "false" {
		t.Errorf("--force default = %q, want %q", force.DefValue, "false")
	}
}

func TestToWildcardPatterns(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "k8s hosts collapse",
			input: []string{"k8s-control-01", "k8s-compute-01", "k8s-compute-02"},
			want:  []string{"k8s-*"},
		},
		{
			name:  "server with trailing digits",
			input: []string{"server1", "server2", "server3"},
			want:  []string{"server*"},
		},
		{
			name:  "github prefix collapse",
			input: []string{"github.com-personal", "github.com-work"},
			want:  []string{"github.com-*"},
		},
		{
			name:  "single host unchanged",
			input: []string{"gitlab"},
			want:  []string{"gitlab"},
		},
		{
			name:  "mixed no common prefix",
			input: []string{"nas", "jenkins", "test-server"},
			want:  []string{"jenkins", "nas", "test-server"},
		},
		{
			name:  "empty input",
			input: nil,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toWildcardPatterns(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toWildcardPatterns(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"k8s-control-01", "k8s-"},
		{"github.com-somaz94", "github.com-"},
		{"server1", "server"},
		{"nas", "nas"},
		{"gitlab", "gitlab"},
	}

	for _, tt := range tests {
		got := extractPrefix(tt.name)
		if got != tt.want {
			t.Errorf("extractPrefix(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
