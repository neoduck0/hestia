package cli

import (
	"errors"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

// deleteCmd implements "hst delete --group <group>", which removes a group
// and its mappings.
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d"},
	Short:   "Delete a group and its mappings",
	Long: `Delete the group given by --group and all of its mappings from the mappings
file.

Files that were already linked are not removed.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		group, err := cmd.Flags().GetString("group")
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(group) == "" {
			log.Fatal(errors.New("group is required"))
		}

		project := backend.NewProject()
		settings := backend.NewSettings()

		if err = project.Delete(settings, group); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	deleteCmd.Flags().StringP("group", "g", "", "group to delete")
	deleteCmd.MarkFlagRequired("group")

	rootCmd.AddCommand(deleteCmd)
}
