package ws

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirm_Approve(t *testing.T) {
	var buf bytes.Buffer
	reader := strings.NewReader("y\n")
	result := confirm(&buf, reader, "Delete?", []string{"/tmp/a"})
	if !result {
		t.Error("expected confirm to return true for 'y' input")
	}
}

func TestConfirm_Reject(t *testing.T) {
	var buf bytes.Buffer
	reader := strings.NewReader("n\n")
	result := confirm(&buf, reader, "Delete?", []string{"/tmp/a"})
	if result {
		t.Error("expected confirm to return false for 'n' input")
	}
}

func TestConfirm_DefaultReject(t *testing.T) {
	var buf bytes.Buffer
	reader := strings.NewReader("\n")
	result := confirm(&buf, reader, "Delete?", []string{"/tmp/a"})
	if result {
		t.Error("expected confirm to return false for empty input")
	}
}

func TestCommonAncestor_Single(t *testing.T) {
	got := commonAncestor([]string{"/home/user/ws/backend/api"})
	if got != "/home/user/ws/backend/api" {
		t.Errorf("expected single path unchanged, got %q", got)
	}
}

func TestCommonAncestor_SameProject(t *testing.T) {
	paths := []string{
		"/home/user/ws/backend/api",
		"/home/user/ws/backend/service",
	}
	got := commonAncestor(paths)
	if got != "/home/user/ws/backend" {
		t.Errorf("expected common parent /home/user/ws/backend, got %q", got)
	}
}

func TestCommonAncestor_MultiProject(t *testing.T) {
	paths := []string{
		"/home/user/ws/backend/api",
		"/home/user/ws/frontend/web",
	}
	got := commonAncestor(paths)
	if got != "/home/user/ws" {
		t.Errorf("expected workspace root /home/user/ws, got %q", got)
	}
}

func TestCommonAncestor_Empty(t *testing.T) {
	got := commonAncestor([]string{})
	if got != "" {
		t.Errorf("expected empty string for no paths, got %q", got)
	}
}

func TestPluralRepos(t *testing.T) {
	if pluralRepos(1) != "repository" {
		t.Error("expected singular for 1")
	}
	if pluralRepos(2) != "repositories" {
		t.Error("expected plural for 2")
	}
}
