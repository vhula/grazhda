package ws

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/targeting"
	wsmod "github.com/vhula/grazhda/internal/workspace"
)

func newPurgeCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var noConfirm bool
	var wsName string
	var all bool

	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Purge a workspace by removing all cloned repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			// ws purge requires explicit targeting
			if wsName == "" && !all {
				fmt.Fprintln(os.Stderr, "ws purge requires --ws <name> or --all")
				os.Exit(1)
			}

			cfgPath := resolveConfigPath()
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}
			if errs := config.Validate(cfg); len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintln(os.Stderr, e)
				}
				return fmt.Errorf("configuration is invalid")
			}

			workspaces, err := targeting.Resolve(cfg, wsName, all)
			if err != nil {
				return err
			}

			// Confirmation prompt (skipped in dry-run)
			if !noConfirm && !dryRun {
				paths := make([]string, 0, len(workspaces))
				for _, ws := range workspaces {
					paths = append(paths, filepath.Join(ws.Path))
				}
				if !confirm(os.Stdout, os.Stdin, paths) {
					fmt.Println("Aborted.")
					return nil
				}
			}

			rep := reporter.NewReporter(os.Stdout, os.Stderr)
			opts := wsmod.RunOptions{
				DryRun:    dryRun,
				Verbose:   verbose,
				NoConfirm: noConfirm,
			}

			for _, ws := range workspaces {
				if err := wsmod.Purge(ws, rep, opts); err != nil {
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
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Target workspace name")
	cmd.Flags().BoolVar(&all, "all", false, "Purge all workspaces")

	return cmd
}
