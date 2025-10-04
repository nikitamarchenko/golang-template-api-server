// Package cmd contain cobra initialization
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	argGlobalLogLevel      = "log-level"
	argGlobalLogShowSource = "log-show-source"
)

var rootCmd = &cobra.Command{ //nolint:gochecknoglobals // ok for cobra
	Use:   "data-app",
	Short: "",
	Long:  ``,
}

// Execute entry point for main func.
func Execute() error {
	return rootCmd.Execute() //nolint:wrapcheck // don't need here
}

func init() { //nolint:gochecknoinits // ok for cobra
	err := rootInitRun()
	if err != nil {
		_, _ = fmt.Printf("Error: cmd.root.init: %v", err) //nolint:forbidigo // we don't have logger here

		os.Exit(ErrorExitCodeCMDInit)
	}
}

func rootInitRun() error {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Global args
	flags := rootCmd.PersistentFlags()

	flags.String(argGlobalLogLevel, "INFO", "log level")

	err := viper.BindPFlag(argGlobalLogLevel, flags.Lookup(argGlobalLogLevel))
	if err != nil {
		return err //nolint:wrapcheck // don't need wrap here
	}

	flags.Bool(argGlobalLogShowSource, false, "log show source")

	err = viper.BindPFlag(argGlobalLogShowSource, flags.Lookup(argGlobalLogShowSource))
	if err != nil {
		return err //nolint:wrapcheck // don't need wrap here
	}

	return nil
}
