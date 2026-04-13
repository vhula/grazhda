package pkgman_test

import (
	"testing"

	"github.com/vhula/grazhda/internal/pkgman"
)

func reg(pkgs ...pkgman.Package) *pkgman.Registry {
	return &pkgman.Registry{Packages: pkgs}
}

func pkg(name string, deps ...string) pkgman.Package {
	return pkgman.Package{Name: name, DependsOn: deps}
}

func names(ordered []pkgman.Package) []string {
	out := make([]string, len(ordered))
	for i, p := range ordered {
		out[i] = p.Name
	}
	return out
}

// ─── Resolve ────────────────────────────────────────────────────────────────

func TestResolve_Single(t *testing.T) {
	r := reg(pkg("a"))
	got, err := pkgman.Resolve(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "a" {
		t.Fatalf("expected [a], got %v", names(got))
	}
}

func TestResolve_LinearChain(t *testing.T) {
	// a → b → c  (c must come first)
	r := reg(pkg("a", "b"), pkg("b", "c"), pkg("c"))
	got, err := pkgman.Resolve(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	order := names(got)
	if order[0] != "c" || order[1] != "b" || order[2] != "a" {
		t.Fatalf("expected [c b a], got %v", order)
	}
}

func TestResolve_Diamond(t *testing.T) {
	// d → b, d → c; b → a; c → a
	r := reg(
		pkg("a"),
		pkg("b", "a"),
		pkg("c", "a"),
		pkg("d", "b", "c"),
	)
	got, err := pkgman.Resolve(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	// a must appear before b, c; b and c before d.
	order := names(got)
	pos := func(n string) int {
		for i, p := range order {
			if p == n {
				return i
			}
		}
		return -1
	}
	if pos("a") > pos("b") || pos("a") > pos("c") {
		t.Fatalf("a must precede b and c; got %v", order)
	}
	if pos("b") > pos("d") || pos("c") > pos("d") {
		t.Fatalf("b and c must precede d; got %v", order)
	}
}

func TestResolve_CycleDetected(t *testing.T) {
	r := reg(pkg("a", "b"), pkg("b", "a"))
	_, err := pkgman.Resolve(r, nil)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestResolve_UnknownDep(t *testing.T) {
	r := reg(pkg("a", "missing"))
	_, err := pkgman.Resolve(r, nil)
	if err == nil {
		t.Fatal("expected unknown-package error, got nil")
	}
}

func TestResolve_SubsetExpandsDeps(t *testing.T) {
	// Ask for only "b", but "b" depends on "a" — expect [a b].
	r := reg(pkg("a"), pkg("b", "a"), pkg("c"))
	got, err := pkgman.Resolve(r, []string{"b"})
	if err != nil {
		t.Fatal(err)
	}
	order := names(got)
	if len(order) != 2 {
		t.Fatalf("expected 2 packages, got %v", order)
	}
	if order[0] != "a" || order[1] != "b" {
		t.Fatalf("expected [a b], got %v", order)
	}
}

func pkgV(name, version string, deps ...string) pkgman.Package {
	return pkgman.Package{Name: name, Version: version, DependsOn: deps}
}

// ─── Versioned depends_on ────────────────────────────────────────────────────

func TestResolve_VersionedDep_Satisfied(t *testing.T) {
	// jdk depends on sdkman@1.2.3; sdkman is at 1.2.3 — should resolve.
	r := reg(pkgV("sdkman", "1.2.3"), pkgV("jdk", "17", "sdkman@1.2.3"))
	got, err := pkgman.Resolve(r, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	order := names(got)
	if order[0] != "sdkman" || order[1] != "jdk" {
		t.Fatalf("expected [sdkman jdk], got %v", order)
	}
}

func TestResolve_VersionedDep_Mismatch(t *testing.T) {
	// jdk requires sdkman@2.0.0 but registry has sdkman@1.2.3 — should error.
	r := reg(pkgV("sdkman", "1.2.3"), pkgV("jdk", "17", "sdkman@2.0.0"))
	_, err := pkgman.Resolve(r, nil)
	if err == nil {
		t.Fatal("expected version mismatch error, got nil")
	}
}

func TestResolve_UnversionedDep_OnVersionedPkg(t *testing.T) {
	// Unversioned dep on a package that has a version — should still resolve.
	r := reg(pkgV("sdkman", "1.2.3"), pkgV("jdk", "17", "sdkman"))
	got, err := pkgman.Resolve(r, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if names(got)[0] != "sdkman" {
		t.Fatalf("expected sdkman first, got %v", names(got))
	}
}

func TestResolveReverse_LinearChain(t *testing.T) {
	r := reg(pkg("a", "b"), pkg("b", "c"), pkg("c"))
	got, err := pkgman.ResolveReverse(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	order := names(got)
	if order[0] != "a" || order[1] != "b" || order[2] != "c" {
		t.Fatalf("expected [a b c], got %v", order)
	}
}
