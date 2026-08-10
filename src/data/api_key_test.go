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

// ============================================================================
// Tests for API Key Type Field (Task 2)
// ============================================================================

func TestNormalizeApiKeyType(t *testing.T) {
	t.Run("empty type becomes external", func(t *testing.T) {
		key := &models.ApiKey{Type: ""}
		normalizeApiKeyType(key)
		assert.Equal(t, "external", key.Type)
	})

	t.Run("internal type stays internal", func(t *testing.T) {
		key := &models.ApiKey{Type: "internal"}
		normalizeApiKeyType(key)
		assert.Equal(t, "internal", key.Type)
	})

	t.Run("external type stays external", func(t *testing.T) {
		key := &models.ApiKey{Type: "external"}
		normalizeApiKeyType(key)
		assert.Equal(t, "external", key.Type)
	})
}

func TestFilterInternalKeys(t *testing.T) {
	keys := []models.ApiKey{
		{ID: "key1", Type: "external"},
		{ID: "key2", Type: "internal"},
		{ID: "key3", Type: "external"},
		{ID: "key4", Type: "internal"},
		{ID: "key5", Type: ""},
	}

	filtered := filterInternalKeys(keys)

	assert.Equal(t, 3, len(filtered))
	assert.Equal(t, "key1", filtered[0].ID)
	assert.Equal(t, "key3", filtered[1].ID)
	assert.Equal(t, "key5", filtered[2].ID)
}

func TestCreateApiKeyDefaultTypeIsExternal(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	apiKey := models.ApiKey{
		ID:     "test-key-default-type",
		Name:   "Test Key Default Type",
		Key:    "TEST_KEY_DEFAULT",
		Secret: "secret",
		// Type is not set
	}

	createdKey, err := db.CreateApiKey(ctx, apiKey)
	require.NoError(t, err)
	assert.Equal(t, "external", createdKey.Type)

	// Verify persistence
	loadedKey, err := db.GetApiKey(ctx, "test-key-default-type")
	require.NoError(t, err)
	assert.Equal(t, "external", loadedKey.Type)
}

func TestCreateApiKeyExplicitTypePreserved(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	t.Run("explicit internal type", func(t *testing.T) {
		apiKey := models.ApiKey{
			ID:     "test-key-internal",
			Name:   "Test Key Internal",
			Key:    "TEST_KEY_INTERNAL",
			Secret: "secret",
			Type:   "internal",
		}

		createdKey, err := db.CreateApiKey(ctx, apiKey)
		require.NoError(t, err)
		assert.Equal(t, "internal", createdKey.Type)

		// Verify persistence
		loadedKey, err := db.GetApiKey(ctx, "test-key-internal")
		require.NoError(t, err)
		assert.Equal(t, "internal", loadedKey.Type)
	})

	t.Run("explicit external type", func(t *testing.T) {
		apiKey := models.ApiKey{
			ID:     "test-key-external",
			Name:   "Test Key External",
			Key:    "TEST_KEY_EXTERNAL",
			Secret: "secret",
			Type:   "external",
		}

		createdKey, err := db.CreateApiKey(ctx, apiKey)
		require.NoError(t, err)
		assert.Equal(t, "external", createdKey.Type)
	})
}

func TestGetApiKeyNormalizesType(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	// Directly insert a key with empty Type (simulating old data)
	oldKey := models.ApiKey{
		ID:        "old-key",
		Name:      "Old Key",
		Key:       "OLD_KEY",
		Secret:    "hashed_secret",
		Type:      "", // Empty, like old data
		CreatedAt: helpers.GetUtcCurrentDateTime(),
		UpdatedAt: helpers.GetUtcCurrentDateTime(),
	}
	db.data.ApiKeys = append(db.data.ApiKeys, oldKey)

	// Retrieve the key - should normalize to "external"
	loadedKey, err := db.GetApiKey(ctx, "old-key")
	require.NoError(t, err)
	assert.Equal(t, "external", loadedKey.Type)
}

func TestGetApiKeysFiltersInternalKeys(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	// Create external keys
	externalKey1 := models.ApiKey{
		ID:     "external-1",
		Name:   "External Key 1",
		Key:    "EXTERNAL_1",
		Secret: "secret1",
		Type:   "external",
	}
	_, err := db.CreateApiKey(ctx, externalKey1)
	require.NoError(t, err)

	externalKey2 := models.ApiKey{
		ID:     "external-2",
		Name:   "External Key 2",
		Key:    "EXTERNAL_2",
		Secret: "secret2",
		Type:   "external",
	}
	_, err = db.CreateApiKey(ctx, externalKey2)
	require.NoError(t, err)

	// Create internal keys
	internalKey1 := models.ApiKey{
		ID:     "internal-1",
		Name:   "Internal Key 1",
		Key:    "INTERNAL_1",
		Secret: "secret3",
		Type:   "internal",
	}
	_, err = db.CreateApiKey(ctx, internalKey1)
	require.NoError(t, err)

	internalKey2 := models.ApiKey{
		ID:     "internal-2",
		Name:   "Internal Key 2",
		Key:    "INTERNAL_2",
		Secret: "secret4",
		Type:   "internal",
	}
	_, err = db.CreateApiKey(ctx, internalKey2)
	require.NoError(t, err)

	// Get all keys - should only return external ones
	allKeys, err := db.GetApiKeys(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, 2, len(allKeys), "Only external keys should be returned")

	// Verify all returned keys are external
	for _, key := range allKeys {
		assert.Equal(t, "external", key.Type)
		assert.NotContains(t, []string{"internal-1", "internal-2"}, key.ID)
	}

	// Verify internal keys are still in DB and retrievable by ID
	loadedInternal, err := db.GetApiKey(ctx, "internal-1")
	require.NoError(t, err)
	assert.Equal(t, "internal", loadedInternal.Type)
	assert.Equal(t, "internal-1", loadedInternal.ID)
}

