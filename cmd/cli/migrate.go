package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/somaz94/bash-pilot/internal/migrate"
	"github.com/somaz94/bash-pilot/internal/report"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Export/import SSH and Git config across machines",
	Long: `Migrate SSH hosts and Git identity profiles between machines.

Usage:
  # Export current config
  bash-pilot migrate export > my-config.json

  # Import on new machine
  bash-pilot migrate import my-config.json --dry-run
  bash-pilot migrate import my-config.json`,
}

var migrateExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export SSH + Git config to portable JSON",
	Long: `Export SSH hosts, key references, and Git profiles to a portable JSON format.
Private keys are NOT included — only names and types.

Usage:
  bash-pilot migrate export > my-config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := migrate.Export("")
		if err != nil {
			return err
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(cfg)
	},
}

var migrateImportCmd = &cobra.Command{
	Use:   "import <config-file>",
	Short: "Import SSH + Git config from portable JSON",
	Long: `Import SSH hosts and Git profiles from a migrate config file.
Paths are automatically translated to the local home directory.
Existing hosts and profiles are not overwritten.

Usage:
  bash-pilot migrate import my-config.json --dry-run
  bash-pilot migrate import my-config.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("cannot read config file: %w", err)
		}

		var cfg migrate.MigrateConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("invalid config file: %w", err)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		f := report.NewFormatter(os.Stdout, output, noColor)

		if output == "json" {
			result, err := migrate.Import(&cfg, dryRun)
			if err != nil {
				return err
			}
			return f.JSON(result)
		}

		title := "MIGRATE IMPORT"
		if dryRun {
			title = "MIGRATE IMPORT (dry-run)"
		}

		f.Header(title)

		result, err := migrate.Import(&cfg, dryRun)
		if err != nil {
			return err
		}

		fmt.Print(migrate.FormatImportResult(result))
		f.Footer()

		fmt.Println()
		if dryRun && (result.SSHHostsAdded > 0 || result.GitConfigWritten) {
			f.Println("Run without --dry-run to apply.")
		}

		return nil
	},
}

func init() {
	migrateImportCmd.Flags().Bool("dry-run", false, "Preview changes without applying")
	migrateCmd.AddCommand(migrateExportCmd)
	migrateCmd.AddCommand(migrateImportCmd)
	rootCmd.AddCommand(migrateCmd)
}
