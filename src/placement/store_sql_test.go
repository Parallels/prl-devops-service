package placement

// Store behaviour tests. The point of every one of these is the window the
// feature exists to close: the host snapshot is fixed, so any admitted burst
// that exceeds it is a real overcommit, not a flake.

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTestStore(t *testing.T) (*sqlStore, *gorm.DB) {
	t.Helper()
	db := testutil.OpenMemoryDB(t)
	testutil.MigrateSchema(t, db)

	return newSQLStore(db), db
}

// seedHost writes the host row the way a hardware refresh would: the numbers
// under test are the cached snapshot, which placement must never trust further
// than one transaction.
func seedHost(t *testing.T, db *gorm.DB, id string, cpu int64, memoryMiB float64) {
	t.Helper()
	host := &data_models.OrchestratorHost{
		ID: id,
		Resources: &data_models.HostResources{
			TotalAvailable: data_models.HostResourceItem{
				LogicalCpuCount: cpu,
				MemorySize:      memoryMiB,
			},
			Total: data_models.HostResourceItem{TotalVms: 0},
		},
	}
	require.NoError(t, db.Create(host).Error)
}

// seedHostWithResources is the variant for the dimensions beyond cpu/memory
// (MAX_VMS_PER_HOST and its mode). It writes through Create rather than an
// Update so the resources JSON serializer runs.
func seedHostWithResources(t *testing.T, db *gorm.DB, id string, resources *data_models.HostResources) {
	t.Helper()
	require.NoError(t, db.Create(&data_models.OrchestratorHost{ID: id, Resources: resources}).Error)
}

func reserve(jobID, hostID, name, owner string, cpu int64, memory float64) Reservation {
	return Reservation{
		JobID: jobID, HostID: hostID, VmName: name, Owner: owner,
		CPU: cpu, Memory: memory, VMCount: 1,
	}
}

// TestAdmitExactFitAndOneOverRefused is the boundary: capacity is not
// "greater than", it is "greater than or equal", and one CPU over is one too
// many.
func TestAdmitExactFitAndOneOverRefused(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 2048)

	admission, err := store.Admit(ctx, "host-1", reserve("job-a", "host-1", "vm-a", "root", 10, 2048), CapacityView{})
	require.NoError(t, err)
	require.True(t, admission.Admitted, "exact fit must be admitted: %s", admission.Message)

	// The first claim consumed the whole host, so nothing is left for the next.
	admission, err = store.Admit(ctx, "host-1", reserve("job-b", "host-1", "vm-b", "root", 1, 0), CapacityView{})
	require.NoError(t, err)
	assert.False(t, admission.Admitted)
	assert.Equal(t, "cpu", admission.Reason)
	assert.Equal(t, "CPU: required 1 logical CPUs, available 0", admission.Message)
}

// TestAdmitIssueReproMessage is the reported bug as a test: 10 CPUs, two
// concurrent 6-CPU creates. The first wins, the second is refused with the
// exact wording the issue expects.
func TestAdmitIssueReproMessage(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	// The issue's host: 10 logical CPUs, 1 system-reserved, no VMs. The host
	// already publishes that as TotalAvailable = 9 (GetHardwareUsage:
	// total − systemReserved − inUse, main.go:2896-2897).
	seedHost(t, db, "host-1", 9, 8192)

	first, err := store.Admit(ctx, "host-1", reserve("job-a", "host-1", "over-1", "root", 6, 2048), CapacityView{})
	require.NoError(t, err)
	require.True(t, first.Admitted)

	second, err := store.Admit(ctx, "host-1", reserve("job-b", "host-1", "over-2", "root", 6, 2048), CapacityView{})
	require.NoError(t, err)
	assert.False(t, second.Admitted)
	assert.Equal(t, "cpu", second.Reason)
	assert.Equal(t, "CPU: required 6 logical CPUs, available 3", second.Message,
		"the issue's expected message; total_available must never go negative for an accepted create")
}

