// Package cli defines the hst command-line interface on top of package
// backend.
package cli

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// appName is the name of the executable and the root command.
const appName = "hst"

// rootCmd is the top-level hst command. Its --verbose flag enables debug
// logging for all subcommands.
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "Hestia links files from a project to where they belong",
	Long: `Hestia links files from a project to where they belong.

A project is a directory containing a .hestia directory, found by searching the
working directory and its ancestors. Its .hestia/mappings.conf file holds named
groups of mappings, each pairing a source path with a destination path.
Relative paths are resolved against the directory containing .hestia, and "~"
expands to the home directory.

Linking a group places each source at its destination, either as a symlink
(the default) or as a copy. Directory sources are linked file by file.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	},
}

// Execute runs the root command and exits with status 1 on failure.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "show debug output")
}
