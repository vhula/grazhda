package pkgman_test

import (
	"testing"

	"github.com/vhula/grazhda/internal/pkgman"
)

func TestPkgLabel_NoVersion(t *testing.T) {
	got := pkgman.PkgLabel(pkgman.Package{Name: "myapp"})
	if got != "myapp" {
		t.Errorf("PkgLabel with no version: want %q, got %q", "myapp", got)
	}
}

func TestPkgLabel_WithVersion(t *testing.T) {
	got := pkgman.PkgLabel(pkgman.Package{Name: "myapp", Version: "1.2.3"})
	if got != "myapp@1.2.3" {
		t.Errorf("PkgLabel with version: want %q, got %q", "myapp@1.2.3", got)
	}
}

func TestPkgLabel_EmptyVersion(t *testing.T) {
	got := pkgman.PkgLabel(pkgman.Package{Name: "tool", Version: ""})
	if got != "tool" {
		t.Errorf("PkgLabel with empty version: want %q, got %q", "tool", got)
	}
}
