package golembic

// OptMetadataTable sets the metadata table name on a generate config.
func OptMetadataTable(table string) GenerateConfigOption {
	return func(gc *GenerateConfig) {
		gc.MetadataTable = table
	}
}

// OptSequence sets the migrations sequence on a generate config.
func OptSequence(migrations *Migrations) GenerateConfigOption {
	return func(gc *GenerateConfig) {
		gc.Sequence = migrations
	}
}

// OptDevelopmentMode sets the development mode flag on a generate config.
func OptDevelopmentMode(mode bool) GenerateConfigOption {
	return func(gc *GenerateConfig) {
		gc.DevelopmentMode = mode
	}
}
