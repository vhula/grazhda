//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// setDetach configures cmd to start in a new session so it is fully detached
// from the parent process and survives the parent exiting.
func setDetach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
