package data

import (
	"Parallels/pd-api-service/data/models"
	"errors"
	"fmt"
	"strings"
)

func (j *JsonDatabase) GetUsers() ([]models.User, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	return j.data.Users, nil
}

func (j *JsonDatabase) GetUser(idOrEmail string) (*models.User, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	for _, user := range j.data.Users {
		if strings.EqualFold(user.ID, idOrEmail) || strings.EqualFold(user.Email, idOrEmail) || strings.EqualFold(user.Username, idOrEmail) {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("User not found")
}

func (j *JsonDatabase) CreateUser(user *models.User) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if u, _ := j.GetUser(user.ID); u != nil {
		return fmt.Errorf("User already exists")
	}

	if u, _ := j.GetUser(user.Email); u != nil {
		return fmt.Errorf("User already exists")
	}

	if u, _ := j.GetUser(user.Username); u != nil {
		return fmt.Errorf("User already exists")
	}

	j.data.Users = append(j.data.Users, *user)
	j.save()
	return nil
}
