package interfaces

// ApiKeysService mirrors api_key method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type ApiKeysService interface {
	GetApiKeys(ctx basecontext.ApiContext, filter string) ([]models.ApiKey, error)
	GetAllApiKeysIncludingInternal(ctx basecontext.ApiContext) ([]models.ApiKey, error)
	GetApiKey(ctx basecontext.ApiContext, idOrName string) (*models.ApiKey, error)
	GetUserByApiKey(ctx basecontext.ApiContext, apiKey string) (*models.User, error)
	CreateApiKey(ctx basecontext.ApiContext, apiKey models.ApiKey) (*models.ApiKey, error)
	DeleteApiKey(ctx basecontext.ApiContext, id string) error
	UpdateKey(ctx basecontext.ApiContext, key models.ApiKey) error
	RevokeKey(ctx basecontext.ApiContext, id string) error
}
