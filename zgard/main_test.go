package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/proto"
	"google.golang.org/grpc"
)

type mockWorkspaceServer struct {
	proto.UnimplementedWorkspaceServiceServer
	proto.UnimplementedDukhServiceServer
	initCalled   bool
	purgeCalled  bool
	stopCalled   bool
	statusCalled bool
}

func (m *mockWorkspaceServer) InitWorkspaces(ctx context.Context, req *proto.InitWorkspacesRequest) (*proto.InitWorkspacesResponse, error) {
	m.initCalled = true
	return &proto.InitWorkspacesResponse{Statuses: []string{"init status"}}, nil
}

func (m *mockWorkspaceServer) PurgeWorkspaces(ctx context.Context, req *proto.PurgeWorkspacesRequest) (*proto.PurgeWorkspacesResponse, error) {
	m.purgeCalled = true
	return &proto.PurgeWorkspacesResponse{Statuses: []string{"purge status"}}, nil
}

func (m *mockWorkspaceServer) GetWorkspaces(ctx context.Context, req *proto.GetWorkspacesRequest) (*proto.GetWorkspacesResponse, error) {
	return &proto.GetWorkspacesResponse{}, nil
}

func (m *mockWorkspaceServer) StopDukh(ctx context.Context, req *proto.StopDukhRequest) (*proto.StopDukhResponse, error) {
	m.stopCalled = true
	return &proto.StopDukhResponse{Status: "dukh server stopping"}, nil
}

func (m *mockWorkspaceServer) StatusDukh(ctx context.Context, req *proto.StatusDukhRequest) (*proto.StatusDukhResponse, error) {
	m.statusCalled = true
	return &proto.StatusDukhResponse{Running: true, Pid: 12345, Status: "dukh server is running"}, nil
}

func startTestServer(t *testing.T, server *mockWorkspaceServer) (string, func()) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	proto.RegisterWorkspaceServiceServer(s, server)
	proto.RegisterDukhServiceServer(s, server)
	go s.Serve(lis)
	addr := lis.Addr().String()
	return addr, func() { s.Stop(); lis.Close() }
}

