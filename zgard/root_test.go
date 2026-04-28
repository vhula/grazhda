package main

import "testing"

func TestRootCmd_HasMainSubcommands(t *testing.T) {
	for _, name := range []string{"ws", "config", "pkg"} {
		if _, _, err := rootCmd.Find([]string{name}); err != nil {
			t.Fatalf("expected root subcommand %q: %v", name, err)
		}
	}
}

func TestRootCmd_HasGlobalFlags(t *testing.T) {
	for _, flag := range []string{"no-color", "quiet"} {
		if f := rootCmd.PersistentFlags().Lookup(flag); f == nil {
			t.Fatalf("expected persistent flag %q", flag)
		}
	}
}
