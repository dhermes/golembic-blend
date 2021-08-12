package golembic

import (
	"github.com/blend/go-sdk/db/migration"
)

// GenerateSuite generates a suite of migrations from a sequence of golembic
// migrations.
func GenerateSuite(_ *Migrations) (*migration.Suite, error) {
	groups := []*migration.Group{
		// TODO: Add group to ensure table exists.
	}
	// TODO: Add group for every registered migration.
	suite := migration.New(migration.OptGroups(groups...))
	return suite, nil
}
