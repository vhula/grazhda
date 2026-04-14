package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
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
	conn, client, err := dial()
	if err != nil {
		fmt.Fprintln(os.Stderr, icolor.Red("✗ "+err.Error()))
		return err
	}
	defer conn.Close()

	resp, err := client.Stop(context.Background(), &dukhpb.StopRequest{})
	if err != nil {
		fmt.Fprintln(os.Stderr, icolor.Red("✗ dukh stop failed: "+err.Error()))
		return err
	}

	fmt.Println(icolor.Green("✓ " + resp.Message))
	return nil
}
