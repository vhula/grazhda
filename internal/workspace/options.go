package workspace

// RunOptions controls the behaviour of workspace operations.
type RunOptions struct {
	DryRun             bool
	Verbose            bool
	Parallel           bool
	ParallelAll        bool // clone all repos across all projects concurrently
	NoConfirm          bool
	CloneDelaySeconds  int  // seconds to sleep after each clone command (sequential mode)
}
