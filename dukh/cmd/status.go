package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	icolor "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/format"
	"github.com/vhula/grazhda/internal/path"
)

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show dukh process status (running, PID, uptime)",
		Long: `# dukh status

Show whether the dukh daemon is currently running.

## Output

The command checks the PID file at ` + "`$GRAZHDA_DIR/run/dukh.pid`" + ` and
verifies the process is alive. If reachable, it also queries the gRPC
endpoint for uptime.

| Symbol | Meaning                         |
|--------|---------------------------------|
| ` + "`●`" + `    | Daemon is running (green)       |
| ` + "`○`" + `    | Daemon is not running (yellow)  |

## Examples

` + "```" + `
$ dukh status
●  dukh: running  (pid 48201, uptime: 2h 14m)
` + "```" + `

` + "```" + `
$ dukh status
○  dukh: not running
` + "```" + `

A stale PID file (process dead) is automatically cleaned up.
`,
		RunE: runDukhStatus,
	}
}

func runDukhStatus(_ *cobra.Command, _ []string) error {
	grazhdaDir, err := path.GrazhdaDir()
	if err != nil {
		return fmt.Errorf("cannot determine grazhda directory: %w", err)
	}

	pid, err := readPIDFile(grazhdaDir)
	if err != nil {
		fmt.Println(icolor.Yellow("○") + "  dukh: not running")
		return nil
	}
	if pid == 0 {
		fmt.Println(icolor.Yellow("○") + "  dukh: not running")
		return nil
	}

	if !isProcessAlive(pid) {
		fmt.Printf("%s  dukh: not running  %s\n",
			icolor.Yellow("○"),
			icolor.Yellow(fmt.Sprintf("(stale pid %d — removing pid file)", pid)),
		)
		_ = os.Remove(filepath.Join(grazhdaDir, "run", "dukh.pid"))
		return nil
	}

	// Try gRPC with a short timeout to get uptime from the live server.
	uptime := tryGetUptime()
	if uptime != "" {
		fmt.Printf("%s  dukh: %s  (pid %d, uptime: %s)\n",
			icolor.Green("●"),
			icolor.Green("running"),
			pid,
			icolor.Blue(uptime),
		)
	} else {
		fmt.Printf("%s  dukh: %s  (pid %d)\n",
			icolor.Green("●"),
			icolor.Green("running"),
			pid,
		)
	}
	return nil
}

// readPIDFile reads and returns the PID stored in $GRAZHDA_DIR/run/dukh.pid.
// Returns 0 if the file does not exist or cannot be parsed.
func readPIDFile(grazhdaDir string) (int, error) {
	pidPath := filepath.Join(grazhdaDir, "run", "dukh.pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

// tryGetUptime attempts a quick gRPC Status call to retrieve the server's
// uptime. Returns an empty string if dukh is unreachable or the call fails.
func tryGetUptime() string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c, err := dial()
	if err != nil {
		return ""
	}
	defer c.Close()

	resp, err := c.Status(ctx, "", false)
	if err != nil {
		return ""
	}
	return format.Uptime(time.Duration(resp.UptimeSeconds) * time.Second)
}
