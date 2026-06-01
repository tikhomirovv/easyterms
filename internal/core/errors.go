// Package core provides shared domain errors for services.
package core

import "errors"

var (
	// ErrInsufficientBalance is returned when the user has no checks left.
	ErrInsufficientBalance = errors.New("insufficient check balance")
	// ErrForbidden is returned when the user does not own the resource.
	ErrForbidden = errors.New("forbidden")
	// ErrInvalidState is returned when the document is not in a valid state for the operation.
	ErrInvalidState = errors.New("invalid document state")
	// ErrNoSources is returned when ingest is attempted without input.
	ErrNoSources = errors.New("document has no sources")
)
