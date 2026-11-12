package parity

// Engine-parity scenarios — orchestrator domain (PBI #425 task 04).
// Transcribed from data/orchestrator_test.go (Tier 1, byte-identical):
// the host CRUD family, active filter (Enabled-only pin), update
// no-diff / details guards, VM seeding THROUGH THE ENGINE API
// (Upsert/Replace full-sync — spec §1, #421 task 14 §2), delete cascade,
// resource aggregation GOLDEN values (absolute numbers on both engines,
// on top of #421's cross-engine equality layer), nested reverse-proxy
// views and the heartbeat (UpdateOrchestratorHostTimestamp) result-
// equality family. Concurrency/NotConnected legs stay with the #421
// suites (D-3: result equality only here — ledger). Aggregation golden
// values are computed once from the fixture and asserted verbatim on
// BOTH engines (legacy golden pattern from TestGetOrchestrator*Resources).

import (
	"sort"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParityOrchestratorInfraWave(t *testing.T) {
	testutil.RunOnEngines(t, append(orchHostScenarios(), orchVMScenarios()...), testutil.LegacyEngine(), testutil.GormEngine())
}

func orchHostScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "orchestrator/host_roundtrip", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot() // orchestrator reads ride the filter layer — root ctx (legacy parity)
			created, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{
				Host: "rt-host.example.com", Description: "orig", State: "healthy",
			})
			require.NoError(t, err)
			byID, err := db.GetOrchestratorHost(ctx, created.ID)
			require.NoError(t, err)
			byName, err := db.GetOrchestratorHost(ctx, "rt-host.example.com") // idOrHost dual lookup
			require.NoError(t, err)
			require.Equal(t, created.ID, byName.ID)
			require.Equal(t, created.ID, byID.ID)
			byID.Description, byID.State = "updated", "unhealthy"
			updated, err := db.UpdateOrchestratorHost(ctx, byID)
			require.NoError(t, err)
			assert.Equal(t, "updated", updated.Description)
			assert.Equal(t, "unhealthy", updated.State)
			require.NoError(t, db.DeleteOrchestratorHost(ctx, "rt-host.example.com")) // delete by NAME
			_, err = db.GetOrchestratorHost(ctx, created.ID)
			assert.Equal(t, data.ErrOrchestratorHostNotFound, err)
		}},
		{Name: "orchestrator/create_guards", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			_, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: ""})
			assert.Equal(t, data.ErrOrchestratorHostEmptyName, err) // 500 quirk preserved
			_, err = db.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: "dup-host.example.com"})
			require.NoError(t, err)
			dup, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: "dup-host.example.com"})
			require.Error(t, err, "duplicate host name rejected")
			assert.Nil(t, dup)
		}},
		{Name: "orchestrator/active_filter", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			_, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: "active-host.example.com"})
			require.NoError(t, err)
			off, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: "inactive-host.example.com"})
			require.NoError(t, err)
			off.Enabled = false
			_, err = db.UpdateOrchestratorHost(ctx, off) // Active == Enabled (state-agnostic pin)
			require.NoError(t, err)
			active, err := db.GetActiveOrchestratorHosts(ctx, "")
			require.NoError(t, err)
			require.Len(t, active, 1)
			assert.Equal(t, "active-host.example.com", active[0].Host)
		}},
		{Name: "orchestrator/update_edges", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: "no-diff-host.example.com"})
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond)
			same, err := db.UpdateOrchestratorHost(ctx, created)
			require.NoError(t, err)
			assert.Equal(t, created.UpdatedAt, same.UpdatedAt, "no-diff update must NOT re-stamp (legacy :207-226)")
			stale := models.OrchestratorHost{ID: "nonexistent-id", Host: "nonexistent.example.com"}
			_, err = db.UpdateOrchestratorHost(ctx, &stale)
			assert.Equal(t, data.ErrOrchestratorHostNotFound, err)
			// details update + rename-into-existing guard (legacy :243-287):
			created.CpuModel, created.Architecture = "Apple M1", "arm64"
			withDetails, err := db.UpdateOrchestratorHostDetails(ctx, created)
			require.NoError(t, err)
			assert.Equal(t, "Apple M1", withDetails.CpuModel)
			other, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{Host: "other-host.example.com"})
			require.NoError(t, err)
			other.Host = "no-diff-host.example.com"
			_, err = db.UpdateOrchestratorHostDetails(ctx, other)
			assert.Error(t, err, "details rename colliding with existing host rejected")
			err = db.DeleteOrchestratorHost(ctx, "nonexistent-host")
			assert.Equal(t, data.ErrOrchestratorHostNotFound, err)
		}},
		{Name: "orchestrator/resources_golden", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			item := func(p, l int64, m, d float64) models.HostResourceItem {
				return models.HostResourceItem{PhysicalCpuCount: p, LogicalCpuCount: l, MemorySize: m, DiskSize: d}
			}
			h1, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{
				Host: "golden-1.example.com", State: "healthy", Enabled: true,
				Resources: &models.HostResources{CpuType: "arm64",
					Total: item(16, 32, 64, 1000), TotalAvailable: item(8, 8, 16, 500),
					TotalInUse: item(4, 4, 8, 0), TotalReserved: item(1, 1, 2, 0),
					SystemReserved: item(1, 1, 1, 0)},
			})
			require.NoError(t, err)
			_, err = db.CreateOrchestratorHost(ctx, models.OrchestratorHost{
				Host: "golden-2.example.com", State: "healthy", Enabled: true,
				Resources: &models.HostResources{CpuType: "arm64",
					Total: item(8, 8, 16, 500), TotalAvailable: item(4, 4, 8, 250),
					TotalInUse: item(2, 2, 4, 0), TotalReserved: item(1, 1, 2, 0)},
			})
			require.NoError(t, err)
			arm := func(p, l int64, m, d float64) map[string]models.HostResourceItem {
				return map[string]models.HostResourceItem{"arm64": item(p, l, m, d)}
			}
			// Golden sums of the fixture — legacy absolute-number pattern
			// (TestGetOrchestratorAvailableResources: 12/12/24/750):
			assert.Equal(t, arm(12, 12, 24, 750), db.GetOrchestratorAvailableResources(ctx))
			assert.Equal(t, arm(24, 40, 80, 1500), db.GetOrchestratorTotalResources(ctx))
			assert.Equal(t, arm(6, 6, 12, 0), db.GetOrchestratorInUseResources(ctx))
			assert.Equal(t, arm(2, 2, 4, 0), db.GetOrchestratorReservedResources(ctx))
			assert.Equal(t, arm(1, 1, 1, 0), db.GetOrchestratorSystemReservedResources(ctx))
			hostRes, err := db.GetOrchestratorHostResources(ctx, h1.ID)
			require.NoError(t, err)
			require.NotNil(t, hostRes)
			assert.Equal(t, int64(16), hostRes.Total.PhysicalCpuCount)
			assert.Equal(t, "arm64", hostRes.CpuType)
		}},
		{Name: "orchestrator/rp_views", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			host, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{
				Host: "rp-view-host.example.com",
				ReverseProxyHosts: []*models.ReverseProxyHost{
					{ID: "rp-1", Host: "proxy1.example.com", Port: "443"},
					{ID: "rp-2", Host: "proxy2.example.com", Port: "443"},
				},
				ReverseProxy: &models.ReverseProxy{Enabled: true, Host: "proxy.example.com", Port: "8443"},
			})
			require.NoError(t, err)
			rps, err := db.GetOrchestratorReverseProxyHosts(ctx, host.ID, "")
			require.NoError(t, err)
			require.Len(t, rps, 2)
			rp, err := db.GetOrchestratorReverseProxyHost(ctx, host.ID, "proxy1.example.com:443")
			require.NoError(t, err)
			assert.Equal(t, "proxy1.example.com", rp.Host) // key = "host:port" (legacy :693)
			cfg, err := db.GetOrchestratorReverseProxyConfig(ctx, host.ID)
			require.NoError(t, err)
			require.NotNil(t, cfg)
			assert.True(t, cfg.Enabled)
			assert.Equal(t, "proxy.example.com", cfg.Host)
			_, err = db.GetOrchestratorReverseProxyHost(ctx, host.ID, "nonexistent")
			assert.Equal(t, data.ErrOrchestratorReverseProxyHostNotFound, err)
		}},
	}
}

func orchVMScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "orchestrator/vms_upsert_replace_cascade", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			// Host-level listing filters healthy+enabled (legacy fixture
			// :366-369 uses exactly this posture — PINNED, both engines):
			host, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{
				Host: "vm-wave-host.example.com", State: "healthy", Enabled: true,
			})
			require.NoError(t, err)
			for _, vm := range []models.VirtualMachine{{ID: "vm-1", Name: "one"}, {ID: "vm-2", Name: "two"}, {ID: "vm-3", Name: "three"}} {
				require.NoError(t, db.UpsertOrchestratorHostVM(ctx, host.ID, vm))
			}
			vms, err := db.GetOrchestratorHostVirtualMachines(ctx, host.ID, "")
			require.NoError(t, err)
			assert.Equal(t, []string{"vm-1", "vm-2", "vm-3"}, sortedVMIDs(vms))
			require.NoError(t, db.UpdateOrchestratorHostVMState(ctx, host.ID, "vm-2", "running"))
			require.NoError(t, db.UpdateOrchestratorHostVMUptime(ctx, host.ID, "vm-2", "3600"))
			vm2, err := db.GetOrchestratorHostVirtualMachine(ctx, host.ID, "vm-2")
			require.NoError(t, err)
			assert.Equal(t, "running", vm2.State)
			_, err = db.GetOrchestratorHostVirtualMachine(ctx, host.ID, "no-vm")
			assert.Equal(t, data.ErrOrchestratorHostVirtualMachineNotFound, err)
			// full-sync: Replace swaps the ENTIRE set (post-state by ID set):
			require.NoError(t, db.ReplaceOrchestratorHostVMs(ctx, host.ID,
				[]models.VirtualMachine{{ID: "vm-2", Name: "two"}, {ID: "vm-4", Name: "four"}}))
			after, err := db.GetOrchestratorHostVirtualMachines(ctx, host.ID, "")
			require.NoError(t, err)
			assert.Equal(t, []string{"vm-2", "vm-4"}, sortedVMIDs(after))
			// host-level list sees the same rows (cascade source of truth):
			all, err := db.GetOrchestratorVirtualMachines(ctx, "")
			require.NoError(t, err)
			assert.Equal(t, []string{"vm-2", "vm-4"}, sortedVMIDs(all))
			// delete-host cascade (legacy :289-308 + spec §1):
			require.NoError(t, db.DeleteOrchestratorHost(ctx, host.ID))
			gone, err := db.GetOrchestratorVirtualMachines(ctx, "")
			require.NoError(t, err)
			assert.Empty(t, gone, "host rows cascade: VM rows gone")
		}},
		{Name: "orchestrator/heartbeat_result_equality", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{
				Host: "hb-host.example.com", State: "healthy", Description: "keep me",
				Architecture: "arm64", Tags: []string{"a"},
			})
			require.NoError(t, err)
			const hb = "2025-11-28T12:30:23.646623Z" // caller stamp stored VERBATIM (legacy :836)
			require.NoError(t, db.UpdateOrchestratorHostTimestamp(ctx, created.ID, hb, ""))
			after, err := db.GetOrchestratorHost(ctx, created.ID)
			require.NoError(t, err)
			assert.Equal(t, hb, after.UpdatedAt)
			assert.Equal(t, "healthy", after.State, "empty state must not touch state")
			assert.Equal(t, "keep me", after.Description)
			assert.Equal(t, []string{"a"}, after.Tags, "preserves other fields")
			require.NoError(t, db.UpdateOrchestratorHostTimestamp(ctx, created.ID, hb, "unhealthy"))
			after, err = db.GetOrchestratorHost(ctx, created.ID)
			require.NoError(t, err)
			assert.Equal(t, "unhealthy", after.State)
			// healthy transition CLEARS the unhealthy bookkeeping fields:
			bad, err := db.CreateOrchestratorHost(ctx, models.OrchestratorHost{
				Host: "hb-bad.example.com", State: "unhealthy",
				LastUnhealthy: "2025-11-28T10:00:00Z", LastUnhealthyErrorMessage: "Connection timeout",
			})
			require.NoError(t, err)
			require.NoError(t, db.UpdateOrchestratorHostTimestamp(ctx, bad.ID, hb, "healthy"))
			recovered, err := db.GetOrchestratorHost(ctx, bad.ID)
			require.NoError(t, err)
			assert.Empty(t, recovered.LastUnhealthy)
			assert.Empty(t, recovered.LastUnhealthyErrorMessage)
			assert.Equal(t, data.ErrOrchestratorHostNotFound,
				db.UpdateOrchestratorHostTimestamp(ctx, "nonexistent-id", hb, "healthy"))
			assert.Equal(t, data.ErrOrchestratorHostEmptyIdOrHost,
				db.UpdateOrchestratorHostTimestamp(ctx, "", hb, "healthy"))
		}},
	}
}

func sortedVMIDs(vms []models.VirtualMachine) []string {
	ids := make([]string, 0, len(vms))
	for _, v := range vms {
		ids = append(ids, v.ID)
	}
	sort.Strings(ids)
	return ids
}
