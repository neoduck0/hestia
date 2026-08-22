package cli

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const appName = "hst"

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "",
	Long:  "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")
		if debug {
			log.SetLevel(log.DebugLevel)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "")
}
