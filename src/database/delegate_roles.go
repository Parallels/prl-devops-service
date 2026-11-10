package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/roles delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetRoles(ctx basecontext.ApiContext, filter string) ([]models.Role, error) {
	return g.legacy.GetRoles(ctx, filter)
}

func (g *GormDatabase) GetRole(ctx basecontext.ApiContext, idOrName string) (*models.Role, error) {
	return g.legacy.GetRole(ctx, idOrName)
}

func (g *GormDatabase) CreateRole(ctx basecontext.ApiContext, role models.Role) (*models.Role, error) {
	return g.legacy.CreateRole(ctx, role)
}

func (g *GormDatabase) AddClaimToRole(ctx basecontext.ApiContext, roleIdOrName string, claimIdOrName string) error {
	return g.legacy.AddClaimToRole(ctx, roleIdOrName, claimIdOrName)
}

func (g *GormDatabase) RemoveClaimFromRole(ctx basecontext.ApiContext, roleIdOrName string, claimIdOrName string) error {
	return g.legacy.RemoveClaimFromRole(ctx, roleIdOrName, claimIdOrName)
}

func (g *GormDatabase) DeleteRole(ctx basecontext.ApiContext, idOrName string) error {
	return g.legacy.DeleteRole(ctx, idOrName)
}

func (g *GormDatabase) UpdateRoleInternalFlag(ctx basecontext.ApiContext, idOrName string, internal bool) error {
	return g.legacy.UpdateRoleInternalFlag(ctx, idOrName, internal)
}

func (g *GormDatabase) UpdateRoleDescription(ctx basecontext.ApiContext, idOrName, description string) error {
	return g.legacy.UpdateRoleDescription(ctx, idOrName, description)
}

func (g *GormDatabase) UpdateRole(ctx basecontext.ApiContext, role *models.Role) (*models.Role, error) {
	return g.legacy.UpdateRole(ctx, role)
}
