package orchestrator

import (
	"errors"
	"github.com/Parallels/prl-devops-service/basecontext"
	"path/filepath"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	dbmodels "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/placement"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/stretchr/testify/require"
)

// The read half of placement: what the requirement checks do with capacity that
// is already promised but not yet visible on the host. The claim half is
// covered in placement's own store tests.

// healthyRoomyHost is the host from the issue: 10 logical CPUs, one kept by the
// system, nothing created yet.
func healthyRoomyHost() dbmodels.OrchestratorHost {
	return dbmodels.OrchestratorHost{
		Host:         "host-1",
		Enabled:      true,
		State:        "healthy",
		Architecture: "arm64",
		Resources: &dbmodels.HostResources{
			TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 9, MemorySize: 8192},
		},
	}
}

func mustCheck(t *testing.T, checks []models.RequirementCheckResult, name string) models.RequirementCheckResult {
	t.Helper()

	check, ok := requirementsCheckByName(models.MachineRequirementsHostResult{Checks: checks}, name)
	require.True(t, ok, "check %q was not evaluated", name)

	return check
}

func TestCheckHostRequirementsChargesCreatesStillInFlight(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := healthyRoomyHost()
	request := models.CreateVirtualMachineRequest{Architecture: "arm64", Owner: "root", Name: "web-1"}
	specs := &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "6", Memory: "4096"}

	// The first create sees the host as the host reported it.
	require.True(t, mustCheck(t, svc.checkHostRequirements(host, request, specs, resourceView{}), checkNameCpu).Passed)

	// The second one arrives before the host has had any chance to re-report, so
	// the snapshot still says 9. The ledger is what makes it say 3 — this is the
	// issue's repro, and the message is what the reporter asked to see.
	second := svc.checkHostRequirements(host, request, specs, resourceView{
		pending: placement.CapacityView{CPU: 6, Memory: 4096, VMCount: 1},
	})
	cpu := mustCheck(t, second, checkNameCpu)
	require.False(t, cpu.Passed)
	require.Equal(t, "CPU: required 6 logical CPUs, available 3", cpu.Message)

	// Memory is the boundary case: exactly the last of it is left, and an exact
	// fit is a fit.
	memory := mustCheck(t, second, checkNameMemory)
	require.True(t, memory.Passed)
	require.Equal(t, "Memory: required 4096 MiB, available 4096 MiB", memory.Message)
}

func TestCheckHostRequirementsChargesOverlappingPendingOnce(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := healthyRoomyHost()
	// The host itself is already holding 6 CPUs for a create it accepted, and
	// placement promised the same 6 to the same operation: the two figures
	// describe one VM, so the host may still be charged 6 and not 12.
	host.Resources.PendingCpu = 6
	host.Resources.PendingMemory = 4096

	request := models.CreateVirtualMachineRequest{Architecture: "arm64", Owner: "root", Name: "web-1"}
	specs := &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "3", Memory: "2048"}

	checks := svc.checkHostRequirements(host, request, specs, resourceView{
		pending: placement.CapacityView{CPU: 6, Memory: 4096},
	})

	cpu := mustCheck(t, checks, checkNameCpu)
	require.True(t, cpu.Passed, "double-charging the overlap would refuse a create that fits: %s", cpu.Message)
	require.Equal(t, "CPU: required 3 logical CPUs, available 3", cpu.Message)
}

func TestCheckHostRequirementsLedgerRaisesPendingVmCount(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := healthyRoomyHost()
	host.Resources.Total = dbmodels.HostResourceItem{TotalVms: 3}
	host.Resources.MaxVmsPerHost = 5

	request := models.CreateVirtualMachineRequest{Architecture: "arm64", Owner: "root", Name: "web-1"}
	specs := &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "1", Memory: "512"}

	// Nothing pending: three of five, room for two more.
	require.True(t, mustCheck(t, svc.checkHostRequirements(host, request, specs, resourceView{}), checkNameVmCount).Passed)

	// Three more promised by placement, invisible to the host until they land.
	checks := svc.checkHostRequirements(host, request, specs, resourceView{
		pending: placement.CapacityView{CPU: 3, Memory: 1536, VMCount: 3},
	})
	vmCount := mustCheck(t, checks, checkNameVmCount)
	require.False(t, vmCount.Passed)
	require.Equal(t, "VM limit: current 3, pending 3, maximum 5", vmCount.Message)
}

func TestCheckHostRequirementsReportsTakenNameVerbatim(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := healthyRoomyHost()
	request := models.CreateVirtualMachineRequest{Architecture: "arm64", Owner: "root", Name: "web-1"}
	specs := &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "2", Memory: "1024"}

	// A taken name is a verdict, not a per-host shortage: it short-circuits to a
	// single check so validateHost reports it on its own rather than burying it
	// under "Insufficient host resources", which is not what is wrong here.
	checks := svc.checkHostRequirements(host, request, specs, resourceView{nameTaken: true})
	require.Len(t, checks, 1)
	require.Equal(t, checkNameVmName, checks[0].Name)
	require.False(t, checks[0].Passed)
	require.Contains(t, checks[0].Message, `"web-1"`)
	require.Contains(t, checks[0].Message, `"root"`)

	ok, apiErr := svc.validateHost(host, request, specs, resourceView{nameTaken: true})
	require.False(t, ok)
	require.NotNil(t, apiErr)
	require.Equal(t, 400, apiErr.Code)
	require.NotContains(t, apiErr.Message, "Insufficient host resources")
	require.Contains(t, apiErr.Message, "already in use")
}

