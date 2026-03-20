package cli

import (
	"fmt"
	"os"

	"github.com/somaz94/bash-pilot/internal/env"
	"github.com/somaz94/bash-pilot/internal/report"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Shell environment diagnostics",
	Long:  "Analyze shell environment — health check, PATH analysis.",
}

var envCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Shell environment health scan",
	Long:  "Check shell, tools, SSH agent, git config, and home directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		result := env.Check()

		f := report.NewFormatter(os.Stdout, output, noColor)

		if output == "json" {
			return f.JSON(result)
		}

		groups, keys := env.GroupFindingsByCategory(result.Findings)

		for _, category := range keys {
			f.Header(fmt.Sprintf("ENV CHECK: %s", category))
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

		ok, warn, errCount := env.SummarizeFindings(result.Findings)
		summary := fmt.Sprintf("Summary: %d ok, %d warnings, %d errors", ok, warn, errCount)
		if errCount > 0 {
			f.Println(f.Fail(summary))
		} else if warn > 0 {
			f.Println(f.Warn(summary))
		} else {
			f.Println(f.OK(summary))
		}

		return nil
	},
}

var envPathCmd = &cobra.Command{
	Use:   "path",
	Short: "PATH analysis (duplicates, missing directories)",
	Long:  "Analyze the PATH environment variable for duplicates, missing directories, and ordering.",
	RunE: func(cmd *cobra.Command, args []string) error {
		result := env.AnalyzePath()

		f := report.NewFormatter(os.Stdout, output, noColor)

		if output == "json" {
			return f.JSON(result)
		}

		f.Header(fmt.Sprintf("PATH ENTRIES (%d total)", result.Total))
		for _, entry := range result.Entries {
			status := f.OK(fmt.Sprintf("[%2d] %s", entry.Index, entry.Path))
			if !entry.Exists {
				status = f.Fail(fmt.Sprintf("[%2d] %s (not found)", entry.Index, entry.Path))
			}
			f.Println(status)
		}
		f.Footer()
		fmt.Println()

		if len(result.Duplicates) > 0 {
			f.Header("DUPLICATES")
			for _, dup := range result.Duplicates {
				f.Println(f.Warn(dup))
			}
			f.Footer()
			fmt.Println()
		}

		if len(result.Missing) > 0 {
			f.Header("MISSING DIRECTORIES")
			for _, m := range result.Missing {
				f.Println(f.Fail(m))
			}
			f.Footer()
			fmt.Println()
		}

		summary := fmt.Sprintf("%d entries, %d duplicates, %d missing", result.Total, len(result.Duplicates), len(result.Missing))
		if len(result.Missing) > 0 || len(result.Duplicates) > 0 {
			f.Println(f.Warn(summary))
		} else {
			f.Println(f.OK(summary))
		}

		return nil
	},
}

func init() {
	envCmd.AddCommand(envCheckCmd)
	envCmd.AddCommand(envPathCmd)
	rootCmd.AddCommand(envCmd)
}
