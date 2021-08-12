package main

import (
	"context"
	"fmt"
	"os"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/logger"

	golembic "github.com/dhermes/golembic-blend"
	"github.com/dhermes/golembic-blend/examples"
)

func run() error {
	migrations, err := examples.AllMigrations()
	if err != nil {
		return err
	}

	m := golembic.NewManager(golembic.OptSequence(migrations))
	suite, err := golembic.GenerateSuite(m)
	if err != nil {
		return err
	}

	suite.Log = logger.All()

	ctx := context.Background()
	pool, err := getPool(ctx)
	if err != nil {
		return err
	}

	return golembic.ApplyDynamic(ctx, suite, pool)
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func getPool(ctx context.Context) (*db.Connection, error) {
	c := db.Config{
		Host:     "127.0.0.1",
		Port:     "23396",
		Database: "golembic",
		Username: "golembic_admin",
		Password: "testpassword_admin",
		SSLMode:  "disable",
	}
	pool, err := db.New(db.OptConfig(c))
	if err != nil {
		return nil, err
	}

	err = pool.Open()
	if err != nil {
		return nil, err
	}

	err = pool.Connection.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
