package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestConfig(t *testing.T, dir, host string, port int, workspacesYAML string) {
	t.Helper()
	configPath := filepath.Join(dir, "config.yaml")
	yamlContent := fmt.Sprintf(`dukh:
  host: %s
  port: %d
zgard:
  config: {}
general:
  install_dir: /tmp
  sources_dir: /tmp/src
  bin_dir: /tmp/bin
workspaces:
%s`, host, port, workspacesYAML)
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestRun_ConfigError(t *testing.T) {
	os.Unsetenv("GRAZHDA_DIR")
	err := run([]string{"zgard", "ws", "init"})
	if err == nil || !strings.Contains(err.Error(), "failed to load config") {
		t.Fatal("expected config load error")
	}
}

func TestRun_NoArgs(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	writeTestConfig(t, tempDir, "localhost", 1234, "  []")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := run([]string{"zgard"})
	_ = w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatal(err)
	}
	output, _ := io.ReadAll(r)
	if !strings.Contains(string(output), "Zgard - The Command CLI") {
		t.Fatal("expected usage output")
	}
}

func TestRun_InvalidCommand(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	writeTestConfig(t, tempDir, "localhost", 1234, "  []")

	err := run([]string{"zgard", "invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatal("expected unknown command error")
	}
}

func TestRun_WsNoSub(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	writeTestConfig(t, tempDir, "localhost", 1234, "  []")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := run([]string{"zgard", "ws"})
	_ = w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatal(err)
	}
	output, _ := io.ReadAll(r)
	if !strings.Contains(string(output), "Usage: zgard ws <subcommand>") {
		t.Fatal("expected ws usage output")
	}
}

func TestRun_WsInvalidSub(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	writeTestConfig(t, tempDir, "localhost", 1234, "  []")

	err := run([]string{"zgard", "ws", "invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown ws subcommand") {
		t.Fatal("expected unknown ws subcommand error")
	}
}

func TestRun_Run(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	writeTestConfig(t, tempDir, "localhost", 1234, "  []")

	if err := run([]string{"zgard", "run"}); err != nil {
		t.Fatal(err)
	}
}

func TestRun_WsInitAndPurge_Local(t *testing.T) {
	tempDir := t.TempDir()
	wsPath := filepath.ToSlash(filepath.Join(tempDir, "workspace-one"))
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")

	workspacesYAML := fmt.Sprintf(`  - name: default
    default: true
    path: %s
    clone_command_template: "git clone --branch {{.Branch}} https://github.com/grazhda/{{.RepoName}} {{.DestDir}}"
    projects: []`, wsPath)
	writeTestConfig(t, tempDir, "localhost", 1234, workspacesYAML)

	if err := run([]string{"zgard", "ws", "init"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.FromSlash(wsPath), "dukh.log")); err != nil {
		t.Fatalf("expected dukh.log to be created: %v", err)
	}

	if err := run([]string{"zgard", "ws", "purge"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.FromSlash(wsPath)); !os.IsNotExist(err) {
		t.Fatalf("expected workspace directory to be removed, err=%v", err)
	}
}
