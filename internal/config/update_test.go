package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/vhula/grazhda/internal/config"
)

// validYAML is a minimal config that passes Validate.
const validYAML = `
workspaces:
  - name: default
    path: /tmp/ws
    clone_command_template: "git clone {{.RepoName}} {{.DestDir}}"
    projects:
      - name: backend
        branch: main
        repositories:
          - name: api
`

// validYAML2 is a second valid config used as a replacement.
const validYAML2 = `
editor: nano
workspaces:
  - name: default
    path: /tmp/ws2
    clone_command_template: "git clone {{.RepoName}} {{.DestDir}}"
    projects:
      - name: frontend
        branch: main
        repositories:
          - name: ui
`

// invalidYAML produces a config that fails Validate (missing name).
const invalidYAML = `
workspaces:
  - name: ""
    path: /tmp/ws
    clone_command_template: "git clone {{.RepoName}} {{.DestDir}}"
`

// writeConfig writes content to cfgPath and returns the path.
func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return p
}

// ── Backup ───────────────────────────────────────────────────────────────────

func TestBackup_CreatesBackupWithSameContent(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)

	bak, err := config.Backup(cfgPath)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if bak != cfgPath+".bak" {
		t.Errorf("backup path = %q, want %q", bak, cfgPath+".bak")
	}
	orig, _ := os.ReadFile(cfgPath)
	got, _ := os.ReadFile(bak)
	if string(orig) != string(got) {
		t.Errorf("backup content differs from original")
	}
}

func TestBackup_FailsWhenSourceMissing(t *testing.T) {
	_, err := config.Backup("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing source")
	}
}

// ── Replace ──────────────────────────────────────────────────────────────────

func TestReplace_ValidConfig_ReplacesFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)

	if err := config.Replace(cfgPath, []byte(validYAML2)); err != nil {
		t.Fatalf("Replace: %v", err)
	}
	data, _ := os.ReadFile(cfgPath)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("loading replaced config: %v", err)
	}
	if len(cfg.Workspaces) == 0 || cfg.Workspaces[0].Path != "/tmp/ws2" {
		t.Errorf("unexpected config after replace: %s", data)
	}
}

func TestReplace_InvalidYAML_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)
	err := config.Replace(cfgPath, []byte(":\tinvalid: yaml: ["))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestReplace_InvalidConfig_ReturnsValidationError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)
	err := config.Replace(cfgPath, []byte(invalidYAML))
	if err == nil {
		t.Fatal("expected validation error")
	}
	var valErr *config.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
	if len(valErr.Errs) == 0 {
		t.Error("expected at least one validation error message")
	}
	if !errors.Is(err, config.ErrInvalid) {
		t.Error("expected error to wrap ErrInvalid")
	}
}

func TestReplace_OriginalPreservedOnValidationFailure(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)
	orig, _ := os.ReadFile(cfgPath)

	_ = config.Replace(cfgPath, []byte(invalidYAML))

	got, _ := os.ReadFile(cfgPath)
	if string(orig) != string(got) {
		t.Error("original config was modified despite validation failure")
	}
}

// ── Merge ────────────────────────────────────────────────────────────────────

func TestMerge_OverridesScalarKeys(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)

	patch := []byte(`editor: emacs`)
	if err := config.Merge(cfgPath, patch); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	cfg, _ := config.Load(cfgPath)
	if cfg.Editor != "emacs" {
		t.Errorf("editor = %q, want %q", cfg.Editor, "emacs")
	}
}

func TestMerge_PreservesUnpatchedKeys(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)

	patch := []byte(`editor: emacs`)
	if err := config.Merge(cfgPath, patch); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	cfg, _ := config.Load(cfgPath)
	if len(cfg.Workspaces) == 0 {
		t.Error("workspaces were lost during scalar-only merge")
	}
}

