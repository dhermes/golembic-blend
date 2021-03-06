package golembic

import (
	"sync"

	"github.com/blend/go-sdk/ex"
)

// NOTE: Ensure that
//       * `Migrations.Since` satisfies `migrationsFilter`.
var (
	_ migrationsFilter = (*Migrations)(nil).Since
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

// RegisterManyOpt attempts to register multiple migrations (in order) with an
// existing sequence. It differs from `RegisterMany()` in that the construction
// of `Migration` objects is handled directly here by taking a slice of
// option slices.
func (m *Migrations) RegisterManyOpt(manyOpts ...[]MigrationOption) error {
	for _, opts := range manyOpts {
		migration, err := NewMigration(opts...)
		if err != nil {
			return err
		}

		err = m.Register(*migration)
		if err != nil {
			return err
		}
	}

	return nil
}

// Root does a linear scan of every migration in the sequence and returns
// the root migration. In the "general" case such a scan would be expensive, but
// the number of migrations should always be a small number.
//
// NOTE: This does not verify or enforce the invariant that there must be
// exactly one migration without a previous migration. This invariant is enforced
// by the exported methods such as `Register()` and `RegisterMany()` and the
// constructor `NewSequence()`.
func (m *Migrations) Root() Migration {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, migration := range m.sequence {
		if migration.Previous == "" {
			return migration
		}
	}

	return Migration{}
}

// All produces the migrations in the sequence, in order.
//
// NOTE: This does not verify or enforce the invariant that there must be
//       exactly one migration without a previous migration. This invariant is
//       enforced by the exported methods such as `Register()` and
//       `RegisterMany()` and the constructor `NewSequence()`.
func (m *Migrations) All() []Migration {
	root := m.Root()

	m.lock.Lock()
	defer m.lock.Unlock()
	result := []Migration{root}
	// Find the unique revision (without validation) that points at the
	// current `previous`.
	previous := root.Revision
	for i := 0; i < len(m.sequence)-1; i++ {
		for _, migration := range m.sequence {
			if migration.Previous != previous {
				continue
			}

			result = append(result, migration)
			previous = migration.Revision
			break
		}
	}

	return result
}

// Since returns the migrations that occur **after** `revision`.
//
// This utilizes `All()` and returns all migrations after the one that
// matches `revision`. If none match, an error will be returned. If
// `revision` is the **last** migration, the migrations returned will be an
// empty slice.
func (m *Migrations) Since(revision string) (int, []Migration, error) {
	all := m.All()
	found := false

	result := []Migration{}
	pastMigrationCount := 0
	for _, migration := range all {
		if found {
			result = append(result, migration)
			continue
		}

		pastMigrationCount++
		if migration.Revision == revision {
			found = true
		}
	}

	if !found {
		err := ex.New(ErrMigrationNotRegistered, ex.OptMessagef("Revision: %q", revision))
		return 0, nil, err
	}

	return pastMigrationCount, result, nil
}

// Revisions produces the revisions in the sequence, in order.
//
// This utilizes `All()` and just extracts the revisions.
func (m *Migrations) Revisions() []string {
	result := []string{}
	for _, migration := range m.All() {
		result = append(result, migration.Revision)
	}
	return result
}

// Get retrieves a revision from the sequence, if present. If not, returns
// `nil`.
func (m *Migrations) Get(revision string) *Migration {
	m.lock.Lock()
	defer m.lock.Unlock()

	migration, ok := m.sequence[revision]
	if ok {
		return &migration
	}

	return nil
}
