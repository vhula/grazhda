package reporter

import "charm.land/log/v2"

// OpResult records the outcome of a single repository operation.
type OpResult struct {
	Workspace string
	Project   string
	Repo      string
	Skipped   bool
	Err       error
}

// Reporter accumulates operation results and produces a summary.
type Reporter struct {
	log     *log.Logger
	results []OpResult
}

// NewReporter creates a Reporter using the provided logger.
func NewReporter(l *log.Logger) *Reporter {
	return &Reporter{log: l}
}

// Record appends an operation result.
func (r *Reporter) Record(res OpResult) {
	r.results = append(r.results, res)
}

// Summary prints a count of successes and failures to the logger.
func (r *Reporter) Summary() {}

// ExitCode returns 0 if all operations succeeded, 1 otherwise.
func (r *Reporter) ExitCode() int {
	for _, res := range r.results {
		if res.Err != nil {
			return 1
		}
	}
	return 0
}
