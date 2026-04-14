package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/workspace"
)

func testWorkspace() config.Workspace {
	return config.Workspace{
		Name:      "test-ws",
		Path:      "/fake/path",
		Structure: "list",
		Projects: []config.Project{
			{
				Name:   "backend",
				Branch: "main",
				Tags:   []string{"go"},
				Repositories: []config.Repository{
					{Name: "org/api"},
					{Name: "org/auth-service", Tags: []string{"security"}},
				},
			},
			{
				Name:   "frontend",
				Branch: "develop",
				Tags:   []string{"js"},
				Repositories: []config.Repository{
					{Name: "org/web-app"},
					{Name: "org/dashboard", Tags: []string{"internal"}},
				},
			},
		},
	}
}

func TestFilteredRepos_NoFilter(t *testing.T) {
	ws := testWorkspace()
	got := workspace.FilteredRepos(ws, workspace.InspectOptions{})

	if len(got) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(got))
	}
	if len(got[0].Repositories) != 2 {
		t.Errorf("expected 2 repos in project 0, got %d", len(got[0].Repositories))
	}
	if len(got[1].Repositories) != 2 {
		t.Errorf("expected 2 repos in project 1, got %d", len(got[1].Repositories))
	}
}

func TestFilteredRepos_ByProject(t *testing.T) {
	ws := testWorkspace()
	got := workspace.FilteredRepos(ws, workspace.InspectOptions{ProjectName: "frontend"})

	if len(got) != 1 {
		t.Fatalf("expected 1 project, got %d", len(got))
	}
	if got[0].Name != "frontend" {
		t.Errorf("expected project 'frontend', got %q", got[0].Name)
	}
	if len(got[0].Repositories) != 2 {
		t.Errorf("expected 2 repos, got %d", len(got[0].Repositories))
	}
}

func TestFilteredRepos_ByRepo(t *testing.T) {
	ws := testWorkspace()
	got := workspace.FilteredRepos(ws, workspace.InspectOptions{RepoName: "auth"})

	if len(got) != 1 {
		t.Fatalf("expected 1 project with matching repo, got %d", len(got))
	}
	if got[0].Name != "backend" {
		t.Errorf("expected project 'backend', got %q", got[0].Name)
	}
	if len(got[0].Repositories) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(got[0].Repositories))
	}
	if got[0].Repositories[0].Name != "org/auth-service" {
		t.Errorf("expected 'org/auth-service', got %q", got[0].Repositories[0].Name)
	}
}

func TestFilteredRepos_ByTags(t *testing.T) {
	ws := testWorkspace()
	// Tag "security" is only on auth-service repo; project "backend" has tag "go"
	// which all its repos inherit. Filter by "security" should match auth-service
	// (repo-level tag) only.
	got := workspace.FilteredRepos(ws, workspace.InspectOptions{Tags: []string{"security"}})

	if len(got) != 1 {
		t.Fatalf("expected 1 project, got %d", len(got))
	}
	if got[0].Name != "backend" {
		t.Errorf("expected project 'backend', got %q", got[0].Name)
	}
	if len(got[0].Repositories) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(got[0].Repositories))
	}
	if got[0].Repositories[0].Name != "org/auth-service" {
		t.Errorf("expected 'org/auth-service', got %q", got[0].Repositories[0].Name)
	}
}

func TestFilteredRepos_ByProjectTag(t *testing.T) {
	ws := testWorkspace()
	// Tag "go" is on the backend project, inherited by all its repos.
	got := workspace.FilteredRepos(ws, workspace.InspectOptions{Tags: []string{"go"}})

	if len(got) != 1 {
		t.Fatalf("expected 1 project, got %d", len(got))
	}
	if got[0].Name != "backend" {
		t.Errorf("expected project 'backend', got %q", got[0].Name)
	}
	if len(got[0].Repositories) != 2 {
		t.Errorf("expected 2 repos (both inherit 'go' tag), got %d", len(got[0].Repositories))
	}
}

func TestResolveRepoInfos_CloneStatus(t *testing.T) {
	base := t.TempDir()

	ws := config.Workspace{
		Name:      "test-ws",
		Path:      base,
		Structure: "list",
		Projects: []config.Project{
			{
				Name:   "proj",
				Branch: "main",
				Repositories: []config.Repository{
					{Name: "org/cloned-repo"},
					{Name: "org/missing-repo"},
				},
			},
		},
	}

	// Create directory for only the first repo.
	clonedPath := filepath.Join(base, "proj", "cloned-repo")
	if err := os.MkdirAll(clonedPath, 0o755); err != nil {
		t.Fatal(err)
	}

	infos := workspace.ResolveRepoInfos(ws, workspace.InspectOptions{})

	if len(infos) != 2 {
		t.Fatalf("expected 2 repo infos, got %d", len(infos))
	}

	// First repo: directory exists → Cloned = true
	if !infos[0].Cloned {
		t.Errorf("expected cloned-repo to have Cloned=true")
	}
	if infos[0].LocalPath != clonedPath {
		t.Errorf("expected LocalPath=%q, got %q", clonedPath, infos[0].LocalPath)
	}
	if infos[0].ProjectName != "proj" {
		t.Errorf("expected ProjectName='proj', got %q", infos[0].ProjectName)
	}

	// Second repo: directory does not exist → Cloned = false
	if infos[1].Cloned {
		t.Errorf("expected missing-repo to have Cloned=false")
	}
	wantMissing := filepath.Join(base, "proj", "missing-repo")
	if infos[1].LocalPath != wantMissing {
		t.Errorf("expected LocalPath=%q, got %q", wantMissing, infos[1].LocalPath)
	}
}
