package ws

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	dukhclient "github.com/vhula/grazhda/dukh/client"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
	icolor "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/format"
	ipath "github.com/vhula/grazhda/internal/path"
)

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show workspace health as monitored by dukh",
		Long: `Query the running **dukh** health monitor and display branch alignment status
for every repository in the targeted workspace.

**dukh** is automatically started if it is not already running. Use **--rescan**
to force a fresh filesystem scan before the report is displayed.

Each repository is shown as one of:

- **aligned** — the actual branch matches the configured branch (green ✓)
- **drifted** — branch mismatch detected (yellow ⚠)
- **missing** — repository directory does not exist on disk (red ✗)`,
		Example: `  # Show status for the default workspace
  zgard ws status

  # Show status for a named workspace
  zgard ws status -n myworkspace

  # Force a fresh workspace scan before reporting
  zgard ws status --rescan`,
		RunE: runWsStatus,
	}
	cmd.Flags().Bool("rescan", false, "Trigger a fresh workspace rescan before reporting (waits for completion)")
	return cmd
}

func runWsStatus(cmd *cobra.Command, _ []string) error {
	rescan, _ := cmd.Flags().GetBool("rescan")

	c, err := dukhclient.NewDefault()
	if err != nil {
		fmt.Fprintln(os.Stderr, icolor.Red("✗ "+err.Error()))
		return err
	}
	defer c.Close()

	ctx := context.Background()
	if rescan {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		fmt.Println(icolor.Blue("⟳ rescanning workspaces…"))
	}

	resp, err := c.Status(ctx, wsName, rescan)
	if err != nil {
		// dukh is not running — attempt to auto-start it.
		resp, err = tryAutoStartAndRetry(c, wsName, rescan)
		if err != nil {
			fmt.Fprintln(os.Stderr, icolor.Red("✗ dukh status failed: "+err.Error()))
			return err
		}
	}

	renderWsStatus(resp)
	return nil
}

// tryAutoStartAndRetry launches `dukh start`, waits for the server to become
// ready, then retries the Status RPC. Returns the response or an error if
// dukh could not be started or did not become ready within the timeout.
func tryAutoStartAndRetry(c *dukhclient.Client, wsName string, rescan bool) (*dukhpb.StatusResponse, error) {
	fmt.Println(icolor.Blue("⟳ dukh is not running — starting…"))

	if err := startDukh(); err != nil {
		return nil, fmt.Errorf("auto-start dukh: %w", err)
	}

	if err := waitForDukh(c, 10*time.Second); err != nil {
		return nil, fmt.Errorf("dukh did not become ready: %w", err)
	}

	fmt.Println(icolor.Green("✓") + " dukh started")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return c.Status(ctx, wsName, rescan)
}

// startDukh executes `dukh start` as a subprocess, reusing the existing
// daemonization logic in the dukh binary.
func startDukh() error {
	dukhBin := ipath.DukhBin()
	cmd := exec.Command(dukhBin, "start")
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// waitForDukh polls the gRPC server until a Status RPC succeeds or the
// timeout elapses.
func waitForDukh(c *dukhclient.Client, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_, err := c.Status(ctx, "", false)
		cancel()
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout after %s", timeout)
}

func renderWsStatus(resp *dukhpb.StatusResponse) {
	uptime := time.Duration(resp.UptimeSeconds) * time.Second
	fmt.Printf("%s  %s  •  uptime: %s\n\n",
		icolor.Blue("Dukh"),
		icolor.Green("running"),
		format.Uptime(uptime),
	)

	var aligned, drifted, missing int

	for _, ws := range resp.Workspaces {
		fmt.Printf("%s %s\n", icolor.Blue("Workspace:"), icolor.Blue(ws.Name))
		for _, proj := range ws.Projects {
			fmt.Printf("  %s %s\n", icolor.Blue("Project:"), proj.Name)
			for _, repo := range proj.Repositories {
				switch {
				case !repo.Exists:
					missing++
					fmt.Printf("    %s %-16s %s\n",
						icolor.Red("✗"),
						repo.Name,
						icolor.Red("(missing)"),
					)
				case !repo.BranchAligned:
					drifted++
					fmt.Printf("    %s %-16s %s → %s  %s\n",
						icolor.Red("✗"),
						repo.Name,
						repo.ConfiguredBranch,
						icolor.Yellow(repo.ActualBranch),
						icolor.Yellow("(branch mismatch)"),
					)
				default:
					aligned++
					fmt.Printf("    %s %-16s %s → %s\n",
						icolor.Green("✓"),
						repo.Name,
						repo.ConfiguredBranch,
						icolor.Green(repo.ActualBranch),
					)
				}
			}
		}
		fmt.Println()
	}

	fmt.Printf("%s %d aligned  %s %d drifted  %s %d missing\n",
		icolor.Green("✓"),
		aligned,
		icolor.Yellow("⚠"),
		drifted,
		icolor.Red("✗"),
		missing,
	)
}
