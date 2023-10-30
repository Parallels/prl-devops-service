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

func ClaimFromDTO(model data_models.Claim) models.UserClaim {
	return models.UserClaim{
		ID:   model.ID,
		Name: model.Name,
	}
}

func ClaimToDTO(model models.UserClaim) data_models.Claim {
	return data_models.Claim{
		ID:   model.ID,
		Name: model.Name,
	}
}

func RoleFromDTO(model data_models.Role) models.UserRole {
	return models.UserRole{
		ID:   model.ID,
		Name: model.Name,
	}
}

func RoleToDTO(model models.UserRole) data_models.Role {
	return data_models.Role{
		ID:   model.ID,
		Name: model.Name,
	}
}
