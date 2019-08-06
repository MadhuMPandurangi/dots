package main

import (
	"os"
	"runtime/debug"
	"time"

	"github.com/fatih/color"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"

	"go.evanpurkhiser.com/dots/config"
)

var (
	sourceConfig   *config.SourceConfig
	sourceLockfile *config.SourceLockfile
)

func loadConfigs(cmd *cobra.Command, args []string) error {
	var err error

	path := config.SourceConfigPath()

	sourceConfig, err = config.LoadSourceConfig(path)
	if err != nil {
		return err
	}

	sourceLockfile, err = config.LoadLockfile(sourceConfig)
	if err != nil {
		return err
	}

	warns := config.SanitizeSourceConfig(sourceConfig)
	for _, err := range warns {
		color.New(color.FgYellow).Fprintf(os.Stderr, "warn: %s\n", err)
	}

	return nil
}

var rootCmd = cobra.Command{
	Use:   "dots",
	Short: "A portable tool for managing a single set of dotfiles",

	SilenceUsage:      true,
	SilenceErrors:     true,
	PersistentPreRunE: loadConfigs,
}

func sentryRecover() {
	err := recover()
	if err == nil {
		return
	}

	sentry.CurrentHub().Recover(err)
	sentry.Flush(time.Second * 5)
	debug.PrintStack()
}

func main() {
	sentry.Init(sentry.ClientOptions{
		Dsn: "https://4c3f2bfcecf64bda8a4729f205e9a540@sentry.io/1522580",
	})

	defer sentryRecover()

	cobra.EnableCommandSorting = false

	rootCmd.AddCommand(&filesCmd)
	rootCmd.AddCommand(&diffCmd)
	rootCmd.AddCommand(&installCmd)
	rootCmd.AddCommand(&configCmd)

	if err := rootCmd.Execute(); err != nil {
		color.New(color.FgRed).Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
