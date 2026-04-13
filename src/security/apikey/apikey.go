package apikey

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/security/password"
)

type ApiKeyValidationResult struct {
	ApiKeyId   string
	ApiKeyName string
	UserID     string
}

type ApiKeyValidationError struct {
	Code    int
	Message string
	Component string
}

func (e *ApiKeyValidationError) Error() string {
	return e.Message
}

func ValidateApiKey(ctx basecontext.ApiContext, dbService interface {
	GetApiKey(ctx basecontext.ApiContext, idOrName string) (*models.ApiKey, error)
}, apiKey string) (*ApiKeyValidationResult, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(apiKey)
	if err != nil {
		return nil, &ApiKeyValidationError{
			Code:    400,
			Message: "Invalid API key format",
			Component: "DecodeApiKey",
		}
	}

	parts := strings.Split(string(decodedKey), ":")
	if len(parts) != 2 {
		return nil, &ApiKeyValidationError{
			Code:    400,
			Message: "API key must be in format base64(key:secret)",
			Component: "ParseApiKey",
		}
	}

	apiKeyKey := parts[0]
	apiKeySecret := parts[1]

	dbApiKey, err := dbService.GetApiKey(ctx, apiKeyKey)
	if err != nil || dbApiKey == nil {
		return nil, &ApiKeyValidationError{
			Code:    401,
			Message: "Invalid API key",
			Component: "GetApiKey",
		}
	}

	if dbApiKey.Revoked {
		return nil, &ApiKeyValidationError{
			Code:    401,
			Message: "Api Key has been revoked",
			Component: "CheckRevoked",
		}
	}

	if dbApiKey.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339Nano, dbApiKey.ExpiresAt)
		if err == nil {
			if time.Now().UTC().After(expiresAt) {
				return nil, &ApiKeyValidationError{
					Code:    401,
					Message: "Api Key has expired",
					Component: "CheckExpiration",
				}
			}
		}
	}

	passwdSvc := password.Get()
	if err := passwdSvc.Compare(apiKeySecret, dbApiKey.ID, dbApiKey.Secret); err != nil {
		return nil, &ApiKeyValidationError{
			Code:    401,
			Message: "Invalid API key secret",
			Component: "CompareApiKey",
		}
	}

	return &ApiKeyValidationResult{
		ApiKeyId:   dbApiKey.ID,
		ApiKeyName: dbApiKey.Name,
		UserID:     dbApiKey.UserID,
	}, nil
}
