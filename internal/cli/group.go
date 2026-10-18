package cli

import (
	"fmt"
	"io"

	"github.com/neoduck0/hestia/internal/backend"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:     "group",
	Aliases: []string{"g"},
	Short:   "Manage groups",
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
