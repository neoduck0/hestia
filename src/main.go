// Command hst is the Hestia command-line tool. It deploys files from a
// project's source tree to destinations on the system by symlinking or
// copying them, as described by the project's mappings file.
package main

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/neoduck0/hestia/src/cli"
)

func main() {
	prepareLogger()
	cli.Execute()
}

// prepareLogger installs a stderr logger with custom level colors as the
// default logger.
func prepareLogger() {
	logger := log.New(os.Stderr)

	styles := log.DefaultStyles()

	styles.Levels[log.DebugLevel] = styles.Levels[log.DebugLevel].Foreground(lipgloss.Color("4"))
	styles.Levels[log.InfoLevel] = styles.Levels[log.InfoLevel].Foreground(lipgloss.Color("6"))
	styles.Levels[log.WarnLevel] = styles.Levels[log.WarnLevel].Foreground(lipgloss.Color("3"))
	styles.Levels[log.ErrorLevel] = styles.Levels[log.ErrorLevel].Foreground(lipgloss.Color("1"))
	styles.Levels[log.FatalLevel] = styles.Levels[log.FatalLevel].Foreground(lipgloss.Color("5"))

	logger.SetStyles(styles)

	log.SetDefault(logger)
}
