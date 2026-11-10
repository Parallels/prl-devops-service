package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/user_config delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetUserConfigs(ctx basecontext.ApiContext, userID string, filter string) ([]models.UserConfig, error) {
	return g.legacy.GetUserConfigs(ctx, userID, filter)
}

func (g *GormDatabase) GetUserConfig(ctx basecontext.ApiContext, userID string, idOrSlug string) (*models.UserConfig, error) {
	return g.legacy.GetUserConfig(ctx, userID, idOrSlug)
}

func (g *GormDatabase) CreateUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error) {
	return g.legacy.CreateUserConfig(ctx, cfg)
}

func (g *GormDatabase) UpsertUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error) {
	return g.legacy.UpsertUserConfig(ctx, cfg)
}

func (g *GormDatabase) UpdateUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error) {
	return g.legacy.UpdateUserConfig(ctx, cfg)
}

func (g *GormDatabase) DeleteUserConfig(ctx basecontext.ApiContext, userID string, idOrSlug string) error {
	return g.legacy.DeleteUserConfig(ctx, userID, idOrSlug)
}
