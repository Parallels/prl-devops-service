package parity

// #425 task 03 — infra wave: VM snapshots. Legacy originals
// (src/data/vm_snapshots_test.go) poked a bare &JsonDatabase{connected:
// true} struct — a legacy-private posture with no engine equivalent;
// converted to the PUBLIC interface contract both engines honour:
// SetListVMSnapshotsByVMId (ctx-free, replace-per-VM semantics) +
// GetListVMSnapshotsByVMId (unknown VM ⇒ EMPTY list, no error —
// vm_snapshots.go:37), and the host-level SetHostVMSnapshots map-upsert
// idempotence + GetHostVMSnapshots unknown-host sentinel.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParitySnapshotsInfraWave(t *testing.T) {
	testutil.RunOnEngines(t, snapshotScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
}

func snap(names ...string) []models.VMSnapshot {
	out := make([]models.VMSnapshot, 0, len(names))
	for _, n := range names {
		out = append(out, models.VMSnapshot{Name: n, ID: n})
	}
	return out
}

func snapshotScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "snapshots/vm_list_replace_per_vm", Run: func(t *testing.T, _ ctxT, db dbT) {
			// vm_snapshots_test.go legs 1-2 (order-faithful read-back), plus
			// the replace-per-VM semantics the legacy set-twice implied.
			require.NoError(t, db.SetListVMSnapshotsByVMId("infra-vm-1", models.VMSnapshots{
				VMId: "infra-vm-1", VMSnapshot: snap("Snapshot1", "Snapshot2"),
			}))
			first, err := db.GetListVMSnapshotsByVMId("infra-vm-1")
			require.NoError(t, err)
			require.Len(t, first, 2)
			assert.Equal(t, "Snapshot1", first[0].Name)
			assert.Equal(t, "Snapshot2", first[1].Name)
			// second write REPLACES (vm_snapshots.go:14-17 index-overwrite)
			require.NoError(t, db.SetListVMSnapshotsByVMId("infra-vm-1", models.VMSnapshots{
				VMId: "infra-vm-1", VMSnapshot: snap("Snapshot3"),
			}))
			replaced, err := db.GetListVMSnapshotsByVMId("infra-vm-1")
			require.NoError(t, err)
			require.Len(t, replaced, 1, "replace semantics, not append")
			assert.Equal(t, "Snapshot3", replaced[0].Name)
			// second VM is independent
			require.NoError(t, db.SetListVMSnapshotsByVMId("infra-vm-2", models.VMSnapshots{
				VMId: "infra-vm-2", VMSnapshot: snap("Other"),
			}))
			untouched, err := db.GetListVMSnapshotsByVMId("infra-vm-1")
			require.NoError(t, err)
			assert.Len(t, untouched, 1)
			// unknown VM ⇒ empty list, NO error (vm_snapshots.go:36-37)
			empty, err := db.GetListVMSnapshotsByVMId("infra-vm-none")
			require.NoError(t, err)
			assert.Empty(t, empty)
		}},
		{Name: "snapshots/host_upsert_idempotent", Run: func(t *testing.T, _ ctxT, db dbT) {
			// PARITY FINDING (pinned, engine-specific): the SQL path ENRICHES
			// every stored snapshot with HostID+VMID on GetHostVMSnapshots;
			// the legacy JSON path preserves the caller's DTO unmodified.
			// #421's suite normalised these fields away — this wave pins the
			// divergence explicitly.
			ctx := authCtxPlain()
			hostSnap := func(hostID, vmID string, names ...string) []models.VMSnapshot {
				out := snap(names...)
				if _, isLegacy := db.(*data.JsonDatabase); !isLegacy {
					for i := range out {
						out[i].HostID = hostID
						out[i].VMID = vmID
					}
				}
				return out
			}
			same := models.VMSnapshots{VMId: "host-vm-a", VMSnapshot: snap("s1", "s2")}
			require.NoError(t, db.SetHostVMSnapshots(ctx, "infra-host-snap", same))
			require.NoError(t, db.SetHostVMSnapshots(ctx, "infra-host-snap", same), "set twice succeeds")
			rec, err := db.GetHostVMSnapshots(ctx, "infra-host-snap")
			require.NoError(t, err)
			require.Len(t, rec.VMSnapshots, 1, "idempotent map upsert, single entry")
			assert.Equal(t, hostSnap("infra-host-snap", "host-vm-a", "s1", "s2"), rec.VMSnapshots["host-vm-a"], "deep-equal round-trip")
			require.NoError(t, db.SetHostVMSnapshots(ctx, "infra-host-snap", models.VMSnapshots{
				VMId: "host-vm-b", VMSnapshot: snap("s9"),
			}))
			grew, err := db.GetHostVMSnapshots(ctx, "infra-host-snap")
			require.NoError(t, err)
			require.Len(t, grew.VMSnapshots, 2, "per-VM map grows, no clobber")
			assert.Equal(t, hostSnap("infra-host-snap", "host-vm-a", "s1", "s2"), grew.VMSnapshots["host-vm-a"])
		}},
		{Name: "snapshots/host_unknown_sentinel", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			_, err := db.GetHostVMSnapshots(ctx, "no-such-host")
			require.Error(t, err)
			assert.Equal(t, datainterfaces.ErrOrchestratorHostsVMSnapshotsNotFound, err)
		}},
	}
}
