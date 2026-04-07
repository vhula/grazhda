//go:build windows

package main

import (
	"os"
)

// isProcessAlive checks if the process with the given PID is alive on Windows
// by attempting to open it. Returns true if the process exists.
func isProcessAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows, FindProcess always succeeds; use a 0-signal probe isn't
	// available so we rely on OpenProcess via os package internals.
	// This is best-effort.
	_ = p
	return true
}
