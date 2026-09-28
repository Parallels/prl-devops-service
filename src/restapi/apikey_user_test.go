package restapi

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/stretchr/testify/require"
)

type apiKeyRestrictedRecord struct{}

func (apiKeyRestrictedRecord) GetRequiredClaims() []string { return []string{"missing"} }
func (apiKeyRestrictedRecord) GetRequiredRoles() []string  { return nil }

func TestAssignedAPIKeyPermissions(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()
	db := data.NewJsonDatabase(ctx, filepath.Join(t.TempDir(), "user-key.json"))
	serviceprovider.NewMockProvider().JsonDatabase = db
	direct, err := db.CreateClaim(ctx, models.Claim{Name: "direct"})
	require.NoError(t, err)
	inherited, err := db.CreateClaim(ctx, models.Claim{Name: "inherited"})
	require.NoError(t, err)
	role, err := db.CreateRole(ctx, models.Role{Name: "key-role", Claims: []models.Claim{*inherited}})
	require.NoError(t, err)
	user, err := db.CreateUser(ctx, models.User{
		ID: "key-permission-user", Name: "Key user", Username: "key-permission-user", Email: "key-permission@example.test",
		Roles:  []models.Role{*role},
		Claims: []models.Claim{*direct},
	})
	require.NoError(t, err)
	_, err = db.CreateRole(ctx, models.Role{Name: constants.SUPER_USER_ROLE})
	require.NoError(t, err)
	admin, err := db.CreateUser(ctx, models.User{
		ID: "key-admin", Name: "Key admin", Username: "key-admin", Email: "key-admin@example.test",
		Roles:  []models.Role{{ID: constants.SUPER_USER_ROLE, Name: constants.SUPER_USER_ROLE}},
		Claims: []models.Claim{*direct},
	})
	require.NoError(t, err)
	for _, tc := range []struct {
		name          string
		userID        string
		roles, claims []string
		op            ComparisonOperation
		allowed       bool
	}{
		{"direct", user.ID, nil, []string{"direct"}, ComparisonOperationAnd, true},
		{"inherited", user.ID, []string{role.ID}, []string{"inherited"}, ComparisonOperationAnd, true},
		{"or", user.ID, nil, []string{"missing", "direct"}, ComparisonOperationOr, true},
		{"and", user.ID, nil, []string{"missing", "direct"}, ComparisonOperationAnd, false},
		{"role denied", user.ID, []string{"missing"}, nil, ComparisonOperationAnd, false},
		{"missing user", "deleted-user", nil, nil, ComparisonOperationAnd, false},
		{"legacy", "", []string{"missing"}, nil, ComparisonOperationAnd, true},
		{"superuser", admin.ID, []string{"missing"}, []string{"missing"}, ComparisonOperationAnd, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := "permission-key-" + tc.name
			_, err := db.CreateApiKey(ctx, models.ApiKey{ID: id, Name: id, Key: id, Secret: "secret", UserID: tc.userID})
			require.NoError(t, err)
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("X-Api-Key", base64.StdEncoding.EncodeToString([]byte(id+":secret")))
			h := ApiKeyAuthorizationMiddlewareAdapter(tc.roles, tc.claims, tc.op, tc.op)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				auth := basecontext.NewBaseContextFromRequest(r).GetAuthorizationContext()
				require.Equal(t, tc.allowed, auth.IsAuthorized)
				if !tc.allowed {
					require.NotNil(t, auth.AuthorizationError)
					return
				}
				if tc.userID != "" {
					require.NotNil(t, auth.User)
					require.Equal(t, tc.userID, auth.User.ID)
					if tc.userID == user.ID {
						require.True(t, auth.HasEffectiveClaim("inherited"))
						require.False(t, data.IsAuthorized(basecontext.NewBaseContextFromRequest(r), apiKeyRestrictedRecord{}))
					} else {
						require.True(t, auth.IsSuperUser)
					}
					require.False(t, auth.IsMicroService)
				} else {
					require.Nil(t, auth.User)
					require.True(t, auth.IsMicroService)
				}
			}))
			h.ServeHTTP(httptest.NewRecorder(), r)
		})
	}
}
