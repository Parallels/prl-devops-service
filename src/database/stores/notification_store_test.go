package stores_test

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/stores"
	"github.com/Parallels/prl-devops-service/database/stores/testhelpers"
	logging "github.com/cjlapao/common-go-logger"
	"github.com/stretchr/testify/assert"
)

func TestNotificationDataStore(t *testing.T) {
	logger := logging.Get()
	logger.Info("Starting notification store tests")
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	store := &stores.NotificationDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	ctx := basecontext.NewBaseContext()
	userID := "test-user-123"

	t.Run("CreateNotification", func(t *testing.T) {
		notification := &models.Notification{
			UserID:  userID,
			Subject: "Test Notification",
			Content: "This is a test notification",
			Type:    "info",
			Channel: "email",
		}

		created, diag := store.CreateNotification(*ctx, notification)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, created)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, userID, created.UserID)
		assert.False(t, created.Read)
	})

	t.Run("GetUnreadCount", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			notification := &models.Notification{
				UserID:  userID,
				Subject: "Unread Notification",
				Content: "Test content",
				Type:    "info",
			}
			store.CreateNotification(*ctx, notification)
		}

		count, diag := store.GetUnreadCount(*ctx, userID)
		assert.False(t, diag.HasErrors())
		assert.True(t, count >= 3)
	})

	t.Run("MarkAsRead", func(t *testing.T) {
		notification := &models.Notification{
			UserID:  userID,
			Subject: "Mark as Read Test",
			Content: "This will be marked as read",
			Type:    "info",
		}
		created, diag := store.CreateNotification(*ctx, notification)
		assert.False(t, diag.HasErrors())

		diag = store.MarkAsRead(*ctx, created.ID, userID)
		assert.False(t, diag.HasErrors())

		var updated models.Notification
		db.First(&updated, "id = ?", created.ID)
		assert.True(t, updated.Read)
		assert.NotNil(t, updated.ReadAt)
	})

	t.Run("MarkAllAsRead", func(t *testing.T) {
		for i := 1; i <= 2; i++ {
			notification := &models.Notification{
				UserID:  userID,
				Subject: "Bulk Read Test",
				Content: "Test content",
				Type:    "info",
			}
			store.CreateNotification(*ctx, notification)
		}

		diag := store.MarkAllAsRead(*ctx, userID)
		assert.False(t, diag.HasErrors())

		count, diag := store.GetUnreadCount(*ctx, userID)
		assert.False(t, diag.HasErrors())
		assert.Equal(t, int64(0), count)
	})

	t.Run("CreateNotification_WithDifferentTypes", func(t *testing.T) {
		types := []string{"info", "warning", "error", "success"}
		for _, notifType := range types {
			notification := &models.Notification{
				UserID:  userID,
				Subject: notifType + " notification",
				Content: "Content for " + notifType,
				Type:    notifType,
			}
			created, diag := store.CreateNotification(*ctx, notification)
			assert.False(t, diag.HasErrors())
			assert.Equal(t, notifType, created.Type)
		}
	})

	t.Run("DeleteNotification", func(t *testing.T) {
		notification := &models.Notification{
			UserID:  userID,
			Subject: "Delete Test",
			Content: "This will be deleted",
			Type:    "info",
		}
		created, diag := store.CreateNotification(*ctx, notification)
		assert.False(t, diag.HasErrors())

		result := db.Delete(created)
		assert.NoError(t, result.Error)
	})
}
