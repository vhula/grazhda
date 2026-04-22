package path_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/path"
)

func TestGrazhdaDir_FromEnv(t *testing.T) {
	t.Setenv("GRAZHDA_DIR", "/custom/grazhda")

	dir, err := path.GrazhdaDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != "/custom/grazhda" {
		t.Errorf("expected /custom/grazhda, got %q", dir)
	}
}

func TestGrazhdaDir_DefaultHome(t *testing.T) {
	t.Setenv("GRAZHDA_DIR", "")

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot determine home dir: %v", err)
	}

	dir, err := path.GrazhdaDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, ".grazhda")
	if dir != want {
		t.Errorf("expected %q, got %q", want, dir)
	}
}

func TestConfigPath_FromEnv(t *testing.T) {
	t.Setenv("GRAZHDA_DIR", "/custom/grazhda")

	p := path.ConfigPath()
	want := filepath.Join("/custom/grazhda", "config.yaml")
	if p != want {
		t.Errorf("expected %q, got %q", want, p)
	}
}

func TestConfigPath_DefaultHome(t *testing.T) {
	t.Setenv("GRAZHDA_DIR", "")

	p := path.ConfigPath()
	if !strings.HasSuffix(p, filepath.Join(".grazhda", "config.yaml")) {
		t.Errorf("expected path ending in .grazhda/config.yaml, got %q", p)
	}
}
