//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/controllers"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/restapi"
	bruteforceguard "github.com/Parallels/prl-devops-service/security/brute_force_guard"
	"github.com/Parallels/prl-devops-service/security/jwt"
	"github.com/Parallels/prl-devops-service/security/password"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"
	"github.com/Parallels/prl-devops-service/startup/seeds"
	"github.com/stretchr/testify/require"
)

const (
	adminEmail    = "root@localhost"
	adminPassword = "RootPass1!Admin"
	testHMACSecret = "integration-test-hmac-secret-long-enough-32chars"
)

// testEnv holds the live test server and associated state for one integration
// test run. Create it with newTestEnv; it registers cleanup via t.Cleanup.
type testEnv struct {
	srv *httptest.Server
	ctx basecontext.ApiContext
}

// newTestEnv boots a complete in-process API server backed by a temporary JSON
// database, seeds the default claims/roles/users, and returns a testEnv whose
// server URL can be used to make real HTTP requests.
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// ── 1. Env vars ─────────────────────────────────────────────────────────
	t.Setenv("JWT_HMACS_SECRET", testHMACSecret)
	t.Setenv("ROOT_PASSWORD", adminPassword)
	// Disable password-complexity requirements so tests can use simple strings.
	t.Setenv("SECURITY_PASSWORD_MIN_PASSWORD_LENGTH", "6")
	t.Setenv("SECURITY_PASSWORD_REQUIRE_UPPERCASE", "false")
	t.Setenv("SECURITY_PASSWORD_REQUIRE_LOWERCASE", "false")
	t.Setenv("SECURITY_PASSWORD_REQUIRE_NUMBER", "false")
	t.Setenv("SECURITY_PASSWORD_REQUIRE_SPECIAL_CHAR", "false")

	// ── 2. Temp DB directory ─────────────────────────────────────────────────
	tmpDir, err := os.MkdirTemp("", "integration-test-*")
	require.NoError(t, err)
	dbFile := filepath.Join(tmpDir, "data.json")

	// ── 3. Reset the global DB singleton so this test gets a fresh database ──
	data.ResetForTesting()

	// ── 4. Global singletons ─────────────────────────────────────────────────
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	config.New(ctx)            // reads env vars on demand
	password.New(ctx)          // picks up relaxed complexity from env
	jwt.New(ctx)               // picks up JWT_HMACS_SECRET from env
	bruteforceguard.New(ctx)   // default brute-force settings

	// ── 5. Service provider with temp DB ─────────────────────────────────────
	provider := serviceprovider.NewMockProvider()
	provider.System = system.New(ctx) // required by registerConfigHandlers
	db := data.NewJsonDatabase(ctx, dbFile)
	provider.JsonDatabase = db
	provider.HardwareId = "integration-test-hardware-id"

	require.NoError(t, db.Connect(ctx))

	// ── 6. Seed default data ──────────────────────────────────────────────────
	require.NoError(t, seeds.SeedDefaultClaims())
	require.NoError(t, seeds.SeedDefaultRoles())
	require.NoError(t, seeds.SeedDefaultRoleClaims())
	require.NoError(t, seeds.SeedDefaultUsers())

	// ── 7. HTTP listener ──────────────────────────────────────────────────────
	listener := restapi.NewHttpListener()
	listener.WithVersion("v1", "/v1", true)

	require.NoError(t, controllers.RegisterV1Handlers(ctx))

	for _, c := range listener.Controllers {
		require.NoError(t, c.Serve())
	}

	srv := httptest.NewServer(listener.Router)

	t.Cleanup(func() {
		srv.Close()
		data.ResetForTesting()
		os.RemoveAll(tmpDir)
	})

	return &testEnv{srv: srv, ctx: ctx}
}

// url returns the full URL for the given API path.
func (e *testEnv) url(path string) string {
	return fmt.Sprintf("%s/v1%s", e.srv.URL, path)
}

// do performs an HTTP request, attaches an optional bearer token, and returns
// the response. The caller is responsible for closing the body.
func (e *testEnv) do(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, e.url(path), bodyReader)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// mustGetToken fetches a JWT for the given credentials or fails the test.
func (e *testEnv) mustGetToken(t *testing.T, email, pass string) string {
	t.Helper()

	resp := e.do(t, http.MethodPost, "/auth/token", map[string]string{
		"email":    email,
		"password": pass,
	}, "")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "token request failed")

	var result struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.NotEmpty(t, result.Token, "expected a non-empty token")
	return result.Token
}

// mustDecodeBody decodes the JSON body of resp into dst, closing the body.
func mustDecodeBody(t *testing.T, resp *http.Response, dst interface{}) {
	t.Helper()
	defer resp.Body.Close()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(dst))
}

// jsonDecode decodes JSON from r into dst. Convenience wrapper used by tests
// that manage the response body lifetime themselves.
func jsonDecode(r io.Reader, dst interface{}) error {
	return json.NewDecoder(r).Decode(dst)
}
