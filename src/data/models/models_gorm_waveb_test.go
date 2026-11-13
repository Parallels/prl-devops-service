package models_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/require"
)

func marshalForGolden(t *testing.T, v interface{}) []byte {
	t.Helper()
	out, err := json.Marshal(v)
	require.NoError(t, err)
	return out
}

func requireEqualStringsWithGolden(t *testing.T, want, got, file string) {
	t.Helper()
	require.Equal(t, strings.TrimSpace(want), got, "golden byte drift: %s", file)
}

// Wave-B golden fixtures (PBI #419 task 03, infra domain). Capture-first:
// this file + testdata/golden/* land BEFORE the gorm tags (separate commit);
// the tag commit must not touch testdata/. Same byte-equal contract as wave A.

func waveBTargets() map[string]func() interface{} {
	return map[string]func() interface{}{
		"catalog_manager.json": func() interface{} {
			return models.CatalogManager{
				ID: "cm-1", Name: "official", URL: "https://catalog.example",
				Internal: true, Active: true, AuthenticationMethod: "basic",
				Username: "svc", Password: "pw", ApiKey: "ak", Global: true,
				RequiredClaims: []string{"catalog.read", "catalog.download"},
				OwnerID:        "u-1", CreatedAt: "2024-01-01 00:00:01", UpdatedAt: "2024-01-02 00:00:02",
			}
		},
		"catalog_manifest.json": func() interface{} {
			return models.CatalogManifest{
				ID: "m-1", CatalogId: "ubuntu-2404", Name: "Ubuntu 24.04", Version: "1.2.3",
				Architecture: "arm64", Description: "test image", Path: "/p", PackFile: "/pp",
				MetadataFile: "/mp", Type: "vm", Size: 1024,
				Provider: &models.CatalogManifestProvider{
					Type: "s3", Host: "s3.example", Port: "443",
					Username: "u", Password: "p", ApiKey: "k",
					Meta: map[string]string{"bucket": "b2", "region": "eu-1"},
				},
				RequiredRoles:  []string{"catalog_admin"},
				RequiredClaims: []string{"catalog.read"},
				Tags:           []string{"linux", "ubuntu"},
				CreatedAt:      "2024-01-01 00:00:00", UpdatedAt: "2024-01-02 00:00:00",
				LastDownloadedAt: "2024-02-01 00:00:00", LastDownloadedUser: "alice",
				IsCompressed: true, PackRelativePath: "rel/p.pak", DownloadCount: 7,
				VirtualMachineContents: []models.CatalogManifestContentItem{
					{IsDir: false, Name: "disk.hdd", Path: "ubuntu/disk.hdd", Checksum: "abc123", Size: 999, CreatedAt: "2024-01-01 00:00:00", UpdatedAt: "2024-01-01 00:00:00"},
				},
				PackContents: []models.CatalogManifestContentItem{
					{IsDir: true, Name: "packdir", Path: "packdir", Size: 0},
				},
				PackSize:                2048,
				MinimumSpecRequirements: &models.MinimumSpecRequirement{Cpu: 2, Memory: 4096, Disk: 20480},
				Tainted:                 true, TaintedBy: "u-2", UnTaintedBy: "u-3", TaintedAt: "2024-03-01 00:00:00",
				Revoked: true, RevokedBy: "u-4", RevokedAt: "2024-04-01 00:00:00",
			}
		},
		"packer_template.json": func() interface{} {
			return models.PackerTemplate{
				ID: "pt-1", Name: "ubuntu-arm64", Owner: "u-1", Hostname: "h1",
				Description: "tpl", PackerFolder: "/packer",
				Variables: map[string]string{"vm_name": "tpl-vm", "iso": "ubuntu.iso"},
				Addons:    []string{"drm", "tools"},
				Specs:     map[string]string{"cpu": "2", "disk": "20480", "memory": "4096"},
				Defaults:  map[string]string{"user": "packer"},
				Internal:  true, UpdatedAt: "2024-01-03 00:00:00", CreatedAt: "2024-01-01 00:00:00",
				RequiredRoles: []string{"packer_admin"}, RequiredClaims: []string{"packer.build"},
			}
		},
		"configuration.json": func() interface{} {
			return models.Configuration{ID: "global"}
		},
		"user_config.json": func() interface{} {
			return models.UserConfig{
				ID: "uc-1", UserID: "u-1", Slug: "default_license", Name: "Default License",
				Type: models.UserConfigValueTypeString, Value: "pro",
				CreatedAt: "2024-01-01 00:00:00", UpdatedAt: "2024-01-02 00:00:00",
				DbRecord: &models.DbRecord{IsLocked: false, LockedBy: "", LockedAt: ""},
			}
		},
		"job.json": func() interface{} {
			return models.Job{
				ID: "j-1", Owner: "u-1", OwnerName: "Alice", OwnerEmail: "a@e.x",
				State: "running", Message: "in flight", Progress: 42,
				JobType: "vm.clone", JobOperation: "clone", IsOrchestratorJob: true,
				Steps: []models.JobStep{
					{Name: "validate", Weight: 1, State: "done", Message: "ok", Unit: "%"},
					{Name: "clone", Weight: 2, State: "running", HasPercentage: true, CurrentPercentage: 50, Value: 5, Total: 10, ETA: "3s", Filename: "f.hdd"},
				},
				Result: "done-ish", ResultRecordId: "vm-77", ResultRecordName: "clone-vm",
				ResultRecordType: "vm", ResultRecordLinkId: "h-1",
				Error: "", CreatedAt: "2024-01-01 00:00:00", UpdatedAt: "2024-01-01 00:00:05",
				DbRecord: &models.DbRecord{IsLocked: true, LockedBy: "worker-1", LockedAt: "2024-01-01 00:00:01"},
			}
		},
		"reverse_proxy.json": func() interface{} {
			return models.ReverseProxy{ID: "rp-1", HostID: "h-1", Enabled: true, Host: "0.0.0.0", Port: "443"}
		},
		"reverse_proxy_host.json": func() interface{} {
			rx, _ := regexp.Compile(`^/api/.*$`) // statically valid pattern
			return models.ReverseProxyHost{
				ID: "rph-1", Name: "app", HostID: "h-1", Host: "app.example", Port: "8443",
				Tls:  &models.ReverseProxyHostTls{Enabled: true, Cert: "cert-pem", Key: "key-pem"},
				Cors: &models.ReverseProxyHostCors{Enabled: true, AllowedOrigins: []string{"https://o1"}, AllowedMethods: []string{"GET", "POST"}, AllowedHeaders: []string{"X-Test"}},
				HttpRoutes: []*models.ReverseProxyHostHttpRoute{
					{ID: "r1", Order: 1, Path: "/api", TargetVmId: "vm-1", TargetHost: "10.0.0.5", TargetPort: "8080", Schema: "https", Pattern: "^/api/.*$", RegexpPattern: rx,
						RequestHeaders: map[string]string{"X-Fwd": "1"}, ResponseHeaders: map[string]string{"X-Srv": "gw"}},
					{ID: "r2", Order: 2, Path: "/static", TargetHost: "10.0.0.6", TargetPort: "80"},
				},
				TcpRoute: &models.ReverseProxyHostTcpRoute{ID: "t1", TargetPort: "2222", TargetHost: "10.0.0.7", TargetVmId: "vm-2"},
			}
		},
		"enrollment_token.json": func() interface{} {
			return models.OrchestratorEnrollmentToken{
				ID: "et-1", Token: "tok-abc", HostName: "orch-1", Used: true,
				ExpiresAt: "2025-01-01 00:00:00", CreatedAt: "2024-01-01 00:00:00",
				DbRecord: &models.DbRecord{IsLocked: false, LockedBy: "u-1", LockedAt: "2024-01-01 00:00:00"},
			}
		},
	}
}

