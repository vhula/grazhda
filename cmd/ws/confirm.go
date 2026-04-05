package ws

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// confirm prints the prompt listing paths and asks the user to confirm deletion.
// reader is injected to allow testing without a real TTY.
func confirm(out io.Writer, reader io.Reader, paths []string) bool {
	fmt.Fprintln(out, "The following directories will be removed:")
	for _, p := range paths {
		fmt.Fprintf(out, "  %s\n", p)
	}
	fmt.Fprint(out, "Confirm? [y/N]: ")

	scanner := bufio.NewScanner(reader)
	if scanner.Scan() {
		answer := strings.TrimSpace(scanner.Text())
		return strings.EqualFold(answer, "y")
	}
	return false
}
