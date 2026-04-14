package data

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/security/password"
)

var (
	ErrApiKeyNotFound      = errors.NewWithCode("API Key not found", 404)
	ErrApiKeyAlreadyExists = errors.NewWithCode("API Key already exists", 500)
)

func GetOwnRecords[T interface{ GetUserID() string }](ctx basecontext.ApiContext, records ...T) []T {
	authContext := ctx.GetAuthorizationContext()
	if authContext == nil || authContext.User == nil {
		return records
	}

	result := make([]T, 0)
	for _, record := range records {
		if record.GetUserID() == authContext.User.ID {
			result = append(result, record)
		}
	}

	return result
}

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

	authContext := ctx.GetAuthorizationContext()
	if authContext == nil || authContext.User == nil {
		return filteredData, nil
	}

	hasFullListClaim := authContext.HasEffectiveClaim(constants.LIST_API_KEY_CLAIM)
	hasOwnListClaim := authContext.HasEffectiveClaim(constants.LIST_OWN_API_KEY_CLAIM)

	if !hasFullListClaim && hasOwnListClaim {
		filteredData = GetOwnRecords(ctx, filteredData...)
	}

	return filteredData, nil
}

func (j *JsonDatabase) GetApiKey(ctx basecontext.ApiContext, idOrName string) (*models.ApiKey, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	for _, apiKey := range j.data.ApiKeys {
		if apiKey.ID == idOrName || strings.EqualFold(apiKey.Name, idOrName) || strings.EqualFold(apiKey.Key, idOrName) {
			authContext := ctx.GetAuthorizationContext()
			if authContext != nil && authContext.User != nil {
				hasFullListClaim := authContext.HasEffectiveClaim(constants.LIST_API_KEY_CLAIM)
				hasOwnListClaim := authContext.HasEffectiveClaim(constants.LIST_OWN_API_KEY_CLAIM)

				if !hasFullListClaim && hasOwnListClaim && apiKey.UserID != authContext.User.ID {
					return nil, ErrApiKeyNotFound
				}
			}

			return &apiKey, nil
		}
	}

	return nil, ErrApiKeyNotFound
}

// GetUserByApiKey retrieves a user linked to an API key. If the key has no associated user
// or the user doesn't exist, it returns the system root user.
func (j *JsonDatabase) GetUserByApiKey(ctx basecontext.ApiContext, apiKey string) (*models.User, error) {
	apiKeyRec, err := j.GetApiKey(ctx, apiKey)
	if err != nil {
		return nil, ErrApiKeyNotFound
	}
	if apiKeyRec.Revoked {
		return nil, ErrApiKeyNotFound
	}
	if apiKeyRec.UserID != "" {
		user, err := j.GetUser(ctx, apiKeyRec.UserID)
		if err != nil || user == nil {
			// User doesn't exist; return root user
			return j.GetUser(ctx, constants.ROOT_USER_ID)
		}
		return user, nil
	}
	// No user linked; return root user
	return j.GetUser(ctx, constants.ROOT_USER_ID)
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
				if IsRecordLocked(j.data.ApiKeys[i].DbRecord) {
					continue
				}
				LockRecord(ctx, j.data.ApiKeys[i].DbRecord)
				j.data.ApiKeys = append(j.data.ApiKeys[:i], j.data.ApiKeys[i+1:]...)
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
			if IsRecordLocked(j.data.ApiKeys[i].DbRecord) {
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