func TestMerge_MergesNestedMaps(t *testing.T) {
	dir := t.TempDir()
	base := `
editor: vim
dukh:
  host: localhost
  port: 8080
workspaces:
  - name: default
    path: /tmp/ws
    clone_command_template: "git clone {{.RepoName}} {{.DestDir}}"
    projects:
      - name: backend
        branch: main
        repositories:
          - name: api
`
	cfgPath := writeConfig(t, dir, base)

	patch := []byte(`
dukh:
  port: 9090
`)
	if err := config.Merge(cfgPath, patch); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	cfg, _ := config.Load(cfgPath)
	if cfg.Dukh.Port != 9090 {
		t.Errorf("dukh.port = %d, want 9090", cfg.Dukh.Port)
	}
	if cfg.Dukh.Host != "localhost" {
		t.Errorf("dukh.host = %q, want %q — unpatch key lost", cfg.Dukh.Host, "localhost")
	}
	if cfg.Editor != "vim" {
		t.Errorf("editor = %q, want %q — unpatch top-level key lost", cfg.Editor, "vim")
	}
}

func TestMerge_SlicesReplacedByPatch(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)

	patch := []byte(`
workspaces:
  - name: default
    path: /tmp/replaced
    clone_command_template: "git clone {{.RepoName}} {{.DestDir}}"
    projects:
      - name: backend
        branch: main
        repositories:
          - name: api
`)
	if err := config.Merge(cfgPath, patch); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	cfg, _ := config.Load(cfgPath)
	if cfg.Workspaces[0].Path != "/tmp/replaced" {
		t.Errorf("workspace path = %q, want /tmp/replaced", cfg.Workspaces[0].Path)
	}
}

func TestMerge_InvalidPatch_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)
	err := config.Merge(cfgPath, []byte(":\tinvalid: yaml: ["))
	if err == nil {
		t.Fatal("expected error for invalid patch YAML")
	}
}

func TestMerge_InvalidResult_ReturnsValidationError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)

	// Patch replaces workspaces with an entry missing a required name.
	patch := []byte(`
workspaces:
  - name: ""
    path: /tmp/ws
    clone_command_template: "git clone {{.RepoName}} {{.DestDir}}"
`)
	err := config.Merge(cfgPath, patch)
	if err == nil {
		t.Fatal("expected validation error")
	}
	var valErr *config.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestMerge_OriginalPreservedOnValidationFailure(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)
	orig, _ := os.ReadFile(cfgPath)

	_ = config.Merge(cfgPath, []byte(`
workspaces:
  - name: ""
    path: /tmp/ws
    clone_command_template: "git clone {{.RepoName}} {{.DestDir}}"
`))

	got, _ := os.ReadFile(cfgPath)
	if string(orig) != string(got) {
		t.Error("original config was modified despite validation failure")
	}
}

// ── Backup + Replace / Merge integration ─────────────────────────────────────

func TestBackupThenReplace_BakExistsAfterSuccess(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)

	bak, err := config.Backup(cfgPath)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if err := config.Replace(cfgPath, []byte(validYAML2)); err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if _, err := os.Stat(bak); os.IsNotExist(err) {
		t.Error(".bak file does not exist after successful replace")
	}
}

func TestBackupThenReplace_BakExistsAfterFailure(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)

	bak, err := config.Backup(cfgPath)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	_ = config.Replace(cfgPath, []byte(invalidYAML))
	if _, err := os.Stat(bak); os.IsNotExist(err) {
		t.Error(".bak file does not exist after failed replace")
	}
}

func TestBackupThenMerge_BakExistsAfterSuccess(t *testing.T) {
	dir := t.TempDir()
	cfgPath := writeConfig(t, dir, validYAML)

	bak, err := config.Backup(cfgPath)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if err := config.Merge(cfgPath, []byte(`editor: nano`)); err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if _, err := os.Stat(bak); os.IsNotExist(err) {
		t.Error(".bak file does not exist after successful merge")
	}
}
