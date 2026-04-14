package main

import (
	"fmt"
	"os"

	icolor "github.com/vhula/grazhda/internal/color"
)

// printErr writes a red "✗ <msg>" line to stderr.
func printErr(msg string) {
	fmt.Fprintln(os.Stderr, icolor.Red("✗ "+msg))
}
