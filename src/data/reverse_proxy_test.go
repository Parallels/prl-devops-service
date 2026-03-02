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
	assert.Equal(t, 1, createdRoute.Order)

	secondRoute := models.ReverseProxyHostHttpRoute{
		ID:   "route-2",
		Path: "/admin",
	}
	createdSecondRoute, err := db.CreateReverseProxyHostHttpRoute(ctx, createdHost.ID, secondRoute)
	require.NoError(t, err)
	assert.Equal(t, 2, createdSecondRoute.Order)

	route.Path = "/v1/api"
	route.ID = createdRoute.ID // since route.ID is overridden
	route.Order = createdRoute.Order
	updatedRoute, err := db.UpdateReverseProxyHostHttpRoute(ctx, createdHost.ID, route)
	require.NoError(t, err)
	assert.Equal(t, "/v1/api", updatedRoute.Path)
	assert.Equal(t, 1, updatedRoute.Order)

	route.Order = 2
	_, err = db.UpdateReverseProxyHostHttpRoute(ctx, createdHost.ID, route)
	require.Error(t, err)
	assert.Equal(t, ErrorReverseProxyHttpRouteOrderUpdate, err)
}

func TestReorderAndDeleteReverseProxyHostHttpRoute(t *testing.T) {
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

	firstRoute, err := db.CreateReverseProxyHostHttpRoute(ctx, createdHost.ID, models.ReverseProxyHostHttpRoute{Path: "/1"})
	require.NoError(t, err)
	secondRoute, err := db.CreateReverseProxyHostHttpRoute(ctx, createdHost.ID, models.ReverseProxyHostHttpRoute{Path: "/2"})
	require.NoError(t, err)
	thirdRoute, err := db.CreateReverseProxyHostHttpRoute(ctx, createdHost.ID, models.ReverseProxyHostHttpRoute{Path: "/3"})
	require.NoError(t, err)
	require.Equal(t, 1, firstRoute.Order)
	require.Equal(t, 2, secondRoute.Order)
	require.Equal(t, 3, thirdRoute.Order)

	loadedHostBeforeReorder, err := db.GetReverseProxyHost(ctx, createdHost.ID)
	require.NoError(t, err)
	require.Len(t, loadedHostBeforeReorder.HttpRoutes, 3)
	routeToMoveID := loadedHostBeforeReorder.HttpRoutes[0].ID

	updatedHost, err := db.ReorderReverseProxyHostHttpRoute(ctx, createdHost.ID, routeToMoveID, 3)
	require.NoError(t, err)
	require.Len(t, updatedHost.HttpRoutes, 3)
	reorderedRouteOrders := map[string]int{}
	reorderedRouteIDs := map[int]string{}
	for _, route := range updatedHost.HttpRoutes {
		reorderedRouteOrders[route.ID] = route.Order
		reorderedRouteIDs[route.Order] = route.ID
	}
	assert.Equal(t, 3, reorderedRouteOrders[routeToMoveID])
	assert.ElementsMatch(t, []int{1, 2, 3}, []int{
		reorderedRouteOrders[updatedHost.HttpRoutes[0].ID],
		reorderedRouteOrders[updatedHost.HttpRoutes[1].ID],
		reorderedRouteOrders[updatedHost.HttpRoutes[2].ID],
	})

	deleteRouteID := reorderedRouteIDs[2]
	err = db.DeleteReverseProxyHostHttpRoute(ctx, createdHost.ID, deleteRouteID)
	require.NoError(t, err)

	loadedHost, err := db.GetReverseProxyHost(ctx, createdHost.ID)
	require.NoError(t, err)
	require.Len(t, loadedHost.HttpRoutes, 2)
	loadedRouteOrders := map[string]int{}
	for _, route := range loadedHost.HttpRoutes {
		loadedRouteOrders[route.ID] = route.Order
	}
	assert.Equal(t, 2, loadedRouteOrders[routeToMoveID])
	assert.ElementsMatch(t, []int{1, 2}, []int{
		loadedHost.HttpRoutes[0].Order,
		loadedHost.HttpRoutes[1].Order,
	})
}
