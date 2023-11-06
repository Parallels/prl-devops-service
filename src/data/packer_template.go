package data

import (
	"strings"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/errors"
)

var (
	ErrPackerTemplateNotFound         = errors.NewWithCode("packer template not found", 404)
	ErrPackerTemplateAlreadyExists    = errors.NewWithCode("machine template already exists", 400)
	ErrRemovingInternalPackerTemplate = errors.NewWithCode("cannot remove internal machine template", 400)
	ErrUpdatingInternalPackerTemplate = errors.NewWithCode("cannot update internal machine template", 400)
)

func (j *JsonDatabase) GetPackerTemplates(ctx basecontext.ApiContext, filter string) ([]models.PackerTemplate, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(j.data.PackerTemplates, dbFilter)
	if err != nil {
		return nil, err
	}

	result := GetAuthorizedRecords(ctx, filteredData...)

	return result, nil
}

func (j *JsonDatabase) GetPackerTemplate(ctx basecontext.ApiContext, nameOrId string) (*models.PackerTemplate, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	packerTemplates, err := j.GetPackerTemplates(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, packerTemplate := range packerTemplates {
		if strings.EqualFold(packerTemplate.Name, nameOrId) || strings.EqualFold(packerTemplate.ID, nameOrId) {
			return &packerTemplate, nil
		}
	}

	return nil, ErrPackerTemplateNotFound
}

func (j *JsonDatabase) AddPackerTemplate(ctx basecontext.ApiContext, template *models.PackerTemplate) (*models.PackerTemplate, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if u, _ := j.GetPackerTemplate(ctx, template.ID); u != nil {
		return nil, ErrPackerTemplateAlreadyExists
	}

	template.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	template.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	// Checking the the required claims and roles exist
	for _, claim := range template.RequiredClaims {
		_, err := j.GetClaim(ctx, claim)
		if err != nil {
			return nil, err
		}
	}
	for _, role := range template.RequiredRoles {
		_, err := j.GetRole(ctx, role)
		if err != nil {
			return nil, err
		}
	}

	j.data.PackerTemplates = append(j.data.PackerTemplates, *template)

	j.Save(ctx)
	return template, nil
}

func (j *JsonDatabase) DeletePackerTemplate(ctx basecontext.ApiContext, nameOrId string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if nameOrId == "" {
		return nil
	}

	for i, template := range j.data.PackerTemplates {
		if strings.EqualFold(template.Name, nameOrId) || strings.EqualFold(template.ID, nameOrId) {
			if template.Internal {
				return ErrRemovingInternalPackerTemplate
			}

			j.data.PackerTemplates = append(j.data.PackerTemplates[:i], j.data.PackerTemplates[i+1:]...)
			j.Save(ctx)
			return nil
		}
	}

	return ErrPackerTemplateNotFound
}

func (j *JsonDatabase) UpdatePackerTemplate(ctx basecontext.ApiContext, template *models.PackerTemplate) (*models.PackerTemplate, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	for i, t := range j.data.PackerTemplates {
		if strings.EqualFold(t.ID, template.ID) {
			if t.Internal {
				return nil, ErrUpdatingInternalPackerTemplate
			}
			j.data.PackerTemplates[i] = *template
			j.Save(ctx)
			return template, nil
		}
	}

	return nil, ErrPackerTemplateNotFound
}
