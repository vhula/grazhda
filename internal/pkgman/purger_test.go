package pkgman_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/pkgman"
)

// ─── Purge ──────────────────────────────────────────────────────────────────

func TestPurge_SinglePackageWithScript(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "removeme", Purge: `echo "purging removeme"`},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, true)

	if err := p.Purge(context.Background(), nil); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}
	if !strings.Contains(out.String(), "Purging removeme") {
		t.Errorf("output should contain Purging header, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "purging removeme") {
		t.Errorf("verbose output should contain script stdout, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "removeme purged") {
		t.Errorf("output should contain purged confirmation, got:\n%s", out.String())
	}
}

func TestPurge_PreCreateDir_RemovesDirectory(t *testing.T) {
	dir := setupGrazhdaDir(t)

	// Create the package directory as the installer would.
	pkgDir := pkgman.PkgDir(dir, "mylib")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Put a file inside to verify recursive removal.
	if err := os.WriteFile(filepath.Join(pkgDir, "data.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "mylib", PreCreateDir: true},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, true)

	if err := p.Purge(context.Background(), nil); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}

	if _, err := os.Stat(pkgDir); !os.IsNotExist(err) {
		t.Errorf("package dir should be removed after purge")
	}
	if !strings.Contains(out.String(), "removed") {
		t.Errorf("output should mention directory removal, got:\n%s", out.String())
	}
}

func TestPurge_RemovesEnvBlocks(t *testing.T) {
	dir := setupGrazhdaDir(t)

	// Pre-populate env file with both pre and post blocks as installer would.
	envPath := pkgman.EnvPath(dir)
	if err := pkgman.UpsertBlock(envPath, "sdk:pre", `export SDK_HOME="/opt/sdk"`); err != nil {
		t.Fatal(err)
	}
	if err := pkgman.UpsertBlock(envPath, "sdk:post", `export PATH="$SDK_HOME/bin:$PATH"`); err != nil {
		t.Fatal(err)
	}
	// Verify blocks exist before purge.
	content := readFile(t, envPath)
	if !strings.Contains(content, "sdk:pre") || !strings.Contains(content, "sdk:post") {
		t.Fatalf("env blocks should exist before purge, got:\n%s", content)
	}

	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{
			Name:           "sdk",
			PreInstallEnv:  `export SDK_HOME="/opt/sdk"`,
			PostInstallEnv: `export PATH="$SDK_HOME/bin:$PATH"`,
		},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, true)

	if err := p.Purge(context.Background(), nil); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}

	content = readFile(t, envPath)
	if strings.Contains(content, "sdk:pre") {
		t.Errorf("pre env block should be removed, got:\n%s", content)
	}
	if strings.Contains(content, "sdk:post") {
		t.Errorf("post env block should be removed, got:\n%s", content)
	}
	if !strings.Contains(out.String(), "removed env block") {
		t.Errorf("output should mention env block removal, got:\n%s", out.String())
	}
}

func TestPurge_Verbose_ShowsScriptOutput(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "chatty", Purge: `echo "visible purge output"`},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, true)

	if err := p.Purge(context.Background(), nil); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}
	if !strings.Contains(out.String(), "visible purge output") {
		t.Errorf("verbose=true should show purge script output, got:\n%s", out.String())
	}
}

func TestPurge_NonVerbose_SuppressesScriptOutput(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "quiet", Purge: `echo "hidden purge"`},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, false)

	if err := p.Purge(context.Background(), nil); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}
	if strings.Contains(out.String(), "hidden purge") {
		t.Errorf("verbose=false should suppress script stdout, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Purging quiet") {
		t.Errorf("non-verbose should still show Purging header, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "quiet purged") {
		t.Errorf("non-verbose should still show purged confirmation, got:\n%s", out.String())
	}
}

func TestPurge_NonVerbose_SpinnerStatusOnSuccess(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "spinpkg", Purge: `echo ok`},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, false)

	if err := p.Purge(context.Background(), nil); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}
	if !strings.Contains(errOut.String(), "[purge] done") {
		t.Errorf("spinner should show done status on errOut, got:\n%s", errOut.String())
	}
}

func TestPurge_NonVerbose_SpinnerStatusOnFailure(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "broken", Purge: `exit 1`},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, false)

	err := p.Purge(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for failing purge script")
	}
	if !strings.Contains(errOut.String(), "[purge] failed") {
		t.Errorf("spinner should show failed status on errOut, got:\n%s", errOut.String())
	}
}

func TestPurge_MultiplePackages_ReverseDependencyOrder(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	// b depends on a → purge order should be b first, then a.
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "a", Purge: `echo "purging a"`},
		{Name: "b", DependsOn: []string{"a"}, Purge: `echo "purging b"`},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, true)

	if err := p.Purge(context.Background(), nil); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}
	output := out.String()
	posB := strings.Index(output, "Purging b")
	posA := strings.Index(output, "Purging a")
	if posB < 0 || posA < 0 {
		t.Fatalf("expected both packages in output, got:\n%s", output)
	}
	if posB >= posA {
		t.Errorf("dependent b should be purged before dependency a, got:\n%s", output)
	}
}

func TestPurge_FailingScript_ReturnsError(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "broken", Purge: `exit 1`},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, true)

	err := p.Purge(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for failing purge script")
	}
	if !strings.Contains(err.Error(), "broken") {
		t.Errorf("error should mention package name, got: %v", err)
	}
}

func TestPurge_SubsetSelection(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "keep", Purge: `echo "keep"`},
		{Name: "remove", Purge: `echo "remove"`},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, true)

	if err := p.Purge(context.Background(), []string{"remove"}); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Purging remove") {
		t.Errorf("should purge requested package, got:\n%s", output)
	}
	if strings.Contains(output, "Purging keep") {
		t.Errorf("should NOT purge unrequested package, got:\n%s", output)
	}
}

func TestPurge_NoEnvFile_NoPanic(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	// Package with env blocks but no .grazhda.env file — should not error.
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{
			Name:           "ghost",
			PreInstallEnv:  `export X=1`,
			PostInstallEnv: `export Y=2`,
		},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, true)

	if err := p.Purge(context.Background(), nil); err != nil {
		t.Fatalf("Purge should not fail when env file is absent: %v", err)
	}
}

func TestPurge_VersionedPackage(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "tool", Version: "2.0", Purge: `echo bye`},
	}}
	p := pkgman.NewPurger(dir, r, &out, &errOut, true)

	if err := p.Purge(context.Background(), nil); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}
	if !strings.Contains(out.String(), "Purging tool@2.0") {
		t.Errorf("output should use versioned label, got:\n%s", out.String())
	}
}
