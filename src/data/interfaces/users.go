package interfaces

// UsersService mirrors users method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type UsersService interface {
	GetUsers(ctx basecontext.ApiContext, filter string) ([]models.User, error)
	GetUser(ctx basecontext.ApiContext, idOrEmail string) (*models.User, error)
	CreateUser(ctx basecontext.ApiContext, user models.User) (*models.User, error)
	UpdateUser(ctx basecontext.ApiContext, key models.User) error
	UpdateUserBlockStatus(ctx basecontext.ApiContext, key models.User) error
	UpdateRootPassword(ctx basecontext.ApiContext, newPassword string) error
	DeleteUser(ctx basecontext.ApiContext, id string) error
	AddRoleToUser(ctx basecontext.ApiContext, userId string, roleId string) error
	RemoveRoleFromUser(ctx basecontext.ApiContext, userId string, roleId string) error
	AddClaimToUser(ctx basecontext.ApiContext, userId string, claimId string) error
	RemoveClaimFromUser(ctx basecontext.ApiContext, userId string, claimId string) error
}
