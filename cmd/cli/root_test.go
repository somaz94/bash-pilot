package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
