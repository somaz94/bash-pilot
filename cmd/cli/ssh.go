package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/somaz94/bash-pilot/internal/report"
	"github.com/somaz94/bash-pilot/internal/ssh"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH host management",
	Long:  "Manage SSH hosts — list, test connectivity, and audit security.",
}

var sshListCmd = &cobra.Command{
	Use:   "list",
	Short: "List SSH hosts with grouping",
	RunE: func(cmd *cobra.Command, args []string) error {
		hosts, err := ssh.ParseConfig(appCfg.SSH.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to parse SSH config: %w", err)
		}

		f := report.NewFormatter(os.Stdout, output, noColor)
		groups := ssh.GroupHosts(hosts, appCfg.SSH)

		if output == "json" {
			return f.JSON(groups)
		}

		for _, g := range groups {
			label := strings.ToUpper(g.Name)
			if g.Label != "" {
				label += " (" + g.Label + ")"
			}
			f.Header(label)

			for _, h := range g.Hosts {
				hostname := h.Hostname
				if hostname == "" {
					hostname = "-"
				}
				user := h.User
				if user == "" {
					user = "-"
				}
				key := h.KeyName()
				if key == "" {
					key = "-"
				}

				line := fmt.Sprintf("  %-25s %-20s %-15s %s", h.Name, hostname, user, key)
				f.Row(line)
			}
			f.Footer()
			fmt.Println()
		}

		return nil
	},
}

var sshPingCmd = &cobra.Command{
	Use:   "ping [pattern]",
	Short: "Test SSH host connectivity (parallel)",
	Long:  "Test TCP connectivity to SSH hosts. Optionally filter by glob pattern.",
	RunE: func(cmd *cobra.Command, args []string) error {
		hosts, err := ssh.ParseConfig(appCfg.SSH.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to parse SSH config: %w", err)
		}

		// Filter by pattern if provided.
		if len(args) > 0 {
			pattern := args[0]
			var filtered []ssh.Host
			for _, h := range hosts {
				matched, _ := filepath.Match(pattern, h.Name)
				if matched {
					filtered = append(filtered, h)
				}
			}
			if len(filtered) == 0 {
				return fmt.Errorf("no hosts matching pattern: %s", pattern)
			}
			hosts = filtered
		}

		f := report.NewFormatter(os.Stdout, output, noColor)
		results := ssh.PingHosts(hosts, appCfg.SSH.Ping.Timeout, appCfg.SSH.Ping.Parallel)

		if output == "json" {
			return f.JSON(results)
		}

		for _, r := range results {
			latency := fmt.Sprintf("%.2fs", r.Latency.Seconds())
			if r.OK {
				f.Println(f.OK(fmt.Sprintf("%-25s %s", r.Host.Name, latency)))
			} else {
				f.Println(f.Fail(fmt.Sprintf("%-25s %s", r.Host.Name, r.Error)))
			}
		}

		return nil
	},
}

var sshAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit SSH config for security issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		hosts, err := ssh.ParseConfig(appCfg.SSH.ConfigFile)
		if err != nil {
			return fmt.Errorf("failed to parse SSH config: %w", err)
		}

		f := report.NewFormatter(os.Stdout, output, noColor)
		result := ssh.Audit(hosts)

		if output == "json" {
			return f.JSON(result)
		}

		for _, finding := range result.Findings {
			switch finding.Severity {
			case ssh.SeverityOK:
				f.Println(f.OK(fmt.Sprintf("%s: %s", finding.Key, finding.Message)))
			case ssh.SeverityWarn:
				f.Println(f.Warn(fmt.Sprintf("%s: %s", finding.Key, finding.Message)))
			case ssh.SeverityFail:
				f.Println(f.Fail(fmt.Sprintf("%s: %s", finding.Key, finding.Message)))
			}
		}

		return nil
	},
}

func init() {
	sshCmd.AddCommand(sshListCmd)
	sshCmd.AddCommand(sshPingCmd)
	sshCmd.AddCommand(sshAuditCmd)
	rootCmd.AddCommand(sshCmd)
}
