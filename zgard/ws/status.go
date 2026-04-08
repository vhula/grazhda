package ws

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
	icolor "github.com/vhula/grazhda/internal/color"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const dukhAddr = "localhost:50501"

func newStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show workspace health as monitored by dukh",
		RunE:  runWsStatus,
	}
	cmd.Flags().Bool("rescan", false, "Trigger a fresh workspace rescan before reporting (waits for completion)")
	return cmd
}

func runWsStatus(cmd *cobra.Command, _ []string) error {
	rescan, _ := cmd.Flags().GetBool("rescan")

	conn, client, err := dialDukh()
	if err != nil {
		fmt.Fprintln(os.Stderr, icolor.Red("✗ "+err.Error()))
		return err
	}
	defer conn.Close()

	ctx := context.Background()
	if rescan {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		fmt.Println(icolor.Blue("⟳ rescanning workspaces…"))
	}

	resp, err := client.Status(ctx, &dukhpb.StatusRequest{
		WorkspaceName: wsName,
		Rescan:        rescan,
	})
	if err != nil {
		// dukh is not running — attempt to auto-start it.
		resp, err = tryAutoStartAndRetry(client, wsName, rescan)
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
func tryAutoStartAndRetry(client dukhpb.DukhServiceClient, wsName string, rescan bool) (*dukhpb.StatusResponse, error) {
	fmt.Println(icolor.Blue("⟳ dukh is not running — starting…"))

	if err := startDukh(); err != nil {
		return nil, fmt.Errorf("auto-start dukh: %w", err)
	}

	if err := waitForDukh(client, 10*time.Second); err != nil {
		return nil, fmt.Errorf("dukh did not become ready: %w", err)
	}

	fmt.Println(icolor.Green("✓") + " dukh started")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.Status(ctx, &dukhpb.StatusRequest{
		WorkspaceName: wsName,
		Rescan:        rescan,
	})
}

// startDukh executes `dukh start` as a subprocess, reusing the existing
// daemonization logic in the dukh binary.
func startDukh() error {
	dukhBin := resolveDukhBin()
	cmd := exec.Command(dukhBin, "start")
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// resolveDukhBin returns the path to the dukh binary. It checks
// $GRAZHDA_DIR/bin/dukh first, then falls back to PATH lookup.
func resolveDukhBin() string {
	grazhdaDir := os.Getenv("GRAZHDA_DIR")
	if grazhdaDir != "" {
		candidate := grazhdaDir + "/bin/dukh"
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if p, err := exec.LookPath("dukh"); err == nil {
		return p
	}
	return "dukh"
}

// waitForDukh polls the gRPC server until a Status RPC succeeds or the
// timeout elapses.
func waitForDukh(client dukhpb.DukhServiceClient, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		_, err := client.Status(ctx, &dukhpb.StatusRequest{})
		cancel()
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout after %s", timeout)
}

// dialDukh opens a gRPC connection to the dukh server.
// The connection is lazy — it will fail on the first RPC if dukh is not running.
func dialDukh() (*grpc.ClientConn, dukhpb.DukhServiceClient, error) {
	conn, err := grpc.NewClient(dukhAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create dukh client for %s: %w", dukhAddr, err)
	}
	return conn, dukhpb.NewDukhServiceClient(conn), nil
}

func renderWsStatus(resp *dukhpb.StatusResponse) {
	uptime := time.Duration(resp.UptimeSeconds) * time.Second
	fmt.Printf("%s  %s  •  uptime: %s\n\n",
		icolor.Blue("Dukh"),
		icolor.Green("running"),
		formatWsUptime(uptime),
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

func formatWsUptime(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
