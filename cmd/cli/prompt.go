package cli

import (
	"fmt"
	"os"

	"github.com/somaz94/bash-pilot/internal/prompt"
	"github.com/somaz94/bash-pilot/internal/report"
	"github.com/spf13/cobra"
)

var (
	promptTheme string
	promptNoK8s bool
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Smart bash prompt with git and k8s context",
	Long:  "Generate and preview a smart bash prompt with git branch, k8s context, and exit code indicator.",
}

var promptInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate prompt script for bash",
	Long: `Generate a bash prompt script to stdout.

Usage:
  # Preview the script
  bash-pilot prompt init

  # Apply immediately
  eval "$(bash-pilot prompt init)"

  # Persist in your shell profile
  echo 'eval "$(bash-pilot prompt init)"' >> ~/.bashrc`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := prompt.Options{
			Theme: parseTheme(promptTheme),
			NoK8s: promptNoK8s,
		}

		script := prompt.GenerateScript(opts)
		fmt.Print(script)
		return nil
	},
}

var promptShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Preview prompt components for current environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := prompt.Options{
			Theme: parseTheme(promptTheme),
			NoK8s: promptNoK8s,
		}

		components := prompt.ShowComponents(opts)

		f := report.NewFormatter(os.Stdout, output, noColor)

		if output == "json" {
			return f.JSON(components)
		}

		f.Header("PROMPT COMPONENTS")
		for _, c := range components {
			f.Println(f.OK(fmt.Sprintf("%-12s %s", c.Name+":", c.Value)))
		}
		f.Footer()

		return nil
	},
}

func parseTheme(s string) prompt.Theme {
	switch s {
	case "full":
		return prompt.ThemeFull
	default:
		return prompt.ThemeMinimal
	}
}

func init() {
	promptCmd.PersistentFlags().StringVar(&promptTheme, "theme", "minimal", "Prompt theme: minimal (git only), full (git + k8s)")
	promptCmd.PersistentFlags().BoolVar(&promptNoK8s, "no-k8s", false, "Disable k8s context in prompt")

	promptCmd.AddCommand(promptInitCmd)
	promptCmd.AddCommand(promptShowCmd)
	rootCmd.AddCommand(promptCmd)
}
