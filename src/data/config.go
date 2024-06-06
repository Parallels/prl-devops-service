package data

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/security"
	"github.com/google/uuid"
)

func (j *JsonDatabase) GetConfiguration(ctx basecontext.ApiContext) (*models.Configuration, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if j.data.Configuration == nil {
		j.data.Configuration = &models.Configuration{
			ID: j.generateId(),
		}
	}

	return j.data.Configuration, nil
}

func (j *JsonDatabase) GetId(ctx basecontext.ApiContext) (string, error) {
	if !j.IsConnected() {
		return "", ErrDatabaseNotConnected
	}

	if j.data.Configuration == nil {
		j.data.Configuration = &models.Configuration{
			ID: j.generateId(),
		}
	}

	return j.data.Configuration.ID, nil
}

func (j *JsonDatabase) SeedId(ctx basecontext.ApiContext) (string, error) {
	if !j.IsConnected() {
		return "", ErrDatabaseNotConnected
	}

	if j.data.Configuration == nil {
		j.data.Configuration = &models.Configuration{
			ID: j.generateId(),
		}
	}
	if j.data.Configuration.ID == "" {
		j.data.Configuration.ID = j.generateId()
	}

	return j.data.Configuration.ID, nil
}

func (j *JsonDatabase) SetId(ctx basecontext.ApiContext, id string) (string, error) {
	if !j.IsConnected() {
		return "", ErrDatabaseNotConnected
	}

	if j.data.Configuration == nil {
		j.data.Configuration = &models.Configuration{}
	}

	if j.data.Configuration.ID == "" {
		j.data.Configuration.ID = id
	} else {
		return "", errors.New("ID already exists")
	}

	return j.data.Configuration.ID, nil
}

func (j *JsonDatabase) generateId() string {
	id, err := security.GenerateCryptoRandomString(32)
	if err != nil {
		id = uuid.New().String()
	}

	return id
}
