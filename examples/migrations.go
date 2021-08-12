package examples

import (
	golembic "github.com/dhermes/golembic-blend"
)

// AllMigrations returns a sequence of migrations to be used as an example.
func AllMigrations() (*golembic.Migrations, error) {
	root, err := golembic.NewMigration(
		golembic.OptRevision("3f34bd961f15"),
		golembic.OptDescription("Create users table"),
		golembic.OptUpFromSQL(createUsersTable),
	)
	if err != nil {
		return nil, err
	}

	migrations, err := golembic.NewSequence(*root)
	if err != nil {
		return nil, err
	}
	err = migrations.RegisterManyOpt(
		[]golembic.MigrationOption{
			golembic.OptPrevious("3f34bd961f15"),
			golembic.OptRevision("464bc456c630"),
			golembic.OptDescription("Seed data in users table"),
			golembic.OptUpFromSQL(seedUsersTable),
		},
		[]golembic.MigrationOption{
			golembic.OptPrevious("464bc456c630"),
			golembic.OptRevision("959456a8af88"),
			golembic.OptDescription("Add city column to users table"),
			golembic.OptUpFromSQL(addUsersCityColumn),
		},
		[]golembic.MigrationOption{
			golembic.OptPrevious("959456a8af88"),
			golembic.OptRevision("57393d6ddb95"),
			golembic.OptDescription("Rename the root user"),
			golembic.OptMilestone(true),
			golembic.OptUpFromSQL(renameRoot),
		},
		[]golembic.MigrationOption{
			golembic.OptPrevious("57393d6ddb95"),
			golembic.OptRevision("4d07dd6af28d"),
			golembic.OptDescription("Add index on user emails (concurrently)"),
			golembic.OptUpConnFromSQL(addUsersEmailIndexConcurrently),
		},
		[]golembic.MigrationOption{
			golembic.OptPrevious("4d07dd6af28d"),
			golembic.OptRevision("2a35ccd628bc"),
			golembic.OptDescription("Create books table"),
			golembic.OptUpFromSQL(createBooksTable),
		},
		[]golembic.MigrationOption{
			golembic.OptPrevious("2a35ccd628bc"),
			golembic.OptRevision("3196713ca7e6"),
			golembic.OptDescription("Create movies table"),
			golembic.OptUpFromSQL(createMoviesTable),
		},
	)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}
