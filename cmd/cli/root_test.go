package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func TestLoadConfig(t *testing.T) {
	const valid = "ssh:\n  ping:\n    timeout: 3s\n"
	const broken = "ssh: [unclosed"
	tests := []struct {
		name        string
		defaultFile string // written to $HOME/.config/bash-pilot/config.yaml when non-empty
		explicit    string // written to a temp file passed as --config when non-empty
		missingFlag bool   // pass a --config path that does not exist
		wantErr     bool
		wantTimeout time.Duration
		wantWarning bool
	}{
		{name: "no default file", wantTimeout: 5 * time.Second},
		{name: "valid default file", defaultFile: valid, wantTimeout: 3 * time.Second},
		{name: "broken default file warns", defaultFile: broken, wantTimeout: 5 * time.Second, wantWarning: true},
		{name: "valid explicit file", explicit: valid, wantTimeout: 3 * time.Second},
		{name: "missing explicit file", missingFlag: true, wantErr: true},
		{name: "broken explicit file", explicit: broken, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			if tt.defaultFile != "" {
				dir := filepath.Join(home, ".config", "bash-pilot")
				if err := os.MkdirAll(dir, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(tt.defaultFile), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			path := ""
			switch {
			case tt.explicit != "":
				path = filepath.Join(t.TempDir(), "config.yaml")
				if err := os.WriteFile(path, []byte(tt.explicit), 0o600); err != nil {
					t.Fatal(err)
				}
			case tt.missingFlag:
				path = filepath.Join(t.TempDir(), "typo.yaml")
			}

			var stderr bytes.Buffer
			cfg, err := loadConfig(path, &stderr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if cfg.SSH.Ping.Timeout != tt.wantTimeout {
				t.Errorf("timeout = %v, want %v", cfg.SSH.Ping.Timeout, tt.wantTimeout)
			}
			if got := strings.Contains(stderr.String(), "warning:"); got != tt.wantWarning {
				t.Errorf("stderr = %q, wantWarning %v", stderr.String(), tt.wantWarning)
			}
		})
	}
}

func TestExecute_RuntimeErrorPrintsNoUsage(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetArgs([]string{"version", "--config", filepath.Join(t.TempDir(), "typo.yaml")})
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		cfgFile, appCfg = "", nil
		versionCmd.SilenceUsage = false
	})

	if err := Execute(); err == nil {
		t.Fatal("expected an error for a missing --config file")
	}
	if s := out.String(); strings.Contains(s, "Usage:") || strings.Count(s, "typo.yaml") != 1 {
		t.Errorf("want the error exactly once and no usage:\n%s", s)
	}
}

// Cobra ignores stray arguments on a command without Args, and validates Args only
// on runnable commands, so every subcommand needs both.
func TestCommandsRejectStrayArgs(t *testing.T) {
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			if sub.Name() == "help" || sub.Name() == "completion" {
				continue
			}
			if !sub.Runnable() || sub.Args == nil {
				t.Errorf("%q: runnable = %v, Args set = %v; want both", sub.CommandPath(), sub.Runnable(), sub.Args != nil)
			} else if !strings.ContainsAny(sub.Use, "<[") {
				if err := sub.Args(sub, []string{"stray"}); err == nil {
					t.Errorf("%q accepted a stray argument", sub.CommandPath())
				}
			}
			walk(sub)
		}
	}
	walk(rootCmd)
}

func TestGroupCommands_TypoFailsBareShowsHelp(t *testing.T) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		cfgFile, appCfg = "", nil
		sshCmd.SilenceUsage = false
	})

	rootCmd.SetArgs([]string{"ssh", "lst"})
	err := Execute()
	if err == nil || !strings.Contains(err.Error(), `unknown command "lst"`) || !strings.Contains(err.Error(), "list") {
		t.Fatalf("ssh lst: err = %v, want unknown command suggesting list", err)
	}

	out.Reset()
	rootCmd.SetArgs([]string{"ssh"})
	if err := Execute(); err != nil || !strings.Contains(out.String(), "Available Commands:") {
		t.Fatalf("bare ssh: err = %v\n%s", err, out.String())
	}
}
