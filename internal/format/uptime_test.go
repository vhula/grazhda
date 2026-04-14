package format

import (
	"testing"
	"time"
)

func TestUptime(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want string
	}{
		{"seconds", 17 * time.Second, "17s"},
		{"minutes", 3*time.Minute + 42*time.Second, "3m 42s"},
		{"hours", 2*time.Hour + 15*time.Minute + 10*time.Second, "2h 15m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uptime(tt.in); got != tt.want {
				t.Fatalf("Uptime(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
