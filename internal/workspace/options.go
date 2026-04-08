package workspace

// RunOptions controls the behaviour of workspace operations.
type RunOptions struct {
	DryRun            bool
	Verbose           bool
	Parallel          bool
	ParallelAll       bool // clone all repos across all projects concurrently
	NoConfirm         bool
	CloneDelaySeconds int    // seconds to sleep after each clone command (sequential mode)
	ProjectName       string // filter: only operate on this project (empty = all)
	RepoName          string // filter: only operate on this repo within ProjectName (empty = all)
}