// TestCaptureGoldenWaveB — same env-gated capture protocol as wave A.
func TestCaptureGoldenWaveB(t *testing.T) {
	if os.Getenv("GOLDEN_CAPTURE") != "1" {
		t.Skip("set GOLDEN_CAPTURE=1 to (re)generate fixtures")
	}
	require.NoError(t, os.MkdirAll("testdata/golden", 0o755))
	for name, build := range waveBTargets() {
		got := marshalForGolden(t, build())
		require.NoError(t, os.WriteFile(filepath.Join("testdata", "golden", name), got, 0o644))
	}
}

func assertGoldenB(t *testing.T, file string, v interface{}) {
	t.Helper()
	got := marshalForGolden(t, v)
	want, err := os.ReadFile(filepath.Join("testdata", "golden", file))
	require.NoError(t, err, "golden fixture missing — capture protocol broken")
	requireEqualStringsWithGolden(t, string(want), string(got), file)
}

func TestCatalogManager_JSONGolden(t *testing.T) {
	assertGoldenB(t, "catalog_manager.json", waveBTargets()["catalog_manager.json"]())
}
func TestCatalogManifest_JSONGolden(t *testing.T) {
	assertGoldenB(t, "catalog_manifest.json", waveBTargets()["catalog_manifest.json"]())
}
func TestPackerTemplate_JSONGolden(t *testing.T) {
	assertGoldenB(t, "packer_template.json", waveBTargets()["packer_template.json"]())
}
func TestConfiguration_JSONGolden(t *testing.T) {
	assertGoldenB(t, "configuration.json", waveBTargets()["configuration.json"]())
}
func TestUserConfig_JSONGolden(t *testing.T) {
	assertGoldenB(t, "user_config.json", waveBTargets()["user_config.json"]())
}
func TestJob_JSONGolden(t *testing.T) {
	assertGoldenB(t, "job.json", waveBTargets()["job.json"]())
}
func TestReverseProxy_JSONGolden(t *testing.T) {
	assertGoldenB(t, "reverse_proxy.json", waveBTargets()["reverse_proxy.json"]())
}
func TestReverseProxyHost_JSONGolden(t *testing.T) {
	assertGoldenB(t, "reverse_proxy_host.json", waveBTargets()["reverse_proxy_host.json"]())
}
func TestEnrollmentToken_JSONGolden(t *testing.T) {
	assertGoldenB(t, "enrollment_token.json", waveBTargets()["enrollment_token.json"]())
}
