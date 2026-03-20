package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration.
type Config struct {
	SSH SSHConfig `yaml:"ssh"`
	Git GitConfig `yaml:"git"`
}

// SSHConfig holds SSH-related settings.
type SSHConfig struct {
	ConfigFile string              `yaml:"config_file,omitempty"`
	Groups     map[string]SSHGroup `yaml:"groups"`
	Ping       PingConfig          `yaml:"ping"`
}

// SSHGroup defines a host grouping rule.
type SSHGroup struct {
	Pattern []string `yaml:"pattern"`
	Label   string   `yaml:"label,omitempty"`
}

// PingConfig controls connectivity check behavior.
type PingConfig struct {
	Timeout  time.Duration `yaml:"timeout"`
	Parallel int           `yaml:"parallel"`
}

// GitConfig holds Git-related settings.
type GitConfig struct {
	Profiles map[string]GitProfile `yaml:"profiles"`
}

// GitProfile represents a git identity configuration.
type GitProfile struct {
	Directory string `yaml:"directory"`
	Email     string `yaml:"email"`
	Key       string `yaml:"key"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		SSH: SSHConfig{
			Groups: map[string]SSHGroup{
				"git": {
					Pattern: []string{"github.com*", "gitlab*", "git-codecommit*"},
				},
				"k8s": {
					Pattern: []string{"k8s-*"},
				},
			},
			Ping: PingConfig{
				Timeout:  5 * time.Second,
				Parallel: 10,
			},
		},
	}
}

// Load reads and parses a YAML config file.
// If path is empty, it tries the default location.
func Load(path string) (*Config, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".config", "bash-pilot", "config.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Apply defaults for zero values.
	if cfg.SSH.Ping.Timeout == 0 {
		cfg.SSH.Ping.Timeout = 5 * time.Second
	}
	if cfg.SSH.Ping.Parallel == 0 {
		cfg.SSH.Ping.Parallel = 10
	}

	return cfg, nil
}
