package golembic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/blend/go-sdk/db"
	dbmigration "github.com/blend/go-sdk/db/migration"
	"github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/logger"
)

// NOTE: Ensure that
//       * `Manager.sinceOrAll` satisfies `migrationsFilter`.
var (
	_ migrationsFilter = (*Manager)(nil).sinceOrAll
)

const (
	// DefaultMetadataTable is the default name for the table used to store
	// metadata about migrations.
	DefaultMetadataTable = "golembic_migrations"
)

// Manager orchestrates database operations done via `Up` / `UpConn` as well as
// supporting operations such as writing rows into a migration metadata table
// during a migration.
type Manager struct {
	// MetadataTable is the name of the table that stores migration metadata.
	// The expected default value (`DefaultMetadataTable`) is
	// "golembic_migrations".
	MetadataTable string
	// Sequence is the collection of registered migrations to be applied,
	// verified, described, etc. by this manager.
	Sequence *Migrations
	// VerifyHistory indicates that the rows **stored** in the migration metadata
	// table should be verified during planning.
	VerifyHistory bool
	// DevelopmentMode is a flag indicating that this manager is currently
	// being run in development mode, so things like extra validation should
	// intentionally be disabled. This is intended for use in testing and
	// development, where an entire database is spun up locally (e.g. in Docker)
	// and migrations will be applied from scratch (including milestones that
	// may not come at the end).
	DevelopmentMode bool
	// Log is used for printing output
	Log logger.Log
}

