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

func TestClaimDataStore(t *testing.T) {
	logger := logging.Get()
	logger.Info("Starting claim store tests")
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	store := &stores.ClaimDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	ctx := basecontext.NewBaseContext()

	t.Run("CreateClaim", func(t *testing.T) {
		claim := &models.Claim{
			Name:        "test.read",
			Description: "Test read permission",
			Group:       "test",
			Resource:    "resource",
			Action:      "read",
		}

		createdClaim, diag := store.CreateClaim(*ctx, claim)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, createdClaim)
		assert.NotEmpty(t, createdClaim.ID)
		assert.Equal(t, "test.read", createdClaim.Name)
		assert.Equal(t, "read", createdClaim.Action)
	})

	t.Run("GetClaims", func(t *testing.T) {
		claims, diag := store.GetClaims(*ctx)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, claims)
		assert.True(t, len(claims) >= 1)
	})

	t.Run("CreateClaim_WithGroup", func(t *testing.T) {
		claim := &models.Claim{
			Name:        "admin.write",
			Description: "Admin write permission",
			Group:       "admin",
			Resource:    "all",
			Action:      "write",
		}

		createdClaim, diag := store.CreateClaim(*ctx, claim)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, createdClaim)
		assert.Equal(t, "admin", createdClaim.Group)
		assert.Equal(t, "all", createdClaim.Resource)
	})

	t.Run("CreateClaim_Internal", func(t *testing.T) {
		claim := &models.Claim{
			Name:        "system.internal",
			Description: "Internal system claim",
			Internal:    true,
		}

		createdClaim, diag := store.CreateClaim(*ctx, claim)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, createdClaim)
		assert.True(t, createdClaim.Internal)
	})
}
