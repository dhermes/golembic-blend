package golembic_test

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/blend/go-sdk/db"
	"github.com/blend/go-sdk/env"
)

var (
	defaultPool     *db.Connection
	defaultPoolLock sync.RWMutex
)

// configDefaults specifies the default configuration; this is intended to match
// the defaults provided in the root `Makefile` for the `deployinator` repo.
func configDefaults() db.Config {
	return db.Config{
		Host:     "127.0.0.1",
		Port:     "23396",
		Database: "golembic",
		Username: "golembic_admin",
		Password: "testpassword_admin",
		SSLMode:  "disable",
	}
}

func resolveAndSetConfig(ctx context.Context, pool *db.Connection, cfg db.Config) error {
	err := (&cfg).Resolve(env.WithVars(ctx, env.Env()))
	if err != nil {
		return err
	}

	pool.Config = cfg
	return nil
}

// setConfig starts from expected defaults but allows for overrides via
// environment variable(s).
func setConfig(ctx context.Context, pool *db.Connection) error {
	cfg := configDefaults()
	return resolveAndSetConfig(ctx, pool, cfg)
}

func requireDB(ctx context.Context) error {
	// Early exit if the pool has already been created and cached.
	pool := defaultDB()
	if pool != nil {
		return nil
	}

	// Create a database connection pool (this should be equivalent to just
	// `pool := &db.Connection{}` but we call `.New()` in case the implementation
	// of `.New()` changes over time).
	pool, err := db.New()
	if err != nil {
		return err
	}

	// Set configuration (likely based on defaults + environment overrides).
	err = setConfig(ctx, pool)
	if err != nil {
		return err
	}

	// Verify the connection string is valid.
	err = pool.Open()
	if err != nil {
		return err
	}

	// Verify that the database is actually running / we can actually connect.
	err = pool.Connection.PingContext(ctx)
	if err != nil {
		return err
	}

	// Cache the connection pool before returning.
	defaultPoolLock.Lock()
	defer defaultPoolLock.Unlock()
	defaultPool = pool
	return nil
}

func defaultDB() *db.Connection {
	defaultPoolLock.RLock()
	defer defaultPoolLock.RUnlock()
	return defaultPool
}

func poolClose() {
	pool := defaultDB()
	if pool == nil {
		return
	}
	err := pool.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error closing pool: %v\n", err)
	}
}
