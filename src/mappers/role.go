package mappers

import (
	"strings"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoRoleToApi(model data_models.Role) models.RoleResponse {
	return models.RoleResponse{
		ID:   model.ID,
		Name: model.Name,
	}
}

func DtoRolesToApi(m []data_models.Role) []models.RoleResponse {
	var roles []models.RoleResponse
	for _, model := range m {
		roles = append(roles, DtoRoleToApi(model))
	}

	return roles
}

func ApiRoleToDto(model models.RoleRequest) data_models.Role {
	return data_models.Role{
		ID:   strings.ToUpper(helpers.NormalizeString(model.Name)),
		Name: model.Name,
	}
}

func ApiRolesToDto(m []models.RoleRequest) []data_models.Role {
	var roles []data_models.Role
	for _, model := range m {
		roles = append(roles, ApiRoleToDto(model))
	}

	return roles
}
