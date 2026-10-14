package cli

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/neoduck0/hestia/src/backend"
	"github.com/neoduck0/hestia/src/style"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show each group's linked mappings and status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		project := backend.NewProject()
		statuses, err := project.Status(backend.NewSettings())
		if err != nil {
			return err
		}
		return printStatuses(cmd.OutOrStdout(), statuses)
	},
}

func printStatuses(w io.Writer, statuses []backend.GroupStatus) error {
	useStyling := style.Enabled(w)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	for _, status := range statuses {
		var label string
		var labelStyle style.Style

		switch state := status.State(); state {
		case backend.GroupNotLinked:
			label, labelStyle.Color = "not linked", style.Red
		case backend.GroupPartiallyLinked:
			label, labelStyle.Color = "partially linked", style.Yellow
		case backend.GroupLinked:
			label, labelStyle.Color = "linked", style.Green
		case backend.GroupEmpty:
			label = "empty"
		default:
			return fmt.Errorf("unknown group state: %d", state)
		}

		name := status.Name
		if useStyling {
			name = style.Style{Bold: true}.Wrap(name)
			label = labelStyle.Wrap(label)
		}
		_, err := fmt.Fprintf(tw, "%s\t%d/%d\t%s\n", name, status.Linked, status.Total, label)
		if err != nil {
			return err
		}
	}
	return tw.Flush()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
