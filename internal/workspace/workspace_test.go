package workspace_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/workspace"
)

func makeWorkspace(t *testing.T) (config.Workspace, string) {
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
					{Name: "auth", Branch: "dev"},
				},
			},
		},
	}
	return ws, tmp
}

// --- Init ---

func TestInit_ClonesRepos(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	mock := &workspace.MockExecutor{}

	err := workspace.Init(ws, mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 clone calls, got %d: %v", len(mock.Calls), mock.Calls)
	}
	if !strings.Contains(mock.Calls[0], "api") {
		t.Errorf("expected first call for 'api', got %q", mock.Calls[0])
	}
	if !strings.Contains(mock.Calls[1], "auth") {
		t.Errorf("expected second call for 'auth', got %q", mock.Calls[1])
	}
}

func TestInit_SkipsExistingRepo(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	mock := &workspace.MockExecutor{}

	err := workspace.Init(ws, mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}

	// Only auth should be cloned; api already exists
	if len(mock.Calls) != 1 {
		t.Errorf("expected 1 clone call (api skipped), got %d", len(mock.Calls))
	}
	if !strings.Contains(out.String(), "⏭") {
		t.Errorf("expected skip symbol for existing repo")
	}
}

func TestInit_DryRun(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	mock := &workspace.MockExecutor{}

	err := workspace.Init(ws, mock, rep, workspace.RunOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}

	if len(mock.Calls) != 0 {
		t.Errorf("expected no calls in dry-run, got %d", len(mock.Calls))
	}
	if !strings.Contains(out.String(), "[DRY RUN]") {
		t.Errorf("expected DRY RUN in output, got: %q", out.String())
	}
}

func TestInit_ContinueOnFailure(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)

	callCount := 0
	mock := &workspace.MockExecutor{}
	// First call fails, second succeeds
	mock.ErrFn = func(call int) error {
		callCount++
		if callCount == 1 {
			return errors.New("clone failed")
		}
		return nil
	}

	err := workspace.Init(ws, mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Init should not return error on repo failure: %v", err)
	}

	// Both repos attempted
	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 calls (continue-on-failure), got %d", len(mock.Calls))
	}
	if rep.ExitCode() == 0 {
		t.Errorf("expected non-zero exit code due to failure")
	}
}

func TestInit_VerboseFlag(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	mock := &workspace.MockExecutor{}

	err := workspace.Init(ws, mock, rep, workspace.RunOptions{Verbose: true})
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}

	if !strings.Contains(out.String(), "→") {
		t.Errorf("expected verbose command arrow in output, got: %q", out.String())
	}
}

func TestInit_Parallel(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	mock := &workspace.MockExecutor{}

	err := workspace.Init(ws, mock, rep, workspace.RunOptions{Parallel: true})
	if err != nil {
		t.Fatalf("Init parallel error: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 clone calls in parallel mode, got %d", len(mock.Calls))
	}
}

// --- Pull ---

func TestPull_PullsExistingRepos(t *testing.T) {
	ws, tmp := makeWorkspace(t)
	// Pre-create repo dirs so pull doesn't skip
	projPath := filepath.Join(tmp, "backend")
	if err := os.MkdirAll(filepath.Join(projPath, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projPath, "auth"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	mock := &workspace.MockExecutor{}

	err := workspace.Pull(ws, mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Pull error: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 pull calls, got %d: %v", len(mock.Calls), mock.Calls)
	}
	if !strings.Contains(mock.Calls[0], "git pull --rebase origin main") {
		t.Errorf("expected pull command for main branch, got %q", mock.Calls[0])
	}
	if !strings.Contains(mock.Calls[1], "git pull --rebase origin dev") {
		t.Errorf("expected pull command for dev branch, got %q", mock.Calls[1])
	}
}

func TestPull_SkipsMissingRepo(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	mock := &workspace.MockExecutor{}

	err := workspace.Pull(ws, mock, rep, workspace.RunOptions{})
	if err != nil {
		t.Fatalf("Pull error: %v", err)
	}

	if len(mock.Calls) != 0 {
		t.Errorf("expected no pull calls for missing repos, got %d", len(mock.Calls))
	}
	if !strings.Contains(out.String(), "⏭") {
		t.Errorf("expected skip symbols for missing repos")
	}
}

func TestPull_DryRun(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	mock := &workspace.MockExecutor{}

	err := workspace.Pull(ws, mock, rep, workspace.RunOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Pull error: %v", err)
	}

	if len(mock.Calls) != 0 {
		t.Errorf("expected no calls in dry-run")
	}
}

// --- Purge ---

func TestPurge_RemovesDirectory(t *testing.T) {
	tmp := t.TempDir()
	wsDir := filepath.Join(tmp, "myws")
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	ws := config.Workspace{Name: "myws", Path: wsDir}
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)

	err := workspace.Purge(ws, rep, workspace.RunOptions{NoConfirm: true})
	if err != nil {
		t.Fatalf("Purge error: %v", err)
	}

	if _, err := os.Stat(wsDir); !os.IsNotExist(err) {
		t.Error("expected directory to be removed")
	}
	if !strings.Contains(out.String(), "✓") {
		t.Errorf("expected success symbol, got: %q", out.String())
	}
}

func TestPurge_SkipsMissingDirectory(t *testing.T) {
	ws := config.Workspace{Name: "ghost", Path: "/nonexistent/path/ghost"}
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)

	err := workspace.Purge(ws, rep, workspace.RunOptions{NoConfirm: true})
	if err != nil {
		t.Fatalf("Purge error: %v", err)
	}

	if !strings.Contains(out.String(), "⏭") {
		t.Errorf("expected skip symbol for missing dir, got: %q", out.String())
	}
	if rep.ExitCode() != 0 {
		t.Error("expected exit code 0 for missing dir")
	}
}

func TestPurge_DryRun(t *testing.T) {
	tmp := t.TempDir()
	wsDir := filepath.Join(tmp, "myws")
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	ws := config.Workspace{Name: "myws", Path: wsDir}
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)

	err := workspace.Purge(ws, rep, workspace.RunOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Purge error: %v", err)
	}

	if _, err := os.Stat(wsDir); os.IsNotExist(err) {
		t.Error("expected directory to still exist after dry-run")
	}
	if !strings.Contains(out.String(), "[DRY RUN]") {
		t.Errorf("expected DRY RUN in output, got: %q", out.String())
	}
}

// Verify summary label rendering
func TestInit_SummaryLabels(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := workspace.NewReporter(&out, &errOut)
	mock := &workspace.MockExecutor{}

	workspace.Init(ws, mock, rep, workspace.RunOptions{}) //nolint:errcheck
	rep.Summary("cloned", false)
	if !strings.Contains(out.String(), "cloned") {
		t.Errorf("expected 'cloned' label in summary, got: %q", out.String())
	}
}

// helper: formatted call count
func callCount(calls []string, substr string) int {
	n := 0
	for _, c := range calls {
		if strings.Contains(c, substr) {
			n++
		}
	}
	return n
}

var _ = fmt.Sprintf // suppress unused import
