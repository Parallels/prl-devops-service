package controllers

import (
	"net/http"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
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
