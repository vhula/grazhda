package exec

import (
	"os/exec"
)

func RunShell(command string, dir string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	return string(output), err
}
