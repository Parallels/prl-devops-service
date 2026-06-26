package interfaces

import (
	"context"

	"gorm.io/gorm"
)

// Store defines the interface that all data stores must implement
// to be managed by the Database Service
type Store interface {
	// Name returns the name of the store
	Name() string
	// Init initializes the store with the given database connection
	Init(ctx context.Context, db *gorm.DB) error
	// Health checks the store health
	Health(ctx context.Context) error
	// IsEnabled checks if the store is enabled
	IsEnabled() bool
	// Dependencies returns the store dependencies
	Dependencies() []string
	// Migrate runs the store-specific migrations
	Migrate() error
}
