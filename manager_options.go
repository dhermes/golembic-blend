package golembic

// OptMetadataTable sets the metadata table name on a manager.
func OptMetadataTable(table string) ManagerOption {
	return func(m *Manager) {
		m.MetadataTable = table
	}
}

// OptSequence sets the migrations sequence on a manager.
func OptSequence(migrations *Migrations) ManagerOption {
	return func(m *Manager) {
		m.Sequence = migrations
	}
}

// OptDevelopmentMode sets the development mode flag on a manager.
func OptDevelopmentMode(mode bool) ManagerOption {
	return func(m *Manager) {
		m.DevelopmentMode = mode
	}
}
