package golembic

// OptManagerMetadataTable sets the metadata table name on a manager.
func OptManagerMetadataTable(table string) ManagerOption {
	return func(m *Manager) error {
		m.MetadataTable = table
		return nil
	}
}

// OptManagerSequence sets the migrations sequence on a manager.
func OptManagerSequence(migrations *Migrations) ManagerOption {
	return func(m *Manager) error {
		m.Sequence = migrations
		return nil
	}
}

// OptDevelopmentMode sets the development mode flag on a manager.
func OptDevelopmentMode(mode bool) ManagerOption {
	return func(m *Manager) error {
		m.DevelopmentMode = mode
		return nil
	}
}
