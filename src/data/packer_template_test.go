package data

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddPackerTemplate(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	template := models.PackerTemplate{
		ID:       "template-1",
		Name:     "Ubuntu 22.04 Setup",
		Internal: false,
	}

	created, err := db.AddPackerTemplate(ctx, &template)
	require.NoError(t, err)
	assert.NotEmpty(t, created.CreatedAt)
	assert.Equal(t, "template-1", created.ID)
	assert.Equal(t, "Ubuntu 22.04 Setup", created.Name)

	loaded, err := db.GetPackerTemplate(ctx, "template-1")
	require.NoError(t, err)
	assert.Equal(t, "Ubuntu 22.04 Setup", loaded.Name)
}

func TestUpdatePackerTemplate(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	template := models.PackerTemplate{
		ID:       "template-2",
		Name:     "Initial Name",
		Internal: false,
	}

	created, err := db.AddPackerTemplate(ctx, &template)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	updatedTemplate := *created
	updatedTemplate.Name = "Updated Name"

	updated, err := db.UpdatePackerTemplate(ctx, &updatedTemplate)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
}

func TestDeletePackerTemplate(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	template := models.PackerTemplate{
		ID:       "template-3",
		Name:     "Delete Me",
		Internal: false,
	}

	_, err := db.AddPackerTemplate(ctx, &template)
	require.NoError(t, err)

	err = db.DeletePackerTemplate(ctx, "template-3")
	require.NoError(t, err)

	_, err = db.GetPackerTemplate(ctx, "template-3")
	require.Error(t, err)
	assert.Equal(t, ErrPackerTemplateNotFound, err)
}

func TestDeleteInternalPackerTemplateFailsWithoutRoot(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	template := models.PackerTemplate{
		ID:       "template-internal",
		Name:     "Internal Setup",
		Internal: true,
	}

	_, err := db.AddPackerTemplate(ctx, &template)
	require.NoError(t, err)

	err = db.DeletePackerTemplate(ctx, "template-internal")
	require.Error(t, err)
	assert.Equal(t, ErrRemovingInternalPackerTemplate, err)
}
