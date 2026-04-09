// Package ui provides terminal rendering helpers for the zgard CLI.
package ui

import (
	"github.com/charmbracelet/glamour"
	clr "github.com/vhula/grazhda/internal/color"
)

// Render processes a Markdown string with glamour for terminal display.
// It auto-detects the terminal's light/dark background and applies an
// appropriate style. When colours are disabled (--no-color or NO_COLOR env)
// plain, unstyled text is produced instead.
// Falls back gracefully to the raw string if rendering fails.
func Render(md string) string {
	if md == "" {
		return ""
	}

	var opts []glamour.TermRendererOption
	if clr.IsDisabled() {
		opts = append(opts, glamour.WithStandardStyle("notty"))
	} else {
		opts = append(opts, glamour.WithAutoStyle())
	}
	opts = append(opts, glamour.WithWordWrap(100))

	r, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		return md
	}

	out, err := r.Render(md)
	if err != nil {
		return md
	}
	return out
}
