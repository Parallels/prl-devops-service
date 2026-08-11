package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCtx(t *testing.T) basecontext.ApiContext {
	t.Helper()
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	return ctx
}

// TestResolveCatalogMachineConnection_NilRequest ensures a nil request
// returns a 400 error.
func TestResolveCatalogMachineConnection_NilRequest(t *testing.T) {
	ctx := newCtx(t)
	conn, err := resolveCatalogMachineConnection(ctx, nil)
	assert.Empty(t, conn)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing catalog manifest request")
}

// TestResolveCatalogMachineConnection_BothSet ensures that providing both
// connection and catalog_manager_id returns a 400 error.
func TestResolveCatalogMachineConnection_BothSet(t *testing.T) {
	ctx := newCtx(t)
	req := &models.CreateCatalogVirtualMachineRequest{
		Connection:       "http://remote-catalog",
		CatalogManagerId: "mgr-id",
	}
	conn, err := resolveCatalogMachineConnection(ctx, req)
	assert.Empty(t, conn)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot both be provided")
}

// TestResolveCatalogMachineConnection_ConnectionSet ensures that when only
// connection is provided it is returned verbatim.
func TestResolveCatalogMachineConnection_ConnectionSet(t *testing.T) {
	ctx := newCtx(t)
	req := &models.CreateCatalogVirtualMachineRequest{
		Connection: "http://remote-catalog",
	}
	conn, err := resolveCatalogMachineConnection(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "http://remote-catalog", conn)
}

// TestResolveCatalogMachineConnection_BothEmpty_CatalogEnabled verifies that
// when both fields are empty and the catalog module is enabled, the function
// returns ("", nil) so the pull service uses the local catalog.
func TestResolveCatalogMachineConnection_BothEmpty_CatalogEnabled(t *testing.T) {
	t.Setenv(constants.ENABLED_MODULES_ENV_VAR, "catalog,api")
	ctx := newCtx(t)
	req := &models.CreateCatalogVirtualMachineRequest{}

	conn, err := resolveCatalogMachineConnection(ctx, req)
	require.NoError(t, err)
	assert.Empty(t, conn, "empty connection string signals local catalog to pull service")
}

// TestResolveCatalogMachineConnection_BothEmpty_CatalogDisabled verifies that
// when both fields are empty and the catalog module is NOT enabled, the
// function returns a 400 error.
func TestResolveCatalogMachineConnection_BothEmpty_CatalogDisabled(t *testing.T) {
	// Ensure catalog is not in the enabled modules list.
	t.Setenv(constants.ENABLED_MODULES_ENV_VAR, "api")
	ctx := newCtx(t)
	req := &models.CreateCatalogVirtualMachineRequest{}

	conn, err := resolveCatalogMachineConnection(ctx, req)
	assert.Empty(t, conn)
	require.Error(t, err)

	type coder interface{ Code() int }
	if ce, ok := err.(coder); ok {
		assert.Equal(t, http.StatusBadRequest, ce.Code())
	}
	assert.Contains(t, err.Error(), "local catalog is not enabled")
}

// setupTestDBForController creates a temporary test database and configures
// the service provider for controller tests
func setupTestDBForController(t *testing.T) (*data.JsonDatabase, string, basecontext.ApiContext) {
	t.Helper()

	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "prl-devops-controller-test-*")
	require.NoError(t, err)

	dbFile := filepath.Join(tmpDir, "test_db.json")

	// Create context
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	// Reset config
	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	// Create and connect database
	db := data.NewJsonDatabase(ctx, dbFile)
	require.NotNil(t, db)
	require.NoError(t, db.Connect(ctx))

	// Set up service provider with test database
	// Use NewMockProvider to ensure globalProvider is initialized
	sp := serviceprovider.NewMockProvider()
	require.NotNil(t, sp)
	sp.JsonDatabase = db

	return db, tmpDir, ctx
}

// cleanupTestDBForController cleans up test database and temp directory
func cleanupTestDBForController(t *testing.T, tmpDir string, db *data.JsonDatabase) {
	t.Helper()

	if db != nil {
		ctx := basecontext.NewBaseContext()
		ctx.DisableLog()
		_ = db.Disconnect(ctx)
	}

	time.Sleep(50 * time.Millisecond)

	err := os.RemoveAll(tmpDir)
	require.NoError(t, err)
}

