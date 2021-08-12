package golembic

import (
	"context"
	"database/sql"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/migration"
)

// ExecConn creates a `migration.Action` that will run a statement with a given
// set of arguments, but ignore the open transaction. This is intended to be
// a `migration.Exec()` equivalent that runs outside of a transaction.
func ExecConn(statement string, args ...interface{}) migration.Action {
	return migration.ActionFunc(func(ctx context.Context, c *db.Connection, _ *sql.Tx) (err error) {
		err = db.IgnoreExecResult(c.Invoke(db.OptContext(ctx)).Exec(statement, args...))
		return
	})
}
