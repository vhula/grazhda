package ws

import (
	"context"
	"fmt"
	"os"
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
	cmd.Flags().StringP("name", "n", "", "Workspace name (default: all)")
	cmd.Flags().Bool("rescan", false, "Trigger a fresh workspace rescan before reporting (waits for completion)")
	return cmd
}

func runWsStatus(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
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
		WorkspaceName: name,
		Rescan:        rescan,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, icolor.Red("✗ dukh status failed: "+err.Error()))
		return err
	}

	renderWsStatus(resp)
	return nil
}

// dialDukh opens a gRPC connection to the dukh server.
func dialDukh() (*grpc.ClientConn, dukhpb.DukhServiceClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//nolint:staticcheck // DialContext is the correct API for this grpc version
	conn, err := grpc.DialContext(ctx, dukhAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot reach dukh at %s — is it running? (%w)", dukhAddr, err)
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
