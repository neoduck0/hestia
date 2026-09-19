package cli

import (
	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

// initCmd implements "hst init", which creates a Hestia project in the
// working directory.
var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Create a Hestia project in the working directory",
	Long: `Create a Hestia project in the working directory.

This creates a .hestia directory containing an empty mappings.conf file. An
existing .hestia directory or mappings file is left as it is.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := backend.Init()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
