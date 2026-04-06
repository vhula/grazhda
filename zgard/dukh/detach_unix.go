//go:build !windows

package dukh

import (
	"os/exec"
	"syscall"
)

// setDetach configures cmd to run in a new process group so it is fully
// detached from the parent process and survives the parent exiting.
func setDetach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
