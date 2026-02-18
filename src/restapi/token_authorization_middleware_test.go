package restapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	data_modules "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/security/jwt"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenAuthorizationMiddlewareAdapter(t *testing.T) {
	// Setup JWT service for testing
	jwtSvc := jwt.Get()

	// Setup ServiceProvider and DB
	sp := serviceprovider.NewMockProvider()

	// Create temp DB file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "data.json")

	// Initialize JsonDatabase
	ctx := basecontext.NewBaseContext()
	db := data.NewJsonDatabase(ctx, dbPath)
	sp.JsonDatabase = db
	db.Connect(ctx)

	// Create default roles
	for _, role := range constants.DefaultRoles {
		_, err := db.CreateRole(ctx, data_modules.Role{
			ID:   role,
			Name: role,
		})
		require.NoError(t, err)
	}

	// Create default claims
	for _, claim := range constants.DefaultClaims {
		_, err := db.CreateClaim(ctx, data_modules.Claim{
			ID:   claim,
			Name: claim,
		})
		require.NoError(t, err)
	}

	// Create test user
	_, err := db.CreateUser(ctx, data_modules.User{
		ID:       "test-user-id",
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
		Password: "password",
	})
	require.NoError(t, err)

	// Create a valid token for testing
	validToken, err := jwtSvc.Sign(map[string]interface{}{
		"id":       "test-user-id",
		"email":    "test@example.com",
		"username": "testuser",
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
		checkAuth      bool
	}{
		{
			name: "Auth via Header",
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", validToken))
			},
			expectedStatus: http.StatusOK,
			checkAuth:      true,
		},
		{
			name: "Auth via access_token Query Param",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("access_token", validToken)
				r.URL.RawQuery = q.Encode()
			},
			expectedStatus: http.StatusOK,
			checkAuth:      true,
		},
		{
			name: "Auth via authorization Query Param",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("authorization", validToken)
				r.URL.RawQuery = q.Encode()
			},
			expectedStatus: http.StatusOK,
			checkAuth:      true,
		},
		{
			name: "Auth via authorization Query Param with Bearer prefix",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("authorization", fmt.Sprintf("Bearer %s", validToken))
				r.URL.RawQuery = q.Encode()
			},
			expectedStatus: http.StatusOK,
			checkAuth:      true,
		},
		{
			name:           "No Auth",
			setupRequest:   func(r *http.Request) {},
			expectedStatus: http.StatusOK, // The middleware passes through but marks as unauthorized if other checks fail, but here we check IsAuthorized
			checkAuth:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock handler to check context
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := basecontext.NewBaseContextFromRequest(r)
				authCtx := ctx.GetAuthorizationContext()

				if tt.checkAuth {
					assert.True(t, authCtx.IsAuthorized, "Expected request to be authorized")
				} else {
					assert.False(t, authCtx.IsAuthorized, "Expected request to NOT be authorized")
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create middleware chain
			authContextMiddleware := AddAuthorizationContextMiddlewareAdapter()
			tokenAuthMiddleware := TokenAuthorizationMiddlewareAdapter(nil, nil)

			// Chain them: context -> token -> handler
			wrappedHandler := authContextMiddleware(tokenAuthMiddleware(handler))

			// Create request
			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)
			tt.setupRequest(req)

			// Record response
			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)
		})
	}
}
