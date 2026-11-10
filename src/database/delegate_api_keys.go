package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/api_keys delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetApiKeys(ctx basecontext.ApiContext, filter string) ([]models.ApiKey, error) {
	return g.legacy.GetApiKeys(ctx, filter)
}

func (g *GormDatabase) GetAllApiKeysIncludingInternal(ctx basecontext.ApiContext) ([]models.ApiKey, error) {
	return g.legacy.GetAllApiKeysIncludingInternal(ctx)
}

func (g *GormDatabase) GetApiKey(ctx basecontext.ApiContext, idOrName string) (*models.ApiKey, error) {
	return g.legacy.GetApiKey(ctx, idOrName)
}

func (g *GormDatabase) GetUserByApiKey(ctx basecontext.ApiContext, apiKey string) (*models.User, error) {
	return g.legacy.GetUserByApiKey(ctx, apiKey)
}

func (g *GormDatabase) CreateApiKey(ctx basecontext.ApiContext, apiKey models.ApiKey) (*models.ApiKey, error) {
	return g.legacy.CreateApiKey(ctx, apiKey)
}

func (g *GormDatabase) DeleteApiKey(ctx basecontext.ApiContext, id string) error {
	return g.legacy.DeleteApiKey(ctx, id)
}

func (g *GormDatabase) UpdateKey(ctx basecontext.ApiContext, key models.ApiKey) error {
	return g.legacy.UpdateKey(ctx, key)
}

func (g *GormDatabase) RevokeKey(ctx basecontext.ApiContext, id string) error {
	return g.legacy.RevokeKey(ctx, id)
}
