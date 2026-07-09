package dbservice

import (
	"context"
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/service"
	"github.com/Parallels/prl-devops-service/database/stores"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	"gorm.io/gorm"
)

// DatabaseService wraps GORM stores and provides a unified interface
type DatabaseService struct {
	db         *gorm.DB
	userStore  stores.UserDataStoreInterface
	roleStore  stores.RoleDataStoreInterface
	claimStore stores.ClaimDataStoreInterface
}

var globalDatabaseService *DatabaseService

// InitDatabase initializes the GORM database and stores using centralized database/service
func InitDatabase(ctx basecontext.ApiContext) (*DatabaseService, error) {
	if globalDatabaseService != nil {
		return globalDatabaseService, nil
	}

	cfg := config.Get()

	// Build database configuration from app config
	dbConfig := buildDatabaseConfig(cfg)

	// Initialize centralized database service
	dbSvc, err := service.Initialize(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database service: %w", err)
	}

	// Verify database health
	if err := dbSvc.Health(context.Background()); err != nil {
		ctx.LogErrorf("Database health check failed: %v", err)
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	// Get the database connection from centralized service
	db := dbSvc.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// Apply custom logger that bridges GORM logs to basecontext
	logLevel := service.ConvertLogLevel(cfg.IsDebugEnabled())
	customLogger := service.NewBaseContextLogger(ctx, logLevel)
	db.Logger = customLogger

	// Auto-migrate
	if err := db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Claim{},
		&models.ApiKey{},
		&models.Configuration{},
		&models.Activity{},
	); err != nil {
		return nil, err
	}

	// Initialize stores
	stdCtx := context.Background()

	userStore := stores.GetUserDataStoreInstance()
	if err := userStore.Init(stdCtx, db); err != nil {
		return nil, err
	}

	roleStore := stores.GetRoleDataStoreInstance()
	if err := roleStore.Init(stdCtx, db); err != nil {
		return nil, err
	}

	claimStore := stores.GetClaimDataStoreInstance()
	if err := claimStore.Init(stdCtx, db); err != nil {
		return nil, err
	}

	globalDatabaseService = &DatabaseService{
		db:         db,
		userStore:  userStore,
		roleStore:  roleStore,
		claimStore: claimStore,
	}

	ctx.LogInfof("Database initialized successfully (type: %s)", dbConfig.Type)
	return globalDatabaseService, nil
}

// buildDatabaseConfig converts app config to database common.Config
func buildDatabaseConfig(cfg *config.Config) common.Config {
	// Read database type from config (sqlite or postgresql)
	dbType := cfg.DatabaseType()

	dbConfig := common.Config{
		Type:    common.DatabaseType(dbType),
		Debug:   cfg.IsDebugEnabled(), // Use app's log level
		Migrate: cfg.IsDatabaseAutoMigrateEnabled(),
		Pool:    common.DefaultPoolConfig(),
	}

	// Configure based on database type
	switch dbConfig.Type {
	case common.SQLite:
		dbPath := "data"
		if cfg.DatabaseFolder() != "" {
			dbPath = cfg.DatabaseFolder()
		}
		dbConfig.SQLite = common.SQLiteConfig{
			StoragePath: dbPath,
			FileName:    "database.db",
		}

	case common.PostgreSQL:
		// Configure PostgreSQL from environment variables
		dbConfig.PostgreSQL = common.PostgreSQLConfig{
			Host:     cfg.GetKey("DATABASE_HOST"),
			Port:     cfg.GetIntKey("DATABASE_PORT"),
			Database: cfg.GetKey("DATABASE_NAME"),
			Username: cfg.GetKey("DATABASE_USERNAME"),
			Password: cfg.GetKey("DATABASE_PASSWORD"),
			SSLMode:  cfg.GetBoolKey("DATABASE_SSL_MODE"),
		}
		// Apply defaults if not set
		if dbConfig.PostgreSQL.Host == "" {
			dbConfig.PostgreSQL.Host = "localhost"
		}
		if dbConfig.PostgreSQL.Port == 0 {
			dbConfig.PostgreSQL.Port = 5432
		}
		if dbConfig.PostgreSQL.Database == "" {
			dbConfig.PostgreSQL.Database = "prl_devops"
		}
		if dbConfig.PostgreSQL.Username == "" {
			dbConfig.PostgreSQL.Username = "postgres"
		}
	}

	return dbConfig
}

