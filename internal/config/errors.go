package config

import (
	"errors"
	"strings"
)

// Sentinel errors for the config package.
// Callers can test with errors.Is() to take programmatic action.
var (
	// ErrNotFound is returned when the configuration file does not exist.
	ErrNotFound = errors.New("configuration file not found")

	// ErrInvalid is returned when the configuration file fails validation.
	ErrInvalid = errors.New("configuration is invalid")

	// ErrKeyNotFound is returned when a dotted-path key does not exist in the config.
	ErrKeyNotFound = errors.New("configuration key not found")
)

// ValidationError is returned by Replace and Merge when the resulting config
// fails structural validation. It wraps ErrInvalid so callers can use errors.Is.
type ValidationError struct {
	Errs []string
}

func (e *ValidationError) Error() string {
	return "configuration is invalid: " + strings.Join(e.Errs, "; ")
}

func (e *ValidationError) Unwrap() error { return ErrInvalid }
