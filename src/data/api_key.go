package data

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/security/password"
)

var (
	ErrApiKeyNotFound      = errors.NewWithCode("API Key not found", 404)
	ErrApiKeyAlreadyExists = errors.NewWithCode("API Key already exists", 500)
)

func (j *JsonDatabase) GetApiKeys(ctx basecontext.ApiContext, filter string) ([]models.ApiKey, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}
	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(j.data.ApiKeys, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (j *JsonDatabase) GetApiKey(ctx basecontext.ApiContext, idOrName string) (*models.ApiKey, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	for _, apiKey := range j.data.ApiKeys {
		if apiKey.ID == idOrName || strings.EqualFold(apiKey.Name, idOrName) || strings.EqualFold(apiKey.Key, idOrName) {
			return &apiKey, nil
		}
	}

	return nil, ErrApiKeyNotFound
}

func (j *JsonDatabase) CreateApiKey(ctx basecontext.ApiContext, apiKey models.ApiKey) (*models.ApiKey, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if a, _ := j.GetApiKey(ctx, apiKey.ID); a != nil {
		return nil, ErrApiKeyAlreadyExists
	}

	if a, _ := j.GetApiKey(ctx, apiKey.Name); a != nil {
		return nil, ErrApiKeyAlreadyExists
	}

	passwdSvc := password.Get()
	hashSecret, err := passwdSvc.Hash(apiKey.Secret, apiKey.ID)
	if err != nil {
		return nil, err
	}
	apiKey.Secret = hashSecret
	apiKey.UpdatedAt = helpers.GetUtcCurrentDateTime()
	apiKey.CreatedAt = helpers.GetUtcCurrentDateTime()
	j.data.ApiKeys = append(j.data.ApiKeys, apiKey)

	return &apiKey, nil
}

func (j *JsonDatabase) DeleteApiKey(ctx basecontext.ApiContext, id string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if id == "" {
		return nil
	}

	for i, apiKey := range j.data.ApiKeys {
		if strings.EqualFold(apiKey.ID, id) || strings.EqualFold(apiKey.Name, id) || strings.EqualFold(apiKey.Key, id) {
			for {
				if j.data.ApiKeys[i].DbRecord.IsLocked {
					continue
				}
				LockRecord(ctx, j.data.ApiKeys[i].DbRecord)
				j.data.ApiKeys = append(j.data.ApiKeys[:i], j.data.ApiKeys[i+1:]...)
				UnlockRecord(ctx, j.data.ApiKeys[i].DbRecord)
				break
			}

			return nil
		}
	}

	return ErrApiKeyNotFound
}

func (j *JsonDatabase) UpdateKey(ctx basecontext.ApiContext, key models.ApiKey) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, apiKey := range j.data.ApiKeys {
		if apiKey.ID != key.ID {
			continue
		}

		for {
			if j.data.ApiKeys[i].DbRecord.IsLocked {
				continue
			}
			LockRecord(ctx, j.data.ApiKeys[i].DbRecord)
			j.data.ApiKeys[i].Revoked = key.Revoked
			j.data.ApiKeys[i].RevokedAt = key.RevokedAt
			j.data.ApiKeys[i].UpdatedAt = helpers.GetUtcCurrentDateTime()
			UnlockRecord(ctx, j.data.ApiKeys[i].DbRecord)
			break
		}

		return nil
	}

	return ErrApiKeyNotFound
}

func (j *JsonDatabase) RevokeKey(ctx basecontext.ApiContext, id string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if id == "" {
		return nil
	}

	key, err := j.GetApiKey(ctx, id)
	if err != nil {
		return err
	}

	if key == nil {
		return ErrApiKeyNotFound
	}

	key.Revoked = true
	key.RevokedAt = helpers.GetUtcCurrentDateTime()
	if err := j.UpdateKey(ctx, *key); err != nil {
		return err
	}

	return nil
}
