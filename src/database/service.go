package database

import (
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/database/interfaces"
	"github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/database/stores"
	apperrors "github.com/Parallels/prl-devops-service/errors"
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

	userOnce sync.Once
	user     interfaces.UserDomain
}

// NewDatabaseService creates a new database service using the builder pattern
func NewDatabaseService(cfg *config.Config, ctx basecontext.ApiContext) (*DatabaseService, *apperrors.Diagnostics) {
	return NewBuilder(cfg, ctx).Build()
}

// Auth returns the authentication domain service (lazy-loaded)
func (s *DatabaseService) Auth() interfaces.AuthDomain {
	s.authOnce.Do(func() {
		s.auth = stores.NewAuthService(
			s.stores.User(),
			s.stores.Role(),
			s.stores.Claim(),
		)
	})
	return s.auth
}

// User returns the user domain service (lazy-loaded)
func (s *DatabaseService) User() interfaces.UserDomain {
	s.userOnce.Do(func() {
		s.user = stores.NewUserDomainService(
			s.stores.User(),
			s.stores.UserConfig(),
		)
	})
	return s.user
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

// ============================================================================
// Convenience methods delegating to domain services
// These provide a cleaner API for controllers
// ============================================================================

// GetUserConfig retrieves a single user config (delegates to User domain)
func (s *DatabaseService) GetUserConfig(ctx basecontext.ApiContext, userID, idOrSlug string) (*models.UserConfig, *apperrors.Diagnostics) {
	return s.User().GetUserConfig(ctx, userID, idOrSlug)
}

// GetUserConfigs retrieves all user configs with filtering (delegates to User domain)
// The filter parameter can be an X-Filter header string which will be parsed into a QueryBuilder
func (s *DatabaseService) GetUserConfigs(ctx basecontext.ApiContext, userID string, filterString string) (*filters.QueryBuilderResponse[models.UserConfig], *apperrors.Diagnostics) {
	var filter *filters.QueryBuilder
	if filterString != "" {
		filter = filters.NewQueryBuilder(filterString)
	}
	return s.User().GetUserConfigs(ctx, userID, filter)
}

// CreateUserConfig creates a new user config (delegates to User domain)
func (s *DatabaseService) CreateUserConfig(ctx basecontext.ApiContext, userID string, config *models.UserConfig) (*models.UserConfig, *apperrors.Diagnostics) {
	return s.User().CreateUserConfig(ctx, userID, config)
}

// UpsertUserConfig creates or updates a user config (delegates to User domain)
// This implements upsert logic: try update first, create if not found
func (s *DatabaseService) UpsertUserConfig(ctx basecontext.ApiContext, config models.UserConfig) (*models.UserConfig, *apperrors.Diagnostics) {
	// Check if config exists
	existing, getDiag := s.User().GetUserConfig(ctx, config.UserID, config.Slug)

	if getDiag != nil && getDiag.HasErrors() {
		// Not found - create new
		return s.User().CreateUserConfig(ctx, config.UserID, &config)
	}

	// Found - update existing
	updateDiag := s.User().UpdateUserConfig(ctx, config.UserID, existing.ID, &config)
	if updateDiag != nil && updateDiag.HasErrors() {
		return nil, updateDiag
	}

	// Return updated config
	return s.User().GetUserConfig(ctx, config.UserID, config.Slug)
}

// UpdateUserConfig updates an existing user config and returns the updated record
func (s *DatabaseService) UpdateUserConfig(ctx basecontext.ApiContext, config models.UserConfig) (*models.UserConfig, *apperrors.Diagnostics) {
	diag := s.User().UpdateUserConfig(ctx, config.UserID, config.ID, &config)
	if diag != nil && diag.HasErrors() {
		return nil, diag
	}

	// Return updated config
	return s.User().GetUserConfig(ctx, config.UserID, config.ID)
}

// DeleteUserConfig deletes a user config (delegates to User domain)
func (s *DatabaseService) DeleteUserConfig(ctx basecontext.ApiContext, userID, idOrSlug string) *apperrors.Diagnostics {
	return s.User().DeleteUserConfig(ctx, userID, idOrSlug)
}

// ============================================================================
// Database lifecycle methods
// ============================================================================

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
