package interfaces

// ClaimsService mirrors claims method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type ClaimsService interface {
	GetClaims(ctx basecontext.ApiContext, filter string) ([]models.Claim, error)
	GetClaim(ctx basecontext.ApiContext, idOrName string) (*models.Claim, error)
	CreateClaim(ctx basecontext.ApiContext, claim models.Claim) (*models.Claim, error)
	DeleteClaim(ctx basecontext.ApiContext, idOrName string) error
	UpdateClaimMetadata(ctx basecontext.ApiContext, idOrName, description, group, resource, action string) error
	UpdateClaim(ctx basecontext.ApiContext, claim *models.Claim) (*models.Claim, error)
}
