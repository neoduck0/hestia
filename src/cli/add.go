package cli

import (
	"errors"
	"strings"

	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <src> <dst>",
	Aliases: []string{"a"},
	Short:   "Add a mapping from src to dst to a group",
	Long: `Add a mapping from src to dst to the group given by --group.

The source must exist. Duplicate destinations are accepted with a warning.
The group must exist unless --create is given.

Relative paths are resolved against the directory containing .hestia, not the
working directory.

Unless --no-portable is given, absolute paths under the home directory are
stored with a leading "~" so the mappings file works for other users.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		group, err := cmd.Flags().GetString("group")
		if err != nil {
			return err
		}
		if strings.TrimSpace(group) == "" {
			return errors.New("group is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		group, err := cmd.Flags().GetString("group")
		if err != nil {
			return err
		}

		noPortable, err := cmd.Flags().GetBool("no-portable")
		if err != nil {
			return err
		}

		create, err := cmd.Flags().GetBool("create")
		if err != nil {
			return err
		}

		project := backend.NewProject()
		settings := backend.NewSettings()

		settings.NoPortable = noPortable

		return project.Add(settings, group, args[0], args[1], create)
	},
}

func init() {
	addCmd.Flags().StringP("group", "g", "", "group to add the mapping to")
	addCmd.MarkFlagRequired("group")

	addCmd.Flags().Bool("create", false, "create the group if it does not exist")

	addCmd.Flags().Bool("no-portable", false, "store paths as given instead of collapsing the home directory to \"~\"")

	rootCmd.AddCommand(addCmd)
}
