package golembic_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/db"

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

	migrations, err := makeSequence(t1, t2)
	it.Nil(err)
	m, err := golembic.NewManager(
		golembic.OptManagerSequence(migrations),
		golembic.OptManagerMetadataTable(mt),
	)
	it.Nil(err)
	suite, err := golembic.GenerateSuite(m)
	it.Nil(err)

	err = golembic.ApplyDynamic(ctx, suite, pool)
	it.Nil(err)
}

func makeSequence(t1, t2 string) (*golembic.Migrations, error) {
	ct1 := fmt.Sprintf("CREATE TABLE %s ( bar TEXT )", golembic.QuoteIdentifier(t1))
	root, err := golembic.NewMigration(
		golembic.OptRevision("aa60f058f5f5"),
		golembic.OptDescription("Create table"),
		golembic.OptUpFromSQL(ct1),
	)
	if err != nil {
		return nil, err
	}

	migrations, err := golembic.NewSequence(*root)
	if err != nil {
		return nil, err
	}

	qt2 := golembic.QuoteIdentifier(t2)
	ct2 := fmt.Sprintf("CREATE TABLE %s ( baz TEXT )", qt2)
	at2 := fmt.Sprintf("ALTER TABLE %s ADD COLUMN quux TEXT", qt2)
	err = migrations.RegisterManyOpt(
		[]golembic.MigrationOption{
			golembic.OptPrevious("aa60f058f5f5"),
			golembic.OptRevision("60a33b9d4c77"),
			golembic.OptDescription("Add second table"),
			golembic.OptUpFromSQL(ct2),
		},
		[]golembic.MigrationOption{
			golembic.OptPrevious("60a33b9d4c77"),
			golembic.OptRevision("ab1208989a3f"),
			golembic.OptDescription("Alter second table"),
			golembic.OptUpFromSQL(at2),
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
