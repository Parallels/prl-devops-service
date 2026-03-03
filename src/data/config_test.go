package data

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfiguration(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	cfg, err := db.GetConfiguration(ctx)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.ID)
}

func TestSetId(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	id, err := db.SetId(ctx, "custom-id")
	require.NoError(t, err)
	assert.Equal(t, "custom-id", id)

	// Second attempt should fail
	_, err = db.SetId(ctx, "another-id")
	require.Error(t, err)
	assert.Equal(t, "ID already exists", err.Error())
}

func TestSeedId(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	id, err := db.SeedId(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	id2, err := db.GetId(ctx)
	require.NoError(t, err)
	assert.Equal(t, id, id2)
}
