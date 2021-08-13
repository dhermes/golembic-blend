package golembic

import (
	"github.com/blend/go-sdk/ex"
	"github.com/blend/go-sdk/logger"
)

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

// OptManagerVerifyHistory sets `VerifyHistory` on a manager.
func OptManagerVerifyHistory(verify bool) ManagerOption {
	return func(m *Manager) error {
		m.VerifyHistory = verify
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

// OptManagerLog sets the logger interface on a manager. If `log` is `nil`code man
// the option will return an error.
func OptManagerLog(log logger.Log) ManagerOption {
	return func(m *Manager) error {
		if log == nil {
			return ex.New(ErrNilInterface)
		}

		m.Log = log
		return nil
	}
}
