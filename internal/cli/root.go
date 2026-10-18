// Package cli defines the hst command-line interface.
package cli

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const appName = "hst"

var rootCmd = &cobra.Command{
	SilenceErrors: true,
	SilenceUsage:  true,
	Use:           appName,
	Short:         "Hestia links files from a project to where they belong",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	},
}

// Execute runs hst, logging errors and exiting with status 1 on failure.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show debug output")
}
