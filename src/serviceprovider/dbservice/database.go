package dbservice

import (
	"context"
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/stores"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Helper to convert Diagnostics to error
func diagToError(diag *apperrors.Diagnostics) error {
	if !diag.HasErrors() {
		return nil
	}
	errors := diag.GetErrors()
	if len(errors) > 0 {
		return fmt.Errorf("%s: %s", errors[0].Code, errors[0].Message)
	}
	return fmt.Errorf("unknown error")
}

// DatabaseService wraps GORM stores and provides a unified interface
type DatabaseService struct {
	db         *gorm.DB
	userStore  stores.UserDataStoreInterface
	roleStore  stores.RoleDataStoreInterface
	claimStore stores.ClaimDataStoreInterface
}

var globalDatabaseService *DatabaseService

// InitDatabase initializes the GORM database and stores
func InitDatabase(ctx basecontext.ApiContext) (*DatabaseService, error) {
	if globalDatabaseService != nil {
		return globalDatabaseService, nil
	}

	cfg := config.Get()
	dbPath := "data/database.db"
	if cfg.DatabaseFolder() != "" {
		dbPath = cfg.DatabaseFolder() + "/database.db"
	}

	// Open database
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

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

	ctx.LogInfof("Database initialized at %s", dbPath)
	return globalDatabaseService, nil
}

// GetDatabaseService returns the global database service instance
func GetDatabaseService() *DatabaseService {
	return globalDatabaseService
}

// User operations
func (s *DatabaseService) GetUsers(ctx basecontext.ApiContext, filter string) ([]models.User, error) {
	// Cast to basecontext.BaseContext since stores expect that type
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	users, diag := s.userStore.GetUsers(*baseCtx)
	if diag.HasErrors() {
		return nil, diagToError(diag)
	}
	return users, nil
}

func (s *DatabaseService) GetUser(ctx basecontext.ApiContext, idOrEmail string) (*models.User, error) {
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
		return nil, diagToError(diag)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found: %s", idOrEmail)
	}
	return user, nil
}

func (s *DatabaseService) CreateUser(ctx basecontext.ApiContext, user models.User) (*models.User, error) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	created, diag := s.userStore.CreateUser(*baseCtx, &user)
	if diag.HasErrors() {
		return nil, diagToError(diag)
	}
	return created, nil
}

func (s *DatabaseService) UpdateUser(ctx basecontext.ApiContext, user *models.User) error {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	diag := s.userStore.UpdateUser(*baseCtx, user)
	if diag.HasErrors() {
		return diagToError(diag)
	}
	return nil
}

func (s *DatabaseService) DeleteUser(ctx basecontext.ApiContext, id string) error {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	diag := s.userStore.DeleteUser(*baseCtx, id)
	if diag.HasErrors() {
		return diagToError(diag)
	}
	return nil
}

func (s *DatabaseService) UpdateUserPassword(ctx basecontext.ApiContext, id string, password string) error {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	diag := s.userStore.UpdateUserPassword(*baseCtx, id, password)
	if diag.HasErrors() {
		return diagToError(diag)
	}
	return nil
}

// Role operations
func (s *DatabaseService) GetUserRoles(ctx basecontext.ApiContext, userId string) ([]models.Role, error) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	roles, diag := s.userStore.GetUserRoles(*baseCtx, userId)
	if diag.HasErrors() {
		return nil, diagToError(diag)
	}
	return roles, nil
}

func (s *DatabaseService) AddUserToRole(ctx basecontext.ApiContext, userId string, roleId string) error {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	diag := s.userStore.AddUserToRole(*baseCtx, userId, roleId)
	if diag.HasErrors() {
		return diagToError(diag)
	}
	return nil
}

func (s *DatabaseService) RemoveRoleFromUser(ctx basecontext.ApiContext, userId string, roleId string) error {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	diag := s.userStore.RemoveUserFromRole(*baseCtx, userId, roleId)
	if diag.HasErrors() {
		return diagToError(diag)
	}
	return nil
}

// Claim operations
func (s *DatabaseService) GetUserClaims(ctx basecontext.ApiContext, userId string) ([]models.Claim, error) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	claims, diag := s.userStore.GetUserClaims(*baseCtx, userId)
	if diag.HasErrors() {
		return nil, diagToError(diag)
	}
	return claims, nil
}

func (s *DatabaseService) AddClaimToUser(ctx basecontext.ApiContext, userId string, claimId string) error {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	diag := s.userStore.AddClaimToUser(*baseCtx, userId, claimId)
	if diag.HasErrors() {
		return diagToError(diag)
	}
	return nil
}

func (s *DatabaseService) RemoveClaimFromUser(ctx basecontext.ApiContext, userId string, claimId string) error {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	diag := s.userStore.RemoveClaimFromUser(*baseCtx, userId, claimId)
	if diag.HasErrors() {
		return diagToError(diag)
	}
	return nil
}

// Role CRUD
func (s *DatabaseService) GetRoles(ctx basecontext.ApiContext, filter string) ([]models.Role, error) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	roles, diag := s.roleStore.GetRoles(*baseCtx)
	if diag.HasErrors() {
		return nil, diagToError(diag)
	}
	return roles, nil
}

func (s *DatabaseService) GetRole(ctx basecontext.ApiContext, idOrName string) (*models.Role, error) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	role, diag := s.roleStore.GetRoleBySlugOrID(*baseCtx, idOrName)
	if diag.HasErrors() {
		return nil, diagToError(diag)
	}
	return role, nil
}

// Claim CRUD
func (s *DatabaseService) GetClaims(ctx basecontext.ApiContext, filter string) ([]models.Claim, error) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	claims, diag := s.claimStore.GetClaims(*baseCtx)
	if diag.HasErrors() {
		return nil, diagToError(diag)
	}
	return claims, nil
}

func (s *DatabaseService) GetClaim(ctx basecontext.ApiContext, idOrName string) (*models.Claim, error) {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}

	claim, diag := s.claimStore.GetClaimByNameOrID(*baseCtx, idOrName)
	if diag.HasErrors() {
		return nil, diagToError(diag)
	}
	return claim, nil
}

// Connect is a no-op for compatibility with JsonDatabase interface
func (s *DatabaseService) Connect(ctx basecontext.ApiContext) error {
	return nil
}

// IsConnected returns true if database is initialized
func (s *DatabaseService) IsConnected() bool {
	return s != nil && s.db != nil
}
