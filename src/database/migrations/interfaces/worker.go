package interfaces

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	apperrors "github.com/Parallels/prl-devops-service/errors"
)

// MigrationWorker defines the interface for database migration workers
type MigrationWorker interface {
	// GetName returns the unique name of this migration
	GetName() string

	// GetDescription returns a human-readable description
	GetDescription() string

	// GetOrder returns the execution order (lower numbers run first)
	GetOrder() int

	// Run executes the migration
	Run(ctx basecontext.BaseContext) *apperrors.Diagnostics

	// Rollback reverts the migration (optional, can return nil if not supported)
	Rollback(ctx basecontext.BaseContext) *apperrors.Diagnostics
}
