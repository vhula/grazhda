package cfg

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/vhula/grazhda/internal/executor"
)

func TestResolveConfigPath_UsesGrazhdaDir(t *testing.T) {
	t.Setenv("GRAZHDA_DIR", "/tmp/grazhda-test")
	got := resolveConfigPath()
	want := filepath.Join("/tmp/grazhda-test", "config.yaml")
	if got != want {
		t.Fatalf("resolveConfigPath = %q, want %q", got, want)
	}
}

func TestNewCmd_HasSubcommands(t *testing.T) {
	cmd := NewCmd()
	for _, name := range []string{"path", "validate", "list", "get", "edit", "replace", "merge"} {
		if _, _, err := cmd.Find([]string{name}); err != nil {
			t.Fatalf("expected subcommand %q: %v", name, err)
		}
	}
}

func TestPathCommand_PrintsResolvedPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	cmd := newPathCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute path cmd: %v", err)
	}

	got := strings.TrimSpace(out.String())
	want := filepath.Join(dir, "config.yaml")
	if got != want {
		t.Fatalf("path output = %q, want %q", got, want)
	}
}

func TestResolveConfigPath_FallbackUsesHome(t *testing.T) {
	t.Setenv("GRAZHDA_DIR", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	if runtimeHome := os.Getenv("HOME"); runtimeHome != home {
		t.Fatalf("expected HOME=%q, got %q", home, runtimeHome)
	}
	got := resolveConfigPath()
	want := filepath.Join(home, ".grazhda", "config.yaml")
	if got != want {
		t.Fatalf("resolveConfigPath fallback = %q, want %q", got, want)
	}
}

func findSubcommand(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
	t.Helper()
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	t.Fatalf("subcommand %q not found", name)
	return nil
}

func TestValidateCmd_HasCorrectUse(t *testing.T) {
	cmd := findSubcommand(t, NewCmd(), "validate")
	if cmd.Use != "validate" {
		t.Fatalf("validate Use = %q, want %q", cmd.Use, "validate")
	}
}

func TestListCmd_HasCorrectUse(t *testing.T) {
	cmd := findSubcommand(t, NewCmd(), "list")
	if cmd.Use != "list" {
		t.Fatalf("list Use = %q, want %q", cmd.Use, "list")
	}
}

func TestGetCmd_HasCorrectUse(t *testing.T) {
	cmd := findSubcommand(t, NewCmd(), "get")
	if cmd.Use != "get <key>" {
		t.Fatalf("get Use = %q, want %q", cmd.Use, "get <key>")
	}
}

func TestGetCmd_RequiresArg(t *testing.T) {
	parent := NewCmd()
	parent.SetArgs([]string{"get"})
	var out bytes.Buffer
	parent.SetOut(&out)
	parent.SetErr(&out)
	err := parent.Execute()
	if err == nil {
		t.Fatal("expected error when get is called with no args")
	}
}

func TestEditCmd_HasCorrectUse(t *testing.T) {
	cmd := findSubcommand(t, NewCmd(), "edit")
	if cmd.Use != "edit" {
		t.Fatalf("edit Use = %q, want %q", cmd.Use, "edit")
	}
}

func TestEditCmd_ErrorWhenConfigMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	mock := &executor.MockExecutor{}
	cmd := newEditCmd(mock)
	cmd.SetArgs(nil)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when config file does not exist")
	}
	if len(mock.Calls) != 0 {
		t.Fatalf("expected no executor calls when config missing, got %v", mock.Calls)
	}
}

func TestEditCmd_InvokesEditorWithConfigPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("editor: myeditor\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	mock := &executor.MockExecutor{}
	cmd := newEditCmd(mock)
	cmd.SetArgs(nil)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("edit cmd: %v", err)
	}

	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 executor call, got %d", len(mock.Calls))
	}
	if !strings.Contains(mock.Calls[0], "myeditor") {
		t.Errorf("expected call to contain editor name, got %q", mock.Calls[0])
	}
	if !strings.Contains(mock.Calls[0], cfgPath) {
		t.Errorf("expected call to contain config path, got %q", mock.Calls[0])
	}
}

func TestEditCmd_FallsBackToEnvEditor(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	t.Setenv("EDITOR", "nano")
	t.Setenv("VISUAL", "")
	t.Setenv("GRAZHDA_EDITOR", "")

	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("editor: \"\"\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	mock := &executor.MockExecutor{}
	cmd := newEditCmd(mock)
	cmd.SetArgs(nil)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("edit cmd: %v", err)
	}

	if len(mock.Calls) != 1 || !strings.HasPrefix(mock.Calls[0], "nano ") {
		t.Errorf("expected call starting with 'nano', got %v", mock.Calls)
	}
}

// ── replace / merge helpers ──────────────────────────────────────────────────

const testValidYAML = `
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

const testValidYAML2 = `
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

const testInvalidYAML = `
workspaces:
  - name: ""
    path: /tmp/ws
    clone_command_template: "git clone {{.RepoName}} {{.DestDir}}"
`

func writeCfg(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return p
}

func writeFromFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write from-file: %v", err)
	}
	return p
}

