package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/somaz94/bash-pilot/internal/env"
	"github.com/somaz94/bash-pilot/internal/git"
	"github.com/somaz94/bash-pilot/internal/report"
	"github.com/somaz94/bash-pilot/internal/ssh"
	"github.com/spf13/cobra"
)

// DoctorResult holds combined diagnostics from all modules.
type DoctorResult struct {
	SSH ssh.AuditResult   `json:"ssh"`
	Git *git.DoctorResult `json:"git"`
	Env *env.CheckResult  `json:"env"`
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Full system diagnostics (SSH + Git + Env)",
	Long:  "Run all diagnostic checks across SSH, Git, and Env modules in a single command.",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := report.NewFormatter(os.Stdout, output, noColor)

		// --- SSH Audit ---
		var sshResult ssh.AuditResult
		configFile := ""
		if appCfg != nil {
			configFile = appCfg.SSH.ConfigFile
		}
		hosts, _ := ssh.ParseConfig(configFile)
		sshResult = ssh.Audit(hosts)

		// --- Git Doctor ---
		gitResult, _ := git.Doctor("~/.gitconfig")

		// --- Env Check ---
		envResult := env.Check()

		if output == "json" {
			return f.JSON(DoctorResult{
				SSH: sshResult,
				Git: gitResult,
				Env: envResult,
			})
		}

		// SSH section
		f.Header("DOCTOR: SSH")
		if len(sshResult.Findings) == 0 {
			f.Println(f.OK("No SSH issues found"))
		}
		for _, finding := range sshResult.Findings {
			switch finding.Severity {
			case ssh.SeverityOK:
				f.Println(f.OK(finding.Message))
			case ssh.SeverityWarn:
				f.Println(f.Warn(finding.Message))
			case ssh.SeverityFail:
				f.Println(f.Fail(finding.Message))
			}
		}
		f.Footer()
		fmt.Println()

		// Git section
		f.Header("DOCTOR: GIT")
		if len(gitResult.Issues) == 0 {
			f.Println(f.OK("No Git issues found"))
		}
		for _, issue := range gitResult.Issues {
			switch issue.Severity {
			case "ok":
				f.Println(f.OK(issue.Message))
			case "warn":
				f.Println(f.Warn(issue.Message))
			case "error":
				f.Println(f.Fail(issue.Message))
			}
		}
		f.Footer()
		fmt.Println()

		// Env section
		groups, keys := env.GroupFindingsByCategory(envResult.Findings)
		for _, category := range keys {
			f.Header(fmt.Sprintf("DOCTOR: ENV (%s)", strings.ToUpper(category)))
			for _, finding := range groups[category] {
				switch finding.Severity {
				case "ok":
					f.Println(f.OK(finding.Message))
				case "warn":
					f.Println(f.Warn(finding.Message))
				case "error":
					f.Println(f.Fail(finding.Message))
				}
			}
			f.Footer()
			fmt.Println()
		}

		// Summary
		sshIssues := 0
		for _, finding := range sshResult.Findings {
			if finding.Severity != ssh.SeverityOK {
				sshIssues++
			}
		}
		gitIssues := 0
		for _, issue := range gitResult.Issues {
			if issue.Severity != "ok" {
				gitIssues++
			}
		}
		_, envWarn, envErr := env.SummarizeFindings(envResult.Findings)
		envIssues := envWarn + envErr

		total := sshIssues + gitIssues + envIssues
		summary := fmt.Sprintf("Total: %d issue(s) — SSH: %d, Git: %d, Env: %d", total, sshIssues, gitIssues, envIssues)
		if total > 0 {
			f.Println(f.Warn(summary))
		} else {
			f.Println(f.OK(summary))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
