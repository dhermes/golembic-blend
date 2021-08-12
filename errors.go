package golembic

import (
	"github.com/blend/go-sdk/ex"
)

var (
	// ErrNotRoot is the error returned when attempting to start a sequence of
	// migrations with a non-root migration.
	ErrNotRoot = ex.Class("Root migration cannot have a previous migration set")
	// ErrMissingRevision is the error returned when attempting to register a migration
	// with no revision.
	ErrMissingRevision = ex.Class("A migration must have a revision")
	// ErrNoPrevious is the error returned when attempting to register a migration
	// with no previous.
	ErrNoPrevious = ex.Class("Cannot register a migration with no previous migration")
	// ErrPreviousNotRegistered is the error returned when attempting to register
	// a migration with a previous that is not yet registered.
	ErrPreviousNotRegistered = ex.Class("Cannot register a migration until previous migration is registered")
	// ErrAlreadyRegistered is the error returned when a migration has already been
	// registered.
	ErrAlreadyRegistered = ex.Class("Migration has already been registered")
)
