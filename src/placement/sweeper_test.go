package placement

// Lifecycle tests: the sweeper's two sweeps, and the release ordering that
// keeps capacity accounted for at every instant (plan D8).

import (
	"errors"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestExpireLeasesRemovesOnlyExpired(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 8192)

	now := time.Now().UTC()
	insert := func(jobID, expires string) {
		require.NoError(t, db.Create(&data_models.ResourceReservation{
			JobID: jobID, HostID: "host-1", Owner: "root", VmName: jobID,
			CPU: 2, VMCount: 1, CreatedAt: expires, UpdatedAt: expires, ExpiresAt: expires,
		}).Error)
	}
	insert("job-stale", now.Add(-2*time.Hour).Format(leaseTimeFormat))
	insert("job-live", now.Add(time.Hour).Format(leaseTimeFormat))

	// A replica killed mid-create must not hold capacity for ever.
	removed, err := store.ExpireLeases(ctx, now)
	require.NoError(t, err)
	assert.Equal(t, int64(1), removed)

	var remaining []string
	require.NoError(t, db.Table("resource_reservations").Pluck("job_id", &remaining).Error)
	assert.Equal(t, []string{"job-live"}, remaining)
}

// TestReconcileOrphansCoversWhatHooksCannot: DetectStaleJobs and
// RecoverOngoingJobs move jobs to terminal states straight in the database, so
// no jobs-manager hook fires and only the reconciler frees the capacity.
func TestReconcileOrphansCoversWhatHooksCannot(t *testing.T) {
	store, db := newTestStore(t)
	ctx := testutil.WithCleanConfig(t)
	seedHost(t, db, "host-1", 10, 8192)

	seedJob(t, db, "job-running", constants.JobStateRunning)
	seedJob(t, db, "job-completed", constants.JobStateCompleted)
	for _, id := range []string{"job-running", "job-completed", "job-vanished"} {
		require.NoError(t, db.Create(&data_models.ResourceReservation{
			JobID: id, HostID: "host-1", Owner: "root", VmName: id, CPU: 1, VMCount: 1,
			ExpiresAt: time.Now().UTC().Add(time.Hour).Format(leaseTimeFormat),
		}).Error)
	}

	removed, err := store.ReconcileOrphans(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), removed, "terminal and missing jobs are orphans; a running one is not")

	var remaining []string
	require.NoError(t, db.Table("resource_reservations").Pluck("job_id", &remaining).Error)
	assert.Equal(t, []string{"job-running"}, remaining)

	// Sweeping twice is a no-op, which is what makes two replicas safe.
	removed, err = store.ReconcileOrphans(ctx)
	require.NoError(t, err)
	assert.Zero(t, removed)
}

func seedJob(t *testing.T, db *gorm.DB, id string, state constants.JobState) {
	t.Helper()
	require.NoError(t, db.Create(&data_models.Job{ID: id, State: state, Owner: "root"}).Error)
}

// ── service-level release ordering, with a fake store ───────────────────────

type fakeStore struct {
	calls       []string
	hostID      string
	releaseErr  error
	nameTaken   bool
	renamedFrom string
	renamedTo   string
	renamed     int64
}

func (f *fakeStore) record(name string) { f.calls = append(f.calls, name) }

func (f *fakeStore) Admit(ctx basecontext.ApiContext, hostID string, res Reservation,
	hostPending CapacityView) (Admission, error) {
	f.record("admit:" + hostID)

	return Admission{Admitted: true}, nil
}

func (f *fakeStore) Release(ctx basecontext.ApiContext, jobID string, outcome Outcome) error {
	f.record("release")

	return f.releaseErr
}

func (f *fakeStore) PendingByHost(ctx basecontext.ApiContext, hostID string) (CapacityView, error) {
	f.record("pending")

	return CapacityView{CPU: 3}, nil
}

func (f *fakeStore) PendingByHosts(ctx basecontext.ApiContext) (map[string]CapacityView, error) {
	return map[string]CapacityView{"host-1": {CPU: 3}}, nil
}

