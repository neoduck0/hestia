package cli

import (
	"errors"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

// addCmd implements "hst add <src> <dst> --group <group>", which adds a
// mapping to a group.
var addCmd = &cobra.Command{
	Use:     "add <src> <dst>",
	Aliases: []string{"a"},
	Short:   "Add a mapping from src to dst to a group",
	Long: `Add a mapping from src to dst to the group given by --group.

The source must exist, and the destination must not already be mapped in any
group. The group must exist unless --create is given, in which case it is
created.

Relative paths are resolved against the directory containing .hestia, not the
working directory.

Unless --no-portable is given, absolute paths under the home directory are
stored with a leading "~" so the mappings file works for other users.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		group, err := cmd.Flags().GetString("group")
		if err != nil {
			log.Fatal(err)
		}

		noPortable, err := cmd.Flags().GetBool("no-portable")
		if err != nil {
			log.Fatal(err)
		}

		create, err := cmd.Flags().GetBool("create")
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(group) == "" {
			log.Fatal(errors.New("group is required"))
		}

		project := backend.NewProject()
		settings := backend.NewSettings()

		settings.NoPortable = noPortable

		if err = project.Add(settings, group, args[0], args[1], create); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	addCmd.Flags().StringP("group", "g", "", "group to add the mapping to")
	addCmd.MarkFlagRequired("group")

	addCmd.Flags().Bool("create", false, "create the group if it does not exist")

	addCmd.Flags().Bool("no-portable", false, "store paths as given instead of collapsing the home directory to \"~\"")

	rootCmd.AddCommand(addCmd)
}
