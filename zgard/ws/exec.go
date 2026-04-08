package ws

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	clr "github.com/vhula/grazhda/internal/color"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

func newExecCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var parallel bool
	var parallelAll bool
	var wsName string
	var all bool
	var projectName string
	var repoName string

	cmd := &cobra.Command{
		Use:   "exec <command> [args...]",
		Short: "Fan out a shell command to all repositories in a workspace",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if repoName != "" && projectName == "" {
				return fmt.Errorf("--repo-name requires --project-name")
			}

			cfgPath := resolveConfigPath()
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}
			if errs := config.Validate(cfg); len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintln(os.Stderr, clr.Red(e))
				}
				return fmt.Errorf("configuration is invalid")
			}

			workspaces, err := workspace.Resolve(cfg, wsName, all)
			if err != nil {
				return err
			}

			command := strings.Join(args, " ")
			exec := executor.OsExecutor{}
			rep := reporter.NewReporter(os.Stdout, os.Stderr)
			opts := workspace.RunOptions{
				DryRun:      dryRun,
				Verbose:     verbose,
				Parallel:    parallel,
				ParallelAll: parallelAll,
				ProjectName: projectName,
				RepoName:    repoName,
			}

			for _, ws := range workspaces {
				if err := workspace.Exec(ws, command, exec, rep, opts); err != nil {
					return err
				}
			}

			label := "executed"
			if dryRun {
				label = "would exec"
			}
			rep.Summary(label, dryRun)
			os.Exit(rep.ExitCode())
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Execute within each project concurrently")
	cmd.Flags().BoolVar(&parallelAll, "parallel-all", false, "Execute across all projects concurrently")
	cmd.Flags().StringVarP(&wsName, "name", "n", "", "Target workspace name (default: default workspace)")
	cmd.Flags().BoolVar(&all, "all", false, "Operate on all workspaces")
	cmd.Flags().StringVar(&projectName, "project-name", "", "Filter to a specific project")
	cmd.Flags().StringVar(&repoName, "repo-name", "", "Filter to a specific repository (requires --project-name)")

	return cmd
}