// TestAdmitChargesConcurrentlyNotExcludes pins plan D3: a reservation is a
// charge, not a host lock. Two small creates keep landing on the same host and
// only the arithmetic moves the third one.
func TestAdmitChargesConcurrentlyNotExcludes(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 8192)
	seedHost(t, db, "host-2", 10, 8192)

	for i, id := range []string{"job-1", "job-2"} {
		admission, err := store.Admit(ctx, "host-1", reserve(id, "host-1", fmt.Sprintf("vm-%d", i), "root", 2, 1024), CapacityView{})
		require.NoError(t, err)
		require.True(t, admission.Admitted, "both 2-CPU creates belong on the same host")
	}

	pending, err := store.PendingByHost(ctx, "host-1")
	require.NoError(t, err)
	assert.Equal(t, int64(4), pending.CPU, "the host must carry both charges")
	assert.Equal(t, int64(2), pending.VMCount)

	// 9 does not fit the 6 left on host-1, so placement is free to try host-2.
	big, err := store.Admit(ctx, "host-1", reserve("job-3", "host-1", "vm-3", "root", 9, 1024), CapacityView{})
	require.NoError(t, err)
	assert.False(t, big.Admitted)
	assert.Equal(t, int64(6), big.Available.CPU)

	elsewhere, err := store.Admit(ctx, "host-2", reserve("job-3", "host-2", "vm-3", "root", 9, 1024), CapacityView{})
	require.NoError(t, err)
	assert.True(t, elsewhere.Admitted, "a refusal on one host must not follow the job to the next")
}

// TestAdmitBurstNeverOverbooks is the race that produced the bug, under -race:
// twenty 2-CPU requests against ten CPUs admit exactly five.
func TestAdmitBurstNeverOverbooks(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 8192)

	const attempts = 20
	var admitted atomic.Int64
	var sum atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			admission, err := store.Admit(ctx, "host-1",
				reserve(fmt.Sprintf("job-%02d", i), "host-1", fmt.Sprintf("vm-%02d", i), "root", 2, 0), CapacityView{})
			if err == nil && admission.Admitted {
				admitted.Add(1)
				sum.Add(2)
			}
		}(i)
	}
	wg.Wait()

	assert.Equal(t, int64(5), admitted.Load(), "ten CPUs admit five 2-CPU creates and no more")
	assert.Equal(t, int64(10), sum.Load())

	pending, err := store.PendingByHost(ctx, "host-1")
	require.NoError(t, err)
	assert.Equal(t, int64(10), pending.CPU)
}

// TestAdmitHostReportedPendingIsNotDoubleCounted pins plan D7: once the host
// publishes its own pending, the max() rule charges the overlapping create
// once. Summing would reject creates that genuinely fit.
func TestAdmitHostReportedPendingIsNotDoubleCounted(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 8192)

	_, err := store.Admit(ctx, "host-1", reserve("job-a", "host-1", "vm-a", "root", 4, 0), CapacityView{})
	require.NoError(t, err)

	// The host now reports the same create as pending: 4, not 8.
	admission, err := store.Admit(ctx, "host-1", reserve("job-b", "host-1", "vm-b", "root", 7, 0),
		CapacityView{CPU: 4, VMCount: 1})
	require.NoError(t, err)
	assert.False(t, admission.Admitted)
	assert.Equal(t, int64(6), admission.Available.CPU, "10 - max(4 ledger, 4 host) = 6")

	admission, err = store.Admit(ctx, "host-1", reserve("job-c", "host-1", "vm-c", "root", 6, 0),
		CapacityView{CPU: 4, VMCount: 1})
	require.NoError(t, err)
	assert.True(t, admission.Admitted, "double charging would have refused this one")
}

// TestAdmitMovesChargeWhenJobReplacesHosts keeps "one charge per job"
// structural: a create that hops hosts updates its row instead of paying twice.
func TestAdmitMovesChargeWhenJobReplacesHosts(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 8192)
	seedHost(t, db, "host-2", 10, 8192)

	_, err := store.Admit(ctx, "host-1", reserve("job-a", "host-1", "vm-a", "root", 6, 0), CapacityView{})
	require.NoError(t, err)

	admission, err := store.Admit(ctx, "host-2", reserve("job-a", "host-2", "vm-a", "root", 6, 0), CapacityView{})
	require.NoError(t, err)
	require.True(t, admission.Admitted, "re-placing must not collide with the job's own pending row")

	first, err := store.PendingByHost(ctx, "host-1")
	require.NoError(t, err)
	assert.Equal(t, int64(0), first.CPU, "the abandoned host must be reimbursed")

	second, err := store.PendingByHost(ctx, "host-2")
	require.NoError(t, err)
	assert.Equal(t, int64(6), second.CPU)

	var rows int64
	require.NoError(t, db.Table("resource_reservations").Count(&rows).Error)
	assert.Equal(t, int64(1), rows, "one job is one charge, wherever it lives")
}

