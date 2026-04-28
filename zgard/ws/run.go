package ws

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

// wsOpFunc is the callback invoked once per workspace. The helper creates the
// executor and reporter so the callback can focus on the actual operation.
type wsOpFunc func(ws config.Workspace, exec executor.Executor, rep *reporter.Reporter, opts workspace.RunOptions) error

// runWorkspaceOp loads the config, resolves workspaces, sets up an executor
// and reporter with the standard flag-driven settings, then invokes op for
// each workspace. On completion it prints a summary and returns an ExitError
// when any repo failed.
func runWorkspaceOp(cmd *cobra.Command, opts workspace.RunOptions, doneLabel, dryLabel string, op wsOpFunc) error {
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

	opts.Context = cmd.Context()

	exec := executor.OsExecutor{}
	rep := reporter.NewReporter(os.Stdout, os.Stderr)
	rep.ShowElapsed = opts.Verbose
	rep.Quiet = rootFlag(cmd, "quiet")
	if opts.DryRun {
		rep.PrintDryRunBanner()
	}

	for _, ws := range workspaces {
		if err := op(ws, exec, rep, opts); err != nil {
			return err
		}
	}

	label := doneLabel
	if opts.DryRun {
		label = dryLabel
	}
	rep.Summary(label, opts.DryRun)
	if code := rep.ExitCode(); code != 0 {
		return reporter.ExitError{Code: code}
	}
	return nil
}