func TestCheckHostRequirementsWithoutPlacementIsUnchanged(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := healthyRoomyHost()
	request := models.CreateVirtualMachineRequest{Architecture: "arm64", Owner: "root", Name: "web-1"}
	specs := &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "9", Memory: "8192"}

	// The zero view must reproduce the pre-feature arithmetic exactly: an exact
	// fit is an exact fit, and no phantom pending quietly shrinks the host.
	ok, apiErr := svc.validateHost(host, request, specs, resourceView{})
	require.True(t, ok, apiErr)
	require.Nil(t, apiErr)
}

func TestReservationFromRequestChargesWhatTheCreateApplies(t *testing.T) {
	request := models.CreateVirtualMachineRequest{Name: "web-1", Owner: "root"}

	res := reservationFromRequest("job-1", "host-1", request, &models.CreateVirtualMachineSpecs{Type: "macvm", Cpu: "4", Memory: "4096"})
	require.Equal(t, "job-1", res.JobID)
	require.Equal(t, "host-1", res.HostID)
	require.Equal(t, "web-1", res.VmName)
	require.Equal(t, "root", res.Owner)
	require.Equal(t, int64(4), res.CPU)
	require.Equal(t, float64(4096), res.Memory)
	require.Equal(t, int64(1), res.VMCount)
	// Only a macOS VM consumes the Apple quota; a pvm must not.
	require.Equal(t, int64(1), res.AppleVM)

	pvm := reservationFromRequest("job-2", "host-1", request, &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "4", Memory: "4096"})
	require.Equal(t, int64(0), pvm.AppleVM)

	// Specs that could not be resolved still occupy a slot: refusing to charge
	// anything at all would let an unmeasurable create bypass the count limit.
	unknown := reservationFromRequest("job-3", "host-1", request, nil)
	require.Equal(t, int64(0), unknown.CPU)
	require.Equal(t, int64(1), unknown.VMCount)
}

func TestPlacementBookkeepingIsInertWhenDisabled(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}

	// Placement is never initialized in this package's tests, which is exactly
	// the state a deployment without a SQL database or with the switch off ends
	// up in. Neither helper may panic or block there.
	require.NotPanics(t, func() {
		svc.queueHostRefresh("host-1")
		svc.queueHostRefresh("")
		svc.releaseCapacity("job-1", placement.OutcomeFailed)
		svc.releaseCapacity("", placement.OutcomeSucceeded)
	})

	// And the read side must degrade to "trust the host's own numbers" rather
	// than to an error that would refuse creates.
	ctx := newCheckTestContext()
	require.NotPanics(t, func() {
		pctx := svc.newPlacementContext(ctx, models.CreateVirtualMachineRequest{Owner: "root", Name: "web-1"}, "job-1")
		require.False(t, pctx.view("host-1").nameTaken)
		require.Equal(t, placement.CapacityView{}, pctx.view("host-1").pending)
	})
}

func TestClaimCapacityIsInertWithoutAJobOrPlacement(t *testing.T) {
	svc := &OrchestratorService{ctx: newCheckTestContext()}
	host := healthyRoomyHost()
	request := models.CreateVirtualMachineRequest{Architecture: "arm64", Owner: "root", Name: "web-1"}
	specs := &models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "6", Memory: "4096"}
	ctx := newCheckTestContext()

	// No job id: nothing to key the charge on, and no terminal hook to release
	// it, so the create proceeds exactly as it did before this feature.
	require.Nil(t, svc.claimCapacity(ctx, host, "", request, specs))
	// Placement uninitialized: same answer, no error surfaced to the caller.
	require.Nil(t, svc.claimCapacity(ctx, host, "job-1", request, specs))
}

// stubStore is the ledger as this package sees it: enough of placement.Store
// to drive the create path without a database, so the claim side of the gate —
// the refusal wording a user reads, the fail-closed posture, the charge a
// create actually carries — is covered where the create path lives instead of
// only inside placement.
type stubStore struct {
	admission placement.Admission
	err       error
	pending   map[string]placement.CapacityView
	nameTaken bool
	hostOf    string

	admitted []placement.Reservation
	released []string
}

func (s *stubStore) Admit(_ basecontext.ApiContext, _ string, res placement.Reservation, _ placement.CapacityView) (placement.Admission, error) {
	if s.err != nil {
		return placement.Admission{}, s.err
	}

	s.admitted = append(s.admitted, res)

	return s.admission, nil
}

func (s *stubStore) Release(_ basecontext.ApiContext, jobID string, _ placement.Outcome) error {
	if s.err != nil {
		return s.err
	}

	s.released = append(s.released, jobID)

	return nil
}

