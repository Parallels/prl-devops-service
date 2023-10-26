package data

import (
	"Parallels/pd-api-service/data/models"
	"errors"
	"strings"
)

var (
	ErrMachineTemplateNotFound      = errors.New("machine Template not found")
	ErrMachineTemplateAlreadyExists = errors.New("machine Template already exists")
)

func (j *JsonDatabase) GetVirtualMachineTemplates() ([]models.VirtualMachineTemplate, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	return j.data.VirtualMachineTemplates, nil
}

func (j *JsonDatabase) GetVirtualMachineTemplate(nameOrId string) (*models.VirtualMachineTemplate, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	for _, template := range j.data.VirtualMachineTemplates {
		if strings.EqualFold(template.Name, nameOrId) || strings.EqualFold(template.ID, nameOrId) {
			return &template, nil
		}
	}

	return nil, ErrMachineTemplateNotFound
}

func (j *JsonDatabase) AddVirtualMachineTemplate(template *models.VirtualMachineTemplate) (*models.VirtualMachineTemplate, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if u, _ := j.GetVirtualMachineTemplate(template.ID); u != nil {
		return nil, ErrMachineTemplateAlreadyExists
	}

	j.data.VirtualMachineTemplates = append(j.data.VirtualMachineTemplates, *template)
	j.save()
	return template, nil
}
