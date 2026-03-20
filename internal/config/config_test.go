package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.SSH.Ping.Timeout != 5*time.Second {
		t.Errorf("default timeout = %v, want 5s", cfg.SSH.Ping.Timeout)
	}
	if cfg.SSH.Ping.Parallel != 10 {
		t.Errorf("default parallel = %d, want 10", cfg.SSH.Ping.Parallel)
	}
	if len(cfg.SSH.Groups) == 0 {
		t.Error("default groups should not be empty")
	}
	if _, ok := cfg.SSH.Groups["git"]; !ok {
		t.Error("default groups should contain 'git'")
	}
}

func TestLoad(t *testing.T) {
	content := `
ssh:
  config_file: /tmp/test-ssh-config
  groups:
    mygroup:
      pattern: ["test-*"]
      label: "Test Group"
  ping:
    timeout: 3s
    parallel: 5
git:
  profiles:
    work:
      directory: ~/work
      email: work@example.com
      key: ~/.ssh/id_rsa_work
`

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.SSH.ConfigFile != "/tmp/test-ssh-config" {
		t.Errorf("SSH.ConfigFile = %q, want %q", cfg.SSH.ConfigFile, "/tmp/test-ssh-config")
	}
	if cfg.SSH.Ping.Timeout != 3*time.Second {
		t.Errorf("Ping.Timeout = %v, want 3s", cfg.SSH.Ping.Timeout)
	}
	if cfg.SSH.Ping.Parallel != 5 {
		t.Errorf("Ping.Parallel = %d, want 5", cfg.SSH.Ping.Parallel)
	}
	if g, ok := cfg.SSH.Groups["mygroup"]; !ok {
		t.Error("expected 'mygroup' in groups")
	} else {
		if g.Label != "Test Group" {
			t.Errorf("mygroup.Label = %q, want %q", g.Label, "Test Group")
		}
		if len(g.Pattern) != 1 || g.Pattern[0] != "test-*" {
			t.Errorf("mygroup.Pattern = %v, want [test-*]", g.Pattern)
		}
	}

	if p, ok := cfg.Git.Profiles["work"]; !ok {
		t.Error("expected 'work' in git profiles")
	} else {
		if p.Email != "work@example.com" {
			t.Errorf("work.Email = %q, want %q", p.Email, "work@example.com")
		}
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Config with zero values should get defaults applied.
	content := `
ssh:
  groups:
    test:
      pattern: ["t-*"]
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.SSH.Ping.Timeout != 5*time.Second {
		t.Errorf("expected default timeout 5s, got %v", cfg.SSH.Ping.Timeout)
	}
	if cfg.SSH.Ping.Parallel != 10 {
		t.Errorf("expected default parallel 10, got %d", cfg.SSH.Ping.Parallel)
	}
}

func TestLoad_NotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("invalid: [yaml: {{{"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(cfgPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoad_DefaultPath(t *testing.T) {
	// Load with empty path should try default location (~/.config/bash-pilot/config.yaml).
	// This will fail since the file doesn't exist, but it shouldn't panic.
	_, err := Load("")
	if err == nil {
		t.Log("default config exists (unexpected but not an error)")
	}
}
