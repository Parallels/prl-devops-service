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
)

func TestWebAuthnDataStore(t *testing.T) {
	logger := logging.Get()
	logger.Info("Starting webauthn store tests")
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	store := &stores.WebAuthnDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	ctx := basecontext.NewBaseContext()
	userID := "test-user-id"

	t.Run("SaveCredential", func(t *testing.T) {
		cred := &models.WebAuthnCredential{
			UserID:          userID,
			CredentialID:    []byte("cred-id-1"),
			PublicKey:       []byte("public-key"),
			AttestationType: "none",
			SignCount:       0,
		}

		diag := store.SaveCredential(*ctx, cred)
		assert.False(t, diag.HasErrors())
	})

	t.Run("GetCredentialsByUser", func(t *testing.T) {
		creds, diag := store.GetCredentialsByUser(*ctx, userID)
		assert.False(t, diag.HasErrors())
		assert.NotEmpty(t, creds)
		assert.NotEmpty(t, creds[0].ID) // ID should be auto-generated
	})

	t.Run("GetCredentialByCredentialID", func(t *testing.T) {
		cred, diag := store.GetCredentialByCredentialID(*ctx, []byte("cred-id-1"))
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, cred)
		assert.Equal(t, userID, cred.UserID)
	})

	t.Run("GetCredentialByCredentialID_NotFound", func(t *testing.T) {
		cred, diag := store.GetCredentialByCredentialID(*ctx, []byte("non-existent"))
		assert.Nil(t, cred)
		// Should return nil, nil for not found
		if diag != nil {
			assert.False(t, diag.HasErrors())
		}
	})
}
