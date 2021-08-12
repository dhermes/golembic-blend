package golembic

const (
	// DefaultMetadataTable is the default name for the table used to store
	// metadata about migrations.
	DefaultMetadataTable = "golembic_migrations"
)

type GenerateConfig struct {
	// MetadataTable is the name of the table that stores migration metadata.
	// The expected default value (`DefaultMetadataTable`) is
	// "golembic_migrations".
	MetadataTable string
	// Sequence is the collection of registered migrations to be applied,
	// verified, described, etc. by this generate config.
	Sequence *Migrations
	// DevelopmentMode is a flag indicating that this generate config is currently
	// being run in development mode, so things like extra validation should
	// intentionally be disabled. This is intended for use in testing and
	// development, where an entire database is spun up locally (e.g. in Docker)
	// and migrations will be applied from scratch (including milestones that
	// may not come at the end).
	DevelopmentMode bool
}

// NewGenerateConfig creates a new config for generating a migrations
// suite.
func NewGenerateConfig(opts ...GenerateConfigOption) GenerateConfig {
	gc := GenerateConfig{MetadataTable: DefaultMetadataTable}
	for _, opt := range opts {
		opt(&gc)
	}
	return gc
}
