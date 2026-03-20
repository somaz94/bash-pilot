package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/somaz94/bash-pilot/internal/report"
	"github.com/somaz94/bash-pilot/internal/snapshot"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup <snapshot-file>",
	Short: "Install missing tools from a snapshot",
	Long: `Compare a snapshot against the current environment and install missing tools.

Usage:
  # Preview what would be installed
  bash-pilot setup my-env.json --dry-run

  # Install missing tools
  bash-pilot setup my-env.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("cannot read snapshot file: %w", err)
		}

		var saved snapshot.Snapshot
		if err := json.Unmarshal(data, &saved); err != nil {
			return fmt.Errorf("invalid snapshot file: %w", err)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		onlyFlag, _ := cmd.Flags().GetString("only")
		only := snapshot.ParseOnly(onlyFlag)

		f := report.NewFormatter(os.Stdout, output, noColor)

		if output == "json" {
			result := snapshot.Execute(&saved, dryRun, only)
			return f.JSON(result)
		}

		if dryRun {
			result := snapshot.Plan(&saved, only)
			f.Header("SETUP PLAN (dry-run)")
			fmt.Print(snapshot.FormatPlan(result))
			f.Footer()

			if len(result.Actions) > 0 {
				pending := 0
				for _, a := range result.Actions {
					if a.Status == "pending" {
						pending++
					}
				}
				fmt.Println()
				f.Println(f.Warn(fmt.Sprintf("%d tool(s) to install, %d skipped",
					pending, result.Summary.Skipped)))
				fmt.Println()
				f.Println("Run without --dry-run to install.")
			}
			return nil
		}

		f.Header("SETUP")
		result := snapshot.Execute(&saved, false, only)
		fmt.Print(snapshot.FormatPlan(result))
		f.Footer()

		fmt.Println()
		summary := fmt.Sprintf("%d installed, %d skipped, %d failed",
			result.Summary.Installed, result.Summary.Skipped, result.Summary.Failed)
		if result.Summary.Failed > 0 {
			f.Println(f.Fail(summary))
		} else if result.Summary.Installed > 0 {
			f.Println(f.OK(summary))
		} else {
			f.Println(f.OK("Nothing to install — environment matches snapshot."))
		}

		return nil
	},
}

func init() {
	setupCmd.Flags().Bool("dry-run", false, "Preview install plan without executing")
	setupCmd.Flags().String("only", "", "Comma-separated sections to install (tools,brew)")
	rootCmd.AddCommand(setupCmd)
}
