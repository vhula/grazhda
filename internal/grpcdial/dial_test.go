package grpcdial

import (
	"testing"

	"github.com/vhula/grazhda/internal/config"
)

func TestAddr_Defaults(t *testing.T) {
	cfg := &config.Config{}
	if got := Addr(cfg); got != "localhost:50501" {
		t.Fatalf("Addr default = %q, want localhost:50501", got)
	}
}

func TestAddr_Configured(t *testing.T) {
	cfg := &config.Config{Dukh: config.DukhConfig{Host: "127.0.0.1", Port: 9000}}
	if got := Addr(cfg); got != "127.0.0.1:9000" {
		t.Fatalf("Addr configured = %q, want 127.0.0.1:9000", got)
	}
}

func TestDefaultAddr(t *testing.T) {
	if got := DefaultAddr(); got != "localhost:50501" {
		t.Fatalf("DefaultAddr = %q, want localhost:50501", got)
	}
}

func TestDialCreatesClientConn(t *testing.T) {
	conn, err := Dial("127.0.0.1:65535")
	if err != nil {
		t.Fatalf("Dial should create lazy conn without immediate network check: %v", err)
	}
	_ = conn.Close()
}
