package cli

import (
	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Create a Hestia project in the working directory",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.NoArgs(cmd, args); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return backend.Init()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
