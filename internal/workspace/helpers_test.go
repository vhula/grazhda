package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/workspace"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot determine home dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{"empty", "", ""},
		{"no tilde", "/usr/local/bin", "/usr/local/bin"},
		{"tilde only", "~", home},
		{"tilde with subpath", "~/projects/foo", filepath.Join(home, "projects/foo")},
		{"tilde mid-path unchanged", "/home/~user", "/home/~user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := workspace.ExpandHome(tt.path)
			if got != tt.want {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestResolveDestName(t *testing.T) {
	tests := []struct {
		name         string
		repoName     string
		localDirName string
		structure    string
		want         string
	}{
		{
			name:         "localDirName takes priority",
			repoName:     "org/repo",
			localDirName: "my-custom-dir",
			structure:    "tree",
			want:         "my-custom-dir",
		},
		{
			name:         "tree structure returns full name",
			repoName:     "org/pack/repo",
			localDirName: "",
			structure:    "tree",
			want:         "org/pack/repo",
		},
		{
			name:         "list structure returns last segment",
			repoName:     "org/pack/repo",
			localDirName: "",
			structure:    "list",
			want:         "repo",
		},
		{
			name:         "list strips .git suffix",
			repoName:     "org/repo.git",
			localDirName: "",
			structure:    "list",
			want:         "repo",
		},
		{
			name:         "list no slash",
			repoName:     "simple-repo",
			localDirName: "",
			structure:    "list",
			want:         "simple-repo",
		},
		{
			name:         "localDirName overrides even in list mode",
			repoName:     "org/repo.git",
			localDirName: "override",
			structure:    "list",
			want:         "override",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := workspace.ResolveDestName("", tt.repoName, tt.localDirName, tt.structure)
			if got != tt.want {
				t.Errorf("ResolveDestName(%q, %q, %q) = %q, want %q",
					tt.repoName, tt.localDirName, tt.structure, got, tt.want)
			}
		})
	}
}

func TestResolveDestNamesForProject(t *testing.T) {
	repos := []config.Repository{
		{Name: "org/backend-api", LocalDirName: ""},
		{Name: "org/frontend.git", LocalDirName: ""},
		{Name: "org/infra/tools", LocalDirName: "custom-tools"},
	}

	t.Run("list structure", func(t *testing.T) {
		got := workspace.ResolveDestNamesForProject(repos, "list")
		want := []string{"backend-api", "frontend", "custom-tools"}
		if len(got) != len(want) {
			t.Fatalf("got %d names, want %d", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("tree structure", func(t *testing.T) {
		got := workspace.ResolveDestNamesForProject(repos, "tree")
		want := []string{"org/backend-api", "org/frontend.git", "custom-tools"}
		if len(got) != len(want) {
			t.Fatalf("got %d names, want %d", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
			}
		}
	})
}
