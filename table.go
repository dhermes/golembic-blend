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
	pkMigrationsTableSQL = `
ALTER TABLE %[1]s
  ADD CONSTRAINT %[2]s PRIMARY KEY (revision)
`
	fkPreviousMigrationsTableSQL = `
ALTER TABLE %[1]s
  ADD CONSTRAINT %[2]s FOREIGN KEY (previous)
  REFERENCES %[1]s(revision)
`
	uqSerialIDSQL = `
ALTER TABLE %[1]s
  ADD CONSTRAINT %[2]s UNIQUE (serial_id)
`
	nonNegativeSerialIDSQL = `
ALTER TABLE %[1]s
  ADD CONSTRAINT %[2]s CHECK (serial_id >= 0)
`
	uqPreviousMigrationsTableSQL = `
ALTER TABLE %[1]s
  ADD CONSTRAINT %[1]s UNIQUE (previous)
	`

	noCyclesMigrationsTableSQL = `
ALTER TABLE %[1]s
  ADD CONSTRAINT %[2]s CHECK (previous != revision)
`
	singleRootMigrationsTableSQL = `
ALTER TABLE %[1]s
  ADD CONSTRAINT %[2]s CHECK
  (
    (serial_id = 0 AND previous IS NULL) OR
    (serial_id != 0 AND previous IS NOT NULL)
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

// pkMigrationsSQL ensures the `revision` is used as the primary key in
// the table.
func pkMigrationsSQL(gc GenerateConfig) string {
	table := gc.MetadataTable
	pkConstraint := fmt.Sprintf("pk_%s_revision", table)

	return fmt.Sprintf(
		pkMigrationsTableSQL,
		providerQuoteIdentifier(table), // [1]
		pkConstraint,                   // [2]
	)
}

// fkPreviousMigrationsSQL ensures the `previous` column is always a foreign
// key to an existing `revision` (or `NULL`).
func fkPreviousMigrationsSQL(gc GenerateConfig) string {
	table := gc.MetadataTable
	fkConstraint := fmt.Sprintf("fk_%s_previous", table)

	return fmt.Sprintf(
		fkPreviousMigrationsTableSQL,
		providerQuoteIdentifier(table), // [1]
		fkConstraint,                   // [2]
	)
}

// uqSerialID ensures the `serial_id` column is UNIQUE.
func uqSerialID(gc GenerateConfig) string {
	table := gc.MetadataTable
	uqConstraint := fmt.Sprintf("uq_%s_serial_id", table)

	return fmt.Sprintf(
		uqSerialIDSQL,
		providerQuoteIdentifier(table), // [1]
		uqConstraint,                   // [2]
	)
}

// nonNegativeSerialID ensures the `serial_id` is not a negative number.
func nonNegativeSerialID(gc GenerateConfig) string {
	table := gc.MetadataTable
	chkConstraint := fmt.Sprintf("chk_%s_serial_id", table)

	return fmt.Sprintf(
		nonNegativeSerialIDSQL,
		providerQuoteIdentifier(table), // [1]
		chkConstraint,                  // [2]
	)
}

// uqPreviousMigrationsSQL ensures the `previous` column is UNIQUE.
func uqPreviousMigrationsSQL(gc GenerateConfig) string {
	table := gc.MetadataTable
	uqConstraint := fmt.Sprintf("uq_%s_previous", table)

	return fmt.Sprintf(
		uqPreviousMigrationsTableSQL,
		providerQuoteIdentifier(table), // [1]
		uqConstraint,                   // [2]
	)
}

// noCyclesMigrationsSQL ensures no cycles can be introduced by having
// `previous` equal to `revision` in a row.
func noCyclesMigrationsSQL(gc GenerateConfig) string {
	table := gc.MetadataTable
	chkConstraint := fmt.Sprintf("chk_%s_previous_neq_revision", table)

	return fmt.Sprintf(
		noCyclesMigrationsTableSQL,
		providerQuoteIdentifier(table), // [1]
		chkConstraint,                  // [2]
	)
}

// singleRootMigrationsSQL exactly **one** root migration (i.e. one with
// `previous=NULL`) can be stored in the table. Additionally it makes sure
// that `serial_id = 0` must be the root as well.
func singleRootMigrationsSQL(gc GenerateConfig) string {
	table := gc.MetadataTable
	nullPreviousIndex := fmt.Sprintf("chk_%s_null_previous", table)

	return fmt.Sprintf(
		singleRootMigrationsTableSQL,
		providerQuoteIdentifier(table),             // [1]
		providerQuoteIdentifier(nullPreviousIndex), // [2]
	)
}

func createMigrationsStatements(gc GenerateConfig) []string {
	_, createTable := createMigrationsSQL(gc)
	return []string{
		createTable,
		pkMigrationsSQL(gc),
		fkPreviousMigrationsSQL(gc),
		uqSerialID(gc),
		nonNegativeSerialID(gc),
		uqPreviousMigrationsSQL(gc),
		noCyclesMigrationsSQL(gc),
		singleRootMigrationsSQL(gc),
	}
}
