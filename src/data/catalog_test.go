package data

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCatalogManifest(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	manifest := models.CatalogManifest{
		ID:           "test-manifest-1",
		Name:         "test-manifest",
		Version:      "1.0.0",
		Architecture: "amd64",
	}

	created, err := db.CreateCatalogManifest(ctx, manifest)
	require.NoError(t, err)
	assert.NotEmpty(t, created.CreatedAt)
	assert.NotEmpty(t, created.UpdatedAt)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "test-manifest", created.Name)

	// Verify persistence
	loaded, err := db.GetCatalogManifestByName(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "test-manifest", loaded.Name)
}

func TestUpdateCatalogManifest(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	manifest := models.CatalogManifest{
		ID:           "test-manifest-2",
		Name:         "test-manifest",
		Version:      "1.0.0",
		Architecture: "amd64",
	}

	created, err := db.CreateCatalogManifest(ctx, manifest)
	require.NoError(t, err)

	// Ensure time difference for UpdatedAt
	time.Sleep(10 * time.Millisecond)

	updatedManifest := *created
	updatedManifest.Name = "updated-manifest"

	updated, err := db.UpdateCatalogManifest(ctx, updatedManifest)
	require.NoError(t, err)
	assert.Equal(t, "updated-manifest", updated.Name)
	assert.NotEqual(t, created.UpdatedAt, updated.UpdatedAt)
}

func TestDeleteCatalogManifest(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	manifest := models.CatalogManifest{
		ID:           "test-manifest-3",
		Name:         "test-manifest",
		Version:      "1.0.0",
		Architecture: "amd64",
	}

	created, err := db.CreateCatalogManifest(ctx, manifest)
	require.NoError(t, err)

	err = db.DeleteCatalogManifest(ctx, created.ID)
	require.NoError(t, err)

	_, err = db.GetCatalogManifestByName(ctx, created.ID)
	require.Error(t, err)
	assert.Equal(t, ErrCatalogManifestNotFound, err)
}

func TestUpdateCatalogManifestTags(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	manifest := models.CatalogManifest{
		ID:           "test-manifest-4",
		Name:         "test-manifest",
		Version:      "1.0.0",
		Architecture: "amd64",
		Tags:         []string{"tag1"},
	}

	created, err := db.CreateCatalogManifest(ctx, manifest)
	require.NoError(t, err)
	assert.Contains(t, created.Tags, "tag1")

	created.Tags = append(created.Tags, "tag2")
	err = db.UpdateCatalogManifestTags(ctx, *created)
	require.NoError(t, err)

	loaded, err := db.GetCatalogManifestByName(ctx, created.ID)
	if err != nil {
		all, allErr := db.GetCatalogManifests(ctx, "")
		t.Logf("Raw data array length: %d", len(db.data.ManifestsCatalog))
		t.Logf("GetCatalogManifests returned err: %v, len=%d", allErr, len(all))
		for _, m := range all {
			t.Logf("In DB: ID=%s, Name=%s (looking for %s)", m.ID, m.Name, created.ID)
		}
	}
	require.NoError(t, err)
	assert.Contains(t, loaded.Tags, "tag1")
	assert.Contains(t, loaded.Tags, "tag2")
}

func TestTaintAndRevokeCatalogManifestVersion(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	manifest := models.CatalogManifest{
		ID:           "test-manifest-5",
		CatalogId:    "catalog-1",
		Version:      "1.0.0",
		Architecture: "amd64",
	}

	_, err := db.CreateCatalogManifest(ctx, manifest)
	require.NoError(t, err)

	// Taint
	tainted, err := db.TaintCatalogManifestVersion(ctx, "catalog-1", "1.0.0")
	require.NoError(t, err)
	assert.True(t, tainted.Tainted)

	// Untaint
	untainted, err := db.UnTaintCatalogManifestVersion(ctx, "catalog-1", "1.0.0")
	require.NoError(t, err)
	assert.False(t, untainted.Tainted)

	// Revoke
	revoked, err := db.RevokeCatalogManifestVersion(ctx, "catalog-1", "1.0.0")
	require.NoError(t, err)
	assert.True(t, revoked.Revoked)
}
