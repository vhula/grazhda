package pkg_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vhula/grazhda/zgard/pkg"
)

func TestPurgeCmd_HasCorrectFlags(t *testing.T) {
	cmd := pkg.NewCmd()
	purgeCmd, _, err := cmd.Find([]string{"purge"})
	if err != nil {
		t.Fatalf("find purge: %v", err)
	}
	for _, name := range []string{"name", "all", "verbose"} {
		if purgeCmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag --%s on purge command", name)
		}
	}
}

func TestPurgeCmd_RequiresNameOrAll(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GRAZHDA_DIR", dir)

	cmd := pkg.NewCmd()
	cmd.SetArgs([]string{"purge"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when neither --name nor --all is provided")
	}
	if !strings.Contains(err.Error(), "--name") && !strings.Contains(err.Error(), "--all") {
		t.Fatalf("expected error mentioning --name or --all, got: %v", err)
	}
}