// TestAdmitRefusesDuplicateNameClusterWide covers the second failure mode: two
// concurrent same-name creates landing on DIFFERENT hosts today produce two
// identical VMs with nothing detecting it.
func TestAdmitRefusesDuplicateNameClusterWide(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 8192)
	seedHost(t, db, "host-2", 10, 8192)

	_, err := store.Admit(ctx, "host-1", reserve("job-a", "host-1", "twin", "root", 2, 0), CapacityView{})
	require.NoError(t, err)

	admission, err := store.Admit(ctx, "host-2", reserve("job-b", "host-2", "twin", "root", 2, 0), CapacityView{})
	require.NoError(t, err)
	assert.False(t, admission.Admitted)
	assert.Equal(t, "vm_name", admission.Reason)

	// A different owner may of course use the same name.
	other, err := store.Admit(ctx, "host-2", reserve("job-c", "host-2", "twin", "someone-else", 2, 0), CapacityView{})
	require.NoError(t, err)
	assert.True(t, other.Admitted)
}

// TestAdmitRefusesExistingVmName closes the create-collides-with-existing-VM
// case, which today only fails host-side after a full pack download.
func TestAdmitRefusesExistingVmName(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 8192)
	require.NoError(t, db.Create(&data_models.VirtualMachine{
		ID: "vm-1", Name: "already-here", User: "root", HostId: "host-1",
	}).Error)

	admission, err := store.Admit(ctx, "host-1", reserve("job-a", "host-1", " already-here ", "root", 2, 0), CapacityView{})
	require.NoError(t, err)
	assert.False(t, admission.Admitted, "name comparison is trimmed, matching the host's own comparison")
	assert.Equal(t, "vm_name", admission.Reason)
}

// TestAdmitEnforcesHostVMCountLimit keeps the MAX_VMS_PER_HOST policy in one
// place (check_host_requirements) rather than letting reservations drift from it.
func TestAdmitEnforcesHostVMCountLimit(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHostWithResources(t, db, "host-1", &data_models.HostResources{
		TotalAvailable:    data_models.HostResourceItem{LogicalCpuCount: 40, MemorySize: 8192},
		Total:             data_models.HostResourceItem{TotalVms: 3},
		MaxVmsPerHost:     5,
		MaxVmsPerHostMode: "any",
	})

	for _, id := range []string{"job-a", "job-b"} {
		admission, err := store.Admit(ctx, "host-1", reserve(id, "host-1", id, "root", 2, 0), CapacityView{})
		require.NoError(t, err)
		require.True(t, admission.Admitted)
	}

	admission, err := store.Admit(ctx, "host-1", reserve("job-c", "host-1", "job-c", "root", 2, 0), CapacityView{})
	require.NoError(t, err)
	assert.False(t, admission.Admitted)
	assert.Equal(t, "vm_count", admission.Reason)
	assert.Equal(t, "VM limit: current 3, pending 3, maximum 5", admission.Message)
}

// TestReleaseIsIdempotentAndPendingIsNeverAnError: release runs from job
// terminal hooks, where a duplicate or an unknown job is routine.
func TestReleaseIsIdempotentAndPendingIsNeverAnError(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 8192)

	_, err := store.Admit(ctx, "host-1", reserve("job-a", "host-1", "vm-a", "root", 6, 0), CapacityView{})
	require.NoError(t, err)

	require.NoError(t, store.Release(ctx, "job-a", OutcomeSucceeded))
	require.NoError(t, store.Release(ctx, "job-a", OutcomeSucceeded), "second release must not error")
	require.NoError(t, store.Release(ctx, "job-never-existed", OutcomeFailed))

	pending, err := store.PendingByHost(ctx, "host-idle")
	require.NoError(t, err, "an idle host is a zero view, not an error")
	assert.Equal(t, CapacityView{}, pending)

	hostID, err := store.HostOf(ctx, "job-a")
	require.NoError(t, err)
	assert.Equal(t, "", hostID, "a released job has no host to refresh")
}

