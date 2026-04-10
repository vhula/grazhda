package ws

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

func newInitCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	var parallel bool
	var cloneDelaySeconds int
	var noConfirm bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a workspace by cloning all repositories",
		Long: `Clone every repository listed in the workspace configuration.

Project directories are created on disk before cloning begins. Repositories
that already exist are **skipped** without error — re-running init is safe.

Use **--parallel** to clone concurrently (recommended for large workspaces),
and **--dry-run** to preview actions without touching the filesystem.`,
		Example: `  # Initialize the default workspace
  zgard ws init

  # Initialize a specific named workspace
  zgard ws init -n myworkspace

  # Clone all workspaces concurrently
  zgard ws init --all --parallel

  # Preview what would be cloned (no changes made)
  zgard ws init --all --dry-run

  # Only clone repositories in a specific project
  zgard ws init -n myworkspace -p backend

  # Only clone a single repository
  zgard ws init -n myworkspace -p backend -r api-service

  # Add a delay between clones (useful for rate-limited hosts)
  zgard ws init --all --clone-delay-seconds 2

  # CI-friendly: parallel, no confirmation prompts
  zgard ws init --all --parallel --no-confirm`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			workspaces, err := workspace.Resolve(cfg, wsName, wsAll)
			if err != nil {
				return err
			}

			if wsName == "" && !wsAll {
				warnDefaultTarget(os.Stderr, workspaces[0])
			}

			exec := executor.OsExecutor{}
			rep := reporter.NewReporter(os.Stdout, os.Stderr)
			rep.ShowElapsed = verbose
			rep.JSONMode = rootFlag(cmd, "json")
			rep.Quiet = rootFlag(cmd, "quiet")
			if dryRun {
				rep.PrintDryRunBanner()
			}
			opts := workspace.RunOptions{
				Context:           cmd.Context(),
				DryRun:            dryRun,
				Verbose:           verbose,
				Parallel:          parallel,
				NoConfirm:         noConfirm,
				CloneDelaySeconds: cloneDelaySeconds,
				ProjectName:       projectName,
				RepoName:          repoName,
				Tags:              tagFilter,
			}

			for _, ws := range workspaces {
				if err := workspace.Init(ws, exec, rep, opts); err != nil {
					return err
				}
			}

			label := "cloned"
			if dryRun {
				label = "would clone"
			}
			rep.Summary(label, dryRun)
			if code := rep.ExitCode(); code != 0 {
				return reporter.ExitError{Code: code}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions without executing them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.Flags().BoolVar(&parallel, "parallel", false, "Clone all repositories concurrently")
	cmd.Flags().IntVar(&cloneDelaySeconds, "clone-delay-seconds", 0, "Seconds to sleep after each clone command (0 = disabled)")
	cmd.Flags().BoolVar(&noConfirm, "no-confirm", false, "Skip confirmation prompts")
	cmd.Flags().BoolVarP(&noConfirm, "yes", "y", false, "Skip confirmation prompts (alias for --no-confirm)")

	return cmd
}
