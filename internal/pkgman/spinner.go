package pkgman

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Spinner renders a braille-style progress indicator to out.
// It is safe to call Stop from any goroutine.
type Spinner struct {
	mu      sync.Mutex
	msg     string
	out     io.Writer
	stop    chan struct{}
	stopped bool
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewSpinner creates and starts a spinner writing to out with the given message.
func NewSpinner(out io.Writer, msg string) *Spinner {
	s := &Spinner{msg: msg, out: out, stop: make(chan struct{})}
	go s.run()
	return s
}

func (s *Spinner) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.mu.Lock()
			msg := s.msg
			s.mu.Unlock()
			fmt.Fprintf(s.out, "\r  %s %s", spinnerFrames[i%len(spinnerFrames)], msg)
			i++
		}
	}
}

// UpdateMsg updates the spinner message without stopping it.
func (s *Spinner) UpdateMsg(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
}

// Stop halts the spinner and clears the spinner line, optionally writing
// a completion symbol and message.
func (s *Spinner) Stop(symbol, msg string) {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	s.mu.Unlock()

	close(s.stop)
	// Clear spinner line and print final status.
	fmt.Fprintf(s.out, "\r%-80s\r", "") // erase the line
	if msg != "" {
		fmt.Fprintf(s.out, "  %s %s\n", symbol, msg)
	}
}
