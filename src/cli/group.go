package cli

import (
	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

// groupCmd is the parent of the group management subcommands.
var groupCmd = &cobra.Command{
	Use:     "group",
	Aliases: []string{"g"},
	Short:   "",
	Long:    "",
	Args:    cobra.NoArgs,
}

// groupAddCmd implements "hst group add", which creates an empty group.
var groupAddCmd = &cobra.Command{
	Use:     "add <group>",
	Aliases: []string{"a"},
	Short:   "",
	Long:    "",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := backend.NewProject()
		settings := backend.NewSettings()

		if err := project.AddGroup(settings, args[0]); err != nil {
			log.Fatal(err)
		}
	},
}

// groupDeleteCmd implements "hst group delete", which removes a group and its
// mappings.
var groupDeleteCmd = &cobra.Command{
	Use:     "delete <group>",
	Aliases: []string{"d"},
	Short:   "",
	Long:    "",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := backend.NewProject()
		settings := backend.NewSettings()

		if err := project.Delete(settings, args[0]); err != nil {
			log.Fatal(err)
		}
	},
}

// groupRenameCmd implements "hst group rename", which renames a group.
var groupRenameCmd = &cobra.Command{
	Use:     "rename <old> <new>",
	Aliases: []string{"r"},
	Short:   "",
	Long:    "",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		project := backend.NewProject()
		settings := backend.NewSettings()

		if err := project.RenameGroup(settings, args[0], args[1]); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	groupCmd.AddCommand(groupAddCmd)
	groupCmd.AddCommand(groupDeleteCmd)
	groupCmd.AddCommand(groupRenameCmd)

	rootCmd.AddCommand(groupCmd)
}
