package interfaces

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/database/models"
	apperrors "github.com/Parallels/prl-devops-service/errors"
)

// UserDomain defines user-related operations with business logic
// Combines user management with user configurations
type UserDomain interface {
	// User profile operations combining user + configs
	GetUserProfile(ctx basecontext.ApiContext, userID string) (*UserProfile, *apperrors.Diagnostics)

	// User config operations (delegates to UserConfigStore but with business logic)
	GetUserConfig(ctx basecontext.ApiContext, userID, idOrSlug string) (*models.UserConfig, *apperrors.Diagnostics)
	GetUserConfigs(ctx basecontext.ApiContext, userID string, filter *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.UserConfig], *apperrors.Diagnostics)
	CreateUserConfig(ctx basecontext.ApiContext, userID string, config *models.UserConfig) (*models.UserConfig, *apperrors.Diagnostics)
	UpdateUserConfig(ctx basecontext.ApiContext, userID, idOrSlug string, config *models.UserConfig) *apperrors.Diagnostics
	DeleteUserConfig(ctx basecontext.ApiContext, userID, idOrSlug string) *apperrors.Diagnostics
}

// UserProfile combines user data with their configurations
type UserProfile struct {
	User    *models.User        `json:"user"`
	Configs []models.UserConfig `json:"configs"`
}
