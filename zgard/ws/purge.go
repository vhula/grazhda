package ws

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

func newPurgeCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var noConfirm bool

	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Purge a workspace by removing all cloned repositories",
		Long: `Remove workspace directories from disk.

Purge requires an explicit target (--name or --all) and cannot fall back
to the default workspace. A confirmation prompt is shown unless
--no-confirm is passed. Use --dry-run to preview the removal.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// ws purge requires explicit targeting — no implicit default.
			if wsName == "" && !wsAll {
				fmt.Fprintln(os.Stderr, clr.Red("ws purge requires --name <name> or --all"))
				os.Exit(1)
			}

			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			workspaces, err := workspace.Resolve(cfg, wsName, wsAll)
			if err != nil {
				return err
			}

			// Confirmation prompt (skipped in dry-run)
			if !noConfirm && !dryRun {
				paths := make([]string, 0, len(workspaces))
				for _, ws := range workspaces {
					paths = append(paths, filepath.Join(ws.Path))
				}
				if !confirm(os.Stdout, os.Stdin, "The following directories will be removed:", paths) {
					fmt.Println(clr.Yellow("Aborted."))
					return nil
				}
			}

			rep := reporter.NewReporter(os.Stdout, os.Stderr)
			opts := workspace.RunOptions{
				DryRun:    dryRun,
				Verbose:   verbose,
				NoConfirm: noConfirm,
			}

			for _, ws := range workspaces {
				if err := workspace.Purge(ws, rep, opts); err != nil {
					return err
				}
			}

			label := "removed"
			if dryRun {
				label = "would remove"
			}
			rep.Summary(label, dryRun)
			os.Exit(rep.ExitCode())
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&noConfirm, "no-confirm", false, "Skip confirmation prompts")

	return cmd
}
