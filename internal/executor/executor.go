package executor

import "os/exec"

// Executor runs shell commands in a given working directory.
type Executor interface {
	Run(dir string, command string) error
}

// OsExecutor runs commands via sh -c using os/exec.
type OsExecutor struct{}

func (e OsExecutor) Run(dir string, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	return cmd.Run()
}
