package golembic

import (
	"context"
	"database/sql"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/migration"
	"github.com/blend/go-sdk/ex"
)

// ApplyDynamic applies a migrations suite. Rather than using a `range`
// over `s.Groups`, it uses a length check, which allows `s.Groups` to
// change dynamically during the iteration.
func ApplyDynamic(ctx context.Context, s *migration.Suite, c *db.Connection) (err error) {
	defer s.WriteStats(ctx)
	defer func() {
		if r := recover(); r != nil {
			err = ex.New(r)
		}
	}()

	for i := 0; i < len(s.Groups); i++ {
		group := s.Groups[i]
		if err = group.Action(migration.WithSuite(ctx, s), c); err != nil {
			return
		}
	}
	return
}

// alwaysPredicate can be used to always run an action.
func alwaysPredicate(_ context.Context, _ *db.Connection, _ *sql.Tx) (bool, error) {
	return true, nil
}
