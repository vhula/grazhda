package workspace_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/executor"
	"github.com/vhula/grazhda/internal/reporter"
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
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

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
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

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
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

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
	rep := reporter.NewReporter(&out, &errOut)

	callCount := 0
	mock := &executor.MockExecutor{}
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
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

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
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Init(ws, mock, rep, workspace.RunOptions{Parallel: true})
	if err != nil {
		t.Fatalf("Init parallel error: %v", err)
	}

	if len(mock.Calls) != 2 {
		t.Errorf("expected 2 clone calls in parallel mode, got %d", len(mock.Calls))
	}
}

func TestInit_ParallelAll(t *testing.T) {
	tmp := t.TempDir()
	ws := config.Workspace{
		Name:                 "test-ws",
		Path:                 tmp,
		CloneCommandTemplate: "echo clone {{.RepoName}} {{.DestDir}}",
		Projects: []config.Project{
			{
				Name:   "frontend",
				Branch: "main",
				Repositories: []config.Repository{
					{Name: "web"},
					{Name: "mobile"},
				},
			},
			{
				Name:   "backend",
				Branch: "main",
				Repositories: []config.Repository{
					{Name: "api"},
				},
			},
		},
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Init(ws, mock, rep, workspace.RunOptions{ParallelAll: true})
	if err != nil {
		t.Fatalf("Init parallel-all error: %v", err)
	}
	if len(mock.Calls) != 3 {
		t.Errorf("expected 3 clone calls (all repos across all projects), got %d", len(mock.Calls))
	}
}

func TestInit_CloneDelay(t *testing.T) {
	ws, _ := makeWorkspace(t)
	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)

	started := make([]int64, 0, 2)
	var mu sync.Mutex
	mock := &executor.MockExecutor{
		ErrFn: func(_ int) error {
			mu.Lock()
			started = append(started, time.Now().UnixMilli())
			mu.Unlock()
			return nil
		},
	}

	const delaySeconds = 1
	err := workspace.Init(ws, mock, rep, workspace.RunOptions{CloneDelaySeconds: delaySeconds})
	if err != nil {
		t.Fatalf("Init clone-delay error: %v", err)
	}
	if len(started) != 2 {
		t.Fatalf("expected 2 clone calls, got %d", len(started))
	}
	gapMs := started[1] - started[0]
	if gapMs < int64(delaySeconds*1000) {
		t.Errorf("expected gap >= %dms between clones, got %dms", delaySeconds*1000, gapMs)
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
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

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
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

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
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Pull(ws, mock, rep, workspace.RunOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Pull error: %v", err)
	}

	if len(mock.Calls) != 0 {
		t.Errorf("expected no calls in dry-run")
	}
}

func TestPull_ParallelAll(t *testing.T) {
	tmp := t.TempDir()
	ws := config.Workspace{
		Name:                 "test-ws",
		Path:                 tmp,
		CloneCommandTemplate: "echo clone {{.RepoName}} {{.DestDir}}",
		Projects: []config.Project{
			{
				Name:   "frontend",
				Branch: "main",
				Repositories: []config.Repository{
					{Name: "web"},
					{Name: "mobile"},
				},
			},
			{
				Name:   "backend",
				Branch: "main",
				Repositories: []config.Repository{
					{Name: "api"},
				},
			},
		},
	}

	// Pre-create all repo dirs so pull doesn't skip them.
	for _, proj := range ws.Projects {
		for _, repo := range proj.Repositories {
			dir := filepath.Join(tmp, proj.Name, repo.Name)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	err := workspace.Pull(ws, mock, rep, workspace.RunOptions{ParallelAll: true})
	if err != nil {
		t.Fatalf("Pull parallel-all error: %v", err)
	}
	if len(mock.Calls) != 3 {
		t.Errorf("expected 3 pull calls (all repos across all projects), got %d", len(mock.Calls))
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
	rep := reporter.NewReporter(&out, &errOut)

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
	rep := reporter.NewReporter(&out, &errOut)

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
	rep := reporter.NewReporter(&out, &errOut)

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
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

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

// --- ResolveDestName ---

func TestResolveDestName_TreeMode(t *testing.T) {
	tmp := t.TempDir()
	got := workspace.ResolveDestName(tmp, "org/pack/repo1", "", config.StructureTree)
	if got != "org/pack/repo1" {
		t.Errorf("tree mode: expected %q, got %q", "org/pack/repo1", got)
	}
}

func TestResolveDestName_TreeModeDefault(t *testing.T) {
	// Empty structure string defaults to tree behaviour.
	tmp := t.TempDir()
	got := workspace.ResolveDestName(tmp, "org/pack/repo1", "", "")
	if got != "org/pack/repo1" {
		t.Errorf("default (empty) structure: expected full name, got %q", got)
	}
}

func TestResolveDestName_ListMode_LastSegment(t *testing.T) {
	tmp := t.TempDir()
	got := workspace.ResolveDestName(tmp, "org/pack/repo1", "", config.StructureList)
	if got != "repo1" {
		t.Errorf("list mode: expected last segment %q, got %q", "repo1", got)
	}
}

// TestResolveDestName_ListMode_AlwaysLastSegmentEvenIfExists verifies that
// list mode always returns the last segment regardless of what is on disk.
// Conflict resolution (skipping) is delegated to the caller (cloneRepo/pullRepo).
func TestResolveDestName_ListMode_AlwaysLastSegmentEvenIfExists(t *testing.T) {
	tmp := t.TempDir()
	// Pre-create "repo1" — the resolver must still return "repo1", not fall back.
	if err := os.MkdirAll(filepath.Join(tmp, "repo1"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := workspace.ResolveDestName(tmp, "org/pack/repo1", "", config.StructureList)
	if got != "repo1" {
		t.Errorf("list mode (dir exists): expected %q (caller skips), got %q", "repo1", got)
	}
}

func TestResolveDestName_ListMode_GitSuffixStripped(t *testing.T) {
	tmp := t.TempDir()
	got := workspace.ResolveDestName(tmp, "org/pack/repo1.git", "", config.StructureList)
	if got != "repo1" {
		t.Errorf("list mode .git strip: expected %q, got %q", "repo1", got)
	}
}

func TestResolveDestName_ListMode_GitSuffixStripped_NoSlash(t *testing.T) {
	tmp := t.TempDir()
	got := workspace.ResolveDestName(tmp, "myrepo.git", "", config.StructureList)
	if got != "myrepo" {
		t.Errorf("list mode .git strip (no slash): expected %q, got %q", "myrepo", got)
	}
}

func TestResolveDestName_LocalDirNameOverridesStructure(t *testing.T) {
	tmp := t.TempDir()
	// localDirName always wins regardless of structure.
	got := workspace.ResolveDestName(tmp, "org/pack/repo1", "myrepo", config.StructureList)
	if got != "myrepo" {
		t.Errorf("localDirName override: expected %q, got %q", "myrepo", got)
	}
}

func TestResolveDestName_ListMode_PlainName(t *testing.T) {
	// Repo name with no slashes behaves the same in both modes.
	tmp := t.TempDir()
	got := workspace.ResolveDestName(tmp, "repo1", "", config.StructureList)
	if got != "repo1" {
		t.Errorf("list mode plain name: expected %q, got %q", "repo1", got)
	}
}

// TestInit_ListStructure verifies that Init uses only the last path segment
// as the clone destination when structure == "list".
func TestInit_ListStructure(t *testing.T) {
	tmp := t.TempDir()
	ws := config.Workspace{
		Name:                 "test-ws",
		Path:                 tmp,
		Structure:            config.StructureList,
		CloneCommandTemplate: "echo clone {{.RepoName}} {{.DestDir}}",
		Projects: []config.Project{
			{
				Name:   "backend",
				Branch: "main",
				Repositories: []config.Repository{
					{Name: "org/pack/repo1"},
				},
			},
		},
	}

	var out, errOut strings.Builder
	rep := reporter.NewReporter(&out, &errOut)
	mock := &executor.MockExecutor{}

	if err := workspace.Init(ws, mock, rep, workspace.RunOptions{}); err != nil {
		t.Fatalf("Init error: %v", err)
	}

	// The clone command's DestDir should end in "repo1", not "org/pack/repo1".
	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.Calls))
	}
	call := mock.Calls[0]
	if strings.HasSuffix(call, filepath.Join("org", "pack", "repo1")) {
		t.Errorf("list mode: DestDir should not include full path, got %q", call)
	}
	if !strings.HasSuffix(call, "repo1") {
		t.Errorf("list mode: expected DestDir to end in 'repo1', got %q", call)
	}
}

// ---------------------------------------------------------------------------
// ResolveDestNamesForProject tests
// ---------------------------------------------------------------------------

func TestResolveDestNamesForProject_TreeMode(t *testing.T) {
repos := []config.Repository{
{Name: "org/pack/repo1"},
{Name: "org/pack/repo2"},
}
got := workspace.ResolveDestNamesForProject(repos, config.StructureTree)
want := []string{"org/pack/repo1", "org/pack/repo2"}
for i, w := range want {
if got[i] != w {
t.Errorf("[%d] tree mode: want %q got %q", i, w, got[i])
}
}
}

func TestResolveDestNamesForProject_ListMode_NoCollision(t *testing.T) {
repos := []config.Repository{
{Name: "org/pack/repo1"},
{Name: "other/pack/repo2"},
}
got := workspace.ResolveDestNamesForProject(repos, config.StructureList)
want := []string{"repo1", "repo2"}
for i, w := range want {
if got[i] != w {
t.Errorf("[%d] list no-collision: want %q got %q", i, w, got[i])
}
}
}

func TestResolveDestNamesForProject_ListMode_Collision(t *testing.T) {
	// Two repos with the same last segment both resolve to that segment.
	// The second clone will be skipped as "already exists" by cloneRepo.
	repos := []config.Repository{
		{Name: "org/pack/repo1"},
		{Name: "other-org/pack/repo1"},
	}
	got := workspace.ResolveDestNamesForProject(repos, config.StructureList)
	if got[0] != "repo1" {
		t.Errorf("[0] want %q got %q", "repo1", got[0])
	}
	// Both resolve to the same last segment; caller handles skip.
	if got[1] != "repo1" {
		t.Errorf("[1] want %q got %q", "repo1", got[1])
	}
}

func TestResolveDestNamesForProject_ListMode_AllReturnLastSegment(t *testing.T) {
	repos := []config.Repository{
		{Name: "a/b/repo"},
		{Name: "c/b/repo"},
		{Name: "d/b/different"},
	}
	got := workspace.ResolveDestNamesForProject(repos, config.StructureList)
	want := []string{"repo", "repo", "different"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("[%d] want %q got %q", i, w, got[i])
		}
	}
}

func TestResolveDestNamesForProject_LocalDirNameOverridesStructure(t *testing.T) {
repos := []config.Repository{
{Name: "org/pack/repo1", LocalDirName: "my-repo"},
}
got := workspace.ResolveDestNamesForProject(repos, config.StructureList)
if got[0] != "my-repo" {
t.Errorf("localDirName override: want %q got %q", "my-repo", got[0])
}
}

func TestResolveDestNamesForProject_LocalDirName_AndListSegment(t *testing.T) {
	// local_dir_name always wins; following repo resolves to its own last segment.
	repos := []config.Repository{
		{Name: "org/pack/repo1", LocalDirName: "custom"},
		{Name: "other/pack/repo1"},
	}
	got := workspace.ResolveDestNamesForProject(repos, config.StructureList)
	if got[0] != "custom" {
		t.Errorf("[0] want %q got %q", "custom", got[0])
	}
	if got[1] != "repo1" {
		t.Errorf("[1] want %q got %q", "repo1", got[1])
	}
}

// --- Init / Pull project+repo filtering ---

func TestInit_ProjectFilter(t *testing.T) {
	ws, _ := makeMultiProjectWorkspace(t)

var out, errOut strings.Builder
rep := reporter.NewReporter(&out, &errOut)
mock := &executor.MockExecutor{}

opts := workspace.RunOptions{ProjectName: "frontend"}
if err := workspace.Init(ws, mock, rep, opts); err != nil {
t.Fatalf("Init error: %v", err)
}

// frontend has 1 repo (web)
if len(mock.Calls) != 1 {
t.Errorf("expected 1 clone call (frontend only), got %d: %v", len(mock.Calls), mock.Calls)
}
if !strings.Contains(out.String(), "frontend") {
t.Error("expected frontend to appear in output")
}
if strings.Contains(out.String(), "backend") {
t.Error("backend should not appear in output")
}
}

func TestInit_RepoFilter(t *testing.T) {
ws, tmp := makeMultiProjectWorkspace(t)
_ = tmp

var out, errOut strings.Builder
rep := reporter.NewReporter(&out, &errOut)
mock := &executor.MockExecutor{}

opts := workspace.RunOptions{ProjectName: "backend", RepoName: "api"}
if err := workspace.Init(ws, mock, rep, opts); err != nil {
t.Fatalf("Init error: %v", err)
}

// backend has 2 repos but only api is targeted
if len(mock.Calls) != 1 {
t.Errorf("expected 1 clone call (api only), got %d: %v", len(mock.Calls), mock.Calls)
}
}

func TestInit_UnknownProject_ReturnsError(t *testing.T) {
ws, _ := makeMultiProjectWorkspace(t)
var out, errOut strings.Builder
rep := reporter.NewReporter(&out, &errOut)
mock := &executor.MockExecutor{}

err := workspace.Init(ws, mock, rep, workspace.RunOptions{ProjectName: "nonexistent"})
if err == nil {
t.Fatal("expected error for unknown project")
}
if len(mock.Calls) != 0 {
t.Error("no commands should have run")
}
}

func TestPull_ProjectFilter(t *testing.T) {
ws, tmp := makeMultiProjectWorkspace(t)
createRepoDirs(t, ws, tmp)

var out, errOut strings.Builder
rep := reporter.NewReporter(&out, &errOut)
mock := &executor.MockExecutor{}

opts := workspace.RunOptions{ProjectName: "frontend"}
if err := workspace.Pull(ws, mock, rep, opts); err != nil {
t.Fatalf("Pull error: %v", err)
}

// frontend has 1 repo (web)
if len(mock.Calls) != 1 {
t.Errorf("expected 1 pull call (frontend only), got %d: %v", len(mock.Calls), mock.Calls)
}
}

func TestPull_RepoFilter(t *testing.T) {
ws, tmp := makeMultiProjectWorkspace(t)
createRepoDirs(t, ws, tmp)

var out, errOut strings.Builder
rep := reporter.NewReporter(&out, &errOut)
mock := &executor.MockExecutor{}

opts := workspace.RunOptions{ProjectName: "backend", RepoName: "auth"}
if err := workspace.Pull(ws, mock, rep, opts); err != nil {
t.Fatalf("Pull error: %v", err)
}

if len(mock.Calls) != 1 {
t.Errorf("expected 1 pull call (auth only), got %d: %v", len(mock.Calls), mock.Calls)
}
}

func TestPull_UnknownProject_ReturnsError(t *testing.T) {
ws, _ := makeMultiProjectWorkspace(t)
var out, errOut strings.Builder
rep := reporter.NewReporter(&out, &errOut)
mock := &executor.MockExecutor{}

err := workspace.Pull(ws, mock, rep, workspace.RunOptions{ProjectName: "nonexistent"})
if err == nil {
t.Fatal("expected error for unknown project")
}
if len(mock.Calls) != 0 {
t.Error("no commands should have run")
}
}
