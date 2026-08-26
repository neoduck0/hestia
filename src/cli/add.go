package cli

import (
	"errors"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "",
	Long:  "",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		group, err := cmd.Flags().GetString("group")
		if err != nil {
			log.Fatal(err)
		}

		noPortable, err := cmd.Flags().GetBool("no-portable")
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(group) == "" {
			log.Fatal(errors.New("group is required"))
		}

		project := backend.NewProject()
		settings := backend.NewSettings()

		settings.NoPortable = noPortable

		if err = project.Add(settings, group, args[0], args[1]); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	addCmd.Flags().StringP("group", "g", "", "")
	addCmd.MarkFlagRequired("group")

	addCmd.Flags().Bool("no-portable", false, "")

	rootCmd.AddCommand(addCmd)
}
