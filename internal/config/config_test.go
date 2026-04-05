package config_test

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/config"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "testdata", name)
}

// --- Load ---

func TestLoad_ValidSingleWorkspace(t *testing.T) {
	cfg, err := config.Load(testdataPath("valid_single_workspace.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(cfg.Workspaces))
	}
	if cfg.Workspaces[0].Name != "test-ws" {
		t.Errorf("expected workspace name 'test-ws', got %q", cfg.Workspaces[0].Name)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !errors.Is(err, errors.Unwrap(err)) && !strings.Contains(err.Error(), "config file") {
		t.Errorf("error should mention config file path, got: %v", err)
	}
}

func TestLoad_ValidMultiWorkspace(t *testing.T) {
	cfg, err := config.Load(testdataPath("valid_multi_workspace.yaml"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Workspaces) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(cfg.Workspaces))
	}
}

// --- DefaultWorkspace ---

func TestDefaultWorkspace_FoundByDefault(t *testing.T) {
	cfg, err := config.Load(testdataPath("valid_multi_workspace.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	ws, err := config.DefaultWorkspace(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Name != "default" {
		t.Errorf("expected default workspace, got %q", ws.Name)
	}
}

func TestDefaultWorkspace_NotFound(t *testing.T) {
	cfg := &config.Config{
		Workspaces: []config.Workspace{{Name: "other", Path: "/tmp"}},
	}
	_, err := config.DefaultWorkspace(cfg)
	if err == nil {
		t.Fatal("expected error for missing default workspace")
	}
	if !strings.Contains(err.Error(), "no default workspace found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// --- Validate ---

func TestValidate_ValidConfig(t *testing.T) {
	cfg, _ := config.Load(testdataPath("valid_single_workspace.yaml"))
	errs := config.Validate(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidate_MissingPath(t *testing.T) {
	cfg := &config.Config{
		Workspaces: []config.Workspace{{
			Name:                 "default",
			CloneCommandTemplate: "git clone {{.RepoName}}",
			Projects: []config.Project{{
				Name:   "p",
				Branch: "main",
			}},
		}},
	}
	errs := config.Validate(cfg)
	if len(errs) == 0 {
		t.Fatal("expected validation errors")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "workspace[0].path: required field missing") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected path error, got: %v", errs)
	}
}

func TestValidate_DuplicateWorkspaceNames(t *testing.T) {
	cfg, err := config.Load(testdataPath("duplicate_workspace_names.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	errs := config.Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e, "workspace names must be unique") && strings.Contains(e, "default") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected duplicate workspace name error, got: %v", errs)
	}
}

func TestValidate_DuplicateProjectNames(t *testing.T) {
	cfg := &config.Config{
		Workspaces: []config.Workspace{{
			Name:                 "default",
			Path:                 "/tmp/ws",
			CloneCommandTemplate: "git clone {{.RepoName}}",
			Projects: []config.Project{
				{Name: "backend", Branch: "main"},
				{Name: "backend", Branch: "main"},
			},
		}},
	}
	errs := config.Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e, "duplicate project name") && strings.Contains(e, "backend") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected duplicate project error, got: %v", errs)
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg, err := config.Load(testdataPath("missing_required_fields.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	errs := config.Validate(cfg)
	if len(errs) < 2 {
		t.Errorf("expected multiple errors, got %d: %v", len(errs), errs)
	}
}

func TestValidate_InvalidTemplate(t *testing.T) {
	cfg, err := config.Load(testdataPath("invalid_template.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	errs := config.Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e, "invalid template") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected invalid template error, got: %v", errs)
	}
}

func TestValidate_MissingBranch(t *testing.T) {
	cfg, err := config.Load(testdataPath("missing_branch.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	errs := config.Validate(cfg)
	found := false
	for _, e := range errs {
		if strings.Contains(e, "branch: required field missing") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing branch error, got: %v", errs)
	}
}

// --- RenderCloneCmd ---

func TestRenderCloneCmd_Basic(t *testing.T) {
	tmpl := "git clone --branch {{.Branch}} https://github.com/org/{{.RepoName}} {{.DestDir}}"
	proj := config.Project{Name: "backend", Branch: "main"}
	repo := config.Repository{Name: "api"}

	cmd, err := config.RenderCloneCmd(tmpl, "/workspace/backend", proj, repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "git clone --branch main https://github.com/org/api /workspace/backend/api"
	if cmd != expected {
		t.Errorf("expected %q, got %q", expected, cmd)
	}
}

func TestRenderCloneCmd_RepoOverridesBranch(t *testing.T) {
	tmpl := "git clone --branch {{.Branch}} https://github.com/org/{{.RepoName}} {{.DestDir}}"
	proj := config.Project{Name: "backend", Branch: "main"}
	repo := config.Repository{Name: "auth", Branch: "dev"}

	cmd, err := config.RenderCloneCmd(tmpl, "/workspace/backend", proj, repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cmd, "dev") {
		t.Errorf("expected branch 'dev', got %q", cmd)
	}
}

func TestRenderCloneCmd_LocalDirName(t *testing.T) {
	tmpl := "git clone {{.RepoName}} {{.DestDir}}"
	proj := config.Project{Name: "backend", Branch: "main"}
	repo := config.Repository{Name: "api", LocalDirName: "api-service"}

	cmd, err := config.RenderCloneCmd(tmpl, "/workspace/backend", proj, repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(cmd, "api-service") {
		t.Errorf("expected local_dir_name 'api-service' in cmd, got %q", cmd)
	}
	if strings.Contains(cmd, "/api ") || strings.HasSuffix(cmd, "/api") {
		t.Errorf("DestDir should use local_dir_name, not repo name, got %q", cmd)
	}
}

func TestRenderCloneCmd_InvalidTemplate(t *testing.T) {
	tmpl := "git clone {{.Unclosed"
	proj := config.Project{Name: "p", Branch: "main"}
	repo := config.Repository{Name: "r"}

	_, err := config.RenderCloneCmd(tmpl, "/ws", proj, repo)
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}
