package golembic

import (
	"context"
	"database/sql"
)

// UpMigration defines a function interface to be used for up / forward
// migrations. The SQL transaction will be started **before** `UpMigration`
// is invoked and will be committed **after** the `UpMigration` exits without
// error. In addition to the contents of `UpMigration`, a row will be written
// to the migrations metadata table as part of the transaction.
//
// The expectation is that the migration runs SQL statements within the
// transaction. If a migration cannot run inside a transaction, e.g. a
// `CREATE UNIQUE INDEX CONCURRENTLY` statement, then the `UpMigration`
// interface should be used.
type UpMigration = func(context.Context, *sql.Tx) error

// UpMigrationConn defines a function interface to be used for up / forward
// migrations. This is the non-transactional form of `UpMigration` and
// should only be used in rare situations.
type UpMigrationConn = func(context.Context, *sql.Conn) error

// GenerateConfigOption describes options used to create a new generate config.
type GenerateConfigOption = func(*GenerateConfig)

// MigrationOption describes options used to create a new migration.
type MigrationOption = func(*Migration) error

// ApplyOption describes options used to create an apply configuration.
type ApplyOption = func(*ApplyConfig) error
