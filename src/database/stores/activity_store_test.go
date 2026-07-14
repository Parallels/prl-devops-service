package stores_test

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/stores"
	"github.com/Parallels/prl-devops-service/database/stores/testhelpers"
	logging "github.com/cjlapao/common-go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivityDataStore(t *testing.T) {
	logger := logging.Get()
	logger.Info("Starting activity store tests")
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	store := &stores.ActivityDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	ctx := basecontext.NewBaseContext()

	t.Run("CreateActivity", func(t *testing.T) {
		activity := &models.Activity{
			ActivityType: "test_action",
			Module:       "test_module",
			ActorType:    "user",
			ActorID:      "user-1",
			ActorName:    "Test User",
			Success:      true,
			Message:      "Test Details",
			Service:      "test_service",
		}

		createdActivity, diag := store.CreateActivity(*ctx, activity)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, createdActivity)
		assert.NotEmpty(t, createdActivity.ID)
		assert.Equal(t, models.ActivityType("test_action"), createdActivity.ActivityType)
	})

	t.Run("GetActivityByID", func(t *testing.T) {
		activity := &models.Activity{
			ActivityType: "get_by_id_action",
			ActorID:      "user-2",
			ActorName:    "Test User 2",
			ActorType:    "user",
			Message:      "Get By ID",
			Module:       "test",
			Service:      "test",
			Success:      true,
		}

		createdActivity, diag := store.CreateActivity(*ctx, activity)
		require.False(t, diag.HasErrors())

		retrievedActivity, diag := store.GetActivityByID(*ctx, createdActivity.ID)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, retrievedActivity)
		assert.Equal(t, createdActivity.ID, retrievedActivity.ID)
	})

	t.Run("GetActivityByID_NotFound", func(t *testing.T) {
		retrievedActivity, diag := store.GetActivityByID(*ctx, "non-existent-id")
		assert.Nil(t, retrievedActivity)
		// Should return nil, nil for not found as per refactoring
		if diag != nil {
			assert.False(t, diag.HasErrors())
		}
	})

	t.Run("GetActivities", func(t *testing.T) {
		activities, diag := store.GetActivities(*ctx)
		assert.False(t, diag.HasErrors())
		assert.True(t, len(activities) >= 2) // We created 2 activities above
	})
}
