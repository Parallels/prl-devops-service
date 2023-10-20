package data

import (
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"errors"
	"fmt"
	"strings"
)

func (j *JsonDatabase) GetRoles() ([]models.UserRole, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	return j.data.Roles, nil
}

func (j *JsonDatabase) GetRole(idOrName string) (*models.UserRole, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	for _, role := range j.data.Roles {
		if strings.EqualFold(role.ID, idOrName) || strings.EqualFold(role.Name, idOrName) {
			return &role, nil
		}
	}

	return nil, fmt.Errorf("Role not found")
}

func (j *JsonDatabase) CreateRole(role *models.UserRole) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if role.ID == "" {
		role.ID = helpers.GenerateId()
	}

	if role.Name == "" {
		return errors.New("role does not contain a name")
	}

	if u, _ := j.GetUser(role.ID); u != nil {
		return fmt.Errorf("Role %s already exists with ID %s", role.Name, role.ID)
	}

	if u, _ := j.GetUser(role.Name); u != nil {
		return fmt.Errorf("Role %s already exists with ID %s", role.Name, role.ID)
	}

	j.data.Roles = append(j.data.Roles, *role)
	j.save()
	return nil
}
