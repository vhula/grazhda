//go:build !windows

package main

import "syscall"

// isProcessAlive returns true if a process with the given PID exists and is
// running. It sends signal 0, which does not kill the process but fails if the
// process does not exist or is not reachable.
func isProcessAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil
}
