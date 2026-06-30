package stores_test

import (
	"context"
	"testing"

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

func TestMessageDataStore(t *testing.T) {
	service.Initialize(models.LogConfig{})
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	store := &stores.MessageDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	err = store.Migrate()
	assert.NoError(t, err)

	ctx := context.Background()

	t.Run("CreateMessage", func(t *testing.T) {
		message := &models.Message{
			Type:    "email",
			Payload: "{}",
			Status:  models.MessageStatusPending,
		}

		err := store.CreateMessage(ctx, message)
		assert.NoError(t, err)
		assert.NotEmpty(t, message.ID)
	})

	t.Run("GetPendingMessages", func(t *testing.T) {
		message := &models.Message{
			Type:    "push",
			Payload: "{}",
			Status:  models.MessageStatusPending,
		}
		err := store.CreateMessage(ctx, message)
		require.NoError(t, err)

		messages, err := store.GetPendingMessages(ctx, 10)
		assert.NoError(t, err)
		assert.NotEmpty(t, messages)

		found := false
		for _, m := range messages {
			if m.ID == message.ID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("GetWorkerByName_NotFound", func(t *testing.T) {
		// New behavior: returns nil, nil for not found
		worker, err := store.GetWorkerByName(ctx, "non-existent-worker")
		assert.NoError(t, err)
		assert.Nil(t, worker)
	})
}
