package system

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestExternalIPUsesBoundedHTTPS(t *testing.T) {
	for _, fallback := range []bool{false, true} {
		name := "curl"
		if fallback {
			name = "wget-fallback"
		}
		t.Run(name, func(t *testing.T) {
			binDir := t.TempDir()
			curlScript := "#!/bin/sh\n[ \"$*\" = '--fail --silent --show-error --connect-timeout 3 --max-time 5 https://ifconfig.me/ip' ] || exit 2\nprintf '203.0.113.7\\n'\n"
			if fallback {
				curlScript = "#!/bin/sh\nexit 1\n"
			}
			wgetScript := "#!/bin/sh\n[ \"$*\" = '-qO- --timeout=5 --tries=1 https://ifconfig.me/ip' ] || exit 2\nprintf '203.0.113.7\\n'\n"
			for executable, script := range map[string]string{"curl": curlScript, "wget": wgetScript} {
				if err := os.WriteFile(filepath.Join(binDir, executable), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", binDir)
			ctx := basecontext.NewBaseContext()
			svc := New(ctx)
			externalIP, err := svc.getUniversalExternalIp(ctx)
			assert.NoError(t, err)
			assert.Equal(t, "203.0.113.7", externalIP)
			assert.Equal(t, externalIP, svc.cache.ExternalIpAddress)
		})
	}
}

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
