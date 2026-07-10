package dbservice

import (
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/serviceprovider/dbservice/domains/auth"
	"github.com/Parallels/prl-devops-service/serviceprovider/dbservice/interfaces"
	"gorm.io/gorm"
)

var (
	globalDatabaseService *DatabaseService
	dbMutex               sync.RWMutex
)

// DatabaseService provides domain-oriented database operations
type DatabaseService struct {
	db     *gorm.DB
	stores *StoreRegistry

	// Domain services with lazy initialization
	authOnce sync.Once
	auth     interfaces.AuthDomain
}

// NewDatabaseService creates a new database service using the builder pattern
func NewDatabaseService(cfg *config.Config, ctx basecontext.ApiContext) (*DatabaseService, *apperrors.Diagnostics) {
	return NewBuilder(cfg, ctx).Build()
}

// Auth returns the authentication domain service (lazy-loaded)
func (s *DatabaseService) Auth() interfaces.AuthDomain {
	s.authOnce.Do(func() {
		s.auth = auth.NewService(
			s.stores.User(),
			s.stores.Role(),
			s.stores.Claim(),
		)
	})
	return s.auth
}

// WithTransaction executes fn within a database transaction
// If fn returns an error, the transaction is rolled back
// If fn returns nil, the transaction is committed
func (s *DatabaseService) WithTransaction(ctx basecontext.ApiContext, fn func(tx *gorm.DB) error) error {
	return s.db.WithContext(ctx.Context()).Transaction(fn)
}

// DB returns the underlying database connection (use with caution)
func (s *DatabaseService) DB() *gorm.DB {
	return s.db
}

// Stores returns the store registry
func (s *DatabaseService) Stores() *StoreRegistry {
	return s.stores
}

// IsConnected returns true if database is initialized
func (s *DatabaseService) IsConnected() bool {
	return s != nil && s.db != nil
}

// InitDatabase initializes the global database service
// This should be called once during application startup
func InitDatabase(ctx basecontext.ApiContext) (*DatabaseService, *apperrors.Diagnostics) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if globalDatabaseService != nil {
		ctx.LogDebugf("Database service already initialized")
		return globalDatabaseService, nil
	}

	cfg := config.Get()
	db, diag := NewDatabaseService(cfg, ctx)
	if diag != nil && diag.HasErrors() {
		ctx.LogErrorf("Failed to initialize database service: %v", diag.GetErrors())
		return nil, diag
	}

	globalDatabaseService = db
	ctx.LogInfof("Database service initialized successfully")
	return db, nil
}

// GetDatabaseService returns the global database service instance
// Returns nil if InitDatabase has not been called
func GetDatabaseService() *DatabaseService {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	return globalDatabaseService
}

// ResetDatabaseService clears the global instance (for testing only)
func ResetDatabaseService() {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	globalDatabaseService = nil
}
