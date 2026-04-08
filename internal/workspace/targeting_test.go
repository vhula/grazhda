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


func makePartialMatchWorkspace() config.Workspace {
return config.Workspace{
Name: "ws",
Path: "/tmp/ws",
Projects: []config.Project{
{
Name: "backend",
Repositories: []config.Repository{
{Name: "ORG/PACK/my-cool-backend-lol"},
{Name: "ORG/PACK/other-service"},
},
},
},
}
}

func TestValidateFilters_PartialMatch_Substring(t *testing.T) {
ws := makePartialMatchWorkspace()
for _, filter := range []string{"cool", "backend", "my", "lol", "my-cool-backend-lol", "ORG/PACK/my-cool-backend-lol"} {
opts := workspace.RunOptions{ProjectName: "backend", RepoName: filter}
if err := workspace.ValidateFilters(ws, opts); err != nil {
t.Errorf("filter %q: expected nil, got %v", filter, err)
}
}
}

func TestValidateFilters_PartialMatch_NoMatch(t *testing.T) {
ws := makePartialMatchWorkspace()
opts := workspace.RunOptions{ProjectName: "backend", RepoName: "zzznope"}
if err := workspace.ValidateFilters(ws, opts); err == nil {
t.Fatal("expected error for unmatched filter")
}
}

func TestValidateFilters_PartialMatch_MultipleReposMatchIsValid(t *testing.T) {
ws := makePartialMatchWorkspace()
// "service" does not appear in both, but "ORG" does — both repos share the namespace prefix.
opts := workspace.RunOptions{ProjectName: "backend", RepoName: "PACK"}
if err := workspace.ValidateFilters(ws, opts); err != nil {
t.Fatalf("multiple matches must be valid: %v", err)
}
}

func TestCountMatchingRepos_Single(t *testing.T) {
ws := makePartialMatchWorkspace()
opts := workspace.RunOptions{ProjectName: "backend", RepoName: "lol"}
if n := workspace.CountMatchingRepos(ws, opts); n != 1 {
t.Errorf("expected 1, got %d", n)
}
}

func TestCountMatchingRepos_Multiple(t *testing.T) {
ws := makePartialMatchWorkspace()
opts := workspace.RunOptions{ProjectName: "backend", RepoName: "PACK"}
if n := workspace.CountMatchingRepos(ws, opts); n != 2 {
t.Errorf("expected 2 (both repos share PACK namespace), got %d", n)
}
}

func TestCountMatchingRepos_NoFilter(t *testing.T) {
ws := makePartialMatchWorkspace()
opts := workspace.RunOptions{ProjectName: "backend"}
if n := workspace.CountMatchingRepos(ws, opts); n != 0 {
t.Errorf("expected 0 when no RepoName filter, got %d", n)
}
}
