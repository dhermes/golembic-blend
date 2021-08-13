package golembic

import (
	"context"
	"database/sql"

	"github.com/blend/go-sdk/db"
)

// readAllMigration performs a SQL query and reads all rows into a
// `Migration` slice, under the assumption that three columns -- revision,
// previous and created_at -- are being returned for the query (in that order).
// For example, the query
//
//   SELECT revision, previous, created_at FROM golembic_migrations
//
// would satisfy this. A more "focused" query would return the latest migration
// applied
//
//   SELECT
//     revision,
//     previous,
//     created_at
//   FROM
//     golembic_migrations
//   ORDER BY
//     serial_id DESC
//   LIMIT 1
func readAllMigration(ctx context.Context, pool *db.Connection, tx *sql.Tx, query string, args ...interface{}) ([]Migration, error) {
	invocation := pool.Invoke(db.OptContext(ctx), db.OptTx(tx))
	q := invocation.Query(query, args...)

	var migrations []Migration
	err := q.OutMany(&migrations)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}
