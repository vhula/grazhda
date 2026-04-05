package workspace

import "sync"

// MockExecutor records Run calls for use in tests.
// Set Err for a static error, or ErrFn for per-call dynamic errors.
// Thread-safe for use in parallel tests.
type MockExecutor struct {
	mu    sync.Mutex
	Calls []string
	Err   error
	ErrFn func(callIndex int) error
}

func (m *MockExecutor) Run(dir string, command string) error {
	m.mu.Lock()
	m.Calls = append(m.Calls, command)
	idx := len(m.Calls)
	errFn := m.ErrFn
	m.mu.Unlock()

	if errFn != nil {
		return errFn(idx)
	}
	return m.Err
}
