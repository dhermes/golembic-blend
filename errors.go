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
	// ErrNilInterface is the error returned when a value satisfying an interface
	// is nil in a context where it is not allowed.
	ErrNilInterface = ex.Class("Value satisfying interface was nil")
	// ErrMigrationNotRegistered is the error returned when no migration has been
	// registered for a given revision.
	ErrMigrationNotRegistered = ex.Class("No migration registered for revision")
	// ErrCannotInvokeUp is the error returned when a migration cannot invoke the
	// up function (e.g. if it is `nil`).
	ErrCannotInvokeUp = ex.Class("Cannot invoke up function for a migration")
	// ErrCannotPassMilestone is the error returned when a migration sequence
	// contains a milestone migration that is **NOT** the last step.
	ErrCannotPassMilestone = ex.Class("If a migration sequence contains a milestone, it must be the last migration")
)
