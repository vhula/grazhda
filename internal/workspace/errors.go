package workspace

import "errors"

// Sentinel errors for the workspace package.
// Callers can test with errors.Is() to take programmatic action.
var (
	// ErrNoWorkspace is returned when no matching workspace is found.
	ErrNoWorkspace = errors.New("no matching workspace")

	// ErrNoProject is returned when the requested project does not exist.
	ErrNoProject = errors.New("no matching project")

	// ErrNoRepo is returned when the requested repository does not exist.
	ErrNoRepo = errors.New("no matching repository")

	// ErrFilterInvalid is returned when a filter combination is not valid
	// (e.g. --repo-name without --project-name).
	ErrFilterInvalid = errors.New("invalid filter combination")

	// ErrCancelled is returned when an operation is interrupted by context cancellation.
	ErrCancelled = errors.New("operation cancelled")
)
