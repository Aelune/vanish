package helpers

import(
	"os"
	"strings"
)
// IsColorTerminal a helper function to detect color terminal support
func IsColorTerminal() bool {
	term := os.Getenv("TERM")
	colorTerm := os.Getenv("COLORTERM")

	// Check for color terminal indicators
	if colorTerm != "" {
		return true
	}

	// Common color-supporting terminals
	colorTerms := []string{
		"xterm-color", "xterm-256color", "screen-256color",
		"tmux-256color", "rxvt-unicode-256color",
	}

	for _, ct := range colorTerms {
		if strings.Contains(term, ct) || term == ct {
			return true
		}
	}

	return false
}