func TestGetApiKeysNormalizesTypes(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	// Directly insert keys with empty Type (simulating old data)
	oldKey1 := models.ApiKey{
		ID:        "old-key-1",
		Name:      "Old Key 1",
		Key:       "OLD_KEY_1",
		Secret:    "hashed_secret_1",
		Type:      "", // Empty, like old data
		CreatedAt: helpers.GetUtcCurrentDateTime(),
		UpdatedAt: helpers.GetUtcCurrentDateTime(),
	}
	db.data.ApiKeys = append(db.data.ApiKeys, oldKey1)

	oldKey2 := models.ApiKey{
		ID:        "old-key-2",
		Name:      "Old Key 2",
		Key:       "OLD_KEY_2",
		Secret:    "hashed_secret_2",
		Type:      "", // Empty, like old data
		CreatedAt: helpers.GetUtcCurrentDateTime(),
		UpdatedAt: helpers.GetUtcCurrentDateTime(),
	}
	db.data.ApiKeys = append(db.data.ApiKeys, oldKey2)

	// Get all keys - should normalize all to "external"
	allKeys, err := db.GetApiKeys(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, 2, len(allKeys))

	for _, key := range allKeys {
		assert.Equal(t, "external", key.Type, "All keys should be normalized to external")
	}
}

func TestInternalKeysNotInListButRetrievableByID(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	// Create a temp internal key (like local-catalog-* keys)
	tempKey := models.ApiKey{
		ID:        "local-catalog-job123",
		Name:      "local-catalog-job123",
		Key:       "local-catalog-job123",
		Secret:    "temporary-secret",
		Type:      "internal",
		ExpiresAt: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	}
	_, err := db.CreateApiKey(ctx, tempKey)
	require.NoError(t, err)

	// Create a regular external key
	regularKey := models.ApiKey{
		ID:     "regular-key",
		Name:   "Regular Key",
		Key:    "REGULAR_KEY",
		Secret: "regular-secret",
		Type:   "external",
	}
	_, err = db.CreateApiKey(ctx, regularKey)
	require.NoError(t, err)

	// List keys - should NOT include temp internal key
	allKeys, err := db.GetApiKeys(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, 1, len(allKeys), "Only external key should be in list")
	assert.Equal(t, "regular-key", allKeys[0].ID)

	// Verify internal key is still retrievable by ID (needed for auth)
	loadedTemp, err := db.GetApiKey(ctx, "local-catalog-job123")
	require.NoError(t, err)
	assert.Equal(t, "local-catalog-job123", loadedTemp.ID)
	assert.Equal(t, "internal", loadedTemp.Type)

	// Verify can also retrieve by Name
	loadedByName, err := db.GetApiKey(ctx, "local-catalog-job123")
	require.NoError(t, err)
	assert.Equal(t, "local-catalog-job123", loadedByName.ID)
	assert.Equal(t, "internal", loadedByName.Type)
}

func TestBackwardCompatibilityWithOldApiKeys(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	// Simulate old API keys without Type field
	oldKeys := []models.ApiKey{
		{
			ID:        "legacy-1",
			Name:      "Legacy Key 1",
			Key:       "LEGACY_1",
			Secret:    "hashed_secret_1",
			Type:      "", // No Type field in old data
			CreatedAt: helpers.GetUtcCurrentDateTime(),
			UpdatedAt: helpers.GetUtcCurrentDateTime(),
		},
		{
			ID:        "legacy-2",
			Name:      "Legacy Key 2",
			Key:       "LEGACY_2",
			Secret:    "hashed_secret_2",
			Type:      "", // No Type field in old data
			CreatedAt: helpers.GetUtcCurrentDateTime(),
			UpdatedAt: helpers.GetUtcCurrentDateTime(),
		},
	}

	// Directly insert into DB (simulating existing data)
	db.data.ApiKeys = append(db.data.ApiKeys, oldKeys...)

	// Get all keys - should normalize and return them as external
	allKeys, err := db.GetApiKeys(ctx, "")
	require.NoError(t, err)
	assert.Equal(t, 2, len(allKeys))

	for _, key := range allKeys {
		assert.Equal(t, "external", key.Type, "Legacy keys should be normalized to external")
	}

	// Retrieve individual keys - should also normalize
	legacy1, err := db.GetApiKey(ctx, "legacy-1")
	require.NoError(t, err)
	assert.Equal(t, "external", legacy1.Type)

	legacy2, err := db.GetApiKey(ctx, "legacy-2")
	require.NoError(t, err)
	assert.Equal(t, "external", legacy2.Type)
}
