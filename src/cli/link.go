package cli

import (
	"errors"

	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/src/backend"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:     "link [group...]",
	Aliases: []string{"l"},
	Short:   "",
	Long:    "",
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
	Run: func(cmd *cobra.Command, args []string) {
		project := backend.NewProject()
		settings := backend.NewSettings()

		copyFiles, err := cmd.Flags().GetBool("copy")
		if err != nil {
			log.Fatal(err)
		}

		symlinkFiles, err := cmd.Flags().GetBool("symlink")
		if err != nil {
			log.Fatal(err)
		}

		if copyFiles {
			err = backend.SetOp(backend.OpCopy, &settings.ForceOp)
		} else if symlinkFiles {
			err = backend.SetOp(backend.OpSymlink, &settings.ForceOp)
		}
		if err != nil {
			log.Fatal(err)
		}

		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			log.Fatal(err)
		}
		settings.SetDryRun(dryRun)

		argsSet := make(map[string]struct{}, len(args))
		for _, arg := range args {
			argsSet[arg] = struct{}{}
		}

		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Fatal(err)
		}

		exclude, err := cmd.Flags().GetBool("exclude")
		if err != nil {
			log.Fatal(err)
		}

		err = project.Link(settings, argsSet, all, exclude)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	linkCmd.Flags().Bool("dry-run", false, "")

	linkCmd.Flags().BoolP("copy", "c", false, "")
	linkCmd.Flags().BoolP("symlink", "s", false, "")
	linkCmd.MarkFlagsMutuallyExclusive("copy", "symlink")

	linkCmd.Flags().BoolP("exclude", "e", false, "")
	linkCmd.Flags().BoolP("all", "a", false, "")
	linkCmd.MarkFlagsMutuallyExclusive("all", "exclude")

	rootCmd.AddCommand(linkCmd)
}
