package mappers

import (
	data_models "Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/models"
)

func UserFromDTO(model data_models.User) models.User {
	user := models.User{
		ID:       model.ID,
		Username: model.Username,
		Name:     model.Name,
		Email:    model.Email,
	}
	for _, role := range model.Roles {
		user.Roles = append(user.Roles, RoleFromDTO(role))
	}
	for _, claim := range model.Claims {
		user.Claims = append(user.Claims, ClaimFromDTO(claim))
	}

	return user
}

func UserToDTO(model models.User) data_models.User {
	user := data_models.User{
		ID:       model.ID,
		Username: model.Username,
		Name:     model.Name,
		Email:    model.Email,
	}
	for _, role := range model.Roles {
		user.Roles = append(user.Roles, RoleToDTO(role))
	}
	for _, claim := range model.Claims {
		user.Claims = append(user.Claims, ClaimToDTO(claim))
	}

	return user
}

func ClaimFromDTO(model data_models.UserClaim) models.UserClaim {
	return models.UserClaim{
		ID:   model.ID,
		Name: model.Name,
	}
}

func ClaimToDTO(model models.UserClaim) data_models.UserClaim {
	return data_models.UserClaim{
		ID:   model.ID,
		Name: model.Name,
	}
}

func RoleFromDTO(model data_models.UserRole) models.UserRole {
	return models.UserRole{
		ID:   model.ID,
		Name: model.Name,
	}
}

func RoleToDTO(model models.UserRole) data_models.UserRole {
	return data_models.UserRole{
		ID:   model.ID,
		Name: model.Name,
	}
}
