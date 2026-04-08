package workspace_test

import (
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/workspace"
)

func makeTargetingConfig() *config.Config {
	return &config.Config{
		Workspaces: []config.Workspace{
			{Name: "default", Default: true, Path: "/tmp/default"},
			{Name: "myws", Path: "/tmp/myws"},
			{Name: "other", Path: "/tmp/other"},
		},
	}
}

func TestResolve_DefaultWorkspace(t *testing.T) {
	cfg := makeTargetingConfig()
	wss, err := workspace.Resolve(cfg, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wss) != 1 || wss[0].Name != "default" {
		t.Errorf("expected default workspace, got %v", wss)
	}
}

func TestResolve_NamedWorkspace(t *testing.T) {
	cfg := makeTargetingConfig()
	wss, err := workspace.Resolve(cfg, "myws", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wss) != 1 || wss[0].Name != "myws" {
		t.Errorf("expected myws, got %v", wss)
	}
}

func TestResolve_NotFound(t *testing.T) {
	cfg := makeTargetingConfig()
	_, err := workspace.Resolve(cfg, "nonexistent", false)
	if err == nil {
		t.Fatal("expected error for nonexistent workspace")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("expected error to mention workspace name, got: %v", err)
	}
}

func TestResolve_All(t *testing.T) {
	cfg := makeTargetingConfig()
	wss, err := workspace.Resolve(cfg, "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wss) != 3 {
		t.Errorf("expected all 3 workspaces, got %d", len(wss))
	}
}

func TestResolve_MutuallyExclusive(t *testing.T) {
	cfg := makeTargetingConfig()
	_, err := workspace.Resolve(cfg, "myws", true)
	if err == nil {
		t.Fatal("expected error for --name + --all")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolve_NoDefaultWorkspace(t *testing.T) {
	cfg := &config.Config{
		Workspaces: []config.Workspace{
			{Name: "ws1", Path: "/tmp/ws1"},
		},
	}
	_, err := workspace.Resolve(cfg, "", false)
	if err == nil {
		t.Fatal("expected error when no default workspace")
	}
	if !strings.Contains(err.Error(), "no default workspace found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolve_MultiWorkspaceFixture(t *testing.T) {
	cfg, err := config.Load("../testdata/valid_multi_workspace.yaml")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	wss, err := workspace.Resolve(cfg, "", false)
	if err != nil {
		t.Fatalf("default resolve: %v", err)
	}
	if len(wss) != 1 || wss[0].Name != "default" {
		t.Errorf("expected default workspace, got %v", wss)
	}

	wss, err = workspace.Resolve(cfg, "secondary", false)
	if err != nil {
		t.Fatalf("named resolve: %v", err)
	}
	if wss[0].Name != "secondary" {
		t.Errorf("expected secondary, got %v", wss)
	}

	wss, err = workspace.Resolve(cfg, "", true)
	if err != nil {
		t.Fatalf("all resolve: %v", err)
	}
	if len(wss) != 2 {
		t.Errorf("expected 2 workspaces, got %d", len(wss))
	}
}

func makeFilterWorkspace() config.Workspace {
	return config.Workspace{
		Name: "ws",
		Path: "/tmp/ws",
		Projects: []config.Project{
			{
				Name: "backend",
				Repositories: []config.Repository{
					{Name: "api"},
					{Name: "auth"},
				},
			},
		},
	}
}

func TestValidateFilters_NoFilter(t *testing.T) {
	ws := makeFilterWorkspace()
	if err := workspace.ValidateFilters(ws, workspace.RunOptions{}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateFilters_ProjectFound(t *testing.T) {
	ws := makeFilterWorkspace()
	opts := workspace.RunOptions{ProjectName: "backend"}
	if err := workspace.ValidateFilters(ws, opts); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateFilters_ProjectNotFound(t *testing.T) {
	ws := makeFilterWorkspace()
	opts := workspace.RunOptions{ProjectName: "nope"}
	err := workspace.ValidateFilters(ws, opts)
	if err == nil {
		t.Fatal("expected error for unknown project")
	}
	if !strings.Contains(err.Error(), "nope") || !strings.Contains(err.Error(), "ws") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestValidateFilters_RepoFound(t *testing.T) {
	ws := makeFilterWorkspace()
	opts := workspace.RunOptions{ProjectName: "backend", RepoName: "api"}
	if err := workspace.ValidateFilters(ws, opts); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateFilters_RepoNotFound(t *testing.T) {
	ws := makeFilterWorkspace()
	opts := workspace.RunOptions{ProjectName: "backend", RepoName: "nope"}
	err := workspace.ValidateFilters(ws, opts)
	if err == nil {
		t.Fatal("expected error for unknown repo")
	}
	if !strings.Contains(err.Error(), "nope") || !strings.Contains(err.Error(), "backend") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func makeListFilterWorkspace() config.Workspace {
return config.Workspace{
Name:      "ws",
Path:      "/tmp/ws",
Structure: "list",
Projects: []config.Project{
{
Name: "backend",
Repositories: []config.Repository{
{Name: "org/team/api"},
{Name: "org/team/auth.git"},
},
},
},
}
}

func TestValidateFilters_ListStructure_MatchesBySegment(t *testing.T) {
ws := makeListFilterWorkspace()
opts := workspace.RunOptions{ProjectName: "backend", RepoName: "api"}
if err := workspace.ValidateFilters(ws, opts); err != nil {
t.Fatalf("expected nil for list-structure segment match, got %v", err)
}
}

func TestValidateFilters_ListStructure_MatchesBySegmentGitSuffix(t *testing.T) {
ws := makeListFilterWorkspace()
opts := workspace.RunOptions{ProjectName: "backend", RepoName: "auth"}
if err := workspace.ValidateFilters(ws, opts); err != nil {
t.Fatalf("expected nil for list-structure .git strip match, got %v", err)
}
}

func TestValidateFilters_ListStructure_NoMatchForFullPath(t *testing.T) {
ws := makeListFilterWorkspace()
// Using full path as filter — still matches via exact name comparison.
opts := workspace.RunOptions{ProjectName: "backend", RepoName: "org/team/api"}
if err := workspace.ValidateFilters(ws, opts); err != nil {
t.Fatalf("exact name should still match: %v", err)
}
}

func TestValidateFilters_ListStructure_NoMatch(t *testing.T) {
ws := makeListFilterWorkspace()
opts := workspace.RunOptions{ProjectName: "backend", RepoName: "nope"}
if err := workspace.ValidateFilters(ws, opts); err == nil {
t.Fatal("expected error for unmatched repo in list-structure workspace")
}
}
