package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vhula/grazhda/config"
	pb "github.com/vhula/grazhda/dukh/proto"
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
	err, cfg := run([]string{"dukh", "stop"})
	if err != nil {
		t.Error(err)
	}
	if cfg != nil {
		t.Error("expected cfg to be nil")
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

func TestConstructGitCommand(t *testing.T) {
	tmpl := "git clone --branch {{.Branch}} https://github.com/grazhda/{{.RepoName}} {{.DestDir}}"
	cmd := constructGitCommand(tmpl, "test-repo", "/tmp/project", "main", "")
	expected := "git clone --branch main https://github.com/grazhda/test-repo test-repo"
	if cmd != expected {
		t.Errorf("expected %s, got %s", expected, cmd)
	}
	cmd2 := constructGitCommand(tmpl, "test-repo", "/tmp/project", "dev", "custom-dir")
	expected2 := "git clone --branch dev https://github.com/grazhda/test-repo custom-dir"
	if cmd2 != expected2 {
		t.Errorf("expected %s, got %s", expected2, cmd2)
	}
}

func TestInitWorkspaces(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Dukh: struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}{Host: "localhost", Port: 50501},
		Zgard: struct {
			Config map[string]interface{} `yaml:"config"`
		}{Config: map[string]interface{}{}},
		General: struct {
			InstallDir string `yaml:"install_dir"`
			SourcesDir string `yaml:"sources_dir"`
			BinDir     string `yaml:"bin_dir"`
		}{
			InstallDir: "/tmp",
			SourcesDir: "/tmp/src",
			BinDir:     "/tmp/bin",
		},
		Workspaces: []config.WorkspaceConfig{
			{
				Name:                 "test-ws",
				Default:              true,
				Path:                 filepath.Join(tempDir, "test-ws"),
				CloneCommandTemplate: "echo cloned {{.RepoName}} to {{.DestDir}}",
				Projects: []config.ProjectConfig{
					{
						Name: "test-project",
						Subprojects: []config.SubprojectConfig{
							{
								Branch: "main",
								Repositories: []config.RepositoryConfig{
									{Name: "repo1"},
									{Name: "repo2", LocalDirName: "custom-repo2"},
								},
							},
						},
					},
				},
			},
		},
	}
	server := &workspaceServer{
		workspaces: make(map[string]*pb.Workspace),
		config:     cfg,
	}
	resp, err := server.InitWorkspaces(context.Background(), &pb.InitWorkspacesRequest{})
	if err != nil {
		t.Error(err)
	}
	if len(resp.Statuses) == 0 {
		t.Error("expected statuses")
	}
	// Check workspace dir created
	if _, err := os.Stat(filepath.Join(tempDir, "test-ws")); os.IsNotExist(err) {
		t.Error("workspace dir not created")
	}
	// Check log file
	logPath := filepath.Join(tempDir, "test-ws", "dukh.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("log file not created")
	}
	// Check project dir
	projectDir := filepath.Join(tempDir, "test-ws", "test-project")
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Error("project dir not created")
	}
	// Check workspace in map
	if len(server.workspaces) != 1 {
		t.Error("workspace not added to map")
	}
}

func TestPurgeWorkspaces(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Dukh: struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}{Host: "localhost", Port: 50501},
		Zgard: struct {
			Config map[string]interface{} `yaml:"config"`
		}{Config: map[string]interface{}{}},
		General: struct {
			InstallDir string `yaml:"install_dir"`
			SourcesDir string `yaml:"sources_dir"`
			BinDir     string `yaml:"bin_dir"`
		}{
			InstallDir: "/tmp",
			SourcesDir: "/tmp/src",
			BinDir:     "/tmp/bin",
		},
		Workspaces: []config.WorkspaceConfig{
			{
				Name:                 "test-ws",
				Default:              true,
				Path:                 filepath.Join(tempDir, "test-ws"),
				CloneCommandTemplate: "echo cloned {{.RepoName}}",
				Projects:             []config.ProjectConfig{},
			},
		},
	}
	server := &workspaceServer{
		workspaces: make(map[string]*pb.Workspace),
		config:     cfg,
	}
	// First init
	server.InitWorkspaces(context.Background(), &pb.InitWorkspacesRequest{})
	// Then purge
	resp, err := server.PurgeWorkspaces(context.Background(), &pb.PurgeWorkspacesRequest{})
	if err != nil {
		t.Error(err)
	}
	if len(resp.Statuses) == 0 {
		t.Error("expected statuses")
	}
	// Check dir deleted
	if _, err := os.Stat(filepath.Join(tempDir, "test-ws")); !os.IsNotExist(err) {
		t.Error("workspace dir not deleted")
	}
	// Check map empty
	if len(server.workspaces) != 0 {
		t.Error("workspace not removed from map")
	}
}

func TestGetWorkspaces(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Dukh: struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}{Host: "localhost", Port: 50501},
		Zgard: struct {
			Config map[string]interface{} `yaml:"config"`
		}{Config: map[string]interface{}{}},
		General: struct {
			InstallDir string `yaml:"install_dir"`
			SourcesDir string `yaml:"sources_dir"`
			BinDir     string `yaml:"bin_dir"`
		}{
			InstallDir: "/tmp",
			SourcesDir: "/tmp/src",
			BinDir:     "/tmp/bin",
		},
		Workspaces: []config.WorkspaceConfig{
			{
				Name:                 "test-ws",
				Default:              true,
				Path:                 filepath.Join(tempDir, "test-ws"),
				CloneCommandTemplate: "echo cloned {{.RepoName}}",
				Projects:             []config.ProjectConfig{},
			},
		},
	}
	server := &workspaceServer{
		workspaces: make(map[string]*pb.Workspace),
		config:     cfg,
	}
	// Init first
	server.InitWorkspaces(context.Background(), &pb.InitWorkspacesRequest{})
	// Get
	resp, err := server.GetWorkspaces(context.Background(), &pb.GetWorkspacesRequest{Ids: []string{"test-ws"}})
	if err != nil {
		t.Error(err)
	}
	if len(resp.Workspaces) != 1 {
		t.Error("expected 1 workspace")
	}
	if resp.Workspaces[0].Id != "test-ws" {
		t.Error("wrong workspace id")
	}
}
