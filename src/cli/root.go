// Package cli defines the hst command-line interface on top of package
// backend.
package cli

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const appName = "hst"

// rootCmd is the top-level hst command. Its --verbose flag enables debug
// logging for all subcommands. Handlers silence usage only after Cobra's
// argument and flag validation, then return runtime errors for Execute to log.
var rootCmd = &cobra.Command{
	SilenceErrors: true,
	Use:           appName,
	Short:         "Hestia links files from a project to where they belong",
	Long: `Hestia links files from a project to where they belong.

A project is a directory containing a .hestia directory, found by searching the
working directory and its ancestors. Its .hestia/mappings.conf file holds named
groups of mappings, each pairing a source path with a destination path.
Relative paths are resolved against the directory containing .hestia, and "~"
expands to the home directory.

Linking a group places each source at its destination, either as a symlink
(the default) or as a copy. Directory sources are linked file by file.

Argument-count and flag errors show command usage. Errors during command
execution are reported without usage.`,
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

// Execute runs the root command, logging errors and exiting with status 1
// on failure.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show debug output")
}