// GetDatabaseService returns the global database service instance
func GetDatabaseService() *DatabaseService {
	return globalDatabaseService
}

// User operations
func (s *DatabaseService) GetUsers(ctx basecontext.ApiContext, filter string) ([]models.User, *apperrors.Diagnostics) {
	// Cast to basecontext.BaseContext since stores expect that type
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.GetUsers(*baseCtx)
}

func (s *DatabaseService) GetUser(ctx basecontext.ApiContext, idOrEmail string) (*models.User, *apperrors.Diagnostics) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	// Try by ID first
	user, diag := s.userStore.GetUserByID(*baseCtx, idOrEmail)
	if diag != nil && !diag.HasErrors() && user != nil {
		return user, nil
	}

	// Try by username
	user, diag = s.userStore.GetUserByUsername(*baseCtx, idOrEmail)
	if diag != nil && diag.HasErrors() {
		return nil, diag
	}
	if user == nil {
		diag = apperrors.NewDiagnostics("get_user")
		diag.AddError("user_not_found", fmt.Sprintf("user not found: %s", idOrEmail), "dbservice", nil)
		return nil, diag
	}
	return user, nil
}

func (s *DatabaseService) CreateUser(ctx basecontext.ApiContext, user models.User) (*models.User, *apperrors.Diagnostics) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.CreateUser(*baseCtx, &user)
}

func (s *DatabaseService) UpdateUser(ctx basecontext.ApiContext, user *models.User) *apperrors.Diagnostics {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.UpdateUser(*baseCtx, user)
}

func (s *DatabaseService) DeleteUser(ctx basecontext.ApiContext, id string) *apperrors.Diagnostics {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.DeleteUser(*baseCtx, id)
}

func (s *DatabaseService) UpdateUserPassword(ctx basecontext.ApiContext, id string, password string) *apperrors.Diagnostics {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.UpdateUserPassword(*baseCtx, id, password)
}

// Role operations
func (s *DatabaseService) GetUserRoles(ctx basecontext.ApiContext, userId string) ([]models.Role, *apperrors.Diagnostics) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.GetUserRoles(*baseCtx, userId)
}

func (s *DatabaseService) AddUserToRole(ctx basecontext.ApiContext, userId string, roleId string) *apperrors.Diagnostics {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.AddUserToRole(*baseCtx, userId, roleId)
}

func (s *DatabaseService) RemoveRoleFromUser(ctx basecontext.ApiContext, userId string, roleId string) *apperrors.Diagnostics {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.RemoveUserFromRole(*baseCtx, userId, roleId)
}

// Claim operations
func (s *DatabaseService) GetUserClaims(ctx basecontext.ApiContext, userId string) ([]models.Claim, *apperrors.Diagnostics) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.GetUserClaims(*baseCtx, userId)
}

func (s *DatabaseService) AddClaimToUser(ctx basecontext.ApiContext, userId string, claimId string) *apperrors.Diagnostics {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.AddClaimToUser(*baseCtx, userId, claimId)
}

func (s *DatabaseService) RemoveClaimFromUser(ctx basecontext.ApiContext, userId string, claimId string) *apperrors.Diagnostics {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.userStore.RemoveClaimFromUser(*baseCtx, userId, claimId)
}

// Role CRUD
func (s *DatabaseService) GetRoles(ctx basecontext.ApiContext, filter string) ([]models.Role, *apperrors.Diagnostics) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.roleStore.GetRoles(*baseCtx)
}

func (s *DatabaseService) GetRole(ctx basecontext.ApiContext, idOrName string) (*models.Role, *apperrors.Diagnostics) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.roleStore.GetRoleBySlugOrID(*baseCtx, idOrName)
}

// Claim CRUD
func (s *DatabaseService) GetClaims(ctx basecontext.ApiContext, filter string) ([]models.Claim, *apperrors.Diagnostics) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.claimStore.GetClaims(*baseCtx)
}

func (s *DatabaseService) GetClaim(ctx basecontext.ApiContext, idOrName string) (*models.Claim, *apperrors.Diagnostics) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	return s.claimStore.GetClaimByNameOrID(*baseCtx, idOrName)
}

// Connect is a no-op for compatibility with JsonDatabase interface
func (s *DatabaseService) Connect(ctx basecontext.ApiContext) error {
	return nil
}

// IsConnected returns true if database is initialized
func (s *DatabaseService) IsConnected() bool {
	return s != nil && s.db != nil
}
