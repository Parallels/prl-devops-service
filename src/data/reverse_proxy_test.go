package data

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnableAndDisableProxyConfig(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	_, err := db.UpdateReverseProxy(ctx, models.ReverseProxy{})
	require.NoError(t, err)

	enabled, err := db.EnableProxyConfig(ctx)
	require.NoError(t, err)
	assert.True(t, enabled.Enabled)

	disabled, err := db.DisableProxyConfig(ctx)
	require.NoError(t, err)
	assert.False(t, disabled.Enabled)
}

func TestUpdateReverseProxy(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	rp := models.ReverseProxy{
		Enabled: true,
		Host:    "example.com",
		Port:    "443",
	}

	updated, err := db.UpdateReverseProxy(ctx, rp)
	require.NoError(t, err)
	assert.True(t, updated.Enabled)
	assert.Equal(t, "example.com", updated.Host)
	assert.Equal(t, "443", updated.Port)

	loaded, err := db.GetReverseProxyConfig(ctx)
	require.NoError(t, err)
	assert.True(t, loaded.Enabled)
	assert.Equal(t, "example.com", loaded.Host)
}

func TestCreateAndGetReverseProxyHost(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	rpHost := models.ReverseProxyHost{
		ID:   "host-1",
		Host: "app.example.com",
		Port: "8080",
	}

	created, err := db.CreateReverseProxyHost(ctx, rpHost)
	require.NoError(t, err)
	assert.Equal(t, "app.example.com", created.Host)

	loaded, err := db.GetReverseProxyHost(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "app.example.com", loaded.Host)
}

func TestUpdateReverseProxyHost(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	rpHost := models.ReverseProxyHost{
		ID:   "host-2",
		Host: "app.example.com",
		Port: "8080",
	}

	created, err := db.CreateReverseProxyHost(ctx, rpHost)
	require.NoError(t, err)

	updatedHost := *created
	updatedHost.Port = "8443"

	updated, err := db.UpdateReverseProxyHost(ctx, &updatedHost)
	require.NoError(t, err)
	assert.Equal(t, "8443", updated.Port)
}

func TestConfigureTLSAndCORS(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	rpHost := models.ReverseProxyHost{
		ID:   "host-tls",
		Host: "secure.example.com",
		Port: "443",
	}

	created, err := db.CreateReverseProxyHost(ctx, rpHost)
	require.NoError(t, err)

	tlsConfig := models.ReverseProxyHostTls{
		Enabled: true,
	}

	configured, err := db.ConfigureReverseProxyHostTls(ctx, created.ID, tlsConfig)
	require.NoError(t, err)
	assert.NotNil(t, configured.Tls)
	assert.True(t, configured.Tls.Enabled)

	corsConfig := models.ReverseProxyHostCors{
		Enabled: true,
	}

	configuredCors, err := db.ConfigureReverseProxyHostCors(ctx, created.ID, corsConfig)
	require.NoError(t, err)
	assert.NotNil(t, configuredCors.Cors)
	assert.True(t, configuredCors.Cors.Enabled)
}

func TestCreateAndUpdateReverseProxyHostHttpRoute(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	rpHost := models.ReverseProxyHost{
		ID:   "host-route",
		Host: "app.example.com",
		Port: "8080",
	}

	createdHost, err := db.CreateReverseProxyHost(ctx, rpHost)
	require.NoError(t, err)

	route := models.ReverseProxyHostHttpRoute{
		ID:   "route-1",
		Path: "/api",
	}

	createdRoute, err := db.CreateReverseProxyHostHttpRoute(ctx, createdHost.ID, route)
	require.NoError(t, err)
	assert.Equal(t, "/api", createdRoute.Path)

	route.Path = "/v1/api"
	route.ID = createdRoute.ID // since route.ID is overridden
	updatedRoute, err := db.UpdateReverseProxyHostHttpRoute(ctx, createdHost.ID, route)
	require.NoError(t, err)
	assert.Equal(t, "/v1/api", updatedRoute.Path)
}
