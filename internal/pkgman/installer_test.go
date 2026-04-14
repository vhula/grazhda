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

func setupGrazhdaDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "pkgs"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// ─── Install ────────────────────────────────────────────────────────────────

func TestInstall_SinglePackageWithScript(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "hello", Install: `echo "hello from install"`},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	if err := inst.Install(context.Background(), nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	if !strings.Contains(out.String(), "Installing hello") {
		t.Errorf("output should contain Installing header, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "hello from install") {
		t.Errorf("verbose output should contain script stdout, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "hello installed") {
		t.Errorf("output should contain installed confirmation, got:\n%s", out.String())
	}
}

func TestInstall_PreCreateDir(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "mylib", PreCreateDir: true, Install: `test -d "$PKG_DIR"`},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	if err := inst.Install(context.Background(), nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	pkgDir := pkgman.PkgDir(dir, "mylib")
	info, err := os.Stat(pkgDir)
	if err != nil {
		t.Fatalf("package dir should exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected directory, got file")
	}
	if !strings.Contains(out.String(), "created") {
		t.Errorf("output should mention directory creation, got:\n%s", out.String())
	}
}

func TestInstall_PreAndPostInstallEnv(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{
			Name:          "sdk",
			PreInstallEnv: `export SDK_HOME="$GRAZHDA_DIR/pkgs/sdk"`,
			Install:       `echo "SDK_HOME is $SDK_HOME"`,
			PostInstallEnv: `export PATH="$SDK_HOME/bin:$PATH"`,
		},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	if err := inst.Install(context.Background(), nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	envPath := pkgman.EnvPath(dir)
	envContent := readFile(t, envPath)

	if !strings.Contains(envContent, "# === BEGIN GRAZHDA: sdk:pre ===") {
		t.Errorf("env file should contain pre block marker, got:\n%s", envContent)
	}
	if !strings.Contains(envContent, `export SDK_HOME=`) {
		t.Errorf("env file should contain pre-install-env content, got:\n%s", envContent)
	}
	if !strings.Contains(envContent, "# === BEGIN GRAZHDA: sdk:post ===") {
		t.Errorf("env file should contain post block marker, got:\n%s", envContent)
	}
	if !strings.Contains(envContent, `export PATH=`) {
		t.Errorf("env file should contain post-install-env content, got:\n%s", envContent)
	}

	// Verify the output mentions the env writes.
	if !strings.Contains(out.String(), "wrote pre-install-env block") {
		t.Errorf("output should mention pre-install-env write, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "wrote post-install-env block") {
		t.Errorf("output should mention post-install-env write, got:\n%s", out.String())
	}
}

func TestInstall_NoInstallScript_OnlyEnvBlocks(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{
			Name:          "envonly",
			PreInstallEnv: `export ENVONLY_VAR=1`,
		},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	if err := inst.Install(context.Background(), nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	envPath := pkgman.EnvPath(dir)
	envContent := readFile(t, envPath)
	if !strings.Contains(envContent, "ENVONLY_VAR=1") {
		t.Errorf("env file should contain env var, got:\n%s", envContent)
	}
	if !strings.Contains(out.String(), "envonly installed") {
		t.Errorf("output should confirm installation, got:\n%s", out.String())
	}
}

func TestInstall_Verbose_ShowsScriptOutput(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "chatty", Install: `echo "visible output"`},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	if err := inst.Install(context.Background(), nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	if !strings.Contains(out.String(), "visible output") {
		t.Errorf("verbose=true should show script output, got:\n%s", out.String())
	}
}

func TestInstall_NonVerbose_SuppressesScriptOutput(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "quiet", Install: `echo "hidden output"`},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, false)

	if err := inst.Install(context.Background(), nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	if strings.Contains(out.String(), "hidden output") {
		t.Errorf("verbose=false should suppress script stdout, got:\n%s", out.String())
	}
	// The status header and confirmation should still appear.
	if !strings.Contains(out.String(), "Installing quiet") {
		t.Errorf("non-verbose should still show Installing header, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "quiet installed") {
		t.Errorf("non-verbose should still show installed confirmation, got:\n%s", out.String())
	}
}

func TestInstall_NonVerbose_SpinnerStatusOnSuccess(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "spinpkg", Install: `echo ok`},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, false)

	if err := inst.Install(context.Background(), nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	// Spinner writes to errOut; after success it should print "[install] done".
	if !strings.Contains(errOut.String(), "[install] done") {
		t.Errorf("spinner should show done status on errOut, got:\n%s", errOut.String())
	}
}

func TestInstall_NonVerbose_SpinnerStatusOnFailure(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "broken", Install: `exit 1`},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, false)

	err := inst.Install(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for failing install script")
	}
	if !strings.Contains(errOut.String(), "[install] failed") {
		t.Errorf("spinner should show failed status on errOut, got:\n%s", errOut.String())
	}
}

func TestInstall_MultiplePackages_DependencyOrder(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	// b depends on a; both should install in order a, b.
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "a", Install: `echo "installing a"`},
		{Name: "b", DependsOn: []string{"a"}, Install: `echo "installing b"`},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	if err := inst.Install(context.Background(), nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	output := out.String()
	posA := strings.Index(output, "Installing a")
	posB := strings.Index(output, "Installing b")
	if posA < 0 || posB < 0 {
		t.Fatalf("expected both packages in output, got:\n%s", output)
	}
	if posA >= posB {
		t.Errorf("package a should be installed before b, got:\n%s", output)
	}
}

func TestInstall_SubsetSelection(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "base", Install: `echo "base"`},
		{Name: "tool", DependsOn: []string{"base"}, Install: `echo "tool"`},
		{Name: "other", Install: `echo "other"`},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	// Install only "tool" — should pull in "base" but not "other".
	if err := inst.Install(context.Background(), []string{"tool"}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Installing base") {
		t.Errorf("should install dependency base, got:\n%s", output)
	}
	if !strings.Contains(output, "Installing tool") {
		t.Errorf("should install requested tool, got:\n%s", output)
	}
	if strings.Contains(output, "Installing other") {
		t.Errorf("should NOT install unrequested other, got:\n%s", output)
	}
}

func TestInstall_FailingScript_ReturnsError(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "broken", Install: `exit 1`},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	err := inst.Install(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for failing install script")
	}
	if !strings.Contains(err.Error(), "broken") {
		t.Errorf("error should mention package name, got: %v", err)
	}
}

func TestInstall_PreInstallEnv_VisibleToScript(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{
			Name:          "envcheck",
			PreInstallEnv: `export MY_TEST_VAR="hello_from_pre"`,
			Install:       `echo "GOT=$MY_TEST_VAR"`,
		},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	if err := inst.Install(context.Background(), nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	if !strings.Contains(out.String(), "GOT=hello_from_pre") {
		t.Errorf("install script should see pre_install_env vars, got:\n%s", out.String())
	}
}

func TestInstall_UnknownPackage_ReturnsError(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "a"},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	err := inst.Install(context.Background(), []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown package")
	}
}

func TestInstall_VersionedPackage(t *testing.T) {
	dir := setupGrazhdaDir(t)
	var out, errOut bytes.Buffer
	r := &pkgman.Registry{Packages: []pkgman.Package{
		{Name: "tool", Version: "3.1", Install: `echo "ver=$VERSION"`},
	}}
	inst := pkgman.NewInstaller(dir, r, &out, &errOut, true)

	if err := inst.Install(context.Background(), nil); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	if !strings.Contains(out.String(), "Installing tool@3.1") {
		t.Errorf("output should use versioned label, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "ver=3.1") {
		t.Errorf("VERSION env var should be set, got:\n%s", out.String())
	}
}
