package workspace

import "context"

// RunOptions controls the behaviour of workspace operations.
type RunOptions struct {
	Context           context.Context // optional; defaults to context.Background()
	DryRun            bool
	Verbose           bool
	Parallel          bool
	NoConfirm         bool
	CloneDelaySeconds int      // seconds to sleep after each clone command (sequential mode)
	ProjectName       string   // filter: only operate on this project (empty = all)
	RepoName          string   // filter: only operate on this repo within ProjectName (empty = all)
	Tags              []string // filter: only operate on repos matching any of these tags (empty = all)
}

// ctx returns opts.Context, falling back to context.Background() if nil.
func (opts RunOptions) ctx() context.Context {
	return ctxOr(opts.Context)
}

// InspectOptions controls inspection commands (diff, stats, search).
type InspectOptions struct {
	Context     context.Context // optional; defaults to context.Background()
	Parallel    bool
	ProjectName string
	RepoName    string
	Tags        []string // filter: only repos matching any of these tags (empty = all)
	Verbose     bool
}

// ctx returns opts.Context, falling back to context.Background() if nil.
func (opts InspectOptions) ctx() context.Context {
	return ctxOr(opts.Context)
}

// SearchOptions extends InspectOptions with search-specific configuration.
type SearchOptions struct {
	InspectOptions
	Pattern string
	Glob    bool // match filenames instead of content
	Regex   bool // treat Pattern as a regular expression
}

// ctxOr returns ctx if non-nil, otherwise context.Background().
func ctxOr(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
