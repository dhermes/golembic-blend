package main

import (
	"context"
	"fmt"
	"os"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/logger"
	"github.com/spf13/cobra"

	golembic "github.com/dhermes/golembic-blend"
	"github.com/dhermes/golembic-blend/examples"
)

func run(length int) error {
	migrations, err := examples.AllMigrations(length)
	if err != nil {
		return err
	}

	log := logger.All()
	m, err := golembic.NewManager(
		golembic.OptManagerSequence(migrations),
		golembic.OptManagerLog(log),
		golembic.OptManagerVerifyHistory(true),
	)
	if err != nil {
		return err
	}

	suite, err := golembic.GenerateSuite(m)
	if err != nil {
		return err
	}

	suite.Log = log

	ctx := context.Background()
	pool, err := getPool(ctx)
	if err != nil {
		return err
	}

	return golembic.ApplyDynamic(ctx, suite, pool)
}

func root() *cobra.Command {
	length := -1
	cmd := &cobra.Command{
		Use:           "golembic-blend-example",
		Short:         "Run example database migrations via golembic-blend",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(length)
		},
	}

	cmd.PersistentFlags().IntVar(
		&length,
		"length",
		-1,
		"The length of the sequence to be run. Must be one of -1, 1, ..., 7.",
	)

	return cmd
}

func main() {
	cmd := root()
	err := cmd.Execute()
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
