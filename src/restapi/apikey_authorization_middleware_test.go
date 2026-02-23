package restapi

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	log "github.com/cjlapao/common-go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiKeyAuthorizationMiddlewareAdapter_Expiration(t *testing.T) {
	// Setup
	common.Logger = log.Get()
	dbPath := filepath.Join(t.TempDir(), "test_db.json")
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	db := data.NewJsonDatabase(ctx, dbPath)

	sp := serviceprovider.NewMockProvider()
	sp.JsonDatabase = db

	// Ensure DB is connected
	err := db.Connect(ctx)
	require.NoError(t, err)

	// Create Keys
	// 1. Valid Key
	validKey := models.ApiKey{
		ID:        "valid-id",
		Name:      "Valid",
		Key:       "VALID_KEY",
		Secret:    "secret",
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339Nano),
	}
	_, err = db.CreateApiKey(ctx, validKey)
	require.NoError(t, err)

	// 2. Expired Key
	expiredKey := models.ApiKey{
		ID:        "expired-id",
		Name:      "Expired",
		Key:       "EXPIRED_KEY",
		Secret:    "secret",
		ExpiresAt: time.Now().Add(-1 * time.Hour).Format(time.RFC3339Nano),
	}
	_, err = db.CreateApiKey(ctx, expiredKey)
	require.NoError(t, err)

	// 3. Forever Key (No Expiration)
	foreverKey := models.ApiKey{
		ID:        "forever-id",
		Name:      "Forever",
		Key:       "FOREVER_KEY",
		Secret:    "secret",
		ExpiresAt: "",
	}
	_, err = db.CreateApiKey(ctx, foreverKey)
	require.NoError(t, err)

	// 4. Revoked Key
	revokedKey := models.ApiKey{
		ID:        "revoked-id",
		Name:      "Revoked",
		Key:       "REVOKED_KEY",
		Secret:    "secret",
		ExpiresAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339Nano),
		Revoked:   true,
	}
	// We can't set Revoked in CreateApiKey directly as it ignores it usually,
	// but let's check CreateApiKey impl.
	// CreateApiKey:
	// apiKey.UpdatedAt = helpers.GetUtcCurrentDateTime()
	// apiKey.CreatedAt = helpers.GetUtcCurrentDateTime()
	// j.data.ApiKeys = append(j.data.ApiKeys, apiKey)
	// It doesn't overwrite Revoked, so setting it in struct works.
	_, err = db.CreateApiKey(ctx, revokedKey)
	require.NoError(t, err)

	// Debug
	k, _ := db.GetApiKey(ctx, "REVOKED_KEY")
	t.Logf("DEBUG: Revoked Key from DB: %+v", k)

	// Helper to make request
	makeRequest := func(key, secret string) *http.Request {
		req, _ := http.NewRequest("GET", "/", nil)
		auth := base64.StdEncoding.EncodeToString([]byte(key + ":" + secret))
		req.Header.Set("X-Api-Key", auth)
		return req
	}

	// Helper to check context
	checkContext := func(key string, expectAuthorized bool, expectedError string) {
		t.Helper()
		req := makeRequest(key, "secret")
		w := httptest.NewRecorder()

		testHandler := ApiKeyAuthorizationMiddlewareAdapter(nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value(constants.AUTHORIZATION_CONTEXT_KEY)
			var authCtx *basecontext.AuthorizationContext
			if val != nil {
				authCtx = val.(*basecontext.AuthorizationContext)
			}
			require.NotNil(t, authCtx, "AuthorizationContext should not be nil")

			if expectAuthorized {
				assert.True(t, authCtx.IsAuthorized, "Key %s should be authorized", key)
				assert.Nil(t, authCtx.AuthorizationError)
			} else {
				assert.False(t, authCtx.IsAuthorized, "Key %s should NOT be authorized", key)
				require.NotNil(t, authCtx.AuthorizationError)
				assert.Contains(t, authCtx.AuthorizationError.ErrorDescription, expectedError)
			}
		}))

		testHandler.ServeHTTP(w, req)
	}

	t.Run("Valid Key", func(t *testing.T) {
		checkContext("VALID_KEY", true, "")
	})

	t.Run("Expired Key", func(t *testing.T) {
		checkContext("EXPIRED_KEY", false, "Api Key has expired")
	})

	t.Run("Forever Key", func(t *testing.T) {
		checkContext("FOREVER_KEY", true, "")
	})

	t.Run("Revoked Key", func(t *testing.T) {
		checkContext("REVOKED_KEY", false, "Api Key has been revoked")
	})
}
