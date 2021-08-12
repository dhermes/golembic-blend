package golembic

// OptMetadataTable sets the metadata table name on a generate config.
func OptMetadataTable(table string) ManagerOption {
	return func(gc *Manager) {
		gc.MetadataTable = table
	}
}

// OptSequence sets the migrations sequence on a generate config.
func OptSequence(migrations *Migrations) ManagerOption {
	return func(gc *Manager) {
		gc.Sequence = migrations
	}
}

// OptDevelopmentMode sets the development mode flag on a generate config.
func OptDevelopmentMode(mode bool) ManagerOption {
	return func(gc *Manager) {
		gc.DevelopmentMode = mode
	}
}
