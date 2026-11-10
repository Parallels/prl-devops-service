package interfaces

// UserConfigsService mirrors user_config method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type UserConfigsService interface {
	GetUserConfigs(ctx basecontext.ApiContext, userID string, filter string) ([]models.UserConfig, error)
	GetUserConfig(ctx basecontext.ApiContext, userID string, idOrSlug string) (*models.UserConfig, error)
	CreateUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error)
	UpsertUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error)
	UpdateUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error)
	DeleteUserConfig(ctx basecontext.ApiContext, userID string, idOrSlug string) error
}
