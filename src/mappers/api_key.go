package mappers

import (
	"time"

	database_models "github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func ApiKeyRequestToDbModel(model models.ApiKeyRequest) database_models.ApiKey {
	mapped := database_models.ApiKey{
		BaseModel: database_models.BaseModel{
			ID: helpers.GenerateId(),
		},
		Name:    model.Name,
		Key:     model.Key,
		Secret:  model.Secret,
		Revoked: model.Revoked,
		UserID:  model.UserID,
	}

	if model.RevokedAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, model.RevokedAt); err == nil {
			mapped.RevokedAt = &t
		}
	}
	if model.ExpiresAt != "" {
		if t, err := time.Parse(time.RFC3339Nano, model.ExpiresAt); err == nil {
			mapped.ExpiresAt = &t
		}
	}

	return mapped
}

func ApiKeyDbModelToApiKeyResponse(m database_models.ApiKey) models.ApiKeyResponse {
	mapped := models.ApiKeyResponse{
		ID:      m.ID,
		Name:    m.Name,
		Key:     m.Key,
		Revoked: m.Revoked,
		UserID:  m.UserID,
	}

	if m.RevokedAt != nil {
		mapped.RevokedAt = m.RevokedAt.Format(time.RFC3339Nano)
	}
	if m.ExpiresAt != nil {
		mapped.ExpiresAt = m.ExpiresAt.Format(time.RFC3339Nano)
	}

	return mapped
}

func ApiKeysDbModelToApiKeyResponse(m []database_models.ApiKey) []models.ApiKeyResponse {
	mapped := make([]models.ApiKeyResponse, 0)
	for _, v := range m {
		mapped = append(mapped, ApiKeyDbModelToApiKeyResponse(v))
	}

	return mapped
}
