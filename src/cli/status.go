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
	Long: `Show every group in mappings-file order with a linked/total mapping count
and a status of linked, partially linked, not linked, or empty.
On terminals, group names are bold and status labels use ANSI colors: green
for linked, yellow for partially linked, red for not linked. Empty groups have
uncolored status labels. Redirected output and TERM=dumb terminals use plain
text. Set NO_COLOR to a non-empty value to disable ANSI styling (for example,
NO_COLOR=1 hst status).

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

// printStatuses writes aligned group names, mapping counts, and states, using
// ANSI styling only when style.Enabled permits it for the output destination.
// Every group name has the same styling escape sequences, so their extra
// width cancels out when tabwriter computes padding. Colored labels are in the
// final column and do not affect the preceding columns' alignment.
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
