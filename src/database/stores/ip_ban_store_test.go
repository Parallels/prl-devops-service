package stores_test

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/stores"
	"github.com/Parallels/prl-devops-service/database/stores/testhelpers"
	logging "github.com/cjlapao/common-go-logger"
	"github.com/stretchr/testify/assert"
)

func TestIpBanDataStore(t *testing.T) {
	logger := logging.Get()
	logger.Info("Starting IP ban store tests")
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	store := &stores.IpBanDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	ctx := basecontext.NewBaseContext()

	t.Run("CreateIpBan", func(t *testing.T) {
		ipBan := &models.IpBan{
			IP:       "192.168.1.100",
			Reason:   "Suspicious activity",
			BanLevel: "global",
			Enabled:  true,
			BannedAt: time.Now(),
		}

		created, diag := store.CreateIpBan(*ctx, ipBan)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, created)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "192.168.1.100", created.IP)
		assert.True(t, created.Enabled)
	})

	t.Run("GetIpBan", func(t *testing.T) {
		ipBan := &models.IpBan{
			IP:       "10.0.0.50",
			Reason:   "Brute force attack",
			BanLevel: "global",
			Enabled:  true,
			BannedAt: time.Now(),
		}
		store.CreateIpBan(*ctx, ipBan)

		retrieved, diag := store.GetIpBan(*ctx, "10.0.0.50")
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, retrieved)
		assert.Equal(t, "10.0.0.50", retrieved.IP)
		assert.Equal(t, "Brute force attack", retrieved.Reason)
	})

	t.Run("GetIpBan_NotFound", func(t *testing.T) {
		retrieved, diag := store.GetIpBan(*ctx, "1.2.3.4")
		assert.Nil(t, retrieved)
		if diag != nil {
			assert.False(t, diag.HasErrors())
		}
	})

	t.Run("GetActiveBans", func(t *testing.T) {
		ips := []string{"203.0.113.1", "203.0.113.2", "203.0.113.3"}
		for _, ip := range ips {
			ipBan := &models.IpBan{
				IP:       ip,
				Reason:   "Test ban",
				BanLevel: "global",
				Enabled:  true,
				BannedAt: time.Now(),
			}
			store.CreateIpBan(*ctx, ipBan)
		}

		activeBans, diag := store.GetActiveBans(*ctx)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, activeBans)
		assert.True(t, len(activeBans) >= 3)
	})

	t.Run("RevokeIpBan", func(t *testing.T) {
		ipBan := &models.IpBan{
			IP:       "172.16.0.1",
			Reason:   "Temporary ban",
			BanLevel: "global",
			Enabled:  true,
			BannedAt: time.Now(),
		}
		created, diag := store.CreateIpBan(*ctx, ipBan)
		assert.False(t, diag.HasErrors())

		diag = store.RevokeIpBan(*ctx, created.IP)
		assert.False(t, diag.HasErrors())

		// GetIpBan only returns active bans, so we need to query directly
		var revoked models.IpBan
		result := db.Where("ip = ?", created.IP).First(&revoked)
		assert.NoError(t, result.Error)
		assert.False(t, revoked.Enabled)
	})

	t.Run("CreateIpBan_WithExpiry", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		ipBan := &models.IpBan{
			IP:        "198.51.100.1",
			Reason:    "Temporary block",
			BanLevel:  "global",
			Enabled:   true,
			BannedAt:  time.Now(),
			ExpiresAt: &expiresAt,
		}

		created, diag := store.CreateIpBan(*ctx, ipBan)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, created.ExpiresAt)
		assert.True(t, created.ExpiresAt.After(time.Now()))
	})

	t.Run("CreateIpBan_IPv6", func(t *testing.T) {
		ipBan := &models.IpBan{
			IP:       "2001:db8::1",
			Reason:   "IPv6 test",
			BanLevel: "global",
			Enabled:  true,
			BannedAt: time.Now(),
		}

		created, diag := store.CreateIpBan(*ctx, ipBan)
		assert.False(t, diag.HasErrors())
		assert.Equal(t, "2001:db8::1", created.IP)
	})

	t.Run("UpdateIpBan", func(t *testing.T) {
		ipBan := &models.IpBan{
			IP:       "10.1.1.1",
			Reason:   "Original reason",
			BanLevel: "global",
			Enabled:  true,
			BannedAt: time.Now(),
		}
		created, diag := store.CreateIpBan(*ctx, ipBan)
		assert.False(t, diag.HasErrors())

		created.Reason = "Updated reason"
		result := db.Save(created)
		assert.NoError(t, result.Error)

		retrieved, diag := store.GetIpBan(*ctx, created.IP)
		assert.False(t, diag.HasErrors())
		assert.Equal(t, "Updated reason", retrieved.Reason)
	})

	t.Run("DeleteIpBan", func(t *testing.T) {
		ipBan := &models.IpBan{
			IP:       "10.2.2.2",
			Reason:   "Delete test",
			BanLevel: "global",
			Enabled:  true,
			BannedAt: time.Now(),
		}
		created, diag := store.CreateIpBan(*ctx, ipBan)
		assert.False(t, diag.HasErrors())

		result := db.Delete(created)
		assert.NoError(t, result.Error)

		retrieved, diag := store.GetIpBan(*ctx, created.IP)
		assert.Nil(t, retrieved)
	})
}
