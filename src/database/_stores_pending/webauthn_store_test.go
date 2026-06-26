package stores_test

import (
	"context"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/database/stores"
	"github.com/cjlapao/common-go-logger/models"
	"github.com/cjlapao/common-go-logger/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWebAuthnDataStore(t *testing.T) {
	service.Initialize(models.LogConfig{})
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	store := &stores.WebAuthnDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	err = store.Migrate()
	assert.NoError(t, err)

	ctx := appctx.NewContext(context.Background())
	userID := "test-user-id"

	t.Run("SaveCredential", func(t *testing.T) {
		cred := &entities.WebAuthnCredential{
			BaseModelWithTenant: common.BaseModelWithTenant{
				ID: "cred-1",
			},
			UserID:          userID,
			CredentialID:    []byte("cred-id-1"),
			PublicKey:       []byte("public-key"),
			AttestationType: "none",
			SignCount:       0,
		}

		diag := store.SaveCredential(ctx, cred)
		assert.False(t, diag.HasErrors())
	})

	t.Run("GetCredentialsByUser", func(t *testing.T) {
		creds, diag := store.GetCredentialsByUser(ctx, userID)
		assert.False(t, diag.HasErrors())
		assert.NotEmpty(t, creds)
		assert.Equal(t, "cred-1", creds[0].ID)
	})

	t.Run("GetCredentialByCredentialID", func(t *testing.T) {
		cred, diag := store.GetCredentialByCredentialID(ctx, []byte("cred-id-1"))
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, cred)
		assert.Equal(t, userID, cred.UserID)
	})

	t.Run("GetCredentialByCredentialID_NotFound", func(t *testing.T) {
		cred, diag := store.GetCredentialByCredentialID(ctx, []byte("non-existent"))
		assert.Nil(t, cred)
		// Should return nil, nil for not found
		if diag != nil {
			assert.False(t, diag.HasErrors())
		}
	})
}
