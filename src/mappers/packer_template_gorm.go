package mappers

import (
	"github.com/Parallels/prl-devops-service/database/filters"
	db_models "github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

// GormPackerTemplateRequestToDto converts API request to GORM database model
func GormPackerTemplateRequestToDto(model models.CreatePackerTemplateRequest) db_models.PackerTemplate {
	return db_models.PackerTemplate{
		ID:             helpers.NormalizeString(model.ID),
		Name:           model.Name,
		Description:    model.Description,
		PackerFolder:   model.PackerFolder,
		Variables:      model.Variables,
		Addons:         model.Addons,
		Specs:          model.Specs,
		Defaults:       model.Defaults,
		Internal:       model.Internal,
		RequiredRoles:  model.RequiredRoles,
		RequiredClaims: model.RequiredClaims,
	}
}

// GormPackerTemplateDtoToResponse converts GORM database model to API response
func GormPackerTemplateDtoToResponse(m db_models.PackerTemplate) models.PackerTemplateResponse {
	return models.PackerTemplateResponse{
		ID:             m.ID,
		Name:           m.Name,
		Description:    m.Description,
		PackerFolder:   m.PackerFolder,
		Variables:      m.Variables,
		Addons:         m.Addons,
		Specs:          m.Specs,
		Defaults:       m.Defaults,
		Internal:       m.Internal,
		UpdatedAt:      m.UpdatedAt,
		CreatedAt:      m.CreatedAt,
		RequiredRoles:  m.RequiredRoles,
		RequiredClaims: m.RequiredClaims,
	}
}

// GormPackerTemplatesDtoToResponse converts GORM database model slice to API response slice
func GormPackerTemplatesDtoToResponse(m []db_models.PackerTemplate) []models.PackerTemplateResponse {
	mapped := make([]models.PackerTemplateResponse, 0, len(m))
	for _, v := range m {
		mapped = append(mapped, GormPackerTemplateDtoToResponse(v))
	}
	return mapped
}

// GormPackerTemplatesQueryResponseToResponse converts QueryBuilderResponse to API response slice
func GormPackerTemplatesQueryResponseToResponse(qbr *filters.QueryBuilderResponse[db_models.PackerTemplate]) []models.PackerTemplateResponse {
	if qbr == nil || qbr.Items == nil {
		return []models.PackerTemplateResponse{}
	}
	return GormPackerTemplatesDtoToResponse(qbr.Items)
}
