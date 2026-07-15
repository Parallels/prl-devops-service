package stores

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/database/interfaces"
	"github.com/Parallels/prl-devops-service/database/models"
	apperrors "github.com/Parallels/prl-devops-service/errors"
)

// Compile-time interface compliance check
var _ interfaces.UserDomain = (*UserDomainService)(nil)

// UserDomainService handles user-related domain operations with business logic
// For simple CRUD operations, use db.Stores() directly instead of adding pass-through methods here
type UserDomainService struct {
	userStore       UserDataStoreInterface
	userConfigStore UserConfigDataStoreInterface
}

// NewUserDomainService creates a new user domain service
func NewUserDomainService(
	userStore UserDataStoreInterface,
	userConfigStore UserConfigDataStoreInterface,
) *UserDomainService {
	return &UserDomainService{
		userStore:       userStore,
		userConfigStore: userConfigStore,
	}
}

// toBaseContext converts ApiContext to BaseContext
func toBaseContextUser(ctx basecontext.ApiContext) *basecontext.BaseContext {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}
	return baseCtx
}

// GetUserProfile retrieves complete user profile including configurations
// This is a business logic method that combines data from multiple stores
func (s *UserDomainService) GetUserProfile(ctx basecontext.ApiContext, userID string) (*interfaces.UserProfile, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("user_domain_get_profile")
	baseCtx := toBaseContextUser(ctx)

	// Get user
	user, userDiag := s.userStore.GetUserByID(*baseCtx, userID)
	if userDiag != nil && userDiag.HasErrors() {
		diag.Append(userDiag)
		return nil, diag
	}

	// Get user configs
	configsResp, configDiag := s.userConfigStore.Find(*baseCtx, userID, nil)
	if configDiag != nil && configDiag.HasErrors() {
		// Configs are optional - log error but don't fail
		ctx.LogWarnf("Failed to load user configs: %v", configDiag.GetErrors())
	}

	configs := []models.UserConfig{}
	if configsResp != nil {
		configs = configsResp.Items
	}

	return &interfaces.UserProfile{
		User:    user,
		Configs: configs,
	}, nil
}

// GetUserConfig retrieves a single user config with validation
func (s *UserDomainService) GetUserConfig(ctx basecontext.ApiContext, userID, idOrSlug string) (*models.UserConfig, *apperrors.Diagnostics) {
	baseCtx := toBaseContextUser(ctx)
	return s.userConfigStore.Get(*baseCtx, userID, idOrSlug)
}

// GetUserConfigs retrieves all configs for a user with optional filtering
func (s *UserDomainService) GetUserConfigs(ctx basecontext.ApiContext, userID string, filter *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.UserConfig], *apperrors.Diagnostics) {
	baseCtx := toBaseContextUser(ctx)
	return s.userConfigStore.Find(*baseCtx, userID, filter)
}

// CreateUserConfig creates a new user config with validation
func (s *UserDomainService) CreateUserConfig(ctx basecontext.ApiContext, userID string, config *models.UserConfig) (*models.UserConfig, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("user_domain_create_config")
	baseCtx := toBaseContextUser(ctx)

	// Ensure the config belongs to the correct user
	config.UserID = userID

	// Validate that user exists (business logic)
	_, userDiag := s.userStore.GetUserByID(*baseCtx, userID)
	if userDiag != nil && userDiag.HasErrors() {
		diag.Append(userDiag)
		return nil, diag
	}

	// Create the config
	return s.userConfigStore.Create(*baseCtx, config)
}

// UpdateUserConfig updates a user config with validation
func (s *UserDomainService) UpdateUserConfig(ctx basecontext.ApiContext, userID, idOrSlug string, config *models.UserConfig) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("user_domain_update_config")
	baseCtx := toBaseContextUser(ctx)

	// Verify the config exists and belongs to the user
	existing, getDiag := s.userConfigStore.Get(*baseCtx, userID, idOrSlug)
	if getDiag != nil && getDiag.HasErrors() {
		diag.Append(getDiag)
		return diag
	}

	// Update only allowed fields (prevent user_id change)
	existing.Name = config.Name
	existing.Type = config.Type
	existing.Value = config.Value

	return s.userConfigStore.Update(*baseCtx, existing)
}

// DeleteUserConfig deletes a user config with validation
func (s *UserDomainService) DeleteUserConfig(ctx basecontext.ApiContext, userID, idOrSlug string) *apperrors.Diagnostics {
	baseCtx := toBaseContextUser(ctx)
	return s.userConfigStore.Delete(*baseCtx, userID, idOrSlug)
}