func (f *fakeStore) HostOf(ctx basecontext.ApiContext, jobID string) (string, error) {
	f.record("hostof")

	return f.hostID, nil
}

func (f *fakeStore) NameTaken(ctx basecontext.ApiContext, owner, vmName, exceptJobID string) (bool, error) {
	return f.nameTaken, nil
}

func (f *fakeStore) ExpireLeases(ctx basecontext.ApiContext, now time.Time) (int64, error) {
	f.record("expire")

	return 0, nil
}

func (f *fakeStore) ReconcileOrphans(ctx basecontext.ApiContext) (int64, error) {
	f.record("reconcile")

	return 0, nil
}

func (f *fakeStore) RenameNode(ctx basecontext.ApiContext, from, to string) (int64, error) {
	f.renamedFrom, f.renamedTo = from, to
	f.record("rename")

	return f.renamed, nil
}

// TestReleaseOnSuccessRefreshesBeforeDeleting pins the ordering that stops the
// bug reappearing one step later: the host's real TotalInUse must be persisted
// before the reservation stops covering it.
func TestReleaseOnSuccessRefreshesBeforeDeleting(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	store := &fakeStore{hostID: "host-1"}
	svc := New(ctx, store)
	t.Cleanup(Shutdown)

	var seen []string
	svc.SetRefreshHostHook(func(_ basecontext.ApiContext, hostID string) error {
		seen = append(seen, "refresh:"+hostID)
		store.record("refresh")

		return nil
	})

	require.NoError(t, svc.ReleaseJob(ctx, "job-a", OutcomeSucceeded))
	assert.Equal(t, []string{"hostof", "refresh", "release"}, store.calls)
	assert.Equal(t, []string{"refresh:host-1"}, seen)
}

// TestReleaseOnFailureDeletesImmediatelyAndStillRefreshes: the refresh is
// best effort and must never be able to strand capacity.
func TestReleaseOnFailureDeletesImmediatelyAndStillRefreshes(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	store := &fakeStore{hostID: "host-1"}
	svc := New(ctx, store)
	t.Cleanup(Shutdown)

	svc.SetRefreshHostHook(func(_ basecontext.ApiContext, _ string) error {
		store.record("refresh")

		return errors.New("host unreachable")
	})

	require.NoError(t, svc.ReleaseJob(ctx, "job-a", OutcomeFailed),
		"a refresh failure must not fail the release")
	assert.Equal(t, []string{"hostof", "release", "refresh"}, store.calls)
}

func TestReleaseWithoutRefreshHookStillReleases(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	store := &fakeStore{hostID: "host-1"}
	svc := New(ctx, store)
	t.Cleanup(Shutdown)

	require.NoError(t, svc.ReleaseJob(ctx, "job-a", OutcomeSucceeded))
	assert.Equal(t, []string{"hostof", "release"}, store.calls)
}

func TestReleasePropagatesStoreFailure(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	store := &fakeStore{hostID: "host-1", releaseErr: errors.New("db down")}
	svc := New(ctx, store)
	t.Cleanup(Shutdown)

	assert.Error(t, svc.ReleaseJob(ctx, "job-a", OutcomeFailed))
}

// TestReserveBypassesWhenDisabled: the rollback lever has to leave the create
// path behaving exactly as it did before the feature existed.
func TestReserveBypassesWhenDisabled(t *testing.T) {
	t.Setenv(constants.PlacementReservationEnabledEnvVar, "false")
	ctx := testutil.WithCleanConfig(t)
	store := &fakeStore{hostID: "host-1"}
	svc := New(ctx, store)
	t.Cleanup(Shutdown)

	admission, err := svc.Reserve(ctx, "host-1", Reservation{JobID: "job-a", VmName: "vm", Owner: "root"})
	require.NoError(t, err)
	assert.True(t, admission.Admitted)
	assert.True(t, admission.Bypassed)
	assert.Empty(t, store.calls, "disabled means the ledger is not touched at all")

	assert.Equal(t, CapacityView{}, svc.PendingForHost(ctx, "host-1"))
	assert.Empty(t, svc.PendingForAllHosts(ctx))
	assert.False(t, svc.NameTaken(ctx, "root", "vm", ""))
}

