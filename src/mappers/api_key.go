package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func ApiKeyRequestToDto(model models.ApiKeyRequest) data_models.ApiKey {
	mapped := data_models.ApiKey{
		ID:        helpers.GenerateId(),
		Name:      model.Name,
		Key:       model.Key,
		Secret:    model.Secret,
		Revoked:   model.Revoked,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		RevokedAt: model.RevokedAt,
	}

	return mapped
}

func ApiKeyDtoToApiKeyResponse(m data_models.ApiKey) models.ApiKeyResponse {
	mapped := models.ApiKeyResponse{
		ID:      m.ID,
		Name:    m.Name,
		Key:     m.Key,
		Revoked: m.Revoked,
	}

	return mapped
}

func ApiKeysDtoToApiKeyResponse(m []data_models.ApiKey) []models.ApiKeyResponse {
	mapped := make([]models.ApiKeyResponse, 0)
	for _, v := range m {
		mapped = append(mapped, ApiKeyDtoToApiKeyResponse(v))
	}

	return mapped
}
