package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/somaz94/bash-pilot/internal/config"
	"github.com/somaz94/bash-pilot/internal/ssh"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate config from existing SSH config",
	Long:  "Analyze ~/.ssh/config and generate ~/.config/bash-pilot/config.yaml with auto-detected groups.",
	RunE: func(cmd *cobra.Command, args []string) error {
		sshConfigPath := appCfg.SSH.ConfigFile

		hosts, err := ssh.ParseConfig(sshConfigPath)
		if err != nil {
			return fmt.Errorf("failed to parse SSH config: %w", err)
		}

		if len(hosts) == 0 {
			fmt.Println("No hosts found in SSH config.")
			return nil
		}

		// Auto-detect groups from hosts.
		defaultCfg := config.Default()
		groups := ssh.GroupHosts(hosts, defaultCfg.SSH)

		// Build config from detected groups.
		cfg := config.Config{
			SSH: config.SSHConfig{
				Groups: make(map[string]config.SSHGroup),
				Ping: config.PingConfig{
					Timeout:  defaultCfg.SSH.Ping.Timeout,
					Parallel: defaultCfg.SSH.Ping.Parallel,
				},
			},
		}

		for _, g := range groups {
			if len(g.Hosts) == 0 {
				continue
			}
			var patterns []string
			for _, h := range g.Hosts {
				patterns = append(patterns, h.Name)
			}
			cfg.SSH.Groups[g.Name] = config.SSHGroup{
				Pattern: patterns,
			}
		}

		data, err := yaml.Marshal(&cfg)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		// Determine output path.
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		cfgDir := filepath.Join(home, ".config", "bash-pilot")
		cfgPath := filepath.Join(cfgDir, "config.yaml")

		// Check if config already exists.
		if _, err := os.Stat(cfgPath); err == nil {
			fmt.Printf("Config already exists: %s\n", cfgPath)
			fmt.Println("Use --force to overwrite.")
			force, _ := cmd.Flags().GetBool("force")
			if !force {
				fmt.Println("\nGenerated config (preview):")
				fmt.Println("---")
				fmt.Print(string(data))
				return nil
			}
		}

		// Create directory and write config.
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		if err := os.WriteFile(cfgPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		fmt.Printf("Config generated: %s\n", cfgPath)
		fmt.Printf("Detected %d groups from %d hosts:\n", len(cfg.SSH.Groups), len(hosts))
		for name, group := range cfg.SSH.Groups {
			fmt.Printf("  %-10s %d hosts\n", name, len(group.Pattern))
		}
		fmt.Println("\nEdit the config to customize group patterns and labels.")

		return nil
	},
}

func init() {
	initCmd.Flags().Bool("force", false, "overwrite existing config file")
	rootCmd.AddCommand(initCmd)
}
