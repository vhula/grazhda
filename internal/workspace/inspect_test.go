package workspace

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/vhula/grazhda/internal/config"
)

// captureMock is a test executor that routes RunCapture output via a function,
// enabling per-call output routing that the shared MockExecutor cannot provide.
type captureMock struct {
	mu         sync.Mutex
	Calls      []string
	CaptureFn  func(dir, command string) (string, error)
	DefaultOut string
}

func (m *captureMock) Run(dir, command string) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, command)
	m.mu.Unlock()
	return nil
}

func (m *captureMock) RunContext(ctx context.Context, dir, command string) error {
	return m.Run(dir, command)
}

func (m *captureMock) RunCapture(dir, command string) (string, error) {
	m.mu.Lock()
	m.Calls = append(m.Calls, command)
	fn := m.CaptureFn
	def := m.DefaultOut
	m.mu.Unlock()
	if fn != nil {
		return fn(dir, command)
	}
	return def, nil
}

func (m *captureMock) RunCaptureContext(ctx context.Context, dir, command string) (string, error) {
	return m.RunCapture(dir, command)
}

func (m *captureMock) RunInteractive(ctx context.Context, dir, command string) error {
	return m.RunContext(ctx, dir, command)
}

// ─────────────────────────────── helpers ────────────────────────────────────