// TestBuildLocalCatalogConnection_CleanDNS tests normal DNS hostname
func TestBuildLocalCatalogConnection_CleanDNS(t *testing.T) {
	db, tmpDir, ctx := setupTestDBForController(t)
	defer cleanupTestDBForController(t, tmpDir, db)

	t.Setenv(constants.ORCHESTRATOR_PUBLIC_URL, "my-service-orchestrator.com")
	t.Setenv(constants.API_PORT_ENV_VAR, "9999")
	// Note: TLS enabled but without valid certs will default to HTTP
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "false")

	// Reload config to pick up env vars
	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	connStr, keyName, err := buildLocalCatalogConnection(ctx, "caller-123", "job-456")

	require.NoError(t, err)
	assert.Equal(t, "local-catalog-job-456", keyName)

	// Verify connection string format: host=<secret>@http://my-service-orchestrator.com:9999
	assert.Contains(t, connStr, "@http://my-service-orchestrator.com:9999")
	assert.True(t, strings.HasPrefix(connStr, "host="))

	// Verify API key was created
	apiKey, err := db.GetApiKey(ctx, keyName)
	require.NoError(t, err)
	assert.NotNil(t, apiKey)
	assert.Equal(t, "internal", apiKey.Type)
	assert.Equal(t, "caller-123", apiKey.UserID)
}

// TestBuildLocalCatalogConnection_HTTPSPrefix tests DNS with https:// prefix
func TestBuildLocalCatalogConnection_HTTPSPrefix(t *testing.T) {
	db, tmpDir, ctx := setupTestDBForController(t)
	defer cleanupTestDBForController(t, tmpDir, db)

	t.Setenv(constants.ORCHESTRATOR_PUBLIC_URL, "https://my-orchestrator.com")
	t.Setenv(constants.API_PORT_ENV_VAR, "8080")
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "false")

	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	connStr, keyName, err := buildLocalCatalogConnection(ctx, "user-1", "job-1")

	require.NoError(t, err)
	assert.Equal(t, "local-catalog-job-1", keyName)

	// Should strip https:// prefix and use http (TLS disabled)
	assert.Contains(t, connStr, "@http://my-orchestrator.com:8080")
	assert.NotContains(t, connStr, "https://https://")
}

// TestBuildLocalCatalogConnection_HTTPPrefix tests DNS with http:// prefix
func TestBuildLocalCatalogConnection_HTTPPrefix(t *testing.T) {
	db, tmpDir, ctx := setupTestDBForController(t)
	defer cleanupTestDBForController(t, tmpDir, db)

	t.Setenv(constants.ORCHESTRATOR_PUBLIC_URL, "http://orchestrator.local")
	t.Setenv(constants.API_PORT_ENV_VAR, "7777")
	// TLS disabled for this test
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "false")

	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	connStr, _, err := buildLocalCatalogConnection(ctx, "admin", "job-999")

	require.NoError(t, err)
	// Should strip http:// prefix and use http (TLS disabled)
	assert.Contains(t, connStr, "@http://orchestrator.local:7777")
	assert.NotContains(t, connStr, "http://http://")
}

// TestBuildLocalCatalogConnection_TrailingSlash tests DNS with trailing slash
func TestBuildLocalCatalogConnection_TrailingSlash(t *testing.T) {
	db, tmpDir, ctx := setupTestDBForController(t)
	defer cleanupTestDBForController(t, tmpDir, db)

	t.Setenv(constants.ORCHESTRATOR_PUBLIC_URL, "my-service.com/")
	t.Setenv(constants.API_PORT_ENV_VAR, "5000")
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "false")

	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	connStr, _, err := buildLocalCatalogConnection(ctx, "test-user", "job-abc")

	require.NoError(t, err)
	// Should strip trailing slash
	assert.Contains(t, connStr, "@http://my-service.com:5000")
	assert.NotContains(t, connStr, "my-service.com/:")
}

// TestBuildLocalCatalogConnection_WithPort tests DNS with port number
func TestBuildLocalCatalogConnection_WithPort(t *testing.T) {
	db, tmpDir, ctx := setupTestDBForController(t)
	defer cleanupTestDBForController(t, tmpDir, db)

	t.Setenv(constants.ORCHESTRATOR_PUBLIC_URL, "orchestrator.example.com:8888")
	t.Setenv(constants.API_PORT_ENV_VAR, "9999")
	// TLS disabled
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "false")

	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	connStr, _, err := buildLocalCatalogConnection(ctx, "user", "job-xyz")

	require.NoError(t, err)
	// Should strip port 8888 and use apiPort 9999 with HTTP
	assert.Contains(t, connStr, "@http://orchestrator.example.com:9999")
	assert.NotContains(t, connStr, ":8888")
}

// TestBuildLocalCatalogConnection_LocalhostDefault tests empty/whitespace falls back to localhost
func TestBuildLocalCatalogConnection_LocalhostDefault(t *testing.T) {
	db, tmpDir, ctx := setupTestDBForController(t)
	defer cleanupTestDBForController(t, tmpDir, db)

	// Don't set ORCHESTRATOR_PUBLIC_URL, should default to localhost
	t.Setenv(constants.API_PORT_ENV_VAR, "3000")
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "false")

	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	connStr, _, err := buildLocalCatalogConnection(ctx, "default-user", "job-default")

	require.NoError(t, err)
	// Should use localhost default
	assert.Contains(t, connStr, "@http://localhost:3000")
}

