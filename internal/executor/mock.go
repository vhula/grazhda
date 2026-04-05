package executor

// MockExecutor records Run calls for use in tests.
type MockExecutor struct {
	Calls []string
	Err   error
}

func (m *MockExecutor) Run(dir string, command string) error {
	m.Calls = append(m.Calls, command)
	return m.Err
}