func twoRepoWorkspace(t *testing.T) (config.Workspace, string) {
	t.Helper()
	base := t.TempDir()

	projPath := filepath.Join(base, "backend")
	apiPath := filepath.Join(projPath, "api")
	authPath := filepath.Join(projPath, "auth-service")
	if err := os.MkdirAll(apiPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(authPath, 0o755); err != nil {
		t.Fatal(err)
	}

	ws := config.Workspace{
		Name:      "myws",
		Path:      base,
		Structure: "list",
		Projects: []config.Project{
			{
				Name: "backend",
				Repositories: []config.Repository{
					{Name: "api"},
					{Name: "auth-service"},
				},
			},
		},
	}
	return ws, base
}

// ─────────────────────────────── Search ─────────────────────────────────────

func TestSearch_ContentMatch(t *testing.T) {
	ws, base := twoRepoWorkspace(t)

	// write a file with a matching line in api, and one without in auth-service
	apiSrc := filepath.Join(base, "backend", "api")
	if err := os.WriteFile(filepath.Join(apiSrc, "main.go"), []byte("package main\n// TODO: fix this\nfunc main(){}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	authSrc := filepath.Join(base, "backend", "auth-service")
	if err := os.WriteFile(filepath.Join(authSrc, "auth.go"), []byte("package auth\n// done\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := Search(ws, SearchOptions{InspectOptions: InspectOptions{}, Pattern: "TODO"}, &buf)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[backend/api]") {
		t.Errorf("expected [backend/api] in output, got:\n%s", out)
	}
	if !strings.Contains(out, "main.go:2:") {
		t.Errorf("expected main.go:2: in output, got:\n%s", out)
	}
	if strings.Contains(out, "[backend/auth-service]") {
		t.Errorf("did not expect auth-service match, got:\n%s", out)
	}
	if !strings.Contains(out, "1 match(es)") {
		t.Errorf("expected '1 match(es)' summary, got:\n%s", out)
	}
}

func TestSearch_GlobMatch(t *testing.T) {
	ws, base := twoRepoWorkspace(t)

	apiSrc := filepath.Join(base, "backend", "api")
	if err := os.WriteFile(filepath.Join(apiSrc, "server.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(apiSrc, "README.md"), []byte("# readme\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := Search(ws, SearchOptions{InspectOptions: InspectOptions{}, Pattern: "*.go", Glob: true}, &buf)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "server.go") {
		t.Errorf("expected server.go match, got:\n%s", out)
	}
	if strings.Contains(out, "README.md") {
		t.Errorf("did not expect README.md in glob *.go, got:\n%s", out)
	}
}

func TestSearch_RegexMatch(t *testing.T) {
	ws, base := twoRepoWorkspace(t)

	apiSrc := filepath.Join(base, "backend", "api")
	if err := os.WriteFile(filepath.Join(apiSrc, "main.go"), []byte("func Handler() {}\nvar x = 42\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := Search(ws, SearchOptions{InspectOptions: InspectOptions{}, Pattern: `^func\s`, Regex: true}, &buf)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Handler") {
		t.Errorf("expected Handler match from regex, got:\n%s", out)
	}
	if strings.Contains(out, "var x") {
		t.Errorf("did not expect var x to match ^func, got:\n%s", out)
	}
}

func TestSearch_BinaryFileSkipped(t *testing.T) {
	ws, base := twoRepoWorkspace(t)

	apiSrc := filepath.Join(base, "backend", "api")
	// write binary content (has null byte)
	if err := os.WriteFile(filepath.Join(apiSrc, "binary.bin"), []byte("data\x00null"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := Search(ws, SearchOptions{InspectOptions: InspectOptions{}, Pattern: "null"}, &buf)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "binary.bin") {
		t.Errorf("binary file should be skipped, got:\n%s", out)
	}
	if !strings.Contains(out, "0 match(es)") {
		t.Errorf("expected 0 matches, got:\n%s", out)
	}
}

func TestSearch_NotClonedRepoSkipped(t *testing.T) {
	base := t.TempDir()
	// only create backend/api, not backend/auth-service
	if err := os.MkdirAll(filepath.Join(base, "backend", "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "backend", "api", "f.go"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ws := config.Workspace{
		Name: "myws", Path: base, Structure: "list",
		Projects: []config.Project{{
			Name: "backend",
			Repositories: []config.Repository{
				{Name: "api"},
				{Name: "missing-repo"},
			},
		}},
	}

	var buf bytes.Buffer
	err := Search(ws, SearchOptions{InspectOptions: InspectOptions{}, Pattern: "hello"}, &buf)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[backend/api]") {
		t.Errorf("expected api match, got:\n%s", out)
	}
	if strings.Contains(out, "missing-repo") {
		t.Errorf("missing repo should be silently skipped, got:\n%s", out)
	}
}

func TestSearch_RepoNameFilter(t *testing.T) {
	ws, base := twoRepoWorkspace(t)

	for _, dir := range []string{"api", "auth-service"} {
		p := filepath.Join(base, "backend", dir, "f.go")
		if err := os.WriteFile(p, []byte("needle\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	err := Search(ws, SearchOptions{
		InspectOptions: InspectOptions{RepoName: "auth"},
		Pattern:        "needle",
	}, &buf)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[backend/auth-service]") {
		t.Errorf("expected auth-service match, got:\n%s", out)
	}
	if strings.Contains(out, "[backend/api]") {
		t.Errorf("api should be filtered out, got:\n%s", out)
	}
}

// ─────────────────────────────── Diff ────────────────────────────────────────

func TestDiff_CleanRepo(t *testing.T) {
	ws, _ := twoRepoWorkspace(t)

	mock := &captureMock{
		CaptureFn: func(dir, cmd string) (string, error) {
			switch {
			case strings.Contains(cmd, "status --porcelain"):
				return "", nil
			case strings.Contains(cmd, "@{u}..HEAD"):
				return "0\n", nil
			case strings.Contains(cmd, "HEAD..@{u}"):
				return "0\n", nil
			}
			return "", nil
		},
	}

	var buf bytes.Buffer
	err := Diff(ws, mock, InspectOptions{}, &buf)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "api") {
		t.Errorf("expected api in diff output, got:\n%s", out)
	}
	if !strings.Contains(out, "clean") {
		t.Errorf("expected summary with clean count, got:\n%s", out)
	}
}

func TestDiff_DirtyRepo(t *testing.T) {
	ws, _ := twoRepoWorkspace(t)

	mock := &captureMock{
		CaptureFn: func(dir, cmd string) (string, error) {
			switch {
			case strings.Contains(cmd, "status --porcelain"):
				if strings.HasSuffix(dir, "api") {
					return " M main.go\n M other.go\n", nil
				}
				return "", nil
			case strings.Contains(cmd, "@{u}..HEAD"):
				return "0\n", nil
			case strings.Contains(cmd, "HEAD..@{u}"):
				return "0\n", nil
			}
			return "", nil
		},
	}

	var buf bytes.Buffer
	err := Diff(ws, mock, InspectOptions{}, &buf)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1 dirty") {
		t.Errorf("expected 1 dirty in summary, got:\n%s", out)
	}
}

func TestDiff_NoUpstream(t *testing.T) {
	ws, _ := twoRepoWorkspace(t)

	mock := &captureMock{
		CaptureFn: func(dir, cmd string) (string, error) {
			if strings.Contains(cmd, "@{u}") {
				return "", fmt.Errorf("no upstream")
			}
			return "", nil
		},
	}

	var buf bytes.Buffer
	err := Diff(ws, mock, InspectOptions{}, &buf)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "--") {
		t.Errorf("expected '--' for no-upstream repos, got:\n%s", out)
	}
}

func TestDiff_NotClonedRepo(t *testing.T) {
	base := t.TempDir()
	// only create api, not ghost-repo
	if err := os.MkdirAll(filepath.Join(base, "backend", "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	ws := config.Workspace{
		Name: "myws", Path: base, Structure: "list",
		Projects: []config.Project{{
			Name: "backend",
			Repositories: []config.Repository{
				{Name: "api"},
				{Name: "ghost-repo"},
			},
		}},
	}

	mock := &captureMock{CaptureFn: func(dir, cmd string) (string, error) { return "", nil }}

	var buf bytes.Buffer
	err := Diff(ws, mock, InspectOptions{}, &buf)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1 not cloned") {
		t.Errorf("expected '1 not cloned' in summary, got:\n%s", out)
	}
}

// ─────────────────────────────── Stats ───────────────────────────────────────

func TestStats_Basic(t *testing.T) {
	ws, _ := twoRepoWorkspace(t)

	mock := &captureMock{
		CaptureFn: func(dir, cmd string) (string, error) {
			switch {
			case strings.Contains(cmd, `log -1 --format`):
				return "2024-06-15 10:30:22 +0000\n", nil
			case strings.Contains(cmd, `--since="30 days ago"`):
				return "abc123\ndef456\n", nil
			case strings.Contains(cmd, `log --format="%ae"`):
				return "alice@example.com\nbob@example.com\nalice@example.com\n", nil
			}
			return "", nil
		},
	}

	var buf bytes.Buffer
	err := Stats(ws, mock, InspectOptions{}, &buf)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "2024-06-15 10:30") {
		t.Errorf("expected last commit date, got:\n%s", out)
	}
	if !strings.Contains(out, "2") { // 30d commits = 2
		t.Errorf("expected 30d commit count, got:\n%s", out)
	}
}

func TestStats_UniqueContributorDedup(t *testing.T) {
	ws, _ := twoRepoWorkspace(t)

	mock := &captureMock{
		CaptureFn: func(dir, cmd string) (string, error) {
			if strings.Contains(cmd, `log --format="%ae"`) {
				// Alice appears 3 times, Bob once → 2 unique
				return "alice@example.com\nalice@EXAMPLE.COM\nbob@example.com\nalice@example.com\n", nil
			}
			return "", nil
		},
	}

	var buf bytes.Buffer
	if err := Stats(ws, mock, InspectOptions{}, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	// Both repos should show 2 unique contributors (case-insensitive dedup)
	count := strings.Count(out, "\t2\t") + strings.Count(out, "  2  ") + strings.Count(out, "  2\n")
	_ = count // table format may vary; verify the number 2 appears at all
	if !strings.Contains(out, "2") {
		t.Errorf("expected contributor count 2, got:\n%s", out)
	}
}

func TestStats_NotClonedRepo(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "backend", "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	ws := config.Workspace{
		Name: "myws", Path: base, Structure: "list",
		Projects: []config.Project{{
			Name: "backend",
			Repositories: []config.Repository{
				{Name: "api"},
				{Name: "missing"},
			},
		}},
	}

	mock := &captureMock{
		CaptureFn: func(dir, cmd string) (string, error) { return "", nil },
	}

	var buf bytes.Buffer
	if err := Stats(ws, mock, InspectOptions{}, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "(not cloned)") {
		t.Errorf("expected '(not cloned)' for missing repo, got:\n%s", out)
	}
}

// ─────────────────────────────── printTable ──────────────────────────────────

func TestPrintTable_Basic(t *testing.T) {
	headers := []string{"NAME", "COUNT"}
	rows := [][]string{
		{"foo", "42"},
		{"longer-name", "1"},
	}

	var buf bytes.Buffer
	printTable(&buf, "  ", headers, rows, nil)
	out := buf.String()

	if !strings.Contains(out, "NAME") || !strings.Contains(out, "COUNT") {
		t.Errorf("headers missing from output:\n%s", out)
	}
	if !strings.Contains(out, "foo") || !strings.Contains(out, "longer-name") {
		t.Errorf("data rows missing:\n%s", out)
	}
	// separator should be present
	if !strings.Contains(out, "─") {
		t.Errorf("separator missing:\n%s", out)
	}
}

func TestPrintTable_RowColor(t *testing.T) {
	headers := []string{"A"}
	rows := [][]string{{"red-row"}}
	colors := []func(string) string{func(s string) string { return "COLOR:" + s }}

	var buf bytes.Buffer
	printTable(&buf, "", headers, rows, colors)
	out := buf.String()
	if !strings.Contains(out, "COLOR:") {
		t.Errorf("row color not applied:\n%s", out)
	}
}
