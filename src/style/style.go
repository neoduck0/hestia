// Package style provides ANSI text styling and terminal detection for CLI output
// and logging. Callers use Enabled for their destination before applying styles.
package style

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/x/term"
)

// Color is an ANSI foreground color code. Its zero value leaves the
// foreground color unchanged.
type Color string

// Standard ANSI foreground colors.
const (
	Red    Color = "31"
	Green  Color = "32"
	Yellow Color = "33"
)

// Style combines ANSI text attributes for reuse across output destinations.
// The zero value leaves text unchanged. Attributes can be combined, for example:
//
//	Style{Color: Green, Bold: true, Italic: true}.Wrap("linked")
type Style struct {
	Color  Color
	Bold   bool
	Italic bool
}

// Enabled reports whether ANSI styling should be enabled for the actual
// output destination. TERM=dumb, a non-empty NO_COLOR, and non-terminal writers
// (including writers without a file descriptor) disable styling.
func Enabled(w io.Writer) bool {
	if os.Getenv("TERM") == "dumb" || os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(interface{ Fd() uintptr })
	return ok && term.IsTerminal(f.Fd())
}

// Wrap applies the style to text and resets terminal styling afterward so
// attributes do not leak into subsequent output. Wrap always applies the style;
// callers should check Enabled to decide whether to use it.
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
