package ws

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/targeting"
	wsmod "github.com/vhula/grazhda/internal/workspace"
)

func newInitCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var parallel bool
	var noConfirm bool
	var wsName string
	var all bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a workspace by cloning all repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			exec := executor.OsExecutor{}
			rep := reporter.NewReporter(os.Stdout, os.Stderr)
			opts := wsmod.RunOptions{
				DryRun:    dryRun,
				Verbose:   verbose,
				Parallel:  parallel,
				NoConfirm: noConfirm,
			}

			for _, ws := range workspaces {
				if err := wsmod.Init(ws, exec, rep, opts); err != nil {
					return err
				}
			}

			label := "cloned"
			if dryRun {
				label = "would clone"
			}
			rep.Summary(label, dryRun)
			os.Exit(rep.ExitCode())
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Clone repositories concurrently")
	cmd.Flags().BoolVar(&noConfirm, "no-confirm", false, "Skip confirmation prompts")
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Target workspace name (default: default workspace)")
	cmd.Flags().BoolVar(&all, "all", false, "Operate on all workspaces")

	return cmd
}
