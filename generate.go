package golembic

import (
	"context"
	"database/sql"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/migration"
)

// GenerateSuite generates a suite of migrations from a sequence of golembic
// migrations.
func GenerateSuite(m *Manager) (*migration.Suite, error) {
	statements := createMigrationsStatements(m)
	groups := []*migration.Group{
		migration.NewGroupWithAction(
			migration.TableNotExists(m.MetadataTable),
			migration.Statements(statements...),
		),
	}
	pa := planAction{}
	groups = append(groups, migration.NewGroupWithAction(
		migration.Guard("Plan sequence", alwaysPredicate),
		&pa,
	))
	// TODO: Add group for every registered migration.
	pa.Suite = migration.New(migration.OptGroups(groups...))
	return pa.Suite, nil
}

// planAction is a meta-action. It determines a plan (dynamically) for
// **more** work to be done and then appends it to the groups in an existing
// suite.
type planAction struct {
	Suite *migration.Suite
}

// Action carries out the planning and updates `Suite.Groups` accordingly.
func (pa *planAction) Action(ctx context.Context, pool *db.Connection, tx *sql.Tx) error {
	if pa.Suite == nil || pa.Suite.Groups == nil {
		return nil
	}

	// TODO: Actually do the planning here based on largest `serial_id`
	pa.Suite.Groups = append(pa.Suite.Groups, migration.NewGroupWithAction(
		migration.Guard("Sequence item 1", alwaysPredicate),
		&planAction{},
	))
	return nil
}
