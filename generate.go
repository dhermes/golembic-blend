package golembic

import (
	"context"
	"database/sql"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/db/migration"
	"github.com/blend/go-sdk/ex"
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
	groups = append(groups, migration.NewGroup(
		migration.OptGroupActions(&pa),
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

	PlanEventWrite(ctx, pa.Suite.Log, "", "Determine migrations that need to be applied", "")

	migrations, err := pa.m.Plan(ctx, pool, tx, OptApplyVerifyHistory(pa.m.VerifyHistory))
	if err != nil {
		return err
	}

	//  m.ApplyMigration(ctx, migration)
	for _, mi := range migrations {
		pa.Suite.Groups = append(pa.Suite.Groups, migration.NewGroup(
			migration.OptGroupActions(&applyAction{m: pa.m, Migration: mi}),
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
	err := aa.m.ApplyMigration(ctx, pool, tx, aa.Migration)
	suite := migration.GetContextSuite(ctx)

	if err != nil {
		if suite != nil {
			suite.Failed++
			suite.Total++
			PlanEventWrite(ctx, aa.m.Log, aa.Migration.Revision, aa.Migration.ExtendedDescription(), PlanStatusFailed)
			return err
		}
		return err
	}

	if suite != nil {
		suite.Applied++
		suite.Total++
		PlanEventWrite(ctx, aa.m.Log, aa.Migration.Revision, aa.Migration.ExtendedDescription(), PlanStatusApplied)
	}

	return nil
}

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
