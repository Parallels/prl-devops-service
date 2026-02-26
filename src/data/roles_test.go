package data

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRole(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	role := models.Role{
		Name:     "API_CONSUMER",
		Internal: false,
	}

	created, err := db.CreateRole(ctx, role)
	require.NoError(t, err)
	assert.Equal(t, "API_CONSUMER", created.ID)
	assert.Equal(t, "API_CONSUMER", created.Name)

	loaded, err := db.GetRole(ctx, "API_CONSUMER")
	require.NoError(t, err)
	assert.Equal(t, "API_CONSUMER", loaded.Name)
}

func TestUpdateRole(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	role := models.Role{
		Name:     "ROLE_UPDATER",
		Internal: false,
	}

	created, err := db.CreateRole(ctx, role)
	require.NoError(t, err)

	updatedRole := *created
	updatedRole.Name = "Updated Name"

	updated, err := db.UpdateRole(ctx, &updatedRole)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
}

func TestDeleteRole(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	role := models.Role{
		Name:     "ROLE_DELETER",
		Internal: false,
	}

	_, err := db.CreateRole(ctx, role)
	require.NoError(t, err)

	err = db.DeleteRole(ctx, "ROLE_DELETER")
	require.NoError(t, err)

	_, err = db.GetRole(ctx, "ROLE_DELETER")
	require.Error(t, err)
	assert.Equal(t, ErrRoleNotFound, err)
}

func TestDeleteInternalRoleFailsWithoutRoot(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	role := models.Role{
		Name:     "SYSTEM_ADMIN",
		Internal: true,
	}

	_, err := db.CreateRole(ctx, role)
	require.NoError(t, err)

	err = db.DeleteRole(ctx, "SYSTEM_ADMIN")
	require.Error(t, err)
	assert.Equal(t, ErrRemoveInternalRole, err)
}
