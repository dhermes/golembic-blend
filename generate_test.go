package golembic_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/bufferutil"
	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/logger"

	golembic "github.com/dhermes/golembic-blend"
)

// NOTE: Ensure that
//       * `jsonNoTimestamp` satisfies `logger.WriteFormatter`.
var (
	_ logger.WriteFormatter = (*jsonNoTimestamp)(nil)
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

	migrations, err := makeSequence(t1, t2, 3, false)
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
	mt := fmt.Sprintf("bar_%s_migrations", suffix)
	t1 := fmt.Sprintf("bar1_%s", suffix)
	t2 := fmt.Sprintf("bar2_%s", suffix)
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

	migrations, err := makeSequence(t1, t2, 3, false)
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

	// Same failure, but with `--verify-history` turned on
	mVerify, err := golembic.NewManager(
		golembic.OptManagerSequence(migrations),
		golembic.OptManagerMetadataTable(mt),
		golembic.OptManagerLog(log),
		golembic.OptManagerVerifyHistory(true),
	)
	it.Nil(err)
	suite, err = golembic.GenerateSuite(mVerify)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Equal("Migration stored in SQL doesn't match sequence", fmt.Sprintf("%v", err))
	logLines = []string{
		fmt.Sprintf("[db.migration] -- skipped -- Check table does not exist: %s", mt),
		"[db.migration] -- plan -- Determine migrations that need to be applied",
		"[db.migration] -- failed -- Sequence has 3 migrations but 4 are stored in the table",
		"[db.migration.stats] 0 applied 1 skipped 0 failed 1 total",
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()

	// Replace last revision with nonsense
	statement = fmt.Sprintf(
		"DELETE FROM %s WHERE revision IN ($1, $2)",
		golembic.QuoteIdentifier(mt),
	)
	_, err = pool.Invoke(db.OptContext(ctx)).Exec(statement, "not-in-sequence", "60a33b9d4c77")
	it.Nil(err)
	statement = fmt.Sprintf(
		"INSERT INTO %s (serial_id, revision, previous) VALUES ($1, $2, $3)",
		golembic.QuoteIdentifier(mt),
	)
	_, err = pool.Invoke(db.OptContext(ctx)).Exec(statement, 2, "not-in-sequence", "ab1208989a3f")
	it.Nil(err)

	suite, err = golembic.GenerateSuite(m)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Equal(`No migration registered for revision; Revision: "not-in-sequence"`, fmt.Sprintf("%v", err))
	logLines = []string{
		fmt.Sprintf("[db.migration] -- skipped -- Check table does not exist: %s", mt),
		"[db.migration] -- plan -- Determine migrations that need to be applied",
		"[db.migration.stats] 0 applied 1 skipped 0 failed 1 total",
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()

	// Replace last revision with nonsense, but with `--verify-history` turned on
	suite, err = golembic.GenerateSuite(mVerify)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Equal("Migration stored in SQL doesn't match sequence", fmt.Sprintf("%v", err))
	logLines = []string{
		fmt.Sprintf("[db.migration] -- skipped -- Check table does not exist: %s", mt),
		"[db.migration] -- plan -- Determine migrations that need to be applied",
		`[db.migration] -- failed -- Stored migration 2: "not-in-sequence:ab1208989a3f" does not match migration "60a33b9d4c77:ab1208989a3f" in sequence`,
		"[db.migration.stats] 0 applied 1 skipped 0 failed 1 total",
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()
}

func TestGenerateSuite_Milestone(t *testing.T) {
	it := assert.New(t)

	ctx := context.TODO()
	pool := defaultDB()
	it.NotNil(pool)

	suffix := anyLowercase(6)
	mt := fmt.Sprintf("baz_%s_migrations", suffix)
	t1 := fmt.Sprintf("baz1_%s", suffix)
	t2 := fmt.Sprintf("baz2_%s", suffix)
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

	// Run **just** the first migration
	migrations1, err := makeSequence(t1, t2, 1, true)
	it.Nil(err)
	m1, err := golembic.NewManager(
		golembic.OptManagerSequence(migrations1),
		golembic.OptManagerMetadataTable(mt),
		golembic.OptManagerLog(log),
	)
	it.Nil(err)
	suite, err := golembic.GenerateSuite(m1)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Nil(err)
	logLines := []string{
		fmt.Sprintf("[db.migration] -- applied -- Check table does not exist: %s", mt),
		"[db.migration] -- plan -- Determine migrations that need to be applied",
		"[db.migration] -- aa60f058f5f5 -- Create first table",
		"[db.migration.stats] 2 applied 0 skipped 0 failed 2 total",
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()

	// **Try** to run all 3 migrations, with a milestone in the middle
	migrations3, err := makeSequence(t1, t2, 3, true)
	it.Nil(err)
	m3, err := golembic.NewManager(
		golembic.OptManagerSequence(migrations3),
		golembic.OptManagerMetadataTable(mt),
		golembic.OptManagerLog(log),
	)
	it.Nil(err)
	suite, err = golembic.GenerateSuite(m3)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Equal("If a migration sequence contains a milestone, it must be the last migration", fmt.Sprintf("%v", err))
	logLines = []string{
		fmt.Sprintf("[db.migration] -- skipped -- Check table does not exist: %s", mt),
		"[db.migration] -- plan -- Determine migrations that need to be applied",
		"[db.migration] -- failed -- Revision ab1208989a3f (1 / 2 migrations)",
		"[db.migration.stats] 0 applied 1 skipped 0 failed 1 total",
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()

	// Apply **only** milestone before trying all 3 migrations
	migrations2, err := makeSequence(t1, t2, 2, true)
	it.Nil(err)
	m2, err := golembic.NewManager(
		golembic.OptManagerSequence(migrations2),
		golembic.OptManagerMetadataTable(mt),
		golembic.OptManagerLog(log),
	)
	it.Nil(err)
	suite, err = golembic.GenerateSuite(m2)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Nil(err)
	logLines = []string{
		fmt.Sprintf("[db.migration] -- skipped -- Check table does not exist: %s", mt),
		"[db.migration] -- plan -- Determine migrations that need to be applied",
		"[db.migration] -- ab1208989a3f -- Alter first table [MILESTONE]",
		"[db.migration.stats] 1 applied 1 skipped 0 failed 2 total",
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()

	// **Finally** apply all 3 migrations
	suite, err = golembic.GenerateSuite(m3)
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

func TestGenerateSuite_FailedDDL(t *testing.T) {
	it := assert.New(t)

	ctx := context.TODO()
	pool := defaultDB()
	it.NotNil(pool)

	suffix := anyLowercase(6)
	mt := fmt.Sprintf("quux_%s_migrations", suffix)
	t1 := fmt.Sprintf("quux1_%s", suffix)
	t.Cleanup(func() {
		err1 := dropTable(ctx, pool, mt)
		err2 := dropTable(ctx, pool, t1)
		it.Nil(err1)
		it.Nil(err2)
	})

	qt1 := golembic.QuoteIdentifier(t1)
	ct1 := fmt.Sprintf("CREATE TABLE %s ( bar TEXT )", qt1)
	root, err := golembic.NewMigration(
		golembic.OptRevision("af808e6e4d5b"),
		golembic.OptDescription("Create table first time"),
		golembic.OptUpFromSQL(ct1),
	)
	it.Nil(err)
	migrations, err := golembic.NewSequence(*root)
	it.Nil(err)
	err = migrations.RegisterManyOpt(
		[]golembic.MigrationOption{
			golembic.OptPrevious("af808e6e4d5b"),
			golembic.OptRevision("52d1d91b4f7e"),
			golembic.OptDescription("Create table second time"),
			golembic.OptUpFromSQL(ct1),
		},
	)
	it.Nil(err)

	var logBuffer bytes.Buffer
	of := newJSONNoTimestamp()
	log := logger.Memory(&logBuffer, logger.OptFormatter(of))
	m, err := golembic.NewManager(
		golembic.OptManagerSequence(migrations),
		golembic.OptManagerMetadataTable(mt),
		golembic.OptManagerLog(log),
	)
	it.Nil(err)
	suite, err := golembic.GenerateSuite(m)
	it.Nil(err)
	err = golembic.ApplyDynamic(ctx, suite, pool)
	expected := fmt.Sprintf("ERROR: relation %q already exists (SQLSTATE 42P07); %s", t1, ct1)
	it.Equal(expected, fmt.Sprintf("%v", err))

	logLines := []string{
		fmt.Sprintf(`{"body":"Check table does not exist: %s","flag":"db.migration","labels":null,"result":"applied"}`, mt),
		`{"body":"Determine migrations that need to be applied","flag":"db.migration","labels":null,"result":"plan"}`,
		`{"body":"Create table first time","flag":"db.migration","labels":null,"revision":"af808e6e4d5b","status":"applied"}`,
		`{"body":"Create table second time","flag":"db.migration","labels":null,"revision":"52d1d91b4f7e","status":"failed"}`,
		`{"applied":2,"failed":1,"flag":"db.migration.stats","skipped":0,"total":3}`,
		"",
	}
	it.Equal(strings.Join(logLines, "\n"), logBuffer.String())
	logBuffer.Reset()
}

func makeSequence(t1, t2 string, length int, milestone bool) (*golembic.Migrations, error) {
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
	opts := [][]golembic.MigrationOption{
		{
			golembic.OptPrevious("aa60f058f5f5"),
			golembic.OptRevision("ab1208989a3f"),
			golembic.OptMilestone(milestone),
			golembic.OptDescription("Alter first table"),
			golembic.OptUpFromSQL(at1),
		},
		{
			golembic.OptPrevious("ab1208989a3f"),
			golembic.OptRevision("60a33b9d4c77"),
			golembic.OptDescription("Add second table"),
			golembic.OptUpFromSQL(ct2),
		},
	}
	// NOTE: This will panic for many values of `length`, caller must vet.
	opts = opts[:length-1]
	err = migrations.RegisterManyOpt(opts...)
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

type jsonNoTimestamp struct {
	Wrapped *logger.JSONOutputFormatter
}

func newJSONNoTimestamp() *jsonNoTimestamp {
	jf := &logger.JSONOutputFormatter{
		BufferPool: bufferutil.NewPool(logger.DefaultBufferPoolSize),
	}
	return &jsonNoTimestamp{Wrapped: jf}
}

func (jnt jsonNoTimestamp) GetScopeFields(ctx context.Context, e logger.Event) map[string]interface{} {
	fields := jnt.Wrapped.GetScopeFields(ctx, e)
	delete(fields, logger.FieldTimestamp)
	return fields
}

func (jw jsonNoTimestamp) WriteFormat(ctx context.Context, output io.Writer, e logger.Event) error {
	buffer := jw.Wrapped.BufferPool.Get()
	defer jw.Wrapped.BufferPool.Put(buffer)

	encoder := json.NewEncoder(buffer)
	if jw.Wrapped.Pretty {
		encoder.SetIndent(jw.Wrapped.PrettyPrefixOrDefault(), jw.Wrapped.PrettyIndentOrDefault())
	}
	if decomposer, ok := e.(logger.JSONWritable); ok {
		fields := jw.Wrapped.CombineFields(jw.GetScopeFields(ctx, e), decomposer.Decompose())
		if err := encoder.Encode(fields); err != nil {
			return err
		}
	} else if err := encoder.Encode(e); err != nil {
		return err
	}
	_, err := io.Copy(output, buffer)
	return err
}
