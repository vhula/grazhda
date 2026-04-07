//go:build windows

package main

import "os/exec"

// setDetach is a no-op on Windows; use the Windows Job API or sc.exe for
// true daemon behaviour if needed.
func setDetach(_ *exec.Cmd) {}
