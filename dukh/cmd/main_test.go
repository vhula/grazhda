package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRun_NoArgs(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err, cfg := run([]string{"dukh"})
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Error(err)
	}
	if cfg != nil {
		t.Error("expected cfg to be nil")
	}
	output, _ := io.ReadAll(r)
	if !strings.Contains(string(output), "Dukh - The Worker CLI") {
		t.Error("expected usage output")
	}
}

func TestRun_Start_ConfigError(t *testing.T) {
	os.Unsetenv("GRAZHDA_DIR")
	err, _ := run([]string{"dukh", "start"})
	if err == nil || !strings.Contains(err.Error(), "failed to load config") {
		t.Error("expected config load error")
	}
}

func TestRun_Stop(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")

	cmd := exec.Command("sh", "-c", "sleep 30")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	pidFilePath := filepath.Join(tempDir, "dukh.pid")
	if err := os.WriteFile(pidFilePath, []byte(strconv.Itoa(cmd.Process.Pid)+"\n"), 0644); err != nil {
		_ = cmd.Process.Kill()
		t.Fatal(err)
	}

	err, cfg := run([]string{"dukh", "stop"})
	if err != nil {
		t.Error(err)
	}
	if cfg != nil {
		t.Error("expected cfg to be nil")
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("expected process to be stopped")
	}
}

func TestRun_Stop_NoPID(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")

	err, _ := run([]string{"dukh", "stop"})
	if err == nil || !strings.Contains(err.Error(), "failed to stop dukh server") {
		t.Error("expected stop error when pid file is missing")
	}
}

func TestRun_Status(t *testing.T) {
	err, cfg := run([]string{"dukh", "status"})
	if err != nil {
		t.Error(err)
	}
	if cfg != nil {
		t.Error("expected cfg to be nil")
	}
}

func TestRun_Invalid(t *testing.T) {
	err, _ := run([]string{"dukh", "invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Error("expected unknown command error")
	}
}