// NewManager creates a new manager for orchestrating migrations.
func NewManager(opts ...ManagerOption) (*Manager, error) {
	m := &Manager{MetadataTable: DefaultMetadataTable}
	for _, opt := range opts {
		err := opt(m)
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}

// InsertMigration inserts a migration into the migrations metadata table.
func (m *Manager) InsertMigration(ctx context.Context, pool *db.Connection, tx *sql.Tx, migration Migration) error {
	if migration.Previous == "" {
		statement := fmt.Sprintf(
			"INSERT INTO %s (serial_id, revision, previous) VALUES (0, %s, NULL)",
			providerQuoteIdentifier(m.MetadataTable),
			providerQueryParameter(1),
		)
		_, err := pool.Invoke(db.OptContext(ctx), db.OptTx(tx)).Exec(statement, migration.Revision)
		return err
	}

	statement := fmt.Sprintf(
		"INSERT INTO %s (serial_id, revision, previous) VALUES (%s, %s, %s)",
		providerQuoteIdentifier(m.MetadataTable),
		providerQueryParameter(1),
		providerQueryParameter(2),
		providerQueryParameter(3),
	)
	_, err := pool.Invoke(db.OptContext(ctx), db.OptTx(tx)).Exec(
		statement,
		migration.serialID, // Parameter 1
		migration.Revision, // Parameter 2
		migration.Previous, // Parameter 3
	)
	return err
}

// ApplyMigration creates a transaction that runs the "Up" migration.
func (m *Manager) ApplyMigration(ctx context.Context, pool *db.Connection, tx *sql.Tx, migration Migration) (err error) {
	err = migration.InvokeUp(ctx, pool, tx)
	if err != nil {
		return
	}

	err = m.InsertMigration(ctx, pool, tx, migration)
	if err != nil {
		return
	}

	return
}

// filterMigrations applies a filter function that takes the revision of the
// last applied migration to determine a set of migrations to run.
//
// NOTE: This assumes, but does not check, that the migrations metadata table
// exists.
func (m *Manager) filterMigrations(ctx context.Context, pool *db.Connection, tx *sql.Tx, filter migrationsFilter, verifyHistory bool) (int, []Migration, error) {
	latest, _, err := m.latestMaybeVerify(ctx, pool, tx, verifyHistory)
	if err != nil {
		return 0, nil, err
	}

	pastMigrationCount, migrations, err := filter(latest)
	if err != nil {
		return 0, nil, err
	}

	if len(migrations) == 0 {
		format := "No migrations to run; latest revision: %s"

		// Add `milestoneSuffix`, if we can detect `latest` is a milestone.
		migration := m.Sequence.Get(latest)
		if migration != nil && migration.Milestone {
			format += milestoneSuffix
		}

		body := fmt.Sprintf(format, latest)
		PlanEventWrite(ctx, m.Log, "", body, "")
		return pastMigrationCount, nil, nil
	}

	return pastMigrationCount, migrations, nil
}

func (m *Manager) validateMilestones(ctx context.Context, pastMigrationCount int, migrations []Migration) error {
	// Early exit if no migrations have been run yet. This **assumes** that the
	// database is being brought up from scratch.
	if pastMigrationCount == 0 {
		return nil
	}

	count := len(migrations)
	// Ensure all (but the last) are not a milestone.
	for i := 0; i < count-1; i++ {
		migration := migrations[i]
		if !migration.Milestone {
			continue
		}

		body := fmt.Sprintf("Revision %s (%d / %d migrations)", migration.Revision, i+1, count)
		suiteWrite(ctx, m.Log, "failed", body)
		err := ex.New(ErrCannotPassMilestone)

		// In development mode, log the error message but don't return an error.
		if m.DevelopmentMode {
			logger.MaybeDebugf(m.Log, "Ignoring error in development mode")
			logger.MaybeDebugf(m.Log, "  %s", err)
			continue
		}

		return err
	}

	return nil
}

// Plan gathers (and verifies) all migrations that have not yet been applied.
func (m *Manager) Plan(ctx context.Context, pool *db.Connection, tx *sql.Tx, opts ...ApplyOption) ([]Migration, error) {
	ac, err := NewApplyConfig(opts...)
	if err != nil {
		return nil, err
	}

	pastMigrationCount, migrations, err := m.filterMigrations(ctx, pool, tx, m.sinceOrAll, ac.VerifyHistory)
	if err != nil {
		return nil, err
	}

	if migrations == nil {
		return nil, nil
	}

	err = m.validateMilestones(ctx, pastMigrationCount, migrations)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}

func (m *Manager) sinceOrAll(revision string) (int, []Migration, error) {
	if revision == "" {
		return 0, m.Sequence.All(), nil
	}

	return m.Sequence.Since(revision)
}

// Latest determines the revision and timestamp of the most recently applied
// migration.
//
// NOTE: This assumes, but does not check, that the migrations metadata table
// exists.
func (m *Manager) Latest(ctx context.Context, pool *db.Connection, tx *sql.Tx) (revision string, createdAt time.Time, err error) {
	query := fmt.Sprintf(
		"SELECT revision, previous, created_at FROM %s ORDER BY serial_id DESC LIMIT 1",
		providerQuoteIdentifier(m.MetadataTable),
	)
	rows, err := readAllMigration(ctx, pool, tx, query)
	if err != nil {
		return
	}

	if len(rows) == 0 {
		return
	}

	// NOTE: Here we trust that the query is sufficient to guarantee that
	//       `len(rows) == 1`.
	revision = rows[0].Revision
	createdAt = rows[0].createdAt
	return
}

// latestMaybeVerify determines the latest applied migration and verifies all of the
// migration history if `verifyHistory` is true.
func (m *Manager) latestMaybeVerify(ctx context.Context, pool *db.Connection, tx *sql.Tx, verifyHistory bool) (revision string, createdAt time.Time, err error) {
	if !verifyHistory {
		revision, createdAt, err = m.Latest(ctx, pool, tx)
		return
	}

	history, _, err := m.verifyHistory(ctx, pool, tx)
	if err != nil {
		return
	}

	if len(history) == 0 {
		return
	}

	revision = history[len(history)-1].Revision
	createdAt = history[len(history)-1].createdAt
	return
}

// verifyHistory retrieves a full history of migrations and compares it against
// the sequence of registered migrations. If they match (up to the end of the
// history, the registered sequence can be longer), this will return with no
// error and include slices of the history and the registered migrations.
func (m *Manager) verifyHistory(ctx context.Context, pool *db.Connection, tx *sql.Tx) (history, registered []Migration, err error) {
	query := fmt.Sprintf(
		"SELECT revision, previous, created_at FROM %s ORDER BY serial_id ASC",
		providerQuoteIdentifier(m.MetadataTable),
	)
	history, err = readAllMigration(ctx, pool, tx, query)
	if err != nil {
		return
	}

	registered = m.Sequence.All()
	if len(history) > len(registered) {
		body := fmt.Sprintf("Sequence has %d migrations but %d are stored in the table", len(registered), len(history))
		suiteWrite(ctx, m.Log, "failed", body)
		err = ex.New(ErrMigrationMismatch)
		return
	}

	for i, row := range history {
		expected := registered[i]
		if !row.Like(expected) {
			body := fmt.Sprintf("Stored migration %d: %q does not match migration %q in sequence", i, row.Compact(), expected.Compact())
			suiteWrite(ctx, m.Log, "failed", body)
			err = ex.New(ErrMigrationMismatch)
			return
		}
	}

	return
}

func suiteWrite(ctx context.Context, log logger.Log, result, body string) {
	logger.MaybeTriggerContext(
		ctx, log,
		dbmigration.NewEvent(result, body, dbmigration.GetContextLabels(ctx)...),
	)
}
