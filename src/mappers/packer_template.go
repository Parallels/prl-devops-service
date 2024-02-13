package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoPackerTemplateFromApiCreateRequest(m models.CreatePackerTemplateRequest) data_models.PackerTemplate {
	data := data_models.PackerTemplate{
		ID:             helpers.NormalizeString(m.ID),
		Name:           m.Name,
		Description:    m.Description,
		PackerFolder:   m.PackerFolder,
		Variables:      m.Variables,
		Addons:         m.Addons,
		Specs:          m.Specs,
		Defaults:       m.Defaults,
		Internal:       m.Internal,
		RequiredRoles:  m.RequiredRoles,
		RequiredClaims: m.RequiredClaims,
	}

	return data
}

func DtoPackerTemplatesFromApiCreateResponse(m []models.CreatePackerTemplateRequest) []data_models.PackerTemplate {
	data := make([]data_models.PackerTemplate, 0)
	for _, v := range m {
		data = append(data, DtoPackerTemplateFromApiCreateRequest(v))
	}

	return data
}

func DtoPackerTemplateToApResponse(m data_models.PackerTemplate) models.PackerTemplateResponse {
	data := models.PackerTemplateResponse{
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

	return data
}

func DtoPackerTemplatesToApResponse(m []data_models.PackerTemplate) []models.PackerTemplateResponse {
	data := make([]models.PackerTemplateResponse, 0)
	for _, v := range m {
		data = append(data, DtoPackerTemplateToApResponse(v))
	}

	return data
}
