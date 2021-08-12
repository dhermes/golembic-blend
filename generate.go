package golembic

import (
	"github.com/blend/go-sdk/db/migration"
)

// GenerateSuite generates a suite of migrations from a sequence of golembic
// migrations.
func GenerateSuite(gc GenerateConfig) (*migration.Suite, error) {
	statements := createMigrationsStatements(gc)
	groups := []*migration.Group{
		migration.NewGroupWithAction(
			migration.TableNotExists(gc.MetadataTable),
			migration.Statements(statements...),
		),
	}
	// TODO: Add group for every registered migration.
	suite := migration.New(migration.OptGroups(groups...))
	return suite, nil
}
