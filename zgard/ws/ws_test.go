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