func (s *stubStore) PendingByHost(_ basecontext.ApiContext, hostID string) (placement.CapacityView, error) {
	return s.pending[hostID], s.err
}

func (s *stubStore) PendingByHosts(_ basecontext.ApiContext) (map[string]placement.CapacityView, error) {
	return s.pending, s.err
}

func (s *stubStore) HostOf(_ basecontext.ApiContext, _ string) (string, error) {
	return s.hostOf, s.err
}

func (s *stubStore) NameTaken(_ basecontext.ApiContext, _, _, _ string) (bool, error) {
	return s.nameTaken, s.err
}

func (s *stubStore) ExpireLeases(_ basecontext.ApiContext, _ time.Time) (int64, error) {
	return 0, s.err
}

func (s *stubStore) ReconcileOrphans(_ basecontext.ApiContext) (int64, error) {
	return 0, s.err
}

func (s *stubStore) RenameNode(_ basecontext.ApiContext, _, _ string) (int64, error) {
	return 0, s.err
}

func useStubPlacement(t *testing.T, store placement.Store) *placement.Service {
	t.Helper()

	service := placement.New(newCheckTestContext(), store)
	t.Cleanup(placement.Shutdown)

	return service
}

func TestClaimCapacityTurnsARefusalIntoTheReportersMessage(t *testing.T) {
	store := &stubStore{admission: placement.Admission{
		Admitted:  false,
		Available: placement.CapacityView{CPU: 3},
		Reason:    constants.PlacementReasonCPU,
		Message:   "CPU: required 6 logical CPUs, available 3",
	}}
	useStubPlacement(t, store)
	svc := &OrchestratorService{ctx: newCheckTestContext()}

	failure := svc.claimCapacity(newCheckTestContext(), healthyRoomyHost(), "job-1",
		models.CreateVirtualMachineRequest{Architecture: "arm64", Owner: "root", Name: "web-1"},
		&models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "6", Memory: "4096"})

	require.NotNil(t, failure)
	require.Equal(t, 400, failure.Code)
	// The exact sentence from the issue, under the same prefix every other
	// resource shortage in this path carries.
	require.Equal(t, "Insufficient host resources: CPU: required 6 logical CPUs, available 3", failure.Message)

	require.Len(t, store.admitted, 1)
	require.Equal(t, "job-1", store.admitted[0].JobID)
	require.Equal(t, int64(6), store.admitted[0].CPU)
	require.Equal(t, int64(1), store.admitted[0].VMCount)
}

func TestClaimCapacityFailsClosedWhenTheLedgerCannotBeRead(t *testing.T) {
	store := &stubStore{err: errors.New("database is gone")}
	useStubPlacement(t, store)
	svc := &OrchestratorService{ctx: newCheckTestContext()}

	failure := svc.claimCapacity(newCheckTestContext(), healthyRoomyHost(), "job-1",
		models.CreateVirtualMachineRequest{Architecture: "arm64", Owner: "root", Name: "web-1"},
		&models.CreateVirtualMachineSpecs{Type: "pvm", Cpu: "6", Memory: "4096"})

	// An unreadable ledger is not an unlimited one: the create is refused rather
	// than allowed to overbook a host nobody can measure.
	require.NotNil(t, failure)
	require.Equal(t, 503, failure.Code)
	require.Empty(t, store.admitted)
}

func TestGetHostResourcesReportsTheOutstandingCharge(t *testing.T) {
	ctx := newCheckTestContext()
	db := data.NewJsonDatabase(ctx, filepath.Join(t.TempDir(), "host-resources.json"))
	//nolint:staticcheck // legacy engine drain by design; the whole engine retires in v1.3.0 (#426)
	t.Cleanup(func() { _ = db.Disconnect(ctx) })
	serviceprovider.NewMockProvider().JsonDatabase = db

	host, err := db.CreateOrchestratorHost(ctx, dbmodels.OrchestratorHost{
		Host:      "host-1",
		Enabled:   true,
		State:     "healthy",
		Resources: &dbmodels.HostResources{PendingCpu: 2, TotalAvailable: dbmodels.HostResourceItem{LogicalCpuCount: 9}},
	})
	require.NoError(t, err)

	// The host is holding 2 CPUs for itself, placement promised 6: the larger,
	// not the sum, because both can describe the same creates.
	useStubPlacement(t, &stubStore{pending: map[string]placement.CapacityView{
		host.ID: {CPU: 6, Memory: 4096, VMCount: 1},
	}})
	svc := &OrchestratorService{ctx: ctx}

	resources, err := svc.GetHostResources(ctx, host.ID)
	require.NoError(t, err)
	require.Equal(t, int64(6), resources.PendingCpu)
	require.Equal(t, float64(4096), resources.PendingMemory)

	// And the cached host still reports the host's own figure: the ledger must
	// never be read back into the scheduling decision as if the host had said it.
	cached, err := db.GetOrchestratorHost(ctx, host.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), cached.Resources.PendingCpu)
}
