package parity

// #425 task 02 — auth wave, api keys: 13 scenarios converted from
// data/api_key_test.go (:15/:42/:67/:98/:143/:227/:309/:334/:375/:400/:467/
// :507/:556). Excluded pure-function pair: TestNormalizeApiKeyType (:272)
// and TestFilterInternalKeys (:292) — pure domain fns, engine-agnostic,
// stay Tier 1 (ledger rows say "pure").
//
// The three old-data rows (get_api_key_normalizes_type,
// get_api_keys_normalizes_types, backward_compatibility_with_old_api_keys)
// seeded via private db.data slicing in the legacy originals; here they
// run on the hook engines (seed_auth_test.go header, E-4): JSON pre-Write
// vs raw INSERT, both landing Type=="" rows no public CreateApiKey can
// produce. Assertions are the legacy ones verbatim.

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParityApiKeysAuthWave(t *testing.T) {
	testutil.RunOnEngines(t, apiKeyScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
	// Old-data trio on the pre-seed engines (E-4):
	testutil.RunOnEngines(t, apiKeyOldDataScenarios(),
		authLegacyEngine(preWriteLegacyKeys(legacyOldKeyIDs)),
		authGormEngine(insertRawLegacyKeys(legacyOldKeyIDs)))
}

var legacyOldKeyIDs = []string{"old-key", "old-key-1", "old-key-2", "legacy-1", "legacy-2"}

func apiKeyScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "api_keys/create_with_expiration", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			expiryDate := helpers.GetUtcCurrentDateTime()
			createdKey, err := db.CreateApiKey(ctx, models.ApiKey{
				ID: "test-key-1", Name: "Test Key 1", Key: "TEST_KEY_1",
				Secret: "secret", ExpiresAt: expiryDate,
			})
			require.NoError(t, err)
			assert.Equal(t, expiryDate, createdKey.ExpiresAt)
			loadedKey, err := db.GetApiKey(ctx, "test-key-1")
			require.NoError(t, err)
			assert.Equal(t, expiryDate, loadedKey.ExpiresAt)
		}},
		{Name: "api_keys/create_without_expiration", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			createdKey, err := db.CreateApiKey(ctx, models.ApiKey{
				ID: "test-key-2", Name: "Test Key 2", Key: "TEST_KEY_2", Secret: "secret",
			})
			require.NoError(t, err)
			assert.Empty(t, createdKey.ExpiresAt)
			loadedKey, err := db.GetApiKey(ctx, "test-key-2")
			require.NoError(t, err)
			assert.Empty(t, loadedKey.ExpiresAt)
		}},
		{Name: "api_keys/expiration_logic", Run: func(t *testing.T, _ ctxT, _ dbT) {
			// Pure string-date round-trip (api_key_test.go:67) — the #419
			// typing contract; runs identically on both legs.
			validKey := models.ApiKey{ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339Nano)}
			expiredKey := models.ApiKey{ExpiresAt: time.Now().Add(-1 * time.Hour).Format(time.RFC3339Nano)}
			foreverKey := models.ApiKey{ExpiresAt: ""}
			expiresAt, err := time.Parse(time.RFC3339Nano, validKey.ExpiresAt)
			require.NoError(t, err)
			assert.False(t, time.Now().UTC().After(expiresAt))
			expiresAt, err = time.Parse(time.RFC3339Nano, expiredKey.ExpiresAt)
			require.NoError(t, err)
			assert.True(t, time.Now().UTC().After(expiresAt))
			assert.Empty(t, foreverKey.ExpiresAt)
		}},
		{Name: "api_keys/with_user_ownership", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db)
			user := models.User{ID: helpers.GenerateId(), Username: "testuser",
				Name: "Test User", Email: "test@example.com", Password: "password123"}
			_, err := db.CreateUser(ctx, user)
			require.NoError(t, err)
			createdKey, err := db.CreateApiKey(ctx, models.ApiKey{
				ID: "test-key-user", Name: "Test Key with User", Key: "TEST_KEY_USER",
				Secret: "secret", UserID: user.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, user.ID, createdKey.UserID)
			loadedKey, err := db.GetApiKey(ctx, "test-key-user")
			require.NoError(t, err)
			assert.Equal(t, user.ID, loadedKey.UserID)
		}},
		{Name: "api_keys/get_api_keys_with_user_filtering", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db)
			user1 := models.User{ID: helpers.GenerateId(), Username: "user1",
				Name: "User One", Email: "user1@example.com", Password: "password123"}
			_, err := db.CreateUser(ctx, user1)
			require.NoError(t, err)
			user2 := models.User{ID: helpers.GenerateId(), Username: "user2",
				Name: "User Two", Email: "user2@example.com", Password: "password456"}
			_, err = db.CreateUser(ctx, user2)
			require.NoError(t, err)
			for _, k := range []models.ApiKey{
				{ID: "key-user1", Name: "Key for User 1", Key: "KEY_USER1", Secret: "secret1", UserID: user1.ID},
				{ID: "key-user2", Name: "Key for User 2", Key: "KEY_USER2", Secret: "secret2", UserID: user2.ID},
				{ID: "key-no-user", Name: "Key without User", Key: "KEY_NO_USER", Secret: "secret3"},
			} {
				_, err = db.CreateApiKey(ctx, k)
				require.NoError(t, err)
			}
			allKeys, err := db.GetApiKeys(ctx, "")
			require.NoError(t, err)
			assert.Equal(t, 3, len(allKeys))
			for _, key := range allKeys {
				switch key.ID {
				case "key-user1":
					assert.Equal(t, user1.ID, key.UserID)
				case "key-user2":
					assert.Equal(t, user2.ID, key.UserID)
				case "key-no-user":
					assert.Empty(t, key.UserID)
				}
			}
		}},
		{Name: "api_keys/create_without_user_id_auto_assigns_to_user", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db)
			user := models.User{ID: helpers.GenerateId(), Username: "testuser",
				Name: "Test User", Email: "test@example.com", Password: "password123"}
			_, err := db.CreateUser(ctx, user)
			require.NoError(t, err)
			createdKey, err := db.CreateApiKey(ctx, models.ApiKey{
				ID: "test-key-no-user-id", Name: "Test Key No User ID",
				Key: "TEST_KEY_NO_USER_ID", Secret: "secret",
			})
			require.NoError(t, err)
			assert.Empty(t, createdKey.UserID) // data layer does not auto-assign
		}},
		{Name: "api_keys/create_default_type_is_external", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			createdKey, err := db.CreateApiKey(ctx, models.ApiKey{
				ID: "test-key-default-type", Name: "Test Key Default Type",
				Key: "TEST_KEY_DEFAULT", Secret: "secret",
			})
			require.NoError(t, err)
			assert.Equal(t, "external", createdKey.Type)
			loadedKey, err := db.GetApiKey(ctx, "test-key-default-type")
			require.NoError(t, err)
			assert.Equal(t, "external", loadedKey.Type)
		}},
		{Name: "api_keys/create_explicit_type_preserved", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			createdKey, err := db.CreateApiKey(ctx, models.ApiKey{
				ID: "test-key-internal", Name: "Test Key Internal",
				Key: "TEST_KEY_INTERNAL", Secret: "secret", Type: "internal",
			})
			require.NoError(t, err)
			assert.Equal(t, "internal", createdKey.Type)
			loadedKey, err := db.GetApiKey(ctx, "test-key-internal")
			require.NoError(t, err)
			assert.Equal(t, "internal", loadedKey.Type)
			createdKey, err = db.CreateApiKey(ctx, models.ApiKey{
				ID: "test-key-external", Name: "Test Key External",
				Key: "TEST_KEY_EXTERNAL", Secret: "secret", Type: "external",
			})
			require.NoError(t, err)
			assert.Equal(t, "external", createdKey.Type)
		}},
		{Name: "api_keys/get_api_keys_filters_internal_keys", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			for _, k := range []models.ApiKey{
				{ID: "external-1", Name: "External Key 1", Key: "EXTERNAL_1", Secret: "secret1", Type: "external"},
				{ID: "external-2", Name: "External Key 2", Key: "EXTERNAL_2", Secret: "secret2", Type: "external"},
				{ID: "internal-1", Name: "Internal Key 1", Key: "INTERNAL_1", Secret: "secret3", Type: "internal"},
				{ID: "internal-2", Name: "Internal Key 2", Key: "INTERNAL_2", Secret: "secret4", Type: "internal"},
			} {
				_, err := db.CreateApiKey(ctx, k)
				require.NoError(t, err)
			}
			allKeys, err := db.GetApiKeys(ctx, "")
			require.NoError(t, err)
			assert.Equal(t, 2, len(allKeys), "Only external keys should be returned")
			for _, key := range allKeys {
				assert.Equal(t, "external", key.Type)
				assert.NotContains(t, []string{"internal-1", "internal-2"}, key.ID)
			}
			loadedInternal, err := db.GetApiKey(ctx, "internal-1")
			require.NoError(t, err)
			assert.Equal(t, "internal", loadedInternal.Type)
			assert.Equal(t, "internal-1", loadedInternal.ID)
		}},
		{Name: "api_keys/internal_keys_not_in_list_but_retrievable_by_id", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			_, err := db.CreateApiKey(ctx, models.ApiKey{
				ID: "local-catalog-job123", Name: "local-catalog-job123",
				Key: "local-catalog-job123", Secret: "temporary-secret",
				Type: "internal", ExpiresAt: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
			})
			require.NoError(t, err)
			_, err = db.CreateApiKey(ctx, models.ApiKey{
				ID: "regular-key", Name: "Regular Key", Key: "REGULAR_KEY",
				Secret: "regular-secret", Type: "external",
			})
			require.NoError(t, err)
			allKeys, err := db.GetApiKeys(ctx, "")
			require.NoError(t, err)
			assert.Equal(t, 1, len(allKeys), "Only external key should be in list")
			assert.Equal(t, "regular-key", allKeys[0].ID)
			loadedTemp, err := db.GetApiKey(ctx, "local-catalog-job123")
			require.NoError(t, err)
			assert.Equal(t, "local-catalog-job123", loadedTemp.ID)
			assert.Equal(t, "internal", loadedTemp.Type)
		}},
	}
}

func apiKeyOldDataScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "api_keys/get_api_key_normalizes_type", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			loadedKey, err := db.GetApiKey(ctx, "old-key")
			require.NoError(t, err)
			assert.Equal(t, "external", loadedKey.Type)
		}},
		{Name: "api_keys/get_api_keys_normalizes_types", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			allKeys, err := db.GetApiKeys(ctx, "")
			require.NoError(t, err)
			assert.Equal(t, 5, len(allKeys)) // full seeded old cohort
			for _, key := range allKeys {
				assert.Equal(t, "external", key.Type, "All keys should be normalized to external")
			}
		}},
		{Name: "api_keys/backward_compatibility_with_old_api_keys", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			allKeys, err := db.GetApiKeys(ctx, "")
			require.NoError(t, err)
			assert.Equal(t, 5, len(allKeys))
			for _, key := range allKeys {
				assert.Equal(t, "external", key.Type, "Legacy keys should be normalized to external")
			}
			legacy1, err := db.GetApiKey(ctx, "legacy-1")
			require.NoError(t, err)
			assert.Equal(t, "external", legacy1.Type)
			legacy2, err := db.GetApiKey(ctx, "legacy-2")
			require.NoError(t, err)
			assert.Equal(t, "external", legacy2.Type)
		}},
	}
}
