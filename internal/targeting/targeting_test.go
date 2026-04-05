package targeting_test

import (
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/config"
	"github.com/vhula/grazhda/internal/targeting"
)

func makeConfig() *config.Config {
	return &config.Config{
		Workspaces: []config.Workspace{
			{Name: "default", Default: true, Path: "/tmp/default"},
			{Name: "myws", Path: "/tmp/myws"},
			{Name: "other", Path: "/tmp/other"},
		},
	}
}

func TestResolve_DefaultWorkspace(t *testing.T) {
	cfg := makeConfig()
	wss, err := targeting.Resolve(cfg, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wss) != 1 || wss[0].Name != "default" {
		t.Errorf("expected default workspace, got %v", wss)
	}
}

func TestResolve_NamedWorkspace(t *testing.T) {
	cfg := makeConfig()
	wss, err := targeting.Resolve(cfg, "myws", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wss) != 1 || wss[0].Name != "myws" {
		t.Errorf("expected myws, got %v", wss)
	}
}

func TestResolve_NotFound(t *testing.T) {
	cfg := makeConfig()
	_, err := targeting.Resolve(cfg, "nonexistent", false)
	if err == nil {
		t.Fatal("expected error for nonexistent workspace")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("expected error to mention workspace name, got: %v", err)
	}
}

func TestResolve_All(t *testing.T) {
	cfg := makeConfig()
	wss, err := targeting.Resolve(cfg, "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wss) != 3 {
		t.Errorf("expected all 3 workspaces, got %d", len(wss))
	}
}

func TestResolve_MutuallyExclusive(t *testing.T) {
	cfg := makeConfig()
	_, err := targeting.Resolve(cfg, "myws", true)
	if err == nil {
		t.Fatal("expected error for --ws + --all")
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
	_, err := targeting.Resolve(cfg, "", false)
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

	wss, err := targeting.Resolve(cfg, "", false)
	if err != nil {
		t.Fatalf("default resolve: %v", err)
	}
	if len(wss) != 1 || wss[0].Name != "default" {
		t.Errorf("expected default workspace, got %v", wss)
	}

	wss, err = targeting.Resolve(cfg, "secondary", false)
	if err != nil {
		t.Fatalf("named resolve: %v", err)
	}
	if wss[0].Name != "secondary" {
		t.Errorf("expected secondary, got %v", wss)
	}

	wss, err = targeting.Resolve(cfg, "", true)
	if err != nil {
		t.Fatalf("all resolve: %v", err)
	}
	if len(wss) != 2 {
		t.Errorf("expected 2 workspaces, got %d", len(wss))
	}
}
