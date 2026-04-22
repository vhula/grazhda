package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	icolor "github.com/vhula/grazhda/internal/color"
)

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running dukh workspace monitor",
		Long: `# dukh stop

Gracefully stop the running dukh daemon via a gRPC call.

The daemon receives the stop request, finishes any in-flight scan, removes
its PID file, and exits cleanly.

## Example

` + "```" + `
$ dukh stop
✓ dukh stopped
` + "```" + `

If the daemon is not running, the command prints an error and exits with a
non-zero code.
`,
		RunE: runStop,
	}
}

func runStop(_ *cobra.Command, _ []string) error {
	c, err := dial()
	if err != nil {
		printErr(err.Error())
		return err
	}
	defer c.Close()

	msg, err := c.Stop(context.Background())
	if err != nil {
		printErr("dukh stop failed: " + err.Error())
		return err
	}

	fmt.Println(icolor.Green("✓ " + msg))
	return nil
}
