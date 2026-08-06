package mappers

import (
	"github.com/Parallels/prl-devops-service/database/filters"
	db_models "github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

// GormUserConfigRequestToDto converts API request to GORM database model
func GormUserConfigRequestToDto(userID string, model models.UserConfigRequest) db_models.UserConfig {
	return db_models.UserConfig{
		BaseModel: db_models.BaseModel{
			ID: helpers.GenerateId(),
		},
		UserID: userID,
		Slug:   model.Slug,
		Name:   model.Name,
		Type:   db_models.UserConfigValueType(model.Type),
		Value:  model.Value,
	}
}

// GormUserConfigDtoToResponse converts GORM database model to API response
func GormUserConfigDtoToResponse(m db_models.UserConfig) models.UserConfigResponse {
	return models.UserConfigResponse{
		ID:        m.ID,
		UserID:    m.UserID,
		Slug:      m.Slug,
		Name:      m.Name,
		Type:      models.UserConfigValueType(m.Type),
		Value:     m.Value,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// GormUserConfigsDtoToResponse converts GORM database model slice to API response slice
func GormUserConfigsDtoToResponse(m []db_models.UserConfig) []models.UserConfigResponse {
	mapped := make([]models.UserConfigResponse, 0, len(m))
	for _, v := range m {
		mapped = append(mapped, GormUserConfigDtoToResponse(v))
	}
	return mapped
}

// GormUserConfigsQueryResponseToResponse converts QueryBuilderResponse to API response slice
func GormUserConfigsQueryResponseToResponse(qbr *filters.QueryBuilderResponse[db_models.UserConfig]) []models.UserConfigResponse {
	if qbr == nil || qbr.Items == nil {
		return []models.UserConfigResponse{}
	}
	return GormUserConfigsDtoToResponse(qbr.Items)
}