// ── replace ──────────────────────────────────────────────────────────────────

func TestReplaceCmd_HasCorrectUse(t *testing.T) {
	cmd := findSubcommand(t, NewCmd(), "replace")
	if cmd.Use != "replace" {
		t.Fatalf("replace Use = %q, want %q", cmd.Use, "replace")
	}
}

func TestReplaceCmd_RequiresFromFile(t *testing.T) {
	cmd := newReplaceCmd()
	cmd.SetArgs(nil)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --from-file is missing")
	}
}

func TestReplaceCmd_SucceedsWithValidFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	writeCfg(t, dir, testValidYAML)
	from := writeFromFile(t, dir, "new.yaml", testValidYAML2)

	cmd := newReplaceCmd()
	cmd.SetArgs([]string{"--from-file", from})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("replace: %v (output: %s)", err, out.String())
	}
	if !strings.Contains(out.String(), "Config replaced") {
		t.Errorf("expected success message, got: %s", out.String())
	}
}

func TestReplaceCmd_CreatesBackupFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	cfgPath := writeCfg(t, dir, testValidYAML)
	from := writeFromFile(t, dir, "new.yaml", testValidYAML2)

	cmd := newReplaceCmd()
	cmd.SetArgs([]string{"--from-file", from})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute() //nolint:errcheck

	if _, err := os.Stat(cfgPath + ".bak"); os.IsNotExist(err) {
		t.Error(".bak file does not exist after replace")
	}
}

func TestReplaceCmd_BackupContainsOriginal(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	cfgPath := writeCfg(t, dir, testValidYAML)
	orig, _ := os.ReadFile(cfgPath)
	from := writeFromFile(t, dir, "new.yaml", testValidYAML2)

	cmd := newReplaceCmd()
	cmd.SetArgs([]string{"--from-file", from})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute() //nolint:errcheck

	bak, _ := os.ReadFile(cfgPath + ".bak")
	if string(orig) != string(bak) {
		t.Error(".bak content differs from original")
	}
}

func TestReplaceCmd_OriginalPreservedOnValidationFailure(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	cfgPath := writeCfg(t, dir, testValidYAML)
	orig, _ := os.ReadFile(cfgPath)
	from := writeFromFile(t, dir, "bad.yaml", testInvalidYAML)

	cmd := newReplaceCmd()
	cmd.SetArgs([]string{"--from-file", from})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute() //nolint:errcheck

	got, _ := os.ReadFile(cfgPath)
	if string(orig) != string(got) {
		t.Error("original config was modified despite validation failure")
	}
}

// ── merge ────────────────────────────────────────────────────────────────────

func TestMergeCmd_HasCorrectUse(t *testing.T) {
	cmd := findSubcommand(t, NewCmd(), "merge")
	if cmd.Use != "merge" {
		t.Fatalf("merge Use = %q, want %q", cmd.Use, "merge")
	}
}

func TestMergeCmd_RequiresFromFile(t *testing.T) {
	cmd := newMergeCmd()
	cmd.SetArgs(nil)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when --from-file is missing")
	}
}

func TestMergeCmd_SucceedsWithValidPatch(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	writeCfg(t, dir, testValidYAML)
	from := writeFromFile(t, dir, "patch.yaml", "editor: nano\n")

	cmd := newMergeCmd()
	cmd.SetArgs([]string{"--from-file", from})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("merge: %v (output: %s)", err, out.String())
	}
	if !strings.Contains(out.String(), "Config merged") {
		t.Errorf("expected success message, got: %s", out.String())
	}
}

func TestMergeCmd_CreatesBackupFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	cfgPath := writeCfg(t, dir, testValidYAML)
	from := writeFromFile(t, dir, "patch.yaml", "editor: nano\n")

	cmd := newMergeCmd()
	cmd.SetArgs([]string{"--from-file", from})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute() //nolint:errcheck

	if _, err := os.Stat(cfgPath + ".bak"); os.IsNotExist(err) {
		t.Error(".bak file does not exist after merge")
	}
}

func TestMergeCmd_AppliesPatch(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	cfgPath := writeCfg(t, dir, testValidYAML)
	from := writeFromFile(t, dir, "patch.yaml", "editor: nano\n")

	cmd := newMergeCmd()
	cmd.SetArgs([]string{"--from-file", from})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("merge: %v", err)
	}

	data, _ := os.ReadFile(cfgPath)
	if !strings.Contains(string(data), "nano") {
		t.Errorf("expected 'nano' in merged config, got:\n%s", data)
	}
}

func TestMergeCmd_OriginalPreservedOnValidationFailure(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)
	cfgPath := writeCfg(t, dir, testValidYAML)
	orig, _ := os.ReadFile(cfgPath)
	from := writeFromFile(t, dir, "bad.yaml", testInvalidYAML)

	cmd := newMergeCmd()
	cmd.SetArgs([]string{"--from-file", from})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.Execute() //nolint:errcheck

	got, _ := os.ReadFile(cfgPath)
	if string(orig) != string(got) {
		t.Error("original config was modified despite validation failure")
	}
}
