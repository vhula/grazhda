// Package color provides simple terminal color helpers for CLI output.
// Colors are disabled automatically when output is not a TTY or when the
// NO_COLOR environment variable is set (https://no-color.org).
package color

import "github.com/fatih/color"

// Terminal color formatters. Each variable is a SprintFunc that wraps its
// argument in the corresponding ANSI colour sequence. When colour is disabled
// (via [Disable] or the NO_COLOR env var) they return the input unchanged.
var (
	Green  = color.New(color.FgGreen).SprintFunc()
	Red    = color.New(color.FgRed).SprintFunc()
	Yellow = color.New(color.FgYellow).SprintFunc()
	Blue   = color.New(color.FgHiBlue).SprintFunc()
)

// Disable turns off all terminal colors. It should be called once when
// the --no-color flag is supplied. Both fatih/color and any package that
// reads the NO_COLOR environment variable (e.g. lipgloss) will comply.
func Disable() {
	color.NoColor = true
}

// IsDisabled reports whether color output is currently disabled.
func IsDisabled() bool { return color.NoColor }
