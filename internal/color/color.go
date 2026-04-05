// Package color provides simple terminal color helpers for CLI output.
// Colors are disabled automatically when output is not a TTY or when the
// NO_COLOR environment variable is set (https://no-color.org).
package color

import "github.com/fatih/color"

var (
	Green  = color.New(color.FgGreen).SprintFunc()
	Red    = color.New(color.FgRed).SprintFunc()
	Yellow = color.New(color.FgYellow).SprintFunc()
	Blue   = color.New(color.FgHiBlue).SprintFunc()
)
