package interfaces

import (
	"context"

	"gorm.io/gorm"
)

// DatabaseServiceInterface defines the contract for the database service
type DatabaseServiceInterface interface {
	// Name returns the service name
	Name() string

	// Init initializes the service
	Init(ctx context.Context) error

	// IsEnabled returns true if the service is enabled
	IsEnabled() bool

	// Dependencies returns the service dependencies
	Dependencies() []string

	// GetDB returns the underlying gorm.DB connection
	GetDB() *gorm.DB

	// Close closes the database connection
	Close() error

	// InitStores initializes all registered stores with the database connection
	InitStores(ctx context.Context) error

	// Health checks the database health
	Health(ctx context.Context) error
}
