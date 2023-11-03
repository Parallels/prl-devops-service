package mappers

import (
	data_models "Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"strings"
)

func DtoClaimToApi(model data_models.Claim) models.ClaimResponse {
	return models.ClaimResponse{
		ID:   model.ID,
		Name: model.Name,
	}
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
