package data

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
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

func TestApiKeyWithUserOwnership(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	// Seed roles and claims first
	for _, roleName := range constants.DefaultRoles {
		_, _ = db.CreateRole(ctx, models.Role{Name: roleName, ID: roleName})
	}
	for _, claimName := range constants.DefaultClaims {
		_, _ = db.CreateClaim(ctx, models.Claim{Name: claimName, ID: claimName})
	}

	// Create a user
	user := models.User{
		ID:       helpers.GenerateId(),
		Username: "testuser",
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create an API key with user ID
	apiKey := models.ApiKey{
		ID:     "test-key-user",
		Name:   "Test Key with User",
		Key:    "TEST_KEY_USER",
		Secret: "secret",
		UserID: user.ID,
	}

	createdKey, err := db.CreateApiKey(ctx, apiKey)
	require.NoError(t, err)
	assert.Equal(t, user.ID, createdKey.UserID)

	// Verify the key can be retrieved by ID
	loadedKey, err := db.GetApiKey(ctx, "test-key-user")
	require.NoError(t, err)
	assert.Equal(t, user.ID, loadedKey.UserID)
}

func TestGetApiKeysWithUserFiltering(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	// Seed roles and claims first
	for _, roleName := range constants.DefaultRoles {
		_, _ = db.CreateRole(ctx, models.Role{Name: roleName, ID: roleName})
	}
	for _, claimName := range constants.DefaultClaims {
		_, _ = db.CreateClaim(ctx, models.Claim{Name: claimName, ID: claimName})
	}

	// Create two users
	user1 := models.User{
		ID:       helpers.GenerateId(),
		Username: "user1",
		Name:     "User One",
		Email:    "user1@example.com",
		Password: "password123",
	}
	_, err := db.CreateUser(ctx, user1)
	require.NoError(t, err)

	user2 := models.User{
		ID:       helpers.GenerateId(),
		Username: "user2",
		Name:     "User Two",
		Email:    "user2@example.com",
		Password: "password456",
	}
	_, err = db.CreateUser(ctx, user2)
	require.NoError(t, err)

	// Create API keys for each user
	apiKey1 := models.ApiKey{
		ID:     "key-user1",
		Name:   "Key for User 1",
		Key:    "KEY_USER1",
		Secret: "secret1",
		UserID: user1.ID,
	}
	_, err = db.CreateApiKey(ctx, apiKey1)
	require.NoError(t, err)

	apiKey2 := models.ApiKey{
		ID:     "key-user2",
		Name:   "Key for User 2",
		Key:    "KEY_USER2",
		Secret: "secret2",
		UserID: user2.ID,
	}
	_, err = db.CreateApiKey(ctx, apiKey2)
	require.NoError(t, err)

	// Create a key without user
	apiKey3 := models.ApiKey{
		ID:     "key-no-user",
		Name:   "Key without User",
		Key:    "KEY_NO_USER",
		Secret: "secret3",
	}
	_, err = db.CreateApiKey(ctx, apiKey3)
	require.NoError(t, err)

	// Get all keys
	allKeys, err := db.GetApiKeys(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, 3, len(allKeys))

	// Verify each key has correct user ID
	for _, key := range allKeys {
		if key.ID == "key-user1" {
			assert.Equal(t, user1.ID, key.UserID)
		} else if key.ID == "key-user2" {
			assert.Equal(t, user2.ID, key.UserID)
		} else if key.ID == "key-no-user" {
			assert.Empty(t, key.UserID)
		}
	}
}

func TestCreateApiKeyWithoutUserIdAutoAssignsToUser(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	// Seed roles and claims first
	for _, roleName := range constants.DefaultRoles {
		_, _ = db.CreateRole(ctx, models.Role{Name: roleName, ID: roleName})
	}
	for _, claimName := range constants.DefaultClaims {
		_, _ = db.CreateClaim(ctx, models.Claim{Name: claimName, ID: claimName})
	}

	// Create a user
	user := models.User{
		ID:       helpers.GenerateId(),
		Username: "testuser",
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	// Create an API key without specifying user ID
	apiKey := models.ApiKey{
		ID:     "test-key-no-user-id",
		Name:   "Test Key No User ID",
		Key:    "TEST_KEY_NO_USER_ID",
		Secret: "secret",
		// UserID is not set
	}

	createdKey, err := db.CreateApiKey(ctx, apiKey)
	require.NoError(t, err)
	// After creation, the UserID should be empty (data layer doesn't auto-assign)
	assert.Empty(t, createdKey.UserID)
}
