package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/somaz94/bash-pilot/internal/report"
	"github.com/somaz94/bash-pilot/internal/snapshot"
	"github.com/spf13/cobra"
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Capture environment snapshot",
	Long: `Capture a full snapshot of the current environment to JSON.

Usage:
  # Save snapshot to file
  bash-pilot snapshot > my-env.json

  # Preview snapshot summary
  bash-pilot snapshot --summary`,
	RunE: func(cmd *cobra.Command, args []string) error {
		snap := snapshot.Capture()

		summary, _ := cmd.Flags().GetBool("summary")
		if summary {
			f := report.NewFormatter(os.Stdout, output, noColor)
			f.Header("ENVIRONMENT SNAPSHOT")
			fmt.Print(snapshot.FormatSummary(snap))
			f.Footer()
			return nil
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(snap)
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff <snapshot-file>",
	Short: "Compare snapshot against current environment",
	Long: `Compare a saved environment snapshot against the current environment.

Usage:
  bash-pilot diff my-env.json`,
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

		result := snapshot.Diff(&saved)

		f := report.NewFormatter(os.Stdout, output, noColor)

		if output == "json" {
			return f.JSON(result)
		}

		f.Header(fmt.Sprintf("DIFF vs %s (%s)", saved.Hostname, saved.Timestamp))

		for _, sec := range result.Sections {
			hasIssues := false
			for _, e := range sec.Entries {
				if e.Status != "match" {
					hasIssues = true
					break
				}
			}

			if !hasIssues {
				f.Println(f.OK(fmt.Sprintf("[%s] all match", sec.Name)))
				continue
			}

			f.Println(fmt.Sprintf("  [%s]", sec.Name))
			for _, e := range sec.Entries {
				switch e.Status {
				case "match":
					// skip
				case "mismatch":
					f.Println(f.Warn(fmt.Sprintf("  ~ %-20s %s → %s", e.Key, e.Left, e.Right)))
				case "missing":
					f.Println(f.Fail(fmt.Sprintf("  - %-20s %s (not in current)", e.Key, e.Left)))
				case "extra":
					f.Println(f.OK(fmt.Sprintf("  + %-20s %s (new)", e.Key, e.Right)))
				}
			}
		}

		f.Footer()
		fmt.Println()

		summary := fmt.Sprintf("%d match, %d changed, %d missing, %d new",
			result.Summary.Match, result.Summary.Mismatch, result.Summary.Missing, result.Summary.Extra)
		if result.Summary.Mismatch > 0 || result.Summary.Missing > 0 {
			f.Println(f.Warn(summary))
		} else {
			f.Println(f.OK(summary))
		}

		return nil
	},
}

func init() {
	snapshotCmd.Flags().Bool("summary", false, "Show summary instead of full JSON")

	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(diffCmd)
}
