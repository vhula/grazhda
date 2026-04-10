package config

import "errors"

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