// TestBuildLocalCatalogConnection_ComplexScenario tests multiple edge cases combined
func TestBuildLocalCatalogConnection_ComplexScenario(t *testing.T) {
	db, tmpDir, ctx := setupTestDBForController(t)
	defer cleanupTestDBForController(t, tmpDir, db)

	// Combination: https prefix + port + trailing slash
	t.Setenv(constants.ORCHESTRATOR_PUBLIC_URL, "https://my-complex-host.io:7777/")
	t.Setenv(constants.API_PORT_ENV_VAR, "9999")
	// TLS disabled
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "false")

	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	connStr, _, err := buildLocalCatalogConnection(ctx, "complex-user", "job-complex")

	require.NoError(t, err)
	// Should clean all: remove https://, remove :7777, remove /, use :9999 with HTTP
	assert.Contains(t, connStr, "@http://my-complex-host.io:9999")
	assert.NotContains(t, connStr, "https://https://")
	assert.NotContains(t, connStr, ":7777")
	assert.NotContains(t, connStr, "/:")
}

// TestBuildLocalCatalogConnection_WhitespaceHandling tests whitespace in config
func TestBuildLocalCatalogConnection_WhitespaceHandling(t *testing.T) {
	db, tmpDir, ctx := setupTestDBForController(t)
	defer cleanupTestDBForController(t, tmpDir, db)

	t.Setenv(constants.ORCHESTRATOR_PUBLIC_URL, "  orchestrator.space.com  ")
	t.Setenv(constants.API_PORT_ENV_VAR, "4000")
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "false")

	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	connStr, _, err := buildLocalCatalogConnection(ctx, "space-user", "job-space")

	require.NoError(t, err)
	// Should trim whitespace
	assert.Contains(t, connStr, "@http://orchestrator.space.com:4000")
	assert.NotContains(t, connStr, "  ")
}

// TestBuildLocalCatalogConnection_APIKeyExpiration tests that temp keys have proper expiration
func TestBuildLocalCatalogConnection_APIKeyExpiration(t *testing.T) {
	db, tmpDir, ctx := setupTestDBForController(t)
	defer cleanupTestDBForController(t, tmpDir, db)

	t.Setenv(constants.ORCHESTRATOR_PUBLIC_URL, "test-host.local")
	t.Setenv(constants.API_PORT_ENV_VAR, "8080")
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "false")

	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	beforeCreate := time.Now()
	_, keyName, err := buildLocalCatalogConnection(ctx, "test", "job-exp")
	afterCreate := time.Now()

	require.NoError(t, err)

	// Verify key has expiration ~2 hours from now
	apiKey, err := db.GetApiKey(ctx, keyName)
	require.NoError(t, err)
	require.NotEmpty(t, apiKey.ExpiresAt)

	expiresAt, err := time.Parse(time.RFC3339, apiKey.ExpiresAt)
	require.NoError(t, err)

	// Should expire approximately 2 hours from creation (allow 1 minute tolerance)
	expectedExpiry := beforeCreate.Add(2 * time.Hour)
	assert.WithinDuration(t, expectedExpiry, expiresAt, 1*time.Minute)
	assert.True(t, expiresAt.After(afterCreate))
}

// TestBuildLocalCatalogConnection_SecretInConnectionString tests that plaintext secret is used
func TestBuildLocalCatalogConnection_SecretInConnectionString(t *testing.T) {
	db, tmpDir, ctx := setupTestDBForController(t)
	defer cleanupTestDBForController(t, tmpDir, db)

	t.Setenv(constants.ORCHESTRATOR_PUBLIC_URL, "secret-test.com")
	t.Setenv(constants.API_PORT_ENV_VAR, "6000")
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "false")

	cfg := config.New(ctx)
	require.NotNil(t, cfg)

	connStr, keyName, err := buildLocalCatalogConnection(ctx, "secret-user", "job-secret")
	require.NoError(t, err)

	// Extract secret from connection string: host=<secret>@http://...
	parts := strings.Split(connStr, "@")
	require.Len(t, parts, 2, "connection string should have format: host=<secret>@<url>")

	secretPart := strings.TrimPrefix(parts[0], "host=")
	assert.NotEmpty(t, secretPart, "secret should be present in connection string")

	// Verify the secret in connection string is NOT the hashed version stored in DB
	apiKey, err := db.GetApiKey(ctx, keyName)
	require.NoError(t, err)

	// The secret in DB should be hashed, not matching the plaintext in connection string
	assert.NotEqual(t, secretPart, apiKey.Secret, "DB should store hashed secret, not plaintext")
	assert.NotEmpty(t, apiKey.Secret, "DB secret should be present")
}
