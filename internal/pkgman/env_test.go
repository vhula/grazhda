package pkgman_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vhula/grazhda/internal/pkgman"
)

func tmpEnvFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), ".grazhda.env")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// ─── UpsertBlock ────────────────────────────────────────────────────────────

func TestUpsertBlock_CreateNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".grazhda.env")
	if err := pkgman.UpsertBlock(path, "myapp", "export FOO=bar\n"); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "# === BEGIN GRAZHDA: myapp ===") {
		t.Errorf("begin marker missing:\n%s", got)
	}
	if !strings.Contains(got, "export FOO=bar") {
		t.Errorf("content missing:\n%s", got)
	}
	if !strings.Contains(got, "# === END GRAZHDA: myapp ===") {
		t.Errorf("end marker missing:\n%s", got)
	}
}

func TestUpsertBlock_AppendToExisting(t *testing.T) {
	path := tmpEnvFile(t, "export EXISTING=yes\n")
	if err := pkgman.UpsertBlock(path, "tool", "export TOOL_HOME=/opt/tool\n"); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if !strings.Contains(got, "export EXISTING=yes") {
		t.Errorf("original content missing:\n%s", got)
	}
	if !strings.Contains(got, "export TOOL_HOME=/opt/tool") {
		t.Errorf("new content missing:\n%s", got)
	}
}

func TestUpsertBlock_ReplacesExistingBlock(t *testing.T) {
	initial := "# === BEGIN GRAZHDA: tool ===\nexport OLD=yes\n# === END GRAZHDA: tool ===\n"
	path := tmpEnvFile(t, initial)
	if err := pkgman.UpsertBlock(path, "tool", "export NEW=yes\n"); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if strings.Contains(got, "export OLD=yes") {
		t.Errorf("old content should be replaced:\n%s", got)
	}
	if !strings.Contains(got, "export NEW=yes") {
		t.Errorf("new content missing:\n%s", got)
	}
	// Only one begin marker should exist.
	if strings.Count(got, "BEGIN GRAZHDA: tool") != 1 {
		t.Errorf("expected exactly one begin marker:\n%s", got)
	}
}

func TestUpsertBlock_Idempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".grazhda.env")
	content := "export FOO=bar\n"
	if err := pkgman.UpsertBlock(path, "pkg", content); err != nil {
		t.Fatal(err)
	}
	first := readFile(t, path)
	if err := pkgman.UpsertBlock(path, "pkg", content); err != nil {
		t.Fatal(err)
	}
	second := readFile(t, path)
	if first != second {
		t.Errorf("not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// ─── RemoveBlock ─────────────────────────────────────────────────────────────

func TestRemoveBlock_RemovesBlock(t *testing.T) {
	initial := "before\n# === BEGIN GRAZHDA: sdk ===\nexport SDK=1\n# === END GRAZHDA: sdk ===\nafter\n"
	path := tmpEnvFile(t, initial)
	if err := pkgman.RemoveBlock(path, "sdk"); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if strings.Contains(got, "BEGIN GRAZHDA: sdk") {
		t.Errorf("block should be removed:\n%s", got)
	}
	if !strings.Contains(got, "before") || !strings.Contains(got, "after") {
		t.Errorf("surrounding content should be preserved:\n%s", got)
	}
}

func TestRemoveBlock_MissingBlockIsNoop(t *testing.T) {
	path := tmpEnvFile(t, "export A=1\n")
	if err := pkgman.RemoveBlock(path, "nonexistent"); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, path)
	if got != "export A=1\n" {
		t.Errorf("file should be unchanged:\n%s", got)
	}
}

func TestRemoveBlock_MissingFileIsNoop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no.env")
	if err := pkgman.RemoveBlock(path, "anything"); err != nil {
		t.Fatal(err)
	}
}

// ─── HasBlock ────────────────────────────────────────────────────────────────

func TestHasBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".grazhda.env")
	present, err := pkgman.HasBlock(path, "x")
	if err != nil || present {
		t.Fatalf("expected false/nil for missing file, got present=%v err=%v", present, err)
	}

	_ = pkgman.UpsertBlock(path, "x", "export X=1\n")
	present, err = pkgman.HasBlock(path, "x")
	if err != nil || !present {
		t.Fatalf("expected true/nil, got present=%v err=%v", present, err)
	}

	_ = pkgman.RemoveBlock(path, "x")
	present, err = pkgman.HasBlock(path, "x")
	if err != nil || present {
		t.Fatalf("expected false/nil after remove, got present=%v err=%v", present, err)
	}
}
