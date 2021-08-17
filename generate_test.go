package golembic_test

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/logger"

	golembic "github.com/dhermes/golembic-blend"
)

func TestGenerateSuite_HappyPath(t *testing.T) {
	it := assert.New(t)

	ctx := context.TODO()
	pool := defaultDB()
	it.NotNil(pool)

	suffix := anyLowercase(6)
	mt := fmt.Sprintf("foo_%s_migrations", suffix)
	t1 := fmt.Sprintf("foo1_%s", suffix)
	t2 := fmt.Sprintf("foo2_%s", suffix)
	t.Cleanup(func() {
		err1 := dropTable(ctx, pool, mt)
		err2 := dropTable(ctx, pool, t1)
		err3 := dropTable(ctx, pool, t2)
		it.Nil(err1)
		it.Nil(err2)
		it.Nil(err3)
	})
	var logBuffer bytes.Buffer
	log := logger.Memory(&logBuffer)

	migrations, err := makeSequence(t1, t2)
	it.Nil(err)
	m, err := golembic.NewManager(
		golembic.OptManagerSequence(migrations),
		golembic.OptManagerMetadataTable(mt),
		golembic.OptManagerLog(log),
	)
	it.Nil(err)
	suite, err := golembic.GenerateSuite(m)
	it.Nil(err)

	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Nil(err)

	logLines := []string{
		fmt.Sprintf("[db.migration] -- applied -- Check table does not exist: %s", mt),
		"[db.migration] -- plan -- Determine migrations that need to be applied",
		"[db.migration] -- aa60f058f5f5 -- Create first table",
		"[db.migration] -- ab1208989a3f -- Alter first table",
		"[db.migration] -- 60a33b9d4c77 -- Add second table",
		"[db.migration.stats] 4 applied 0 skipped 0 failed 4 total",
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()

	// Run again, should be a no-op
	suite, err = golembic.GenerateSuite(m)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Nil(err)
	logLines = []string{
		fmt.Sprintf("[db.migration] -- skipped -- Check table does not exist: %s", mt),
		"[db.migration] -- plan -- Determine migrations that need to be applied",
		"[db.migration] -- plan -- No migrations to run; latest revision: 60a33b9d4c77",
		"[db.migration.stats] 0 applied 1 skipped 0 failed 1 total",
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()

	// Manually delete the last migration and run again
	err = dropTable(ctx, pool, t2)
	it.Nil(err)
	statement := fmt.Sprintf("DELETE FROM %s WHERE revision = $1", golembic.QuoteIdentifier(mt))
	_, err = pool.Invoke(db.OptContext(ctx)).Exec(statement, "60a33b9d4c77")
	it.Nil(err)

	suite, err = golembic.GenerateSuite(m)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Nil(err)
	logLines = []string{
		fmt.Sprintf("[db.migration] -- skipped -- Check table does not exist: %s", mt),
		"[db.migration] -- plan -- Determine migrations that need to be applied",
		"[db.migration] -- 60a33b9d4c77 -- Add second table",
		"[db.migration.stats] 1 applied 1 skipped 0 failed 2 total",
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()
}

func TestGenerateSuite_StaleCheckout(t *testing.T) {
	it := assert.New(t)

	ctx := context.TODO()
	pool := defaultDB()
	it.NotNil(pool)

	// Happy path once (no need to verify logs, see above)
	suffix := anyLowercase(6)
	mt := fmt.Sprintf("foo_%s_migrations", suffix)
	t1 := fmt.Sprintf("foo1_%s", suffix)
	t2 := fmt.Sprintf("foo2_%s", suffix)
	t.Cleanup(func() {
		err1 := dropTable(ctx, pool, mt)
		err2 := dropTable(ctx, pool, t1)
		err3 := dropTable(ctx, pool, t2)
		it.Nil(err1)
		it.Nil(err2)
		it.Nil(err3)
	})
	var logBuffer bytes.Buffer
	log := logger.Memory(&logBuffer)

	migrations, err := makeSequence(t1, t2)
	it.Nil(err)
	m, err := golembic.NewManager(
		golembic.OptManagerSequence(migrations),
		golembic.OptManagerMetadataTable(mt),
		golembic.OptManagerLog(log),
	)
	it.Nil(err)
	suite, err := golembic.GenerateSuite(m)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Nil(err)
	logBuffer.Reset()

	// Introduce "new" migration that has not yet been seen
	statement := fmt.Sprintf(
		"INSERT INTO %s (serial_id, revision, previous) VALUES ($1, $2, $3)",
		golembic.QuoteIdentifier(mt),
	)
	_, err = pool.Invoke(db.OptContext(ctx)).Exec(statement, 3, "not-in-sequence", "60a33b9d4c77")
	it.Nil(err)

	suite, err = golembic.GenerateSuite(m)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Equal(`No migration registered for revision; Revision: "not-in-sequence"`, fmt.Sprintf("%v", err))
	logLines := []string{
		fmt.Sprintf("[db.migration] -- skipped -- Check table does not exist: %s", mt),
		"[db.migration] -- plan -- Determine migrations that need to be applied",
		"[db.migration.stats] 0 applied 1 skipped 0 failed 1 total",
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()
}

func makeSequence(t1, t2 string) (*golembic.Migrations, error) {
	ct1 := fmt.Sprintf("CREATE TABLE %s ( bar TEXT )", golembic.QuoteIdentifier(t1))
	root, err := golembic.NewMigration(
		golembic.OptRevision("aa60f058f5f5"),
		golembic.OptDescription("Create first table"),
		golembic.OptUpFromSQL(ct1),
	)
	if err != nil {
		return nil, err
	}

	migrations, err := golembic.NewSequence(*root)
	if err != nil {
		return nil, err
	}

	qt1 := golembic.QuoteIdentifier(t1)
	qt2 := golembic.QuoteIdentifier(t2)
	at1 := fmt.Sprintf("ALTER TABLE %s ADD COLUMN quux TEXT", qt1)
	ct2 := fmt.Sprintf("CREATE TABLE %s ( baz TEXT )", qt2)
	err = migrations.RegisterManyOpt(
		[]golembic.MigrationOption{
			golembic.OptPrevious("aa60f058f5f5"),
			golembic.OptRevision("ab1208989a3f"),
			golembic.OptDescription("Alter first table"),
			golembic.OptUpFromSQL(at1),
		},
		[]golembic.MigrationOption{
			golembic.OptPrevious("ab1208989a3f"),
			golembic.OptRevision("60a33b9d4c77"),
			golembic.OptDescription("Add second table"),
			golembic.OptUpFromSQL(ct2),
		},
	)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}

func anyLowercase(n int) string {
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func dropTable(ctx context.Context, pool *db.Connection, name string) error {
	statement := fmt.Sprintf("DROP TABLE IF EXISTS %s", golembic.QuoteIdentifier(name))
	_, err := pool.Invoke(db.OptContext(ctx)).Exec(statement)
	return err
}
