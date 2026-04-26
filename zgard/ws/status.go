package ws

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	dukhclient "github.com/vhula/grazhda/dukh/client"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
	icolor "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/format"
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

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, icolor.Red("✗ "+err.Error()))
		return err
	}

	c, err := dukhclient.Connect(cfg)
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
		fmt.Fprintln(os.Stderr, icolor.Red("✗ dukh status failed: "+err.Error()))
		return err
	}

	renderWsStatus(resp)
	return nil
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
