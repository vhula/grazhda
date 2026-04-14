package ui

import (
	"strings"
	"testing"
)

func TestRender_Empty(t *testing.T) {
	if got := Render(""); got != "" {
		t.Fatalf("expected empty render for empty input, got %q", got)
	}
}

func TestRender_MarkdownIncludesContent(t *testing.T) {
	got := Render("# Hello\n\nworld")
	if !strings.Contains(strings.ToLower(got), "hello") {
		t.Fatalf("expected rendered output to contain markdown content, got %q", got)
	}
}
