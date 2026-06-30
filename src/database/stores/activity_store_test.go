package stores_test

import (
	"context"
	"testing"

	activity_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/stores"
	"github.com/cjlapao/common-go-logger/models"
	"github.com/cjlapao/common-go-logger/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestActivityDataStore(t *testing.T) {
	service.Initialize(models.LogConfig{})
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	store := &stores.ActivityDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	err = store.Migrate()
	assert.NoError(t, err)

	ctx := basecontext.NewContext(context.Background())
	tenantID := "test-tenant-id"

	t.Run("CreateActivity", func(t *testing.T) {
		activity := &models.Activity{
			BaseModelWithTenant: common.BaseModelWithTenant{
				TenantID: tenantID,
			},
			ActivityType: "test_action",
			Module:       "test_module",
			ActorType:    "user",
			ActorID:      "user-1",
			ActorName:    "Test User",
			Success:      true,
			Message:      "Test Details",
		}

		// CreateActivity(ctx, tenantID, activity)
		createdActivity, diag := store.CreateActivity(ctx, tenantID, activity)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, createdActivity)
		assert.NotEmpty(t, createdActivity.ID)
		assert.Equal(t, activity_models.ActivityType("test_action"), createdActivity.ActivityType)
	})

	t.Run("GetActivityByID", func(t *testing.T) {
		activity := &models.Activity{
			BaseModelWithTenant: common.BaseModelWithTenant{
				TenantID: tenantID,
			},
			ActivityType: "get_by_id_action",
			ActorID:      "user-2",
			Message:      "Get By ID",
			Module:       "test",
			Service:      "test",
		}

		createdActivity, diag := store.CreateActivity(ctx, tenantID, activity)
		require.False(t, diag.HasErrors())

		retrievedActivity, diag := store.GetActivityByID(ctx, tenantID, createdActivity.ID)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, retrievedActivity)
		assert.Equal(t, createdActivity.ID, retrievedActivity.ID)
	})

	t.Run("GetActivityByID_NotFound", func(t *testing.T) {
		retrievedActivity, diag := store.GetActivityByID(ctx, tenantID, "non-existent-id")
		assert.Nil(t, retrievedActivity)
		// Should return nil, nil for not found as per refactoring
		if diag != nil {
			assert.False(t, diag.HasErrors())
		}
	})

	t.Run("GetActivities", func(t *testing.T) {
		activities, diag := store.GetActivities(ctx, tenantID)
		assert.False(t, diag.HasErrors())
		assert.True(t, len(activities) >= 2) // We created 2 activities above
	})
}
