package data

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCatalogManager(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	manager := models.CatalogManager{
		ID:                   "test-manager-1",
		Name:                 "Test Manager 1",
		URL:                  "https://example.com/catalog",
		Internal:             false,
		AuthenticationMethod: "None",
	}

	err := db.AddCatalogManager(ctx, manager)
	require.NoError(t, err)

	// Verify persistence
	loaded, err := db.GetCatalogManager("test-manager-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Manager 1", loaded.Name)
}

func TestUpdateCatalogManager(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	manager := models.CatalogManager{
		ID:                   "test-manager-2",
		Name:                 "Test Manager 2",
		URL:                  "https://example.com/catalog",
		Internal:             false,
		AuthenticationMethod: "None",
	}

	err := db.AddCatalogManager(ctx, manager)
	require.NoError(t, err)

	manager.Name = "Updated Manager 2"

	err = db.UpdateCatalogManager(ctx, manager)
	require.NoError(t, err)

	updated, err := db.GetCatalogManager("test-manager-2")
	require.NoError(t, err)
	assert.Equal(t, "Updated Manager 2", updated.Name)
}

func TestDeleteCatalogManager(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	manager := models.CatalogManager{
		ID:                   "test-manager-3",
		Name:                 "Test Manager 3",
		URL:                  "https://example.com/catalog",
		Internal:             false,
		AuthenticationMethod: "None",
	}

	err := db.AddCatalogManager(ctx, manager)
	require.NoError(t, err)

	err = db.DeleteCatalogManager(ctx, "test-manager-3")
	require.NoError(t, err)

	_, err = db.GetCatalogManager("test-manager-3")
	require.Error(t, err)
	assert.Equal(t, ErrCatalogManagerNotFound, err)
}

func TestUpdateInternalCatalogManagerFails(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	manager := models.CatalogManager{
		ID:                   "internal-manager",
		Name:                 "Internal Manager",
		Internal:             true,
		AuthenticationMethod: "None",
	}

	err := db.AddCatalogManager(ctx, manager)
	require.NoError(t, err)

	manager.Name = "Attempt Update"

	err = db.UpdateCatalogManager(ctx, manager)
	require.Error(t, err)
	assert.Equal(t, ErrUpdateInternalCatalogManager, err)
}

func TestDeleteInternalCatalogManagerFailsWithoutRoot(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	manager := models.CatalogManager{
		ID:                   "internal-manager",
		Name:                 "Internal Manager",
		Internal:             true,
		AuthenticationMethod: "None",
	}

	err := db.AddCatalogManager(ctx, manager)
	require.NoError(t, err)

	err = db.DeleteCatalogManager(ctx, "internal-manager")
	require.Error(t, err)
	assert.Equal(t, ErrRemoveInternalCatalogManager, err)
}
