package server_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/vhula/grazhda/dukh/server"
)

// writeMinimalConfig writes a valid but empty-workspace config YAML and returns its path.
func writeMinimalConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("workspaces: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return cfgPath
}

func TestMonitor_StartStop(t *testing.T) {
	cfgPath := writeMinimalConfig(t)
	m := server.NewMonitor(cfgPath, log.New(io.Discard))
	m.Start()

	// Confirm the loop is running by completing a scan round-trip.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.TriggerScanAndWait(ctx); err != nil {
		t.Fatalf("monitor not responsive after Start: %v", err)
	}

	m.Stop()
	// Reaching here confirms Stop returned without hanging.
}

func TestMonitor_Snapshot(t *testing.T) {
	cfgPath := writeMinimalConfig(t)
	m := server.NewMonitor(cfgPath, log.New(io.Discard))
	m.Start()
	defer m.Stop()

	// Ensure at least one scan has completed before reading the snapshot.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.TriggerScanAndWait(ctx); err != nil {
		t.Fatalf("TriggerScanAndWait: %v", err)
	}

	snap := m.Snapshot()
	if snap == nil {
		t.Fatal("expected non-nil snapshot after scan")
	}
	// With an empty-workspace config the snapshot should be an empty slice.
	if len(snap) != 0 {
		t.Fatalf("expected 0 workspaces, got %d", len(snap))
	}
}

func TestMonitor_TriggerScanAndWait(t *testing.T) {
	cfgPath := writeMinimalConfig(t)
	m := server.NewMonitor(cfgPath, log.New(io.Discard))
	m.Start()
	defer m.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.TriggerScanAndWait(ctx); err != nil {
		t.Fatalf("TriggerScanAndWait returned error: %v", err)
	}
}

func TestMonitor_TriggerScanAndWait_Cancelled(t *testing.T) {
	cfgPath := writeMinimalConfig(t)
	m := server.NewMonitor(cfgPath, log.New(io.Discard))
	m.Start()
	defer m.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled
	if err := m.TriggerScanAndWait(ctx); err == nil {
		t.Fatal("expected error from already-cancelled context")
	}
}

func TestMonitor_SnapshotBeforeStart(t *testing.T) {
	m := server.NewMonitor("/nonexistent/config.yaml", log.New(io.Discard))
	snap := m.Snapshot()
	if snap == nil {
		t.Fatal("expected non-nil (empty) snapshot before start")
	}
	if len(snap) != 0 {
		t.Fatalf("expected 0 workspaces before start, got %d", len(snap))
	}
}
