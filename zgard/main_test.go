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

	"github.com/vhula/grazhda/dukh/proto"
	"google.golang.org/grpc"
)

type mockWorkspaceServer struct {
	proto.UnimplementedWorkspaceServiceServer
	initCalled  bool
	purgeCalled bool
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

func startTestServer(t *testing.T, server *mockWorkspaceServer) (string, func()) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	s := grpc.NewServer()
	proto.RegisterWorkspaceServiceServer(s, server)
	go s.Serve(lis)
	addr := lis.Addr().String()
	return addr, func() { s.Stop(); lis.Close() }
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
