package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/claims delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetClaims(ctx basecontext.ApiContext, filter string) ([]models.Claim, error) {
	return g.legacy.GetClaims(ctx, filter)
}

func (g *GormDatabase) GetClaim(ctx basecontext.ApiContext, idOrName string) (*models.Claim, error) {
	return g.legacy.GetClaim(ctx, idOrName)
}

func (g *GormDatabase) CreateClaim(ctx basecontext.ApiContext, claim models.Claim) (*models.Claim, error) {
	return g.legacy.CreateClaim(ctx, claim)
}

func (g *GormDatabase) DeleteClaim(ctx basecontext.ApiContext, idOrName string) error {
	return g.legacy.DeleteClaim(ctx, idOrName)
}

func (g *GormDatabase) UpdateClaimMetadata(ctx basecontext.ApiContext, idOrName, description, group, resource, action string) error {
	return g.legacy.UpdateClaimMetadata(ctx, idOrName, description, group, resource, action)
}

func (g *GormDatabase) UpdateClaim(ctx basecontext.ApiContext, claim *models.Claim) (*models.Claim, error) {
	return g.legacy.UpdateClaim(ctx, claim)
}
