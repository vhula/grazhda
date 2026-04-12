package pkgman

import (
	"fmt"
	"os"
	"strings"
)

const (
	beginMarkerFmt = "# === BEGIN GRAZHDA: %s ==="
	endMarkerFmt   = "# === END GRAZHDA: %s ==="
)

// UpsertBlock writes the given content into a named block in the env file at
// path. If the file does not yet exist it is created. If a block for name
// already exists it is replaced in-place; otherwise the block is appended.
// The operation is idempotent: calling it twice with the same arguments
// produces the same file.
func UpsertBlock(path, name, content string) error {
	beginMarker := fmt.Sprintf(beginMarkerFmt, name)
	endMarker := fmt.Sprintf(endMarkerFmt, name)

	// Read existing file (tolerate ENOENT).
	var rawLines []string
	if data, err := os.ReadFile(path); err == nil {
		rawLines = splitLines(string(data))
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read env file %q: %w", path, err)
	}

	// Locate existing block.
	startIdx, endIdx := findBlock(rawLines, beginMarker, endMarker)

	newBlock := buildBlock(beginMarker, endMarker, content)

	var lines []string
	if startIdx >= 0 {
		// Replace existing block.
		lines = append(rawLines[:startIdx:startIdx], newBlock...)
		lines = append(lines, rawLines[endIdx+1:]...)
	} else {
		// Append block (separated by a blank line if file is non-empty).
		lines = rawLines
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, newBlock...)
	}

	return writeLines(path, lines)
}

// RemoveBlock excises the named block from the env file at path.
// If the block is not found the file is left unchanged. ENOENT is silently
// ignored (nothing to remove).
func RemoveBlock(path, name string) error {
	beginMarker := fmt.Sprintf(beginMarkerFmt, name)
	endMarker := fmt.Sprintf(endMarkerFmt, name)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read env file %q: %w", path, err)
	}

	lines := splitLines(string(data))
	startIdx, endIdx := findBlock(lines, beginMarker, endMarker)
	if startIdx < 0 {
		return nil // nothing to remove
	}

	var out []string
	out = append(out, lines[:startIdx]...)
	// Trim the trailing blank line before the block if present.
	for len(out) > 0 && strings.TrimSpace(out[len(out)-1]) == "" {
		out = out[:len(out)-1]
	}
	tail := lines[endIdx+1:]
	// Trim leading blank lines after the block.
	for len(tail) > 0 && strings.TrimSpace(tail[0]) == "" {
		tail = tail[1:]
	}
	if len(out) > 0 && len(tail) > 0 {
		out = append(out, "") // single blank separator
	}
	out = append(out, tail...)

	return writeLines(path, out)
}

// HasBlock reports whether a named block is present in the env file at path.
func HasBlock(path, name string) (bool, error) {
	beginMarker := fmt.Sprintf(beginMarkerFmt, name)
	endMarker := fmt.Sprintf(endMarkerFmt, name)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read env file %q: %w", path, err)
	}
	lines := splitLines(string(data))
	start, _ := findBlock(lines, beginMarker, endMarker)
	return start >= 0, nil
}

// ─── helpers ────────────────────────────────────────────────────────────────

func buildBlock(beginMarker, endMarker, content string) []string {
	var lines []string
	lines = append(lines, beginMarker)
	for _, l := range splitLines(strings.TrimRight(content, "\n")) {
		lines = append(lines, l)
	}
	lines = append(lines, endMarker)
	return lines
}

// findBlock returns the line indices of the begin/end markers, or (-1, -1).
func findBlock(lines []string, beginMarker, endMarker string) (startIdx, endIdx int) {
	startIdx = -1
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if startIdx < 0 && trimmed == beginMarker {
			startIdx = i
		} else if startIdx >= 0 && trimmed == endMarker {
			return startIdx, i
		}
	}
	return -1, -1
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}

func writeLines(path string, lines []string) error {
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}
	return os.WriteFile(path, []byte(content), 0644)
}
