package data

import (
	"Parallels/pd-api-service/data/models"
	"errors"
	"fmt"
	"strings"
)

func (j *JsonDatabase) GetVirtualMachineTemplates() ([]models.VirtualMachineTemplate, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	return j.data.VirtualMachineTemplates, nil
}

func (j *JsonDatabase) GetVirtualMachineTemplate(nameOrId string) (*models.VirtualMachineTemplate, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	for _, template := range j.data.VirtualMachineTemplates {
		if strings.EqualFold(template.Name, nameOrId) || strings.EqualFold(template.ID, nameOrId) {
			return &template, nil
		}
	}

	return nil, fmt.Errorf("Machine Template not found")
}

func (j *JsonDatabase) AddVirtualMachineTemplate(template *models.VirtualMachineTemplate) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if u, _ := j.GetVirtualMachineTemplate(template.Name); u != nil {
		return fmt.Errorf("Machine Template already exists")
	}

	j.data.VirtualMachineTemplates = append(j.data.VirtualMachineTemplates, *template)
	j.save()
	return nil
}
