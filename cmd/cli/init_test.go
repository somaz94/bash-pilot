package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// captureStdout runs fn with os.Stdout redirected, since init prints via fmt.Printf.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()
	// Drain concurrently so output larger than the pipe buffer cannot block fn.
	done := make(chan string)
	go func() {
		defer func() { _ = r.Close() }()
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()
	fn()
	_ = w.Close()
	return <-done
}

func TestInitCmd_Run(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sshCfg := filepath.Join(t.TempDir(), "ssh_config")
	hosts := "Host k8s-control-01\n  Hostname 192.0.2.10\nHost k8s-compute-01\n  Hostname 192.0.2.11\nHost k8s-compute-02\n  Hostname 192.0.2.12\n"
	if err := os.WriteFile(sshCfg, []byte(hosts), 0o600); err != nil {
		t.Fatal(err)
	}
	bpCfg := filepath.Join(t.TempDir(), "bp.yaml")
	if err := os.WriteFile(bpCfg, []byte("ssh:\n  config_file: "+sshCfg+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = initCmd.Flags().Set("force", "false")
		rootCmd.SetArgs(nil)
		cfgFile, appCfg = "", nil
		initCmd.SilenceUsage = false
	})

	run := func(args ...string) string {
		return captureStdout(t, func() {
			rootCmd.SetArgs(append([]string{"init", "--config", bpCfg}, args...))
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("init %v: %v", args, err)
			}
		})
	}

	first := run()
	if !strings.Contains(first, "k8s        3 hosts") {
		t.Errorf("first run should count 3 hosts for one k8s-* pattern:\n%s", first)
	}
	if second := run(); !strings.Contains(second, "Use --force to overwrite.") {
		t.Errorf("second run should suggest --force:\n%s", second)
	}
	if forced := run("--force"); strings.Contains(forced, "Use --force") {
		t.Errorf("--force run should not suggest --force:\n%s", forced)
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
