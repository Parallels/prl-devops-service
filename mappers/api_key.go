package mappers

import (
	data_models "Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/models"
)

func ApiKeyFromDTO(model data_models.ApiKey) models.ApiKey {
	mapped := models.ApiKey{
		ID:        model.ID,
		Name:      model.Name,
		Key:       model.Key,
		Revoked:   model.Revoked,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		RevokedAt: model.RevokedAt,
	}

	return mapped
}

func ApiKeyToDTO(model models.ApiKey) data_models.ApiKey {
	mapped := data_models.ApiKey{
		ID:        model.ID,
		Name:      model.Name,
		Key:       model.Key,
		Revoked:   model.Revoked,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		RevokedAt: model.RevokedAt,
	}

	return mapped
}
