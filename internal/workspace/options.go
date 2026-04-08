package workspace

// RunOptions controls the behaviour of workspace operations.
type RunOptions struct {
	DryRun            bool
	Verbose           bool
	Parallel          bool
	ParallelAll       bool // clone all repos across all projects concurrently
	NoConfirm         bool
	CloneDelaySeconds int      // seconds to sleep after each clone command (sequential mode)
	ProjectName       string   // filter: only operate on this project (empty = all)
	RepoName          string   // filter: only operate on this repo within ProjectName (empty = all)
	Tags              []string // filter: only operate on repos matching any of these tags (empty = all)
}

// InspectOptions controls inspection commands (diff, stats, search).
type InspectOptions struct {
	Parallel    bool
	ParallelAll bool
	ProjectName string
	RepoName    string
	Tags        []string // filter: only repos matching any of these tags (empty = all)
	Verbose     bool
}

// SearchOptions extends InspectOptions with search-specific configuration.
type SearchOptions struct {
	InspectOptions
	Pattern string
	Glob    bool // match filenames instead of content
	Regex   bool // treat Pattern as a regular expression
}
