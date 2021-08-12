package golembic

import (
	"time"
)

// Migration represents an individual migration to be applied; typically as
// a set of SQL statements.
type Migration struct {
	// Previous is the revision identifier for the migration immediately
	// preceding this one. If absent, this indicates that this migration is
	// the "base" or "root" migration.
	Previous string
	// Revision is an opaque name that uniquely identifies a migration. It
	// is required for a migration to be valid.
	Revision string
	// Description is a long form description of why the migration is being
	// performed. It is intended to be used in "describe" scenarios where
	// a long form "history" of changes is presented.
	Description string
	// Milestone is a flag indicating if the current migration is a milestone.
	// A milestone is a special migration that **must** be the last migration
	// in a sequence whenever applied. This is intended to be used in situations
	// where a change must be "staged" in two (or more parts) and one part
	// must run and "stabilize" before the next migration runs. For example, in
	// a rolling update deploy strategy some changes may not be compatible with
	// "old" and "new" versions of the code that may run simultaneously, so a
	// milestone marks the last point where old / new versions of application
	// code should be expected to be able to interact with the current schema.
	Milestone bool
	// Up is the function to be executed when a migration is being applied. Either
	// this field or `UpConn` are required (not both) and this field should be
	// the default choice in most cases. This function will be run in a transaction
	// that also writes a row to the migrations metadata table to signify that
	// this migration was applied.
	Up UpMigration
	// UpConn is the non-transactional form of `Up`. This should be used in
	// rare situations where a migration cannot run inside a transaction, e.g.
	// a `CREATE UNIQUE INDEX CONCURRENTLY` statement.
	UpConn UpMigrationConn
	// createdAt is stored in the migrations metadata table and represents the
	// moment when the migration was inserted into the table.  It is **not**
	// exported because it is internal to the implementation and should not be
	// specified by calling code.
	createdAt time.Time
	// serialID is an integer used for sorting migrations and will be stored
	// in the migrations table. It is intended to be used for migrations
	// retrieved via a SQL query to the migrations metadata table. It is
	// **not** exported because it is internal to the implementation and should
	// not be specified by calling code.
	serialID uint32
}
