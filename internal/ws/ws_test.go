package ws

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/config"
)

func newTestConfig(workspacePath string, cloneTemplate string) *config.Config {
	cfg := &config.Config{}
	cfg.Workspaces = []config.WorkspaceConfig{
		{
			Name:                 "default",
			Default:              true,
			Path:                 workspacePath,
			CloneCommandTemplate: cloneTemplate,
			Projects: []config.ProjectConfig{
				{
					Name:   "project-a",
					Branch: "main",
					Repositories: []config.RepositoryConfig{
						{Name: "repo-1"},
						{Name: "repo-2", LocalDirName: "repo-2-local"},
					},
				},
			},
		},
	}
	return cfg
}

func containsStatus(statuses []string, needle string) bool {
	for _, s := range statuses {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}

func TestConstructGitCommand(t *testing.T) {
	tmpl := "git clone --branch {{.Branch}} https://github.com/grazhda/{{.RepoName}} {{.DestDir}}"

	cmd, err := constructGitCommand(tmpl, "repo-1", "main", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != "git clone --branch main https://github.com/grazhda/repo-1 repo-1" {
		t.Fatalf("unexpected command: %s", cmd)
	}

	cmd, err = constructGitCommand(tmpl, "repo-2", "dev", "custom-dir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != "git clone --branch dev https://github.com/grazhda/repo-2 custom-dir" {
		t.Fatalf("unexpected command: %s", cmd)
	}
}

func TestConstructGitCommand_InvalidTemplate(t *testing.T) {
	_, err := constructGitCommand("{{ .RepoName", "repo-1", "main", "")
	if err == nil {
		t.Fatal("expected template parse error")
	}
}

func TestInit_CreatesWorkspaceAndLog(t *testing.T) {
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "ws-default")
	cloneTemplate := "echo cloning {{.RepoName}} {{.Branch}} {{.DestDir}}"

	manager := New(newTestConfig(workspacePath, cloneTemplate))
	statuses := manager.Init()

	if !containsStatus(statuses, "initialized workspace default") {
		t.Fatalf("expected initialized status, got: %v", statuses)
	}
	if !containsStatus(statuses, "cloned repo-1 successfully") {
		t.Fatalf("expected cloned status for repo-1, got: %v", statuses)
	}
	if !containsStatus(statuses, "cloned repo-2 successfully") {
		t.Fatalf("expected cloned status for repo-2, got: %v", statuses)
	}

	if _, err := os.Stat(workspacePath); err != nil {
		t.Fatalf("expected workspace dir to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workspacePath, "project-a")); err != nil {
		t.Fatalf("expected project dir to exist: %v", err)
	}

	logFile := filepath.Join(workspacePath, "dukh.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("expected dukh.log to exist: %v", err)
	}
	if !strings.Contains(string(content), "Id: default") {
		t.Fatalf("expected log to contain workspace id, got: %s", string(content))
	}

	if _, ok := manager.items["default"]; !ok {
		t.Fatal("expected workspace to be registered in manager items")
	}
}

func TestInit_TemplateErrorReported(t *testing.T) {
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "ws-default")
	cloneTemplate := "{{ .RepoName"

	manager := New(newTestConfig(workspacePath, cloneTemplate))
	statuses := manager.Init()

	if !containsStatus(statuses, "failed to clone repo-1") {
		t.Fatalf("expected clone failure status for repo-1, got: %v", statuses)
	}
	if !containsStatus(statuses, "failed to clone repo-2") {
		t.Fatalf("expected clone failure status for repo-2, got: %v", statuses)
	}
	if !containsStatus(statuses, "initialized workspace default") {
		t.Fatalf("expected workspace init status despite clone errors, got: %v", statuses)
	}
}

func TestPurge_RemovesWorkspaceDirectoriesAndItems(t *testing.T) {
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, "ws-default")
	cloneTemplate := "echo cloning {{.RepoName}} {{.Branch}} {{.DestDir}}"

	manager := New(newTestConfig(workspacePath, cloneTemplate))
	_ = manager.Init()

	if _, err := os.Stat(workspacePath); err != nil {
		t.Fatalf("expected workspace dir before purge: %v", err)
	}
	if _, ok := manager.items["default"]; !ok {
		t.Fatal("expected workspace in items before purge")
	}

	statuses := manager.Purge()
	if !containsStatus(statuses, "purged default") {
		t.Fatalf("expected purge status, got: %v", statuses)
	}
	if _, err := os.Stat(workspacePath); !os.IsNotExist(err) {
		t.Fatalf("expected workspace dir removed, err=%v", err)
	}
	if _, ok := manager.items["default"]; ok {
		t.Fatal("expected workspace removed from items after purge")
	}
}
