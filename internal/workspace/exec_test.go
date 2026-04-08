package workspace_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
	"github.com/vhula/grazhda/internal/workspace"
)

// makeMultiProjectWorkspace creates a two-project workspace for filter tests.
func makeMultiProjectWorkspace(t *testing.T) (config.Workspace, string) {
	t.Helper()
	tmp := t.TempDir()
	ws := config.Workspace{
		Name:                 "test-ws",
		Path:                 tmp,
		CloneCommandTemplate: "echo clone {{.RepoName}} {{.DestDir}}",
		Projects: []config.Project{
			{
				Name:   "backend",
				Branch: "main",
				Repositories: []config.Repository{
					{Name: "api"},
					{Name: "auth"},
				},
			},
			{
				Name:   "frontend",
				Branch: "main",
				Repositories: []config.Repository{
					{Name: "web"},
				},
			},
		},
	}
	return ws, tmp
}

// createRepoDirs creates all project/repo directories so operations don't skip.
func createRepoDirs(t *testing.T, ws config.Workspace, basePath string) {
	t.Helper()
	for _, proj := range ws.Projects {
		for _, repo := range proj.Repositories {
			dir := filepath.Join(basePath, proj.Name, repo.Name)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// --- Exec ---

func TestExec_ExecutesCommand(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projPath, "auth"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{CaptureOutput: "output line\n"}

	err := workspace.Exec(ws, "echo hi", mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 exec calls, got %d: %v", len(mock.Calls), mock.Calls)
	}
	for _, call := range mock.Calls {
		if call != "echo hi" {
			t.Errorf("unexpected command %q, expected 'echo hi'", call)
		}
	}
	// Output lines should appear in printed output
	if !strings.Contains(out.String(), "output line") {
		t.Errorf("expected captured output in reporter output, got: %q", out.String())
	}
}

func TestExec_SkipsMissingRepo(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Exec(ws, "make test", mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if len(mock.Calls) != 0 {
		t.Errorf("expected no calls for missing repos, got %d", len(mock.Calls))
	}
	if !strings.Contains(out.String(), "⏭") {
		t.Errorf("expected skip symbol for missing repos")
	}
}

func TestExec_DryRun(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Exec(ws, "make test", mock, rep, workspace.RunOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if len(mock.Calls) != 0 {
		t.Errorf("expected no calls in dry-run, got %d", len(mock.Calls))
	}
	if !strings.Contains(out.String(), "[DRY RUN]") {
		t.Errorf("expected [DRY RUN] in output, got: %q", out.String())
	}
}

func TestExec_ContinueOnFailure(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projPath, "auth"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	callCount := 0
	mock := &executor.MockExecutor{
		ErrFn: func(_ int) error {
			callCount++
			if callCount == 1 {
				return errors.New("command failed")
			}
			return nil
		},
	}

	err := workspace.Exec(ws, "make test", mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Exec should not return error on repo failure: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 calls (continue-on-failure), got %d", len(mock.Calls))
	}
	if rep.ExitCode() == 0 {
		t.Error("expected non-zero exit code due to failure")
	}
}

func TestExec_ProjectFilter(t *testing.T) {
	ws, tmp := makeMultiProjectWorkspace(t)
	createRepoDirs(t, ws, tmp)

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	opts := workspace.RunOptions{ProjectName: "backend"}
	err := workspace.Exec(ws, "echo hi", mock, rep, opts)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	// backend has 2 repos, frontend has 1; only backend should be processed
	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 calls (backend only), got %d: %v", len(mock.Calls), mock.Calls)
	}
	if strings.Contains(out.String(), "frontend") {
		t.Errorf("expected no frontend output when --project-name=backend, got: %q", out.String())
	}
}

func TestExec_RepoFilter(t *testing.T) {
	ws, tmp := makeMultiProjectWorkspace(t)
	createRepoDirs(t, ws, tmp)

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	opts := workspace.RunOptions{ProjectName: "backend", RepoName: "api"}
	err := workspace.Exec(ws, "echo hi", mock, rep, opts)
	if err != nil {
		t.Fatalf("Exec error: %v", err)
	}

	if len(mock.Calls) != 1 {
		t.Errorf("expected 1 call (api only), got %d: %v", len(mock.Calls), mock.Calls)
	}
	if !strings.Contains(out.String(), "api") {
		t.Errorf("expected 'api' in output, got: %q", out.String())
	}
}

func TestExec_Parallel(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projPath, "auth"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Exec(ws, "echo hi", mock, rep, workspace.RunOptions{Parallel: true})
	if err != nil {
		t.Fatalf("Exec parallel error: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 exec calls in parallel mode, got %d", len(mock.Calls))
	}
}

// --- Stash ---

func TestStash_StashesRepos(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projPath, "auth"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Stash(ws, mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Stash error: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 stash calls, got %d: %v", len(mock.Calls), mock.Calls)
	}
	for _, call := range mock.Calls {
		if call != "git stash push" {
			t.Errorf("expected 'git stash push', got %q", call)
		}
	}
	if !strings.Contains(out.String(), "stashed") {
		t.Errorf("expected 'stashed' in output, got: %q", out.String())
	}
}

func TestStash_SkipsMissingRepo(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Stash(ws, mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Stash error: %v", err)
	}

	if len(mock.Calls) != 0 {
		t.Errorf("expected no calls for missing repos, got %d", len(mock.Calls))
	}
	if !strings.Contains(out.String(), "⏭") {
		t.Errorf("expected skip symbol for missing repos")
	}
}

func TestStash_DryRun(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Stash(ws, mock, rep, workspace.RunOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Stash error: %v", err)
	}

	if len(mock.Calls) != 0 {
		t.Errorf("expected no calls in dry-run, got %d", len(mock.Calls))
	}
	if !strings.Contains(out.String(), "[DRY RUN]") {
		t.Errorf("expected [DRY RUN] in output, got: %q", out.String())
	}
}

func TestStash_ContinueOnFailure(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projPath, "auth"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	callCount := 0
	mock := &executor.MockExecutor{
		ErrFn: func(_ int) error {
			callCount++
			if callCount == 1 {
				return errors.New("stash failed")
			}
			return nil
		},
	}

	err := workspace.Stash(ws, mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Stash should not return error on repo failure: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 calls (continue-on-failure), got %d", len(mock.Calls))
	}
	if rep.ExitCode() == 0 {
		t.Error("expected non-zero exit code due to failure")
	}
}

func TestStash_ProjectFilter(t *testing.T) {
	ws, tmp := makeMultiProjectWorkspace(t)
	createRepoDirs(t, ws, tmp)

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	opts := workspace.RunOptions{ProjectName: "frontend"}
	err := workspace.Stash(ws, mock, rep, opts)
	if err != nil {
		t.Fatalf("Stash error: %v", err)
	}

	// frontend has 1 repo (web)
	if len(mock.Calls) != 1 {
		t.Errorf("expected 1 stash call (frontend only), got %d: %v", len(mock.Calls), mock.Calls)
	}
}

// --- Checkout ---

func TestCheckout_ChecksOutBranch(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projPath, "auth"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Checkout(ws, "feature-x", mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Checkout error: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 checkout calls, got %d: %v", len(mock.Calls), mock.Calls)
	}
	for _, call := range mock.Calls {
		if call != "git checkout feature-x" {
			t.Errorf("expected 'git checkout feature-x', got %q", call)
		}
	}
	if !strings.Contains(out.String(), "feature-x") {
		t.Errorf("expected branch name in output, got: %q", out.String())
	}
}

func TestCheckout_SkipsMissingRepo(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Checkout(ws, "main", mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Checkout error: %v", err)
	}

	if len(mock.Calls) != 0 {
		t.Errorf("expected no calls for missing repos, got %d", len(mock.Calls))
	}
	if !strings.Contains(out.String(), "⏭") {
		t.Errorf("expected skip symbol for missing repos")
	}
}

func TestCheckout_DryRun(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Checkout(ws, "main", mock, rep, workspace.RunOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Checkout error: %v", err)
	}

	if len(mock.Calls) != 0 {
		t.Errorf("expected no calls in dry-run, got %d", len(mock.Calls))
	}
	if !strings.Contains(out.String(), "[DRY RUN]") {
		t.Errorf("expected [DRY RUN] in output, got: %q", out.String())
	}
}

func TestCheckout_ContinueOnFailure(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projPath, "auth"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	callCount := 0
	mock := &executor.MockExecutor{
		ErrFn: func(_ int) error {
			callCount++
			if callCount == 1 {
				return errors.New("pathspec 'feature-x' did not match any file")
			}
			return nil
		},
	}

	err := workspace.Checkout(ws, "feature-x", mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Checkout should not return error on repo failure: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 calls (continue-on-failure), got %d", len(mock.Calls))
	}
	if rep.ExitCode() == 0 {
		t.Error("expected non-zero exit code due to failure")
	}
}

func TestCheckout_ProjectFilter(t *testing.T) {
	ws, tmp := makeMultiProjectWorkspace(t)
	createRepoDirs(t, ws, tmp)

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	opts := workspace.RunOptions{ProjectName: "backend"}
	err := workspace.Checkout(ws, "main", mock, rep, opts)
	if err != nil {
		t.Fatalf("Checkout error: %v", err)
	}

	// backend has 2 repos, frontend has 1
	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 calls (backend only), got %d: %v", len(mock.Calls), mock.Calls)
	}
}

func TestCheckout_RepoFilter(t *testing.T) {
	ws, tmp := makeMultiProjectWorkspace(t)
	createRepoDirs(t, ws, tmp)

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	opts := workspace.RunOptions{ProjectName: "backend", RepoName: "auth"}
	err := workspace.Checkout(ws, "main", mock, rep, opts)
	if err != nil {
		t.Fatalf("Checkout error: %v", err)
	}

	if len(mock.Calls) != 1 {
		t.Errorf("expected 1 call (auth only), got %d: %v", len(mock.Calls), mock.Calls)
	}
}
