package golembic

import (
	"github.com/blend/go-sdk/db/migration"
)

type GenerateConfig struct {
	// MetadataTable is the name of the table that stores migration metadata.
	// The expected default value (`DefaultMetadataTable`) is
	// "golembic_migrations".
	MetadataTable string
	// Sequence is the collection of registered migrations to be applied,
	// verified, described, etc. by this manager.
	Sequence *Migrations
	// DevelopmentMode is a flag indicating that this manager is currently
	// being run in development mode, so things like extra validation should
	// intentionally be disabled. This is intended for use in testing and
	// development, where an entire database is spun up locally (e.g. in Docker)
	// and migrations will be applied from scratch (including milestones that
	// may not come at the end).
	DevelopmentMode bool
}

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
