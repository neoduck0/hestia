package cli

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show each group's linked mappings and status",
	Long: `Show every group in mappings-file order with a linked/total mapping count
and a status of linked, partially linked, not linked, or empty.

A mapping is linked when its destination is a symlink to its source or a copy
with matching contents and permissions. A directory mapping counts as one
mapping and is linked only when all its files are linked. Empty directories
and directories with any unlinked files are not linked. Groups are partially
linked only when some, but not all, of their mappings are fully linked.
Empty groups are shown as empty with a count of 0/0.

The mappings file is validated, including checking that sources exist.
No files are changed. If there are no groups, nothing is printed.`,
	Args: cobra.NoArgs,
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

// printStatuses writes aligned group names, mapping counts, and states.
func printStatuses(w io.Writer, statuses []backend.GroupStatus) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	for _, status := range statuses {
		var label string

		switch state := status.State(); state {
		case backend.GroupNotLinked:
			label = "not linked"
		case backend.GroupPartiallyLinked:
			label = "partially linked"
		case backend.GroupLinked:
			label = "linked"
		case backend.GroupEmpty:
			label = "empty"
		default:
			return fmt.Errorf("unknown group state: %d", state)
		}

		_, err := fmt.Fprintf(tw, "%s\t%d/%d\t%s\n", status.Name, status.Linked, status.Total, label)
		if err != nil {
			return err
		}
	}
	return tw.Flush()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
