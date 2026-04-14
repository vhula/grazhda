package color

import (
	"testing"

	fcolor "github.com/fatih/color"
)

func TestDisableAndIsDisabled(t *testing.T) {
	prev := IsDisabled()
	t.Cleanup(func() {
		fcolor.NoColor = prev
	})

	fcolor.NoColor = false
	Disable()
	if !IsDisabled() {
		t.Fatal("expected colors to be disabled")
	}
}
