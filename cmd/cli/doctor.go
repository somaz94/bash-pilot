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

		configFile := ""
		if appCfg != nil {
			configFile = appCfg.SSH.ConfigFile
		}
		sshResult := auditSSHConfig(configFile)

		gitResult, _ := git.Doctor(resolveGitConfigPath())

		envResult := env.Check()

		if output == "json" {
			return f.JSON(DoctorResult{
				SSH: sshResult,
				Git: gitResult,
				Env: envResult,
			})
		}

		f.Header("DOCTOR: SSH")
		if len(sshResult.Findings) == 0 {
			f.Println(f.OK("No SSH issues found"))
		}
		for _, finding := range sshResult.Findings {
			f.Println(f.RenderSeverity(string(finding.Severity), finding.Message))
		}
		f.Footer()
		fmt.Println()

		f.Header("DOCTOR: GIT")
		if len(gitResult.Issues) == 0 {
			f.Println(f.OK("No Git issues found"))
		}
		for _, issue := range gitResult.Issues {
			f.Println(f.RenderSeverity(issue.Severity, issue.Message))
		}
		f.Footer()
		fmt.Println()

		groups, keys := env.GroupFindingsByCategory(envResult.Findings)
		for _, category := range keys {
			f.Header(fmt.Sprintf("DOCTOR: ENV (%s)", strings.ToUpper(category)))
			for _, finding := range groups[category] {
				f.Println(f.RenderSeverity(finding.Severity, finding.Message))
			}
			f.Footer()
			fmt.Println()
		}

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

// auditSSHConfig reports an unreadable SSH config as a finding, as git.Doctor
// does for gitconfig, instead of auditing an empty host list.
func auditSSHConfig(path string) ssh.AuditResult {
	hosts, err := ssh.ParseConfig(path)
	if err != nil {
		return ssh.AuditResult{Findings: []ssh.AuditFinding{{
			Severity: ssh.SeverityFail,
			Key:      "config",
			Message:  fmt.Sprintf("Cannot read SSH config: %s", err),
		}}}
	}
	return ssh.Audit(hosts)
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
