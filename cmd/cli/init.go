package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

		defaultCfg := config.Default()
		groups := ssh.GroupHosts(hosts, defaultCfg.SSH)

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
			var names []string
			for _, h := range g.Hosts {
				names = append(names, h.Name)
			}
			cfg.SSH.Groups[g.Name] = config.SSHGroup{
				Pattern: toWildcardPatterns(names),
			}
		}

		data, err := yaml.Marshal(&cfg)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		cfgDir := filepath.Join(home, ".config", "bash-pilot")
		cfgPath := filepath.Join(cfgDir, "config.yaml")

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

		if err := os.MkdirAll(cfgDir, config.PermConfigDir); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		if err := os.WriteFile(cfgPath, data, config.PermConfigFile); err != nil {
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

// toWildcardPatterns folds names sharing an extractPrefix into one "<prefix>*"
// pattern ("k8s-control-01", "k8s-compute-01" → "k8s-*") and keeps the rest
// as-is: "nas" and "nas-svn" stay separate because their prefixes differ.
func toWildcardPatterns(names []string) []string {
	if len(names) <= 1 {
		return names
	}

	sort.Strings(names)

	prefixGroups := make(map[string][]string)
	for _, name := range names {
		prefix := extractPrefix(name)
		prefixGroups[prefix] = append(prefixGroups[prefix], name)
	}

	var patterns []string
	seen := make(map[string]bool)

	for prefix, group := range prefixGroups {
		if len(group) >= 2 && prefix != "" {
			pattern := prefix + "*"
			if !seen[pattern] {
				patterns = append(patterns, pattern)
				seen[pattern] = true
			}
		} else {
			for _, name := range group {
				if !seen[name] {
					patterns = append(patterns, name)
					seen[name] = true
				}
			}
		}
	}

	sort.Strings(patterns)
	return patterns
}

// extractPrefix returns name up to and including its first '-' (preferred over an
// earlier '.'), else its first '.', else name minus trailing digits ("server1" → "server").
func extractPrefix(name string) string {
	if idx := strings.Index(name, "-"); idx > 0 {
		return name[:idx+1]
	}
	if idx := strings.Index(name, "."); idx > 0 {
		return name[:idx+1]
	}
	i := len(name)
	for i > 0 && name[i-1] >= '0' && name[i-1] <= '9' {
		i--
	}
	return name[:i]
}

func init() {
	initCmd.Flags().Bool("force", false, "overwrite existing config file")
	rootCmd.AddCommand(initCmd)
}
