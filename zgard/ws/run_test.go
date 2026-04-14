package ws_test

import (
	"bytes"
	"testing"

	"github.com/vhula/grazhda/zgard/ws"
)

func TestWsSubcommands_AllRegistered(t *testing.T) {
	cmd := ws.NewCmd()
	want := []string{
		"init", "pull", "stash", "checkout", "exec",
		"purge", "list", "search", "diff", "stats", "status",
	}
	registered := make(map[string]bool)
	for _, c := range cmd.Commands() {
		registered[c.Name()] = true
	}
	for _, name := range want {
		if !registered[name] {
			t.Errorf("expected subcommand %q to be registered", name)
		}
	}
}

func TestInitCmd_HasDryRunFlag(t *testing.T) {
	cmd := ws.NewCmd()
	initCmd, _, err := cmd.Find([]string{"init"})
	if err != nil {
		t.Fatalf("find init: %v", err)
	}
	if initCmd.Flags().Lookup("dry-run") == nil {
		t.Error("expected --dry-run flag on init command")
	}
}

func TestPullCmd_HasParallelFlag(t *testing.T) {
	cmd := ws.NewCmd()
	pullCmd, _, err := cmd.Find([]string{"pull"})
	if err != nil {
		t.Fatalf("find pull: %v", err)
	}
	if pullCmd.Flags().Lookup("parallel") == nil {
		t.Error("expected --parallel flag on pull command")
	}
}

func TestCheckoutCmd_RequiresArg(t *testing.T) {
	cmd := ws.NewCmd()
	cmd.SetArgs([]string{"checkout"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when checkout is called with no args")
	}
}

func TestExecCmd_RequiresArg(t *testing.T) {
	cmd := ws.NewCmd()
	cmd.SetArgs([]string{"exec"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when exec is called with no args")
	}
}
