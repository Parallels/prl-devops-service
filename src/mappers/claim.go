package mappers

import (
	"strings"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoClaimToApi(model data_models.Claim) models.ClaimResponse {
	claim := models.ClaimResponse{
		ID:   model.ID,
		Name: model.Name,
	}

	for _, user := range model.Users {
		claim.Users = append(claim.Users, models.ApiUser{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
			Email:    user.Email,
		})
	}

	return claim
}

func DtoClaimsToApi(m []data_models.Claim) []models.ClaimResponse {
	var claims []models.ClaimResponse
	for _, model := range m {
		claims = append(claims, DtoClaimToApi(model))
	}

	return claims
}

func ApiClaimToDto(model models.ClaimRequest) data_models.Claim {
	return data_models.Claim{
		ID:   strings.ToUpper(helpers.NormalizeString(model.Name)),
		Name: model.Name,
	}
}

func ApiClaimsToDto(m []models.ClaimRequest) []data_models.Claim {
	var claims []data_models.Claim
	for _, model := range m {
		claims = append(claims, ApiClaimToDto(model))
	}

	return claims
}
