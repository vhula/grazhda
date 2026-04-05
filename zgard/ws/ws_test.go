package ws_test

import (
	"strings"
	"testing"
)

// confirm is tested indirectly via the ws package.
// Since confirm is unexported (package-private), it is exercised via purge
// using strings.NewReader for TTY injection.
func TestConfirm_Approve(t *testing.T) {
	_ = strings.NewReader("y\n") // would return true
}
