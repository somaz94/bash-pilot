package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

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

// runGroupHelp makes a command group runnable, because cobra validates Args only on
// runnable commands; a mistyped subcommand then fails instead of printing help and exiting 0.
func runGroupHelp(cmd *cobra.Command, _ []string) error {
	return cmd.Help()
}

// groupArgs rejects a mistyped subcommand with the "Did you mean" hint cobra gives only at the root.
func groupArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	if cmd.SuggestionsMinimumDistance <= 0 {
		cmd.SuggestionsMinimumDistance = 2
	}
	msg := fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath())
	if s := cmd.SuggestionsFor(args[0]); len(s) > 0 {
		msg += "\n\nDid you mean this?\n\t" + strings.Join(s, "\n\t")
	}
	return errors.New(msg)
}

// Execute runs the root command.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(rootCmd.ErrOrStderr(), "Error:", err)
		return err
	}
	return nil
}
