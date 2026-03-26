package catalog_test

import (
	"testing"

	catalog_models "github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPullManifestValidate_CatalogEnabled verifies that when the catalog
// module is enabled an empty Connection is allowed (local catalog path).
func TestPullManifestValidate_CatalogEnabled(t *testing.T) {
	t.Setenv(constants.ENABLED_MODULES_ENV_VAR, "catalog,api")
	req := &catalog_models.PullCatalogManifestRequest{
		CatalogId:   "MY_CATALOG",
		Version:     "v1",
		MachineName: "test-vm",
		// Connection intentionally empty — should fall through to local catalog
	}
	err := req.Validate()
	require.NoError(t, err, "empty connection must be accepted when catalog module is enabled")
}

// TestPullManifestValidate_CatalogDisabled verifies that an empty Connection
// returns ErrMissingConnection when the catalog module is NOT enabled.
func TestPullManifestValidate_CatalogDisabled(t *testing.T) {
	t.Setenv(constants.ENABLED_MODULES_ENV_VAR, "api")
	req := &catalog_models.PullCatalogManifestRequest{
		CatalogId:   "MY_CATALOG",
		Version:     "v1",
		MachineName: "test-vm",
	}
	err := req.Validate()
	require.Error(t, err)
	assert.Equal(t, catalog_models.ErrMissingConnection, err)
}

// TestPullManifestValidate_ConnectionSet verifies that when a connection is
// provided, validation passes regardless of catalog module state.
func TestPullManifestValidate_ConnectionSet(t *testing.T) {
	t.Setenv(constants.ENABLED_MODULES_ENV_VAR, "api")
	req := &catalog_models.PullCatalogManifestRequest{
		CatalogId:   "MY_CATALOG",
		Version:     "v1",
		MachineName: "test-vm",
		Connection:  "host=catalog.example.com;username=user;password=pass",
	}
	err := req.Validate()
	require.NoError(t, err)
}
