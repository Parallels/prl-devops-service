package data

import (
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"errors"
	"strings"
	"time"
)

func (j *JsonDatabase) GetApiKeys() ([]models.ApiKey, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	return j.data.ApiKeys, nil
}

func (j *JsonDatabase) GetApiKey(idOrName string) (*models.ApiKey, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	for _, apiKey := range j.data.ApiKeys {
		if apiKey.ID == idOrName || strings.EqualFold(apiKey.Name, idOrName) || strings.EqualFold(apiKey.Key, idOrName) {
			return &apiKey, nil
		}
	}

	return nil, nil
}

func (j *JsonDatabase) CreateApiKey(apiKey *models.ApiKey) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if a, _ := j.GetApiKey(apiKey.ID); a != nil {
		return errors.New("API Key already exists")
	}

	if a, _ := j.GetApiKey(apiKey.Name); a != nil {
		return errors.New("API Key already exists")
	}

	// Hash the password with SHA-256
	apiKey.Secret = helpers.Sha256Hash(apiKey.Secret)
	apiKey.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	apiKey.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	j.data.ApiKeys = append(j.data.ApiKeys, *apiKey)
	j.save()

	return nil
}

func (j *JsonDatabase) RemoveKey(id string) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if id == "" {
		return nil
	}

	for i, apiKey := range j.data.ApiKeys {
		if apiKey.ID == id {
			// Remove the key from the slice
			j.data.ApiKeys = append(j.data.ApiKeys[:i], j.data.ApiKeys[i+1:]...)
			j.save()
			return nil
		}
	}

	return errors.New("API Key not found")
}

func (j *JsonDatabase) UpdateKey(key models.ApiKey) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	for i, apiKey := range j.data.ApiKeys {
		if apiKey.ID == key.ID {
			j.data.ApiKeys[i].Revoked = key.Revoked
			j.data.ApiKeys[i].RevokedAt = key.RevokedAt
			j.data.ApiKeys[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			j.save()
			return nil
		}
	}

	return errors.New("API Key not found")
}

func (j *JsonDatabase) RevokeKey(id string) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if id == "" {
		return nil
	}

	key, err := j.GetApiKey(id)
	if err != nil {
		return err
	}

	if key == nil {
		return errors.New("API Key not found")
	}

	key.Revoked = true
	key.RevokedAt = time.Now().UTC().Format(time.RFC3339Nano)
	j.UpdateKey(*key)
	j.save()

	return nil
}
