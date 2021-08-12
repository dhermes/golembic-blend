package golembic

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/blend/go-sdk/db"
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
	// DevelopmentMode is a flag indicating that this manager is currently
	// being run in development mode, so things like extra validation should
	// intentionally be disabled. This is intended for use in testing and
	// development, where an entire database is spun up locally (e.g. in Docker)
	// and migrations will be applied from scratch (including milestones that
	// may not come at the end).
	DevelopmentMode bool
}

// NewManager creates a new config for generating a migrations
// suite.
func NewManager(opts ...ManagerOption) Manager {
	m := Manager{MetadataTable: DefaultMetadataTable}
	for _, opt := range opts {
		opt(&m)
	}
	return m
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
