package golembic

import (
	"context"
	"database/sql"
	"time"

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

	var mms []migrationModel
	err := q.OutMany(&mms)
	if err != nil {
		return nil, err
	}

	migrations := make([]Migration, len(mms))
	for i, mm := range mms {
		migrations[i] = mm.ToMigration()
	}
	return migrations, nil
}

// migrationModel is a shallow version of `Migration` meant for use with
// database queries.
type migrationModel struct {
	Previous  string    `db:"previous"`
	Revision  string    `db:"revision"`
	CreatedAt time.Time `db:"created_at"`
	SerialID  uint32    `db:"serial_id"`
}

func (mm migrationModel) ToMigration() Migration {
	return Migration{
		Previous:  mm.Previous,
		Revision:  mm.Revision,
		createdAt: mm.CreatedAt,
		serialID:  mm.SerialID,
	}
}
