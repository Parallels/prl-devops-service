package mappers

import (
	"strings"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoRoleToApi(model data_models.Role) models.RoleResponse {
	role := models.RoleResponse{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		Internal:    model.Internal,
		Claims:      []models.ClaimResponse{},
		Users:       []models.ApiUser{},
	}

	for _, c := range model.Claims {
		role.Claims = append(role.Claims, models.ClaimResponse{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Internal:    c.Internal,
			Group:       c.Group,
			Resource:    c.Resource,
			Action:      c.Action,
		})
	}

	for _, user := range model.Users {
		role.Users = append(role.Users, models.ApiUser{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
			Email:    user.Email,
		})
	}

	return role
}

func DtoRolesToApi(m []data_models.Role) []models.RoleResponse {
	var roles []models.RoleResponse
	for _, model := range m {
		roles = append(roles, DtoRoleToApi(model))
	}

	return roles
}

func ApiRoleToDto(model models.RoleRequest) data_models.Role {
	role := data_models.Role{
		ID:   strings.ToUpper(helpers.NormalizeString(model.Name)),
		Name: model.Name,
	}

	for _, claimName := range model.Claims {
		normalized := strings.ToUpper(helpers.NormalizeString(claimName))
		role.Claims = append(role.Claims, data_models.Claim{
			ID:   normalized,
			Name: normalized,
		})
	}

	return role
}

func ApiRolesToDto(m []models.RoleRequest) []data_models.Role {
	var roles []data_models.Role
	for _, model := range m {
		roles = append(roles, ApiRoleToDto(model))
	}

	return roles
}
