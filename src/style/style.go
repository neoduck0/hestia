// Package style provides ANSI styling and terminal detection.
package style

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
)

// Color is an ANSI foreground color code.
type Color string

// Standard ANSI foreground colors.
const (
	Red    Color = "31"
	Green  Color = "32"
	Yellow Color = "33"
)

// Style combines ANSI text attributes; its zero value leaves text unchanged.
type Style struct {
	Color  Color
	Bold   bool
	Italic bool
}

// Enabled reports whether the output is a terminal with ANSI styling enabled.
func Enabled(w io.Writer) bool {
	if os.Getenv("TERM") == "dumb" || os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(f.Fd())
}

// Wrap applies the style to text and resets styling afterward.
func (s Style) Wrap(text string) string {
	var codes []string
	if s.Color != "" {
		codes = append(codes, string(s.Color))
	}
	if s.Bold {
		codes = append(codes, "1")
	}
	if s.Italic {
		codes = append(codes, "3")
	}
	if len(codes) == 0 {
		return text
	}
	return "\x1b[" + strings.Join(codes, ";") + "m" + text + "\x1b[0m"
}
