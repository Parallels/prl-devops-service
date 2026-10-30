package system

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestRefreshDiskInfoReportsFreeSpace(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)

	info := &models.SystemHardwareInfo{}
	svc.refreshDiskInfo(ctx, info)

	assert.Greater(t, info.DiskSize, float64(0))
	assert.Greater(t, info.FreeDiskSize, float64(0))
	assert.LessOrEqual(t, info.FreeDiskSize, info.DiskSize)
}

func TestGetHardwareInfoRefreshesDiskAfterTTL(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)

	// Seed a stale cache entry with bogus disk values.
	svc.cache.HardwareInfo = &models.SystemHardwareInfo{DiskSize: 1, FreeDiskSize: 1, LogicalCpuCount: 42}
	svc.cache.HardwareInfoDiskUpdatedAt = time.Now()

	fresh, err := svc.GetHardwareInfo(ctx)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), fresh.DiskSize, "disk must be served from cache within the TTL")

	svc.cache.HardwareInfoDiskUpdatedAt = time.Now().Add(-2 * diskInfoTTL)
	stale, err := svc.GetHardwareInfo(ctx)
	assert.NoError(t, err)
	assert.Greater(t, stale.DiskSize, float64(1), "disk must be re-read after the TTL")
	assert.Equal(t, 42, stale.LogicalCpuCount, "static fields stay cached")
}

// NOTE: GetOsVersionForLinux and GetOSName use the physical host's OS.
// Asserting that these don't panic on the current hardware is useful.

func TestGetOSNameDoesNotPanic(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)

	name := svc.GetOSName()
	assert.NotEmpty(t, name)
}
