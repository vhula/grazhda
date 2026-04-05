package ws

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	clr "github.com/vhula/grazhda/internal/color"
)

// confirm prints msg, lists paths, then asks the user to confirm.
// reader is injected to allow testing without a real TTY.
func confirm(out io.Writer, reader io.Reader, msg string, paths []string) bool {
	fmt.Fprintln(out, clr.Yellow(msg))
	for _, p := range paths {
		fmt.Fprintf(out, "  %s\n", clr.Yellow(p))
	}
	fmt.Fprint(out, clr.Blue("Confirm? [y/N]: "))

	scanner := bufio.NewScanner(reader)
	if scanner.Scan() {
		answer := strings.TrimSpace(scanner.Text())
		return strings.EqualFold(answer, "y")
	}
	return false
}
