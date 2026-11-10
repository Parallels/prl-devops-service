package interfaces

// RolesService mirrors roles method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type RolesService interface {
	GetRoles(ctx basecontext.ApiContext, filter string) ([]models.Role, error)
	GetRole(ctx basecontext.ApiContext, idOrName string) (*models.Role, error)
	CreateRole(ctx basecontext.ApiContext, role models.Role) (*models.Role, error)
	AddClaimToRole(ctx basecontext.ApiContext, roleIdOrName string, claimIdOrName string) error
	RemoveClaimFromRole(ctx basecontext.ApiContext, roleIdOrName string, claimIdOrName string) error
	DeleteRole(ctx basecontext.ApiContext, idOrName string) error
	UpdateRoleInternalFlag(ctx basecontext.ApiContext, idOrName string, internal bool) error
	UpdateRoleDescription(ctx basecontext.ApiContext, idOrName, description string) error
	UpdateRole(ctx basecontext.ApiContext, role *models.Role) (*models.Role, error)
}
