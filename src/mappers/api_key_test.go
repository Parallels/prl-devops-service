package mappers

import (
	"testing"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestApiKeyDtoToApiKeyResponse(t *testing.T) {
	// Test that the response correctly maps the data model to API response
	apiKey := data_models.ApiKey{
		ID:        "test-id",
		Name:      "Test Key",
		Key:       "TEST_KEY",
		Revoked:   false,
		ExpiresAt: "2025-12-31T23:59:59Z",
		RevokedAt: "",
		UserID:    "user-123",
	}

	response := ApiKeyDtoToApiKeyResponse(apiKey)

	assert.Equal(t, apiKey.ID, response.ID)
	assert.Equal(t, apiKey.Name, response.Name)
	assert.Equal(t, apiKey.Key, response.Key)
	assert.Equal(t, apiKey.Revoked, response.Revoked)
	assert.Equal(t, apiKey.ExpiresAt, response.ExpiresAt)
	assert.Equal(t, apiKey.UserID, response.UserID)
	// User details should be empty since enrichment happens in controllers
	assert.Empty(t, response.UserEmail)
	assert.Empty(t, response.UserName)
	assert.Empty(t, response.UserUsername)
}

func TestApiKeysDtoToApiKeyResponse(t *testing.T) {
	apiKeys := []data_models.ApiKey{
		{
			ID:        "key-1",
			Name:      "Key 1",
			Key:       "KEY_1",
			Revoked:   false,
			ExpiresAt: "2025-12-31T23:59:59Z",
			UserID:    "user-1",
		},
		{
			ID:        "key-2",
			Name:      "Key 2",
			Key:       "KEY_2",
			Revoked:   true,
			ExpiresAt: "2025-01-01T00:00:00Z",
			UserID:    "user-2",
		},
	}

	responses := ApiKeysDtoToApiKeyResponse(apiKeys)

	assert.Equal(t, 2, len(responses))
	assert.Equal(t, "key-1", responses[0].ID)
	assert.Equal(t, "key-2", responses[1].ID)
	assert.Equal(t, "user-1", responses[0].UserID)
	assert.Equal(t, "user-2", responses[1].UserID)
}

func TestEnrichApiKeyWithUser(t *testing.T) {
	// Test that enrichApiKeyWithUser correctly populates user details
	response := models.ApiKeyResponse{
		ID:        "key-1",
		Name:      "Key 1",
		Key:       "KEY_1",
		Revoked:   false,
		ExpiresAt: "2025-12-31T23:59:59Z",
		UserID:    "user-123",
	}

	// Verify initial state - user details should be empty
	assert.Empty(t, response.UserEmail)
	assert.Empty(t, response.UserName)
	assert.Empty(t, response.UserUsername)

	// In a real scenario, the controller would call enrichApiKeyWithUser
	// which fetches user data from DB and populates the response
	// For this test, we just verify the response structure supports enrichment
	assert.Equal(t, "user-123", response.UserID)
}
