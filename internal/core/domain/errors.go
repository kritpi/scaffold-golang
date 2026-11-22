package domain

import "errors"

// Error kinds. httpx maps each to a status with errors.Is. An error that wraps
// none of them is a bug and becomes 500.
var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrUnavailable  = errors.New("dependency unavailable")
	// ErrUnauthenticated: the caller proved no identity.
	ErrUnauthenticated = errors.New("unauthenticated")
)
