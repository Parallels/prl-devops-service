package data

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateApiKeyWithExpiration(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	expiryDate := helpers.GetUtcCurrentDateTime()

	apiKey := models.ApiKey{
		ID:        "test-key-1",
		Name:      "Test Key 1",
		Key:       "TEST_KEY_1",
		Secret:    "secret",
		ExpiresAt: expiryDate,
	}

	createdKey, err := db.CreateApiKey(ctx, apiKey)
	require.NoError(t, err)
	assert.Equal(t, expiryDate, createdKey.ExpiresAt)

	// Verify persistence
	loadedKey, err := db.GetApiKey(ctx, "test-key-1")
	require.NoError(t, err)
	assert.Equal(t, expiryDate, loadedKey.ExpiresAt)
}

func TestCreateApiKeyWithoutExpiration(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	apiKey := models.ApiKey{
		ID:     "test-key-2",
		Name:   "Test Key 2",
		Key:    "TEST_KEY_2",
		Secret: "secret",
		// ExpiresAt is empty
	}

	createdKey, err := db.CreateApiKey(ctx, apiKey)
	require.NoError(t, err)
	assert.Empty(t, createdKey.ExpiresAt)

	// Verify persistence
	loadedKey, err := db.GetApiKey(ctx, "test-key-2")
	require.NoError(t, err)
	assert.Empty(t, loadedKey.ExpiresAt)
}

func TestApiKeyExpirationLogic(t *testing.T) {
	// This test verifies the logic we will use in middleware
	// It doesn't test the middleware itself, but the logic
	// of checking expiration against current time

	validKey := models.ApiKey{
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339Nano),
	}

	expiredKey := models.ApiKey{
		ExpiresAt: time.Now().Add(-1 * time.Hour).Format(time.RFC3339Nano),
	}

	foreverKey := models.ApiKey{
		ExpiresAt: "",
	}

	// Check valid key
	expiresAt, err := time.Parse(time.RFC3339Nano, validKey.ExpiresAt)
	require.NoError(t, err)
	assert.False(t, time.Now().UTC().After(expiresAt))

	// Check expired key
	expiresAt, err = time.Parse(time.RFC3339Nano, expiredKey.ExpiresAt)
	require.NoError(t, err)
	assert.True(t, time.Now().UTC().After(expiresAt))

	// Check forever key
	assert.Empty(t, foreverKey.ExpiresAt)
}
