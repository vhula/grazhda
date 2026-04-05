package workspace_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/workspace"
)

func TestRecord_Success(t *testing.T) {
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)

	rep.Record(workspace.OpResult{Repo: "api", Msg: "cloned (main)"})

	if !strings.Contains(out.String(), "✓") {
		t.Errorf("expected success symbol, got: %q", out.String())
	}
	if !strings.Contains(out.String(), "api") {
		t.Errorf("expected repo name, got: %q", out.String())
	}
	if !strings.Contains(out.String(), "cloned (main)") {
		t.Errorf("expected message, got: %q", out.String())
	}
}

func TestRecord_Skipped(t *testing.T) {
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)

	rep.Record(workspace.OpResult{Repo: "api", Skipped: true, Msg: "already exists, skipped"})

	if !strings.Contains(out.String(), "⏭") {
		t.Errorf("expected skip symbol, got: %q", out.String())
	}
}

func TestRecord_Failed(t *testing.T) {
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)

	rep.Record(workspace.OpResult{Repo: "api", Err: errors.New("exit status 128")})

	if !strings.Contains(out.String(), "✗") {
		t.Errorf("expected failure symbol, got: %q", out.String())
	}
	if !strings.Contains(out.String(), "exit status 128") {
		t.Errorf("expected error message in output, got: %q", out.String())
	}
}

func TestSummary_Counts(t *testing.T) {
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)

	rep.Record(workspace.OpResult{Repo: "api", Msg: "cloned (main)"})
	rep.Record(workspace.OpResult{Repo: "auth", Skipped: true, Msg: "already exists, skipped"})
	rep.Record(workspace.OpResult{Repo: "svc", Err: errors.New("clone failed")})

	rep.Summary("cloned", false)

	sum := out.String()
	if !strings.Contains(sum, "✓ 1 cloned") {
		t.Errorf("expected success count, got: %q", sum)
	}
	if !strings.Contains(sum, "⏭ 1 skipped") {
		t.Errorf("expected skip count, got: %q", sum)
	}
	if !strings.Contains(sum, "✗ 1 failed") {
		t.Errorf("expected failure count, got: %q", sum)
	}
}

func TestSummary_FailureDetails_ToStderr(t *testing.T) {
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)

	rep.Record(workspace.OpResult{Repo: "svc", Err: errors.New("exit 128: repository not found")})
	rep.Summary("cloned", false)

	if !strings.Contains(errOut.String(), "svc") {
		t.Errorf("expected failure detail on stderr, got: %q", errOut.String())
	}
	if !strings.Contains(errOut.String(), "exit 128") {
		t.Errorf("expected error text on stderr, got: %q", errOut.String())
	}
}

func TestSummary_DryRun(t *testing.T) {
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)

	rep.Record(workspace.OpResult{Repo: "api", Msg: "[DRY RUN] would clone (main)"})
	rep.Summary("would clone", true)

	if !strings.Contains(out.String(), "[DRY RUN]") {
		t.Errorf("expected DRY RUN prefix in summary, got: %q", out.String())
	}
}

func TestExitCode_AllSuccess(t *testing.T) {
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	rep.Record(workspace.OpResult{Repo: "api", Msg: "cloned"})
	if rep.ExitCode() != 0 {
		t.Errorf("expected exit code 0")
	}
}

func TestExitCode_WithFailure(t *testing.T) {
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	rep.Record(workspace.OpResult{Repo: "api", Err: errors.New("failed")})
	if rep.ExitCode() != 1 {
		t.Errorf("expected exit code 1")
	}
}

func TestPrintLine(t *testing.T) {
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	rep.PrintLine("Workspace: default")
	rep.PrintLine("  Project: backend")
	if !strings.Contains(out.String(), "Workspace: default") {
		t.Errorf("expected header line, got: %q", out.String())
	}
}