func TestReserveClaimsAndStampsNode(t *testing.T) {
	t.Setenv(constants.OrchestratorNodeIDEnvVar, "qa-node-1")
	ctx := testutil.WithCleanConfig(t)
	store := &fakeStore{hostID: "host-1"}
	svc := New(ctx, store)
	t.Cleanup(Shutdown)

	admission, err := svc.Reserve(ctx, "host-1", Reservation{JobID: "job-a", VmName: "vm", Owner: "root", CPU: 2})
	require.NoError(t, err)
	assert.True(t, admission.Admitted)
	assert.False(t, admission.Bypassed)
	assert.Equal(t, []string{"admit:host-1"}, store.calls)
}

// TestReserveRejectsJoblessReservation: an unclaimable request is a store
// error, and a store error is a refusal, not a pass.
func TestReserveRejectsJoblessReservation(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	svc := New(ctx, &fakeStore{})
	t.Cleanup(Shutdown)

	_, err := svc.Reserve(ctx, "host-1", Reservation{VmName: "vm", Owner: "root"})
	assert.ErrorIs(t, err, ErrStoreUnavailable)
}

// TestServiceWithoutStoreIsInert is the legacy JSON DB posture: no panic, no
// claims, creates proceed.
func TestServiceWithoutStoreIsInert(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	svc := register(&Service{ctx: ctx})
	t.Cleanup(Shutdown)

	assert.False(t, svc.Enabled())
	admission, err := svc.Reserve(ctx, "host-1", Reservation{JobID: "job-a"})
	require.NoError(t, err)
	assert.True(t, admission.Bypassed)
	require.NoError(t, svc.ReleaseJob(ctx, "job-a", OutcomeSucceeded))
	svc.ReconcileOnce(ctx)
}

func TestSweeperRunsBothSweeps(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	store := &fakeStore{}
	svc := New(ctx, store)
	t.Cleanup(Shutdown)

	svc.ReconcileOnce(ctx)
	assert.Equal(t, []string{"expire", "reconcile"}, store.calls)

	// Start/Stop must be safe to call twice and Stop safe before Start.
	svc.StartReconciler(ctx)
	svc.StartReconciler(ctx)
	svc.StopReconciler()
	svc.StopReconciler()
}

// TestRenameNodeIdentityReachesTheStore: the service is the only path to a
// rename, so one that silently no-ops above a working store would strand this
// replica's rows exactly as a missing rename would.
func TestRenameNodeIdentityReachesTheStore(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	store := &fakeStore{renamed: 3}
	svc := New(ctx, store)
	t.Cleanup(Shutdown)

	require.NoError(t, svc.RenameNodeIdentity(ctx, "node-a", "node-b"))
	assert.Equal(t, []string{"rename"}, store.calls)
	assert.Equal(t, "node-a", store.renamedFrom)
	assert.Equal(t, "node-b", store.renamedTo)
}

// TestRenameNodeIdentityWithoutAStoreIsNotAnError: legacy JSON mode has no
// ledger, so there is nothing to strand and the rename has genuinely done all it
// can — the same reading ReleaseJob gives an absent store.
func TestRenameNodeIdentityWithoutAStoreIsNotAnError(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	svc := &Service{ctx: ctx}

	require.NoError(t, svc.RenameNodeIdentity(ctx, "node-a", "node-b"))
}

// TestRenameNodeIdentityIsNotGatedOnEnforcement: rows written while enforcement
// was on outlive switching it off, and they need the rename just as much.
func TestRenameNodeIdentityIsNotGatedOnEnforcement(t *testing.T) {
	ctx := testutil.WithCleanConfig(t)
	t.Setenv(constants.PlacementReservationEnabledEnvVar, "false")
	store := &fakeStore{}
	svc := New(ctx, store)
	t.Cleanup(Shutdown)

	require.False(t, svc.Enabled())
	require.NoError(t, svc.RenameNodeIdentity(ctx, "node-a", "node-b"))
	assert.Equal(t, []string{"rename"}, store.calls)
}
