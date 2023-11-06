package data

import (
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
)

var (
	ErrRoleEmptyNameOrId  = errors.NewWithCode("no role specified", 500)
	ErrRoleEmptyName      = errors.NewWithCode("role name cannot be empty", 500)
	ErrRoleNotFound       = errors.NewWithCode("role not found", 404)
	ErrRemoveInternalRole = errors.NewWithCode("role is internal and cannot be removed", 400)
	ErrUpdateInternalRole = errors.NewWithCode("role is internal and cannot be updated", 400)
)

func (j *JsonDatabase) GetRoles(ctx basecontext.ApiContext, filter string) ([]models.Role, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(j.data.Roles, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (j *JsonDatabase) GetRole(ctx basecontext.ApiContext, idOrName string) (*models.Role, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	roles, err := j.GetRoles(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		if strings.EqualFold(role.ID, idOrName) || strings.EqualFold(role.Name, idOrName) {
			return &role, nil
		}
	}

	return nil, ErrRoleNotFound
}

func (j *JsonDatabase) CreateRole(ctx basecontext.ApiContext, role models.Role) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if role.Name == "" {
		return ErrRoleEmptyName
	}

	role.Name = strings.ToUpper(helpers.NormalizeString(role.Name))
	role.ID = role.Name

	if u, _ := j.GetUser(ctx, role.ID); u != nil {
		return errors.NewWithCodef(400, "role %s already exists with ID %s", role.Name, role.ID)
	}

	j.data.Roles = append(j.data.Roles, role)
	j.Save(ctx)
	return nil
}

func (j *JsonDatabase) DeleteRole(ctx basecontext.ApiContext, idOrName string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if idOrName == "" {
		return ErrRoleEmptyNameOrId
	}

	for i, role := range j.data.Roles {
		if strings.EqualFold(role.ID, idOrName) || strings.EqualFold(role.Name, idOrName) {
			if role.Internal {
				return ErrRemoveInternalRole
			}
			j.data.Roles = append(j.data.Roles[:i], j.data.Roles[i+1:]...)
			j.Save(ctx)
			return nil
		}
	}

	return ErrRoleNotFound
}

func (j *JsonDatabase) UpdateRole(ctx basecontext.ApiContext, role *models.Role) (*models.Role, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if role.ID == "" {
		return nil, ErrRoleEmptyNameOrId
	}

	for i, c := range j.data.Roles {
		if strings.EqualFold(c.ID, role.ID) || strings.EqualFold(c.Name, role.Name) {
			if role.Internal {
				return nil, ErrUpdateInternalRole
			}

			j.data.Roles[i] = *role
			j.Save(ctx)
			return role, nil
		}
	}

	return nil, ErrRoleNotFound
}