func writeTestConfig(t *testing.T, dir, host string, port int) {
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
workspaces: []`, host, port)
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestRun_ConfigError(t *testing.T) {
	os.Unsetenv("GRAZHDA_DIR")
	err := run([]string{"zgard", "ws", "init"})
	if err == nil || !strings.Contains(err.Error(), "failed to load config") {
		t.Error("expected config load error")
	}
}

func TestRun_NoArgs(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	configPath := filepath.Join(tempDir, "config.yaml")
	yamlContent := `dukh:
  host: localhost
  port: 1234
zgard:
  config: {}
general:
  install_dir: /tmp
  sources_dir: /tmp/src
  bin_dir: /tmp/bin
workspaces: []`
	os.WriteFile(configPath, []byte(yamlContent), 0644)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := run([]string{"zgard"})
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Error(err)
	}
	output, _ := io.ReadAll(r)
	if !strings.Contains(string(output), "Zgard - The Command CLI") {
		t.Error("expected usage output")
	}
}

func TestRun_InvalidCommand(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	configPath := filepath.Join(tempDir, "config.yaml")
	yamlContent := `dukh:
  host: localhost
  port: 1234
zgard:
  config: {}
general:
  install_dir: /tmp
  sources_dir: /tmp/src
  bin_dir: /tmp/bin
workspaces: []`
	os.WriteFile(configPath, []byte(yamlContent), 0644)
	err := run([]string{"zgard", "invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Error("expected unknown command error")
	}
}

func TestRun_WsNoSub(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	configPath := filepath.Join(tempDir, "config.yaml")
	yamlContent := `dukh:
  host: localhost
  port: 1234
zgard:
  config: {}
general:
  install_dir: /tmp
  sources_dir: /tmp/src
  bin_dir: /tmp/bin
workspaces: []`
	os.WriteFile(configPath, []byte(yamlContent), 0644)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := run([]string{"zgard", "ws"})
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Error(err)
	}
	output, _ := io.ReadAll(r)
	if !strings.Contains(string(output), "Usage: zgard ws <subcommand>") {
		t.Error("expected ws usage output")
	}
}

func TestRun_WsInvalidSub(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	configPath := filepath.Join(tempDir, "config.yaml")
	yamlContent := `dukh:
  host: localhost
  port: 1234
zgard:
  config: {}
general:
  install_dir: /tmp
  sources_dir: /tmp/src
  bin_dir: /tmp/bin
workspaces: []`
	os.WriteFile(configPath, []byte(yamlContent), 0644)
	err := run([]string{"zgard", "ws", "invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown ws subcommand") {
		t.Error("expected unknown subcommand error")
	}
}

func TestRun_Run(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	configPath := filepath.Join(tempDir, "config.yaml")
	yamlContent := `dukh:
  host: localhost
  port: 1234
zgard:
  config: {}
general:
  install_dir: /tmp
  sources_dir: /tmp/src
  bin_dir: /tmp/bin
workspaces: []`
	os.WriteFile(configPath, []byte(yamlContent), 0644)
	err := run([]string{"zgard", "run"})
	if err != nil {
		t.Error(err)
	}
}

func TestRun_DukhNoSub(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := run([]string{"zgard", "dukh"})
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Error(err)
	}
	output, _ := io.ReadAll(r)
	if !strings.Contains(string(output), "Usage: zgard dukh <subcommand>") {
		t.Error("expected dukh usage output")
	}
}

func TestRun_DukhInvalidSub(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")

	err := run([]string{"zgard", "dukh", "invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown dukh subcommand") {
		t.Error("expected unknown dukh subcommand error")
	}
}

func TestRun_DukhStopNoPIDFile(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")

	err := run([]string{"zgard", "dukh", "stop"})
	if err == nil || !strings.Contains(err.Error(), "failed to load config") {
		t.Error("expected config error for dukh stop")
	}
}

func TestRun_DukhStatusStopped(t *testing.T) {
	server := &mockWorkspaceServer{}
	addr, cleanup := startTestServer(t, server)
	defer cleanup()
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	writeTestConfig(t, tempDir, host, port)

	err := run([]string{"zgard", "dukh", "status"})
	if err != nil {
		t.Fatal(err)
	}
	if !server.statusCalled {
		t.Error("StatusDukh not called")
	}
}

func TestRun_DukhStatusRunning(t *testing.T) {
	server := &mockWorkspaceServer{}
	addr, cleanup := startTestServer(t, server)
	defer cleanup()
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	writeTestConfig(t, tempDir, host, port)

	err := run([]string{"zgard", "dukh", "status"})
	if err != nil {
		t.Fatal(err)
	}
	if !server.statusCalled {
		t.Error("StatusDukh not called")
	}
}

func TestRun_DukhStop(t *testing.T) {
	server := &mockWorkspaceServer{}
	addr, cleanup := startTestServer(t, server)
	defer cleanup()
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	writeTestConfig(t, tempDir, host, port)

	err := run([]string{"zgard", "dukh", "stop"})
	if err != nil {
		t.Fatal(err)
	}
	if !server.stopCalled {
		t.Error("StopDukh not called")
	}
}

func TestRun_WsInit(t *testing.T) {
	server := &mockWorkspaceServer{}
	addr, cleanup := startTestServer(t, server)
	defer cleanup()
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	configPath := filepath.Join(tempDir, "config.yaml")
	yamlContent := fmt.Sprintf(`dukh:
  host: %s
  port: %d
zgard:
  config: {}
general:
  install_dir: /tmp
  sources_dir: /tmp/src
  bin_dir: /tmp/bin
workspaces: []`, host, port)
	os.WriteFile(configPath, []byte(yamlContent), 0644)
	err := run([]string{"zgard", "ws", "init"})
	if err != nil {
		t.Error(err)
	}
	if !server.initCalled {
		t.Error("InitWorkspaces not called")
	}
}

func TestRun_WsPurge(t *testing.T) {
	server := &mockWorkspaceServer{}
	addr, cleanup := startTestServer(t, server)
	defer cleanup()
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	tempDir := t.TempDir()
	os.Setenv("GRAZHDA_DIR", tempDir)
	defer os.Unsetenv("GRAZHDA_DIR")
	configPath := filepath.Join(tempDir, "config.yaml")
	yamlContent := fmt.Sprintf(`dukh:
  host: %s
  port: %d
zgard:
  config: {}
general:
  install_dir: /tmp
  sources_dir: /tmp/src
  bin_dir: /tmp/bin
workspaces: []`, host, port)
	os.WriteFile(configPath, []byte(yamlContent), 0644)
	err := run([]string{"zgard", "ws", "purge"})
	if err != nil {
		t.Error(err)
	}
	if !server.purgeCalled {
		t.Error("PurgeWorkspaces not called")
	}
}
