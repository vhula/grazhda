package executor

import "sync"

// MockExecutor records Run calls for use in tests.
// Set Err for a static error, or ErrFn for per-call dynamic errors.
// Set CaptureOutput to control the string returned by RunCapture.
// Thread-safe for use in parallel tests.
type MockExecutor struct {
	mu            sync.Mutex
	Calls         []string
	Err           error
	ErrFn         func(callIndex int) error
	CaptureOutput string // returned as stdout by RunCapture
}

// Run records the command and returns the configured error.
func (m *MockExecutor) Run(dir string, command string) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, command)
	idx := len(m.Calls)
	errFn := m.ErrFn
	staticErr := m.Err
	m.mu.Unlock()

	if errFn != nil {
		return errFn(idx)
	}
	return staticErr
}

// RunCapture records the command, returns CaptureOutput and the configured error.
func (m *MockExecutor) RunCapture(dir, command string) (string, error) {
	m.mu.Lock()
	m.Calls = append(m.Calls, command)
	idx := len(m.Calls)
	errFn := m.ErrFn
	staticErr := m.Err
	output := m.CaptureOutput
	m.mu.Unlock()

	if errFn != nil {
		return output, errFn(idx)
	}
	return output, staticErr
}
