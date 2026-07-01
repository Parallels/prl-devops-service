package stores_test

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/stores"
	"github.com/Parallels/prl-devops-service/database/stores/testhelpers"
	logging "github.com/cjlapao/common-go-logger"
	"github.com/stretchr/testify/assert"
)

func TestApiKeyDataStore(t *testing.T) {
	logger := logging.Get()
	logger.Info("Starting API key store tests")
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	store := &stores.ApiKeyDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	ctx := basecontext.NewBaseContext()

	t.Run("CreateApiKey", func(t *testing.T) {
		apiKey := &models.ApiKey{
			Name:   "test-api-key",
			Key:    "test-key-123",
			Secret: "test-secret-456",
			UserID: "user-123",
		}

		createdKey, diag := store.CreateApiKey(*ctx, apiKey)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, createdKey)
		assert.NotEmpty(t, createdKey.ID)
		assert.Equal(t, "test-api-key", createdKey.Name)
		assert.Equal(t, "test-key-123", createdKey.Key)
		assert.False(t, createdKey.Revoked)
	})

	t.Run("GetApiKey", func(t *testing.T) {
		// Create an API key first
		apiKey := &models.ApiKey{
			Name:   "get-test-key",
			Key:    "get-key-123",
			Secret: "get-secret-456",
		}
		createdKey, diag := store.CreateApiKey(*ctx, apiKey)
		assert.False(t, diag.HasErrors())

		// Get by ID
		retrievedKey, diag := store.GetApiKey(*ctx, createdKey.ID)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, retrievedKey)
		assert.Equal(t, createdKey.ID, retrievedKey.ID)
		assert.Equal(t, "get-test-key", retrievedKey.Name)
	})

	t.Run("GetApiKey_NotFound", func(t *testing.T) {
		retrievedKey, diag := store.GetApiKey(*ctx, "non-existent-key")
		assert.Nil(t, retrievedKey)
		if diag != nil {
			assert.False(t, diag.HasErrors())
		}
	})

	t.Run("GetApiKeys", func(t *testing.T) {
		keys, diag := store.GetApiKeys(*ctx)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, keys)
		// Should have at least the keys we created above
		assert.True(t, len(keys) >= 2)
	})

	t.Run("UpdateApiKey", func(t *testing.T) {
		// Create an API key
		apiKey := &models.ApiKey{
			Name:   "update-key",
			Key:    "update-key-123",
			Secret: "update-secret-456",
		}
		createdKey, diag := store.CreateApiKey(*ctx, apiKey)
		assert.False(t, diag.HasErrors())

		// Update the key
		createdKey.Name = "updated-key-name"
		diag = store.UpdateApiKey(*ctx, createdKey)
		assert.False(t, diag.HasErrors())

		// Verify update
		updatedKey, diag := store.GetApiKey(*ctx, createdKey.ID)
		assert.False(t, diag.HasErrors())
		assert.Equal(t, "updated-key-name", updatedKey.Name)
	})

	t.Run("RevokeApiKey", func(t *testing.T) {
		// Create an API key
		apiKey := &models.ApiKey{
			Name:   "revoke-key",
			Key:    "revoke-key-123",
			Secret: "revoke-secret-456",
		}
		createdKey, diag := store.CreateApiKey(*ctx, apiKey)
		assert.False(t, diag.HasErrors())

		// Revoke the key
		diag = store.RevokeApiKey(*ctx, createdKey.ID)
		assert.False(t, diag.HasErrors())

		// Verify revoked
		revokedKey, diag := store.GetApiKey(*ctx, createdKey.ID)
		assert.False(t, diag.HasErrors())
		assert.True(t, revokedKey.Revoked)
		assert.NotNil(t, revokedKey.RevokedAt)
	})

	t.Run("DeleteApiKey", func(t *testing.T) {
		// Create an API key
		apiKey := &models.ApiKey{
			Name:   "delete-key",
			Key:    "delete-key-123",
			Secret: "delete-secret-456",
		}
		createdKey, diag := store.CreateApiKey(*ctx, apiKey)
		assert.False(t, diag.HasErrors())

		// Delete the key
		diag = store.DeleteApiKey(*ctx, createdKey.ID)
		assert.False(t, diag.HasErrors())

		// Verify deleted
		deletedKey, diag := store.GetApiKey(*ctx, createdKey.ID)
		assert.Nil(t, deletedKey)
	})

	t.Run("CreateApiKey_WithExpiry", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		apiKey := &models.ApiKey{
			Name:      "expiry-key",
			Key:       "expiry-key-123",
			Secret:    "expiry-secret-456",
			ExpiresAt: &expiresAt,
		}

		createdKey, diag := store.CreateApiKey(*ctx, apiKey)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, createdKey)
		assert.NotNil(t, createdKey.ExpiresAt)
	})
}
