package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"

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
	// Execute prints the error once; cobra would print it a second time.
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Flags are parsed by now, so a flag error still shows usage but a runtime error does not.
		cmd.SilenceUsage = true
		var err error
		appCfg, err = loadConfig(cfgFile, cmd.ErrOrStderr())
		return err
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default ~/.config/bash-pilot/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "color", "output format: color, plain, json, table")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
}

// loadConfig treats a missing default config as "use defaults", a failing
// --config as an error, and a broken default config as a warning plus defaults.
func loadConfig(path string, stderr io.Writer) (*config.Config, error) {
	cfg, err := config.Load(path)
	switch {
	case err == nil:
		return cfg, nil
	case path != "":
		return nil, fmt.Errorf("load config: %w", err)
	case errors.Is(err, fs.ErrNotExist):
		return config.Default(), nil
	default:
		fmt.Fprintf(stderr, "warning: %v; using defaults\n", err)
		return config.Default(), nil
	}
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(rootCmd.ErrOrStderr(), "Error:", err)
		return err
	}
	return nil
}
