package cli

import (
	"fmt"
	"os"

	"github.com/somaz94/bash-pilot/internal/git"
	"github.com/somaz94/bash-pilot/internal/report"
	"github.com/spf13/cobra"
)

var gitConfigFile string

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git multi-profile management",
	Long:  "Manage git identities — profiles, diagnostics, and cleanup.",
}

var gitProfilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List git identity profiles",
	Long:  "Show all git profiles from includeIf directives, with active profile highlighted.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveGitConfigPath()

		profiles, err := git.GetProfiles(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to read git profiles: %w", err)
		}

		f := report.NewFormatter(os.Stdout, output, noColor)

		if output == "json" {
			return f.JSON(profiles)
		}

		if len(profiles) == 0 {
			f.Println("No git profiles found.")
			f.Println("  Tip: Use includeIf directives in ~/.gitconfig for multi-profile setup.")
			return nil
		}

		f.Header("GIT PROFILES")
		for _, p := range profiles {
			status := "  "
			if p.Active {
				status = f.Color(report.Green, "→ ")
			}

			dir := p.Directory
			if dir == "" {
				dir = "(global)"
			}

			line := fmt.Sprintf("%s%-20s %-30s %s", status, p.Name, p.Email, dir)
			if p.SignKey != "" {
				line += fmt.Sprintf("  key:%s", p.SignKey)
			}
			f.Row(line)
		}
		f.Footer()

		return nil
	},
}

var gitDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose gitconfig issues",
	Long:  "Check for duplicate safe.directory entries, missing includeIf targets, and other common problems.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveGitConfigPath()

		result, err := git.Doctor(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to run git doctor: %w", err)
		}

		f := report.NewFormatter(os.Stdout, output, noColor)

		if output == "json" {
			return f.JSON(result)
		}

		f.Header("GIT DOCTOR")
		for _, issue := range result.Issues {
			loc := ""
			if issue.File != "" && issue.Line > 0 {
				loc = fmt.Sprintf(" (%s:%d)", issue.File, issue.Line)
			}

			switch issue.Severity {
			case "ok":
				f.Println(f.OK(fmt.Sprintf("[%s] %s%s", issue.Category, issue.Message, loc)))
			case "warn":
				f.Println(f.Warn(fmt.Sprintf("[%s] %s%s", issue.Category, issue.Message, loc)))
			case "error":
				f.Println(f.Fail(fmt.Sprintf("[%s] %s%s", issue.Category, issue.Message, loc)))
			}
		}
		f.Footer()

		return nil
	},
}

var (
	cleanDryRun bool
)

var gitCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up stale/duplicate gitconfig entries",
	Long:  "Remove duplicate safe.directory entries and stale references from gitconfig.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveGitConfigPath()

		result, err := git.Clean(cfgPath, cleanDryRun)
		if err != nil {
			return fmt.Errorf("failed to clean gitconfig: %w", err)
		}

		f := report.NewFormatter(os.Stdout, output, noColor)

		if output == "json" {
			return f.JSON(result)
		}

		if len(result.Removed) == 0 {
			f.Println(f.OK("No stale or duplicate entries found."))
			return nil
		}

		if result.DryRun {
			f.Header("DRY RUN — entries that would be removed")
		} else {
			f.Header("CLEANED")
		}

		for _, entry := range result.Removed {
			f.Println(f.Warn("  " + entry))
		}
		f.Footer()

		if !result.DryRun && result.BackupDir != "" {
			f.Println(f.OK(fmt.Sprintf("Backup saved to: %s", result.BackupDir)))
		}

		return nil
	},
}

func init() {
	gitCmd.PersistentFlags().StringVar(&gitConfigFile, "gitconfig", "", "path to gitconfig (default: auto-detect)")
	gitCleanCmd.Flags().BoolVar(&cleanDryRun, "dry-run", false, "show what would be removed without changing files")

	gitCmd.AddCommand(gitProfilesCmd)
	gitCmd.AddCommand(gitDoctorCmd)
	gitCmd.AddCommand(gitCleanCmd)
	rootCmd.AddCommand(gitCmd)
}

func resolveGitConfigPath() string {
	if gitConfigFile != "" {
		return gitConfigFile
	}
	return git.DefaultGitConfigPath()
}
