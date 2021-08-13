package examples

import (
	"github.com/blend/go-sdk/ex"
	golembic "github.com/dhermes/golembic-blend"
)

// AllMigrations returns a sequence of migrations to be used as an example.
func AllMigrations(length int) (*golembic.Migrations, error) {
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
	opts := [][]golembic.MigrationOption{
		{
			golembic.OptPrevious("3f34bd961f15"),
			golembic.OptRevision("464bc456c630"),
			golembic.OptDescription("Seed data in users table"),
			golembic.OptUpFromSQL(seedUsersTable),
		},
		{
			golembic.OptPrevious("464bc456c630"),
			golembic.OptRevision("959456a8af88"),
			golembic.OptDescription("Add city column to users table"),
			golembic.OptUpFromSQL(addUsersCityColumn),
		},
		{
			golembic.OptPrevious("959456a8af88"),
			golembic.OptRevision("57393d6ddb95"),
			golembic.OptDescription("Rename the root user"),
			golembic.OptMilestone(true),
			golembic.OptUpFromSQL(renameRoot),
		},
		{
			golembic.OptPrevious("57393d6ddb95"),
			golembic.OptRevision("4d07dd6af28d"),
			golembic.OptDescription("Add index on user emails (concurrently)"),
			golembic.OptUpConnFromSQL(addUsersEmailIndexConcurrently),
		},
		{
			golembic.OptPrevious("4d07dd6af28d"),
			golembic.OptRevision("2a35ccd628bc"),
			golembic.OptDescription("Create books table"),
			golembic.OptUpFromSQL(createBooksTable),
		},
		{
			golembic.OptPrevious("2a35ccd628bc"),
			golembic.OptRevision("3196713ca7e6"),
			golembic.OptDescription("Create movies table"),
			golembic.OptUpFromSQL(createMoviesTable),
		},
	}

	length, err = checkLength(length)
	if err != nil {
		return nil, err
	}

	opts = opts[:length-1]
	err = migrations.RegisterManyOpt(opts...)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}

func checkLength(length int) (int, error) {
	switch length {
	case -1:
		return 7, nil
	case 1, 2, 3, 4, 5, 6, 7:
		return length, nil
	default:
		return 0, ex.New("Invalid sequence length", ex.OptMessagef("Length: %d", length))
	}
}
