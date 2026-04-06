package dukh

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
	icolor "github.com/vhula/grazhda/internal/color"
)

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show workspace health as monitored by dukh",
		RunE:  runStatus,
	}
	cmd.Flags().StringP("name", "n", "", "Workspace name (default: all)")
	cmd.Flags().Bool("rescan", false, "Trigger a fresh workspace rescan before reporting (waits for completion)")
	return cmd
}

func runStatus(cmd *cobra.Command, _ []string) error {
	name, _ := cmd.Flags().GetString("name")
	rescan, _ := cmd.Flags().GetBool("rescan")

	conn, client, err := dial()
	if err != nil {
		fmt.Fprintln(os.Stderr, icolor.Red("✗ "+err.Error()))
		return err
	}
	defer conn.Close()

	ctx := context.Background()
	if rescan {
		// Allow up to 60 s for the scan to complete on large workspaces.
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

	renderStatus(resp)
	return nil
}

func renderStatus(resp *dukhpb.StatusResponse) {
	uptime := time.Duration(resp.UptimeSeconds) * time.Second
	fmt.Printf("%s  %s  •  uptime: %s\n\n",
		icolor.Blue("Dukh"),
		icolor.Green("running"),
		formatUptime(uptime),
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

// formatUptime renders a duration as a human-readable string (e.g. "2h 34m").
func formatUptime(d time.Duration) string {
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
