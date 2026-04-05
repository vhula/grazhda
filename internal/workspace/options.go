package workspace

// RunOptions controls the behaviour of workspace operations.
type RunOptions struct {
	DryRun    bool
	Verbose   bool
	Parallel  bool
	NoConfirm bool
}