// TestAdmitMakesNoNetworkCalls pins plan D10. The per-host lock is held to
// COMMIT; a probe inside it would queue every create for that host behind a
// network round trip, turning a correctness fix into a throughput incident.
func TestAdmitMakesNoNetworkCalls(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 8192)

	var calls atomic.Int64
	previous := http.DefaultTransport
	http.DefaultTransport = countRoundTripper{inner: previous, calls: &calls}
	t.Cleanup(func() { http.DefaultTransport = previous })

	_, err := store.Admit(ctx, "host-1", reserve("job-a", "host-1", "vm-a", "root", 2, 0), CapacityView{})
	require.NoError(t, err)
	require.NoError(t, store.Release(ctx, "job-a", OutcomeSucceeded))

	assert.Zero(t, calls.Load(), "no HTTP may happen inside the admission transaction")
}

type countRoundTripper struct {
	inner http.RoundTripper
	calls *atomic.Int64
}

func (c countRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	c.calls.Add(1)

	return c.inner.RoundTrip(req)
}

// TestStoreWithoutHandleFailsClosed: no SQL backend is "cannot measure", and
// an unmeasurable limit is not an unlimited one.
func TestStoreWithoutHandleFailsClosed(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	store := newSQLStore(nil)

	_, err := store.Admit(ctx, "host-1", reserve("job-a", "host-1", "vm-a", "root", 1, 0), CapacityView{})
	assert.ErrorIs(t, err, ErrStoreUnavailable)
	assert.ErrorIs(t, store.Release(ctx, "job-a", OutcomeFailed), ErrStoreUnavailable)

	_, err = store.PendingByHost(ctx, "host-1")
	assert.ErrorIs(t, err, ErrStoreUnavailable)

	// Including the rename: with no ledger there is nothing to re-attribute, and
	// reporting success would let a rename proceed as if the rows had moved.
	_, err = store.RenameNode(ctx, "node-a", "node-b")
	assert.ErrorIs(t, err, ErrStoreUnavailable)
}

// TestRenameNodeReAttributesOnlyItsOwnClaims covers the ledger after an operator
// renames a replica through the settings service. Nothing filters on node_id
// today, so this is not a correctness repair — it is the difference between a
// ledger that still names a replica that exists and one attributing live work to
// a name nobody can resolve.
func TestRenameNodeReAttributesOnlyItsOwnClaims(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 16, 32768)

	admit := func(jobID, nodeID, vmName string) {
		t.Helper()
		admission, err := store.Admit(ctx, "host-1", Reservation{
			JobID: jobID, HostID: "host-1", VmName: vmName, Owner: "root",
			NodeID: nodeID, CPU: 1, Memory: 512, VMCount: 1,
		}, CapacityView{})
		require.NoError(t, err)
		require.True(t, admission.Admitted)
	}
	admit("job-a1", "node-a", "vm-a1")
	admit("job-a2", "node-a", "vm-a2")
	admit("job-b1", "node-b", "vm-b1")

	moved, err := store.RenameNode(ctx, "node-a", "node-a-renamed")
	require.NoError(t, err)
	assert.Equal(t, int64(2), moved)

	nodes := map[string]string{}
	var rows []data_models.ResourceReservation
	require.NoError(t, db.Find(&rows).Error)
	for _, row := range rows {
		nodes[row.JobID] = row.NodeID
	}
	assert.Equal(t, "node-a-renamed", nodes["job-a1"])
	assert.Equal(t, "node-a-renamed", nodes["job-a2"])
	// Another replica's claims are not this rename's to move.
	assert.Equal(t, "node-b", nodes["job-b1"])

	// Idempotent: a rename interrupted between the two tables is repaired by
	// saving the same value again, not by a transaction spanning domains.
	moved, err = store.RenameNode(ctx, "node-a", "node-a-renamed")
	require.NoError(t, err)
	assert.Equal(t, int64(0), moved)

	// The refusals. An empty side would match rows whose owner was never
	// recorded and stamp them as this replica's work.
	for _, pair := range [][2]string{{"", "node-x"}, {"   ", "node-x"}, {"node-b", ""}, {"node-b", "  "}, {"node-b", "node-b"}} {
		_, err = store.RenameNode(ctx, pair[0], pair[1])
		assert.ErrorIs(t, err, interfaces.ErrNodeIdentityUnusable, "rename %q -> %q", pair[0], pair[1])
	}

	// Surrounding space from an input field is not a different node.
	moved, err = store.RenameNode(ctx, " node-b ", " node-b-2 ")
	require.NoError(t, err)
	assert.Equal(t, int64(1), moved)
}
