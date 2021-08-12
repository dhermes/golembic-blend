package golembic

import (
	"fmt"
)

const (
	createMigrationsTableSQL = `
CREATE TABLE %[1]s (
  serial_id  %[2]s,
  revision   %[3]s,
  previous   %[4]s,
  created_at %[5]s
)
`
)

// CreateTableParameters specifies a set of parameters that are intended
// to be used in a `CREATE TABLE` statement.
type CreateTableParameters struct {
	SerialID  string
	Revision  string
	Previous  string
	CreatedAt string
}

func createMigrationsSQL(gc GenerateConfig) (CreateTableParameters, string) {
	table := gc.MetadataTable
	ctp := providerNewCreateTableParameters()

	statement := fmt.Sprintf(
		createMigrationsTableSQL,
		providerQuoteIdentifier(table), // [1]
		ctp.SerialID,                   // [2]
		ctp.Revision,                   // [3]
		ctp.Previous,                   // [4]
		ctp.CreatedAt,                  // [5]
	)
	return ctp, statement
}

func createMigrationsStatements(gc GenerateConfig) []string {
	_, createTable := createMigrationsSQL(gc)
	return []string{createTable}
}
