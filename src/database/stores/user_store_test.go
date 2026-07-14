package stores_test

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/database/stores"
	"github.com/Parallels/prl-devops-service/database/stores/testhelpers"
	logging "github.com/cjlapao/common-go-logger"
	"github.com/stretchr/testify/assert"
)

func TestUserDataStore(t *testing.T) {
	logger := logging.Get()
	logger.Info("Starting user store tests")
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	store := &stores.UserDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	ctx := basecontext.NewBaseContext()

	t.Run("CreateUser", func(t *testing.T) {
		user := &models.User{
			Username: "testuser",
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "hashedpassword123",
		}

		createdUser, diag := store.CreateUser(*ctx, user)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, createdUser)
		assert.NotEmpty(t, createdUser.ID)
		assert.Equal(t, "testuser", createdUser.Username)
		assert.Equal(t, "test@example.com", createdUser.Email)
	})

	t.Run("GetUserByID", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			Username: "getbyid",
			Name:     "Get By ID",
			Email:    "getbyid@example.com",
			Password: "password",
		}
		createdUser, diag := store.CreateUser(*ctx, user)
		assert.False(t, diag.HasErrors())

		// Get the user
		retrievedUser, diag := store.GetUserByID(*ctx, createdUser.ID)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, retrievedUser)
		assert.Equal(t, createdUser.ID, retrievedUser.ID)
		assert.Equal(t, "getbyid", retrievedUser.Username)
	})

	t.Run("GetUserByID_NotFound", func(t *testing.T) {
		retrievedUser, diag := store.GetUserByID(*ctx, "non-existent-id")
		assert.Nil(t, retrievedUser)
		if diag != nil {
			assert.False(t, diag.HasErrors())
		}
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			Username: "byusername",
			Name:     "By Username",
			Email:    "byusername@example.com",
			Password: "password",
		}
		createdUser, diag := store.CreateUser(*ctx, user)
		assert.False(t, diag.HasErrors())

		// Get by username
		retrievedUser, diag := store.GetUserByUsername(*ctx, "byusername")
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, retrievedUser)
		assert.Equal(t, createdUser.ID, retrievedUser.ID)
	})

	t.Run("GetUsers", func(t *testing.T) {
		users, diag := store.GetUsers(*ctx)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, users)
		assert.True(t, len(users) >= 3)
	})

	t.Run("BlockUser", func(t *testing.T) {
		user := &models.User{
			Username: "blockme",
			Name:     "Block Me",
			Email:    "blockme@example.com",
			Password: "password",
		}
		createdUser, diag := store.CreateUser(*ctx, user)
		assert.False(t, diag.HasErrors())

		diag = store.BlockUser(*ctx, createdUser.ID)
		assert.False(t, diag.HasErrors())

		blockedUser, diag := store.GetUserByID(*ctx, createdUser.ID)
		assert.False(t, diag.HasErrors())
		assert.True(t, blockedUser.Blocked)
	})

	t.Run("DeleteUser", func(t *testing.T) {
		user := &models.User{
			Username: "deleteme",
			Name:     "Delete Me",
			Email:    "deleteme@example.com",
			Password: "password",
		}
		createdUser, diag := store.CreateUser(*ctx, user)
		assert.False(t, diag.HasErrors())

		diag = store.DeleteUser(*ctx, createdUser.ID)
		assert.False(t, diag.HasErrors())

		deletedUser, diag := store.GetUserByID(*ctx, createdUser.ID)
		assert.Nil(t, deletedUser)
	})
}
