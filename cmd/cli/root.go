package cli

import (
	"fmt"
	"os"

	"github.com/somaz94/bash-pilot/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	output  string
	noColor bool
	appCfg  *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "bash-pilot",
	Short: "A powerful CLI toolkit for bash power users",
	Long:  "bash-pilot — SSH management, Git multi-profile, environment health checks, and smart prompt.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		appCfg, err = config.Load(cfgFile)
		if err != nil {
			// Config file is optional; use defaults if not found
			appCfg = config.Default()
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default ~/.config/bash-pilot/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "color", "output format: color, plain, json, table")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
