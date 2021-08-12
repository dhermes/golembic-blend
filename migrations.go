package golembic

import (
	"sync"

	"github.com/blend/go-sdk/ex"
)

// Migrations represents a sequence of migrations to be applied.
type Migrations struct {
	sequence map[string]Migration
	lock     sync.Mutex
}

// NewSequence creates a new sequence of migrations rooted in a single
// base / root migration.
func NewSequence(root Migration) (*Migrations, error) {
	if root.Previous != "" {
		err := ex.New(
			ErrNotRoot,
			ex.OptMessagef("Previous: %q, Revision: %q", root.Previous, root.Revision),
		)
		return nil, err
	}

	if root.Revision == "" {
		return nil, ex.New(ErrMissingRevision)
	}

	m := &Migrations{
		sequence: map[string]Migration{
			root.Revision: root,
		},
		lock: sync.Mutex{},
	}
	return m, nil
}

// Register adds a new migration to an existing sequence of migrations, if
// possible. The new migration must have a previous migration and have a valid
// revision that is not already registered.
func (m *Migrations) Register(migration Migration) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if migration.Previous == "" {
		err := ex.New(
			ErrNoPrevious,
			ex.OptMessagef("Revision: %q", migration.Revision),
		)
		return err
	}

	if _, ok := m.sequence[migration.Previous]; !ok {
		err := ex.New(
			ErrPreviousNotRegistered,
			ex.OptMessagef("Revision: %q, Previous: %q", migration.Revision, migration.Previous),
		)
		return err
	}

	if migration.Revision == "" {
		err := ex.New(
			ErrMissingRevision,
			ex.OptMessagef("Previous: %q", migration.Previous),
		)
		return err
	}

	if _, ok := m.sequence[migration.Revision]; ok {
		err := ex.New(
			ErrAlreadyRegistered,
			ex.OptMessagef("Revision: %q", migration.Revision),
		)
		return err
	}

	// NOTE: This crucially relies on `m.sequence` being locked.
	migration.serialID = uint32(len(m.sequence))
	m.sequence[migration.Revision] = migration
	return nil
}

// RegisterMany attempts to register multiple migrations (in order) with an
// existing sequence.
func (m *Migrations) RegisterMany(ms ...Migration) error {
	for _, migration := range ms {
		err := m.Register(migration)
		if err != nil {
			return err
		}
	}

	return nil
}
