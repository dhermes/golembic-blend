package golembic

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/migration"
)

// NOTE: Ensure that
//       * `planAction` satisfies `migration.Action`.
//       * `applyAction` satisfies `migration.Action`.
var (
	_ migration.Action = (*planAction)(nil)
	_ migration.Action = (*applyAction)(nil)
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
	pa := planAction{m: m}
	groups = append(groups, migration.NewGroupWithAction(
		migration.Guard("Finished planning migrations sequence", alwaysPredicate),
		&pa,
	))

	pa.Suite = migration.New(
		migration.OptGroups(groups...),
		migration.OptLog(m.Log),
	)
	return pa.Suite, nil
}

// planAction is a meta-action. It determines a plan (dynamically) for
// **more** work to be done and then appends it to the groups in an existing
// suite.
type planAction struct {
	m     *Manager
	Suite *migration.Suite
}

// Action carries out the planning and updates `Suite.Groups` accordingly.
func (pa *planAction) Action(ctx context.Context, pool *db.Connection, tx *sql.Tx) error {
	if pa.Suite == nil {
		return nil
	}

	pa.Suite.Write(ctx, "plan   ", "Determine migrations that need to be applied")

	migrations, err := pa.m.Plan(ctx, pool, tx)
	if err != nil {
		return err
	}

	//  m.ApplyMigration(ctx, migration)
	for _, mi := range migrations {
		d := fmt.Sprintf("%s: %s", mi.Revision, mi.ExtendedDescription())
		pa.Suite.Groups = append(pa.Suite.Groups, migration.NewGroupWithAction(
			migration.Guard(d, alwaysPredicate),
			&applyAction{m: pa.m, Migration: mi},
		))
	}

	return nil
}

type applyAction struct {
	m         *Manager
	Migration Migration
}

// Action executes ApplyMigration for a given migration.
func (aa *applyAction) Action(ctx context.Context, pool *db.Connection, tx *sql.Tx) error {
	return aa.m.ApplyMigration(ctx, pool, tx, aa.Migration)
}
