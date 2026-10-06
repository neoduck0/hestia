package cli

import (
	"errors"

	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:     "link [group...]",
	Aliases: []string{"l"},
	Short:   "Link the mappings of one or more groups",
	Long: `Link the mappings of the given groups.

With --all, every group is linked and no groups may be given. With --exclude,
every group except the given ones is linked.

Each source is placed at its destination as a symlink by default, or as a copy
with --copy. Parent directories are created as needed. When symlinking,
destinations that already link to their source are left untouched.

With --dry-run, the mappings file is checked but nothing is linked.`,
	Args: func(cmd *cobra.Command, args []string) error {
		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			return err
		}

		exclude, err := cmd.Flags().GetBool("exclude")
		if err != nil {
			return err
		}

		if all && len(args) > 0 {
			return errors.New("group arguments cannot be used with --all")
		}

		if !all && len(args) == 0 && !exclude {
			return errors.New("no groups specified")
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		project := backend.NewProject()
		settings := backend.NewSettings()

		copyFiles, err := cmd.Flags().GetBool("copy")
		if err != nil {
			return err
		}

		symlinkFiles, err := cmd.Flags().GetBool("symlink")
		if err != nil {
			return err
		}

		if copyFiles {
			err = backend.SetOp(backend.OpCopy, &settings.ForceOp)
		} else if symlinkFiles {
			err = backend.SetOp(backend.OpSymlink, &settings.ForceOp)
		}
		if err != nil {
			return err
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}
		settings.SetDryRun(dryRun)

		argsSet := make(map[string]struct{}, len(args))
		for _, arg := range args {
			argsSet[arg] = struct{}{}
		}

		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			return err
		}

		exclude, err := cmd.Flags().GetBool("exclude")
		if err != nil {
			return err
		}

		return project.Link(settings, argsSet, all, exclude)
	},
}

func init() {
	linkCmd.Flags().Bool("dry-run", false, "check the mappings file without linking anything")

	linkCmd.Flags().BoolP("copy", "c", false, "copy every mapping instead of symlinking")
	linkCmd.Flags().BoolP("symlink", "s", false, "symlink every mapping (the default)")
	linkCmd.MarkFlagsMutuallyExclusive("copy", "symlink")

	linkCmd.Flags().BoolP("exclude", "e", false, "link every group except the given ones")
	linkCmd.Flags().BoolP("all", "a", false, "link every group")
	linkCmd.MarkFlagsMutuallyExclusive("all", "exclude")

	rootCmd.AddCommand(linkCmd)
}
