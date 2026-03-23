package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func UserConfigRequestToDto(userID string, model models.UserConfigRequest) data_models.UserConfig {
	return data_models.UserConfig{
		ID:     helpers.GenerateId(),
		UserID: userID,
		Slug:   model.Slug,
		Name:   model.Name,
		Type:   data_models.UserConfigValueType(model.Type),
		Value:  model.Value,
	}
}

func UserConfigDtoToResponse(m data_models.UserConfig) models.UserConfigResponse {
	return models.UserConfigResponse{
		ID:        m.ID,
		UserID:    m.UserID,
		Slug:      m.Slug,
		Name:      m.Name,
		Type:      models.UserConfigValueType(m.Type),
		Value:     m.Value,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func UserConfigsDtoToResponse(m []data_models.UserConfig) []models.UserConfigResponse {
	mapped := make([]models.UserConfigResponse, 0)
	for _, v := range m {
		mapped = append(mapped, UserConfigDtoToResponse(v))
	}
	return mapped
}
