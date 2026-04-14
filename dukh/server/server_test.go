package server

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	dukhpb "github.com/vhula/grazhda/dukh/proto"
)

func TestWriteAndRemovePID(t *testing.T) {
	dir := t.TempDir()
	if err := WritePID(dir); err != nil {
		t.Fatalf("WritePID failed: %v", err)
	}
	pidFile := filepath.Join(dir, "run", "dukh.pid")
	if _, err := os.Stat(pidFile); err != nil {
		t.Fatalf("expected pid file to exist: %v", err)
	}
	RemovePID(dir)
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Fatalf("expected pid file removed, stat err=%v", err)
	}
}

func TestCurrentBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...) //nolint:gosec
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v (%s)", args, err, string(out))
		}
	}
	run("init")
	run("config", "user.email", "a@b.com")
	run("config", "user.name", "A")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "init")

	branch, err := currentBranch(dir)
	if err != nil {
		t.Fatalf("currentBranch error: %v", err)
	}
	if strings.TrimSpace(branch) == "" {
		t.Fatal("expected non-empty branch")
	}
}

func TestMonitorLoadPeriodDefaultOnMissingConfig(t *testing.T) {
	m := NewMonitor("/missing/config.yaml", log.New(io.Discard))
	if got := m.loadPeriod(); got != defaultPeriod {
		t.Fatalf("loadPeriod = %v, want %v", got, defaultPeriod)
	}
}

func TestServerStatusFromSnapshot(t *testing.T) {
	logger := log.New(io.Discard)
	m := NewMonitor("/unused", logger)
	m.snapshot = []WorkspaceHealth{
		{
			Name: "w1",
			Path: "/tmp/w1",
			Projects: []ProjectHealth{
				{
					Name: "p1",
					Repositories: []RepoHealth{
						{Name: "r1", Path: "/tmp/w1/p1/r1", ConfiguredBranch: "main", ActualBranch: "main", Exists: true, BranchAligned: true},
					},
				},
			},
		},
	}
	s := New(m, logger)
	resp, err := s.Status(context.Background(), &dukhpb.StatusRequest{WorkspaceName: "w1"})
	if err != nil {
		t.Fatalf("Status error: %v", err)
	}
	if len(resp.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(resp.Workspaces))
	}
	if resp.Workspaces[0].Name != "w1" {
		t.Fatalf("unexpected workspace name: %s", resp.Workspaces[0].Name)
	}
}
