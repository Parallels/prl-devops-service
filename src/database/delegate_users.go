package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/users delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetUsers(ctx basecontext.ApiContext, filter string) ([]models.User, error) {
	return g.legacy.GetUsers(ctx, filter)
}

func (g *GormDatabase) GetUser(ctx basecontext.ApiContext, idOrEmail string) (*models.User, error) {
	return g.legacy.GetUser(ctx, idOrEmail)
}

func (g *GormDatabase) CreateUser(ctx basecontext.ApiContext, user models.User) (*models.User, error) {
	return g.legacy.CreateUser(ctx, user)
}

func (g *GormDatabase) UpdateUser(ctx basecontext.ApiContext, key models.User) error {
	return g.legacy.UpdateUser(ctx, key)
}

func (g *GormDatabase) UpdateUserBlockStatus(ctx basecontext.ApiContext, key models.User) error {
	return g.legacy.UpdateUserBlockStatus(ctx, key)
}

func (g *GormDatabase) UpdateRootPassword(ctx basecontext.ApiContext, newPassword string) error {
	return g.legacy.UpdateRootPassword(ctx, newPassword)
}

func (g *GormDatabase) DeleteUser(ctx basecontext.ApiContext, id string) error {
	return g.legacy.DeleteUser(ctx, id)
}

func (g *GormDatabase) AddRoleToUser(ctx basecontext.ApiContext, userId string, roleId string) error {
	return g.legacy.AddRoleToUser(ctx, userId, roleId)
}

func (g *GormDatabase) RemoveRoleFromUser(ctx basecontext.ApiContext, userId string, roleId string) error {
	return g.legacy.RemoveRoleFromUser(ctx, userId, roleId)
}

func (g *GormDatabase) AddClaimToUser(ctx basecontext.ApiContext, userId string, claimId string) error {
	return g.legacy.AddClaimToUser(ctx, userId, claimId)
}

func (g *GormDatabase) RemoveClaimFromUser(ctx basecontext.ApiContext, userId string, claimId string) error {
	return g.legacy.RemoveClaimFromUser(ctx, userId, claimId)
}
