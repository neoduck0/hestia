package cli

import (
	"fmt"
	"io"

	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:     "group",
	Aliases: []string{"g"},
	Short:   "Manage groups",
	Long: `Manage the groups in the mappings file.

A group is a named set of mappings that are linked together.
Use "group list" to print the group names in mappings-file order.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.NoArgs(cmd, args); err != nil {
			return err
		}
		return nil
	},
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the groups in the project",
	Long: `List the groups in the project's mappings file, one name per line.

Groups are printed in mappings-file order, including empty groups. If there
are no groups, nothing is printed. The mappings file is checked using the same
validation as other commands, including checking that mapped sources exist.
No files are changed.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.NoArgs(cmd, args); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		project := backend.NewProject()
		groups, err := project.ListGroups(backend.NewSettings())
		if err != nil {
			return err
		}

		return printGroups(cmd.OutOrStdout(), groups)
	},
}

var groupAddCmd = &cobra.Command{
	Use:     "add <group>",
	Aliases: []string{"a"},
	Short:   "Create an empty group",
	Long: `Create an empty group in the mappings file.

The name must not be blank, have leading or trailing whitespace, or contain
brackets or line breaks, and must not already be used by another group.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		project := backend.NewProject()
		settings := backend.NewSettings()

		return project.AddGroup(settings, args[0])
	},
}

var groupDeleteCmd = &cobra.Command{
	Use:     "delete <group>",
	Aliases: []string{"d"},
	Short:   "Delete a group and its mappings",
	Long: `Delete a group and all of its mappings from the mappings file.

Files that were already linked are not removed.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		project := backend.NewProject()
		settings := backend.NewSettings()

		return project.Delete(settings, args[0])
	},
}

var groupRenameCmd = &cobra.Command{
	Use:     "rename <old> <new>",
	Aliases: []string{"r"},
	Short:   "Rename a group",
	Long: `Rename a group in the mappings file, keeping its mappings.

The new name must follow the same rules as "group add" and must not already be
used by another group.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		project := backend.NewProject()
		settings := backend.NewSettings()

		return project.RenameGroup(settings, args[0], args[1])
	},
}

// printGroups writes one group name per line, stopping at the first write error.
func printGroups(w io.Writer, groups []string) error {
	for _, name := range groups {
		if _, err := fmt.Fprintln(w, name); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupAddCmd)
	groupCmd.AddCommand(groupDeleteCmd)
	groupCmd.AddCommand(groupRenameCmd)

	rootCmd.AddCommand(groupCmd)
}
