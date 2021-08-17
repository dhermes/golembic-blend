package golembic_test

import (
	"fmt"
	"testing"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/ex"

	golembic "github.com/dhermes/golembic-blend"
)

func TestNewSequence(t *testing.T) {
	it := assert.New(t)

	// Root has `Previous` set already
	root := golembic.Migration{
		Previous:    "ab23bc8ab04a",
		Revision:    "61da13775953",
		Description: "The monster",
	}
	migrations, err := golembic.NewSequence(root)
	it.Nil(migrations)
	expected := `Root migration cannot have a previous migration set; Previous: "ab23bc8ab04a", Revision: "61da13775953"`
	it.Equal(expected, fmt.Sprintf("%v", err))

	// Root is missing `Revision`
	root = golembic.Migration{Description: "Absent"}
	migrations, err = golembic.NewSequence(root)
	it.Nil(migrations)
	expected = "A migration must have a revision"
	it.Equal(expected, fmt.Sprintf("%v", err))
}

func TestMigrations_Register(t *testing.T) {
	it := assert.New(t)

	root := golembic.Migration{
		Revision:    "9f67b79c824c",
		Description: "The base",
	}
	migrations, err := golembic.NewSequence(root)
	it.Nil(err)
	it.Equal([]string{"9f67b79c824c"}, migrations.Revisions())

	// Has no `Previous`
	migration := golembic.Migration{
		Revision:    "5827c0b4806e",
		Description: "The second",
	}
	err = migrations.Register(migration)
	expected := `Cannot register a migration with no previous migration; Revision: "5827c0b4806e"`
	it.Equal(expected, fmt.Sprintf("%v", err))
	it.Equal([]string{"9f67b79c824c"}, migrations.Revisions())

	// `Previous` is not registered
	migration = golembic.Migration{
		Previous:    "b7d387d9e5fb",
		Revision:    "0168d5c42a1e",
		Description: "The second",
	}
	err = migrations.Register(migration)
	expected = `Cannot register a migration until previous migration is registered; Revision: "0168d5c42a1e", Previous: "b7d387d9e5fb"`
	it.Equal(expected, fmt.Sprintf("%v", err))
	it.Equal([]string{"9f67b79c824c"}, migrations.Revisions())

	// Has no `Revision`
	migration = golembic.Migration{
		Previous:    "9f67b79c824c",
		Description: "The second",
	}
	err = migrations.Register(migration)
	expected = `A migration must have a revision; Previous: "9f67b79c824c"`
	it.Equal(expected, fmt.Sprintf("%v", err))
	it.Equal([]string{"9f67b79c824c"}, migrations.Revisions())

	// Success
	migration = golembic.Migration{
		Previous:    "9f67b79c824c",
		Revision:    "24d7ac7c42c5",
		Description: "The second",
	}
	err = migrations.Register(migration)
	it.Nil(err)
	it.Equal([]string{"9f67b79c824c", "24d7ac7c42c5"}, migrations.Revisions())

	// `Revision` is already registered
	migration = golembic.Migration{
		Previous:    "24d7ac7c42c5",
		Revision:    "9f67b79c824c",
		Description: "The third",
	}
	err = migrations.Register(migration)
	expected = `Migration has already been registered; Revision: "9f67b79c824c"`
	it.Equal(expected, fmt.Sprintf("%v", err))
	it.Equal([]string{"9f67b79c824c", "24d7ac7c42c5"}, migrations.Revisions())
}

func TestMigrations_RegisterMany(t *testing.T) {
	it := assert.New(t)

	root := golembic.Migration{
		Revision:    "7742529e7d4d",
		Description: "The first",
	}
	migrations, err := golembic.NewSequence(root)
	it.Nil(err)
	it.Equal([]string{"7742529e7d4d"}, migrations.Revisions())

	// Mixed success and failure
	migration1 := golembic.Migration{
		Previous:    "7742529e7d4d",
		Revision:    "0bbe84d9c57f",
		Description: "The second",
	}
	migration2 := golembic.Migration{
		Revision:    "85efe3f39003",
		Description: "The third",
	}
	err = migrations.RegisterMany(migration1, migration2)
	expected := `Cannot register a migration with no previous migration; Revision: "85efe3f39003"`
	it.Equal(expected, fmt.Sprintf("%v", err))
	it.Equal([]string{"7742529e7d4d", "0bbe84d9c57f"}, migrations.Revisions())

	// All success
	migration3 := golembic.Migration{
		Previous:    "0bbe84d9c57f",
		Revision:    "1323ed11b94b",
		Description: "The third",
	}
	err = migrations.RegisterMany(migration3)
	it.Nil(err)
	it.Equal([]string{"7742529e7d4d", "0bbe84d9c57f", "1323ed11b94b"}, migrations.Revisions())
}

func TestMigrations_RegisterManyOpt(t *testing.T) {
	it := assert.New(t)

	root := golembic.Migration{
		Revision:    "e40bb43dbb16",
		Description: "The first",
	}
	migrations, err := golembic.NewSequence(root)
	it.Nil(err)
	it.Equal([]string{"e40bb43dbb16"}, migrations.Revisions())

	// Fail `NewMigration()`
	known := ex.New("WRENCH")
	opt := func(_ *golembic.Migration) error {
		return known
	}
	err = migrations.RegisterManyOpt([]golembic.MigrationOption{opt})
	it.Equal(known, err)
	it.Equal([]string{"e40bb43dbb16"}, migrations.Revisions())

	// Fail `Migrations.Register()`
	err = migrations.RegisterManyOpt(
		[]golembic.MigrationOption{
			golembic.OptRevision("4e964a8c9717"),
			golembic.OptDescription("The second"),
		},
	)
	expected := `Cannot register a migration with no previous migration; Revision: "4e964a8c9717"`
	it.Equal(expected, fmt.Sprintf("%v", err))
}

func TestMigrations_Root(t *testing.T) {
	it := assert.New(t)

	// Empty sequence
	migrations := &golembic.Migrations{}
	root := migrations.Root()
	it.Equal(golembic.Migration{}, root)
}

func TestMigrations_Get(t *testing.T) {
	it := assert.New(t)

	// Empty sequence
	migrations := &golembic.Migrations{}
	migration := migrations.Get("ee6f7cc897ba")
	it.Nil(migration)
}
