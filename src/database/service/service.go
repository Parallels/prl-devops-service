package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/Parallels/prl-devops-service/database/common"
	dbinterfaces "github.com/Parallels/prl-devops-service/database/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	instance *DatabaseService
	once     sync.Once
)

// DatabaseService represents the database service
type DatabaseService struct {
	db               *gorm.DB
	config           *common.Config
	registeredStores []dbinterfaces.Store
	mu               sync.RWMutex
}

// GetInstance returns the singleton instance of the database service
func GetInstance() *DatabaseService {
	return instance
}

// Reset clears the singleton instance (for testing)
func Reset() {
	if instance != nil {
		_ = instance.Close()
		instance = nil
	}
	once = sync.Once{}
}

// Name returns the name of the service
func (s *DatabaseService) Name() string {
	return "database"
}

// Dependencies returns the dependencies of the service
func (s *DatabaseService) Dependencies() []string {
	return []string{}
}

// IsEnabled returns true if the service is enabled
func (s *DatabaseService) IsEnabled() bool {
	return true
}

// Init initializes the database service (for system.Service interface)
func (s *DatabaseService) Init(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Additional initialization logic can go here
	return nil
}

// Initialize bootstraps the database service with the given config
func Initialize(cfg common.Config) (*DatabaseService, error) {
	// Validate configuration
	if diag := cfg.Validate(); diag.HasErrors() {
		return nil, fmt.Errorf("configuration validation failed: %v", diag.GetErrors())
	}

	var initErr error
	once.Do(func() {
		svc := &DatabaseService{}
		if err := svc.initialize(&cfg); err != nil {
			initErr = err
			return
		}

		instance = svc
	})

	if initErr != nil {
		return nil, fmt.Errorf("initialization failed: %w", initErr)
	}

	if instance == nil {
		return nil, fmt.Errorf("database service instance is nil after initialization")
	}

	return instance, nil
}

// initialize sets up the database connection
func (s *DatabaseService) initialize(config *common.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Configure logging
	logLevel := logger.Silent
	if config.Debug {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	var db *gorm.DB
	var err error

	// Initialize database based on type
	switch config.Type {
	case common.SQLite:
		db, err = initializeSQLite(config, gormConfig)
	case common.PostgreSQL:
		db, err = initializePostgreSQL(config, gormConfig)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err != nil {
		return err
	}

	s.db = db
	s.config = config

	return nil
}

// InitStores initializes all registered stores with the database connection
func (s *DatabaseService) InitStores(ctx context.Context) error {
	// This will be populated with actual stores as they are migrated
	// For now, just return nil
	return nil
}

// RegisterStore registers a store with the database service
func (s *DatabaseService) RegisterStore(store dbinterfaces.Store) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registeredStores = append(s.registeredStores, store)
}

// GetStores returns the registered stores
func (s *DatabaseService) GetStores() []dbinterfaces.Store {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.registeredStores
}

// Health checks the health of the database
func (s *DatabaseService) Health(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// GetDB returns the database connection
func (s *DatabaseService) GetDB() *gorm.DB {
	return s.db
}

// Close closes the database connection
func (s *DatabaseService) Close() error {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// New returns a new DatabaseService (helper for system registration)
func New() *DatabaseService {
	return &DatabaseService{}
}
