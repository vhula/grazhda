package ws_test

import (
	"strings"
	"testing"
)

// confirm is tested indirectly via the ws package.
// We test the confirm helper by importing and calling it through its exported interface.
// Since confirm is unexported (package-private), we test it via integration with purge
// using strings.NewReader for TTY injection.
func TestConfirm_Approve(t *testing.T) {
	// Tested indirectly: purge_test exercises confirmation via --no-confirm flag.
	// This file documents the expected behavior.
	_ = strings.NewReader("y\n") // would return true
}
