package controllers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	dbmodels "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestCatalogManagersWithAssignedAPIKey(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()
	db := data.NewJsonDatabase(ctx, filepath.Join(t.TempDir(), "catalog-key.json"))
	serviceprovider.NewMockProvider().JsonDatabase = db
	claim := dbmodels.Claim{ID: constants.CATALOG_MANAGER_LIST_OWN_CLAIM, Name: constants.CATALOG_MANAGER_LIST_OWN_CLAIM}
	_, err := db.CreateClaim(ctx, claim)
	require.NoError(t, err)
	role, err := db.CreateRole(ctx, dbmodels.Role{Name: "catalog-key-role", Claims: []dbmodels.Claim{claim}})
	require.NoError(t, err)
	baseline, err := db.CreateClaim(ctx, dbmodels.Claim{Name: "catalog-key-baseline"})
	require.NoError(t, err)
	user, err := db.CreateUser(ctx, dbmodels.User{
		ID: "catalog-key-user", Name: "Catalog key user", Username: "catalog-key-user", Email: "catalog-key@example.test",
		Roles:  []dbmodels.Role{*role},
		Claims: []dbmodels.Claim{*baseline},
	})
	require.NoError(t, err)
	key := dbmodels.ApiKey{ID: "catalog-key", Key: "catalog-key", Name: "catalog-key", Secret: "test-secret", UserID: user.ID}
	_, err = db.CreateApiKey(ctx, key)
	require.NoError(t, err)
	for _, mgr := range []dbmodels.CatalogManager{
		{ID: "key-owned", Name: "Owned", OwnerID: user.ID, Active: true},
		{ID: "key-other", Name: "Other", OwnerID: "another-user", Active: true},
	} {
		require.NoError(t, db.AddCatalogManager(ctx, mgr))
	}
	request := func(handler restapi.ControllerHandler, id string, claims []string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/catalog-managers", nil)
		r = mux.SetURLVars(r, map[string]string{"id": id})
		r.Header.Set("X-Api-Key", base64.StdEncoding.EncodeToString([]byte(key.Key+":"+key.Secret)))
		// A user key must not gain access by claiming to be a trusted forwarder.
		r.Header.Set("X-SOURCE", "CATALOG_MANAGER_REQUEST")
		r.Header.Set(constants.X_CLAIMS_HEADER, base64.StdEncoding.EncodeToString([]byte(constants.CATALOG_MANAGER_LIST_CLAIM)))
		r.Header.Set(constants.X_SUPER_USER_HEADER, "true")
		wrapped := restapi.ApiKeyAuthorizationMiddlewareAdapter(nil, claims, restapi.ComparisonOperationAnd, restapi.ComparisonOperationOr)(
			restapi.XClaimsMiddlewareAdapter()(restapi.EndAuthorizationMiddlewareAdapter()(http.HandlerFunc(handler))))
		w := httptest.NewRecorder()
		wrapped.ServeHTTP(w, r)
		return w
	}
	claims := []string{constants.CATALOG_MANAGER_LIST_CLAIM, constants.CATALOG_MANAGER_LIST_OWN_CLAIM}
	w := request(GetCatalogManagersHandler(), "", claims)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var managers []struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &managers))
	require.Len(t, managers, 1)
	require.Equal(t, "key-owned", managers[0].ID)
	w = request(GetCatalogManagerByIdHandler(), "key-owned", claims)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	w = request(GetCatalogManagerByIdHandler(), "key-other", claims)
	require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	w = request(GetCatalogManagersHandler(), "", []string{constants.CATALOG_MANAGER_DELETE_CLAIM})
	require.Equal(t, http.StatusUnauthorized, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "permissions")
}
