package golembic

import (
	"context"
	"database/sql"
	"io/ioutil"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/ex"
)

// OptPrevious sets the previous on a migration.
func OptPrevious(previous string) MigrationOption {
	return func(m *Migration) error {
		m.Previous = previous
		return nil
	}
}

// OptRevision sets the revision on a migration.
func OptRevision(revision string) MigrationOption {
	return func(m *Migration) error {
		if revision == "" {
			return ex.New(ErrMissingRevision)
		}

		m.Revision = revision
		return nil
	}
}

// OptDescription sets the description on a migration.
func OptDescription(description string) MigrationOption {
	return func(m *Migration) error {
		m.Description = description
		return nil
	}
}

// OptMilestone sets the milestone flag on a migration.
func OptMilestone(milestone bool) MigrationOption {
	return func(m *Migration) error {
		m.Milestone = milestone
		return nil
	}
}

// OptUp sets the `up` function on a migration.
func OptUp(up UpMigration) MigrationOption {
	return func(m *Migration) error {
		if up == nil {
			return ex.New(ErrNilInterface)
		}

		m.Up = up
		return nil
	}
}

// OptUpFromSQL returns an option that sets the `up` function to execute a
// SQL statement.
func OptUpFromSQL(statement string) MigrationOption {
	up := func(ctx context.Context, pool *db.Connection, tx *sql.Tx) error {
		i := pool.Invoke(db.OptContext(ctx), db.OptTx(tx))
		_, err := i.Exec(statement)
		return err
	}

	return func(m *Migration) error {
		m.Up = up
		return nil
	}
}

// OptUpFromFile returns an option that sets the `up` function to execute a
// SQL statement that is stored in a file.
func OptUpFromFile(filename string) MigrationOption {
	statement, err := ioutil.ReadFile(filename)
	if err != nil {
		return OptAlwaysError(err)
	}

	return OptUpFromSQL(string(statement))
}

// OptUpConn sets the non-transactional `up` function on a migration.
func OptUpConn(up UpMigrationConn) MigrationOption {
	return func(m *Migration) error {
		if up == nil {
			return ex.New(ErrNilInterface)
		}

		m.UpConn = up
		return nil
	}
}

// OptUpConnFromSQL returns an option that sets the non-transctional `up`
// function to execute a SQL statement.
func OptUpConnFromSQL(statement string) MigrationOption {
	up := func(ctx context.Context, pool *db.Connection) error {
		i := pool.Invoke(db.OptContext(ctx))
		_, err := i.Exec(statement)
		return err
	}

	return func(m *Migration) error {
		m.UpConn = up
		return nil
	}
}

// OptUpConnFromFile returns an option that sets the non-transctional `up`
// function to execute a SQL statement that is stored in a file.
func OptUpConnFromFile(filename string) MigrationOption {
	statement, err := ioutil.ReadFile(filename)
	if err != nil {
		return OptAlwaysError(err)
	}

	return OptUpConnFromSQL(string(statement))
}

// OptAlwaysError returns an option that always returns an error.
func OptAlwaysError(err error) MigrationOption {
	return func(m *Migration) error {
		return err
	}
}
