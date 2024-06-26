package mappers

import (
	"strings"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func ApiUserCreateRequestToDto(model models.UserCreateRequest) data_models.User {
	user := data_models.User{
		ID:       helpers.GenerateId(),
		Username: model.Username,
		Password: model.Password,
		Name:     model.Name,
		Email:    model.Email,
	}

	for _, role := range model.Roles {
		user.Roles = append(user.Roles, data_models.Role{ID: strings.ToUpper(helpers.NormalizeString(role)), Name: strings.ToUpper(helpers.NormalizeString(role))})
	}
	for _, claim := range model.Claims {
		user.Claims = append(user.Claims, data_models.Claim{ID: strings.ToUpper(helpers.NormalizeString(claim)), Name: strings.ToUpper(helpers.NormalizeString(claim))})
	}

	return user
}

func ApiUserUpdateRequestToDto(model models.UserUpdateRequest) data_models.User {
	user := data_models.User{
		ID:       helpers.GenerateId(),
		Password: model.Password,
		Name:     model.Name,
		Email:    model.Email,
	}

	return user
}

func DtoUserToApiResponse(model data_models.User) models.ApiUser {
	user := models.ApiUser{
		ID:       model.ID,
		Username: model.Username,
		Name:     model.Name,
		Email:    model.Email,
	}
	for _, role := range model.Roles {
		user.Roles = append(user.Roles, role.ID)
	}
	for _, claim := range model.Claims {
		user.Claims = append(user.Claims, claim.ID)
	}
	if user.Claims == nil {
		user.Claims = []string{}
	}
	if user.Roles == nil {
		user.Roles = []string{}
	}

	return user
}

func DtoUsersToApiResponse(model []data_models.User) []models.ApiUser {
	var users []models.ApiUser
	for _, user := range model {
		users = append(users, DtoUserToApiResponse(user))
	}
	return users
}
