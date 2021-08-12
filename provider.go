package golembic

import (
	"fmt"
)

// providerNewCreateTableParameters is a concrete implementation for PostgreSQL
// table parameters / column constraints. In `github.com/dhermes/golembic`,
// this is abstracted away into the `EngineProvider` interface but we don't
// have a need for that here.
func providerNewCreateTableParameters() CreateTableParameters {
	return CreateTableParameters{
		SerialID:  "INTEGER NOT NULL",
		Revision:  "VARCHAR(32) NOT NULL",
		Previous:  "VARCHAR(32)",
		CreatedAt: "TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP",
	}
}

// providerQuoteIdentifier is a concrete implementation for PostgreSQL
// identifier quoting. In `github.com/dhermes/golembic`, this is abstracted
// away into the `EngineProvider` interface but we don't have a need for that
// here.
func providerQuoteIdentifier(name string) string {
	return QuoteIdentifier(name)
}

// providerQueryParameter is a concrete implementation for PostgreSQL
// integer parameters. In `github.com/dhermes/golembic`, this is abstracted
// away into the `EngineProvider` interface but we don't have a need for that
// here.
func providerQueryParameter(index int) string {
	return fmt.Sprintf("$%d", index)
}
