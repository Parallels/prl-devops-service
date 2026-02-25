package data

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateClaim(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	claim := models.Claim{
		Name:     "READ_ONLY_ACCESS",
		Internal: false,
	}

	created, err := db.CreateClaim(ctx, claim)
	require.NoError(t, err)
	assert.Equal(t, "READ_ONLY_ACCESS", created.ID)
	assert.Equal(t, "READ_ONLY_ACCESS", created.Name)

	loaded, err := db.GetClaim(ctx, "READ_ONLY_ACCESS")
	require.NoError(t, err)
	assert.Equal(t, "READ_ONLY_ACCESS", loaded.Name)
}

func TestUpdateClaim(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	claim := models.Claim{
		Name:     "CLAIM_UPDATER",
		Internal: false,
	}

	created, err := db.CreateClaim(ctx, claim)
	require.NoError(t, err)

	updatedClaim := *created
	updatedClaim.Name = "Updated Name"

	updated, err := db.UpdateClaim(ctx, &updatedClaim)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
}

func TestDeleteClaim(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	claim := models.Claim{
		Name:     "CLAIM_DELETER",
		Internal: false,
	}

	_, err := db.CreateClaim(ctx, claim)
	require.NoError(t, err)

	err = db.DeleteClaim(ctx, "CLAIM_DELETER")
	require.NoError(t, err)

	_, err = db.GetClaim(ctx, "CLAIM_DELETER")
	require.Error(t, err)
	assert.Equal(t, ErrClaimNotFound, err)
}

func TestDeleteInternalClaimFailsWithoutRoot(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	claim := models.Claim{
		Name:     "SYSTEM_CLAIM",
		Internal: true,
	}

	_, err := db.CreateClaim(ctx, claim)
	require.NoError(t, err)

	err = db.DeleteClaim(ctx, "SYSTEM_CLAIM")
	require.Error(t, err)
	assert.Equal(t, ErrRemoveInternalClaim, err)
}
