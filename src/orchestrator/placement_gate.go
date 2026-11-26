package orchestrator

// Placement gate: the orchestrator's side of capacity reservations.
//
// The bug this closes (see .ai-agents/features/vm-placement-reservations/):
// checkHostRequirements read a cached host snapshot, nothing was written
// between the decision and the host actually creating the VM, and two
// concurrent creates both read "N CPUs free" and both were accepted.
//
// Two halves, deliberately separated:
//
//   - the VIEW is a read. It makes the requirement checks honest about what is
//     already promised to creates still in flight, and it is what the pre-flight
//     endpoint reports. It claims nothing.
//   - the CLAIM is the decision. It runs once per host actually dispatched to,
//     inside the placement store's per-host transaction, and it is the only
//     thing that can promise capacity.
//
// A host that refuses a claim is a host whose arithmetic no longer fits — the
// create moves to the next candidate. A reservation never marks a host busy:
// several creates are expected to keep landing on the same good host while it
// has room (plan D3), which is what keeps filterAndSortHosts' cache-locality
// and latency ordering meaningful.

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/placement"
)

// resourceView is what placement knows that the cached host snapshot does not.
// It is passed by value into the requirement checks so no check can quietly
// consult a stale global.
type resourceView struct {
	// pending is capacity already claimed on this host. Combined with the
	// host's own reported pending by max(), never summed — see
	// placement.CombinePending.
	pending placement.CapacityView
	// nameTaken is host-independent: the same create carries the same name to
	// every host, so it is probed once per request rather than per candidate.
	nameTaken bool
}

// placementContext is the read side, computed once per create (or once per
// pre-flight) instead of once per host: one ledger query plus one name probe
// regardless of how many hosts are being evaluated.
type placementContext struct {
	byHost    map[string]placement.CapacityView
	nameTaken bool
}

// view returns what placement knows about one host. A context from a disabled
// or unreadable placement domain yields the zero view, which evaluates a host
// on exactly the numbers the host itself reported.
func (p placementContext) view(hostID string) resourceView {
	return resourceView{pending: p.byHost[hostID], nameTaken: p.nameTaken}
}

// newPlacementContext reads the ledger for the whole candidate set. It never
// fails a create: the authoritative gate is the claim, and a ledger that
// cannot be read must not reject a create that the claim would still admit.
func (s *OrchestratorService) newPlacementContext(ctx basecontext.ApiContext, request models.CreateVirtualMachineRequest, jobID string) placementContext {
	service := placement.Get()
	if !service.Enabled() {
		return placementContext{}
	}

	pctx := placementContext{
		byHost:    service.PendingForAllHosts(ctx),
		nameTaken: service.NameTaken(ctx, request.Owner, request.Name, jobID),
	}
	if len(pctx.byHost) == 0 {
		pctx.byHost = map[string]placement.CapacityView{}
	}

	return pctx
}

// claimCapacity promises capacity for this create on this host. It is the only
// call in the placement path that changes state.
//
// A nil return means "go ahead". A non-nil return means do not dispatch here:
// either the arithmetic refused (400, phrased like validateHost so the
// aggregate message the user sees is unchanged), or the ledger could not be
// read (503) — an unmeasurable limit is not an unlimited one, so placement
// fails closed rather than overbooking.
func (s *OrchestratorService) claimCapacity(ctx basecontext.ApiContext, host data_models.OrchestratorHost,
	jobID string, request models.CreateVirtualMachineRequest, specs *models.CreateVirtualMachineSpecs) *models.ApiErrorResponse {
	if strings.TrimSpace(jobID) == "" {
		// Without a job there is no primary key to hang the charge on, and no
		// terminal-state hook to release it. Logged, not blocked: these are the
		// synchronous internal calls, not the async create API.
		s.ctx.LogDebugf("[Placement] No job id, skipping capacity claim on host %s", host.ID)
		return nil
	}

	service := placement.Get()
	if !service.Enabled() {
		return nil
	}

	admission, err := service.Reserve(ctx, host.ID, reservationFromRequest(jobID, host.ID, request, specs))
	if err != nil {
		ctx.LogErrorf("[Placement] Capacity claim unavailable for job %s on host %s: %v", jobID, host.ID, err)

		return &models.ApiErrorResponse{
			Message: "Cannot verify host resources right now, please retry",
			Code:    503,
		}
	}
	if admission.Admitted {
		return nil
	}

	ctx.LogInfof("[Placement] Capacity claim refused for job %s on host %s: %s", jobID, host.ID, admission.Message)

	return &models.ApiErrorResponse{
		Message: "Insufficient host resources: " + admission.Message,
		Code:    400,
	}
}

// releaseCapacity gives capacity back when a dispatch attempt fails before the
// create is truly in flight. The authoritative release is the job terminal
// hook (jobs.OnJobTerminal); this one is the courtesy that stops a host
// hopping loop from stacking charges. Both are idempotent.
func (s *OrchestratorService) releaseCapacity(jobID string, outcome placement.Outcome) {
	if strings.TrimSpace(jobID) == "" {
		return
	}

	service := placement.Get()
	if !service.Enabled() {
		return
	}
	if err := service.ReleaseJob(s.ctx, jobID, outcome); err != nil {
		s.ctx.LogWarnf("[Placement] Capacity release for job %s failed (lease will reclaim): %v", jobID, err)
	}
}

// reservationFromRequest charges exactly what the create will apply: the same
// resolved specs the requirement checks compared against. Diverging here is how
// an orchestrator-side charge and the host-side gate end up disagreeing about
// what one VM costs.
func reservationFromRequest(jobID, hostID string, request models.CreateVirtualMachineRequest,
	specs *models.CreateVirtualMachineSpecs) placement.Reservation {
	reservation := placement.Reservation{
		JobID:   jobID,
		HostID:  hostID,
		VmName:  request.Name,
		Owner:   request.Owner,
		VMCount: 1,
	}
	if specs == nil {
		return reservation
	}

	reservation.CPU = specs.GetCpuCount()
	reservation.Memory = specs.GetMemorySize()
	if strings.EqualFold(specs.Type, "macvm") {
		reservation.AppleVM = 1
	}

	return reservation
}

// hostReportedPending is what the host itself publishes as in flight: the
// CPU/memory a create has been admitted for on that host and has not applied
// yet, plus its pending VM counts. Both sources are needed because each sees
// what the other cannot — the ledger knows about creates placement promised and
// the host has not seen yet, the host knows about a create submitted straight
// against it — and CombinePending charges the overlap once rather than twice.
//
// Zero while VM_CPU_ADMISSION_ENABLED is off, which leaves the pre-feature
// numbers intact for hosts without the gate.
func hostReportedPending(resources *data_models.HostResources) placement.CapacityView {
	if resources == nil {
		return placement.CapacityView{}
	}

	return placement.CapacityView{
		CPU:     resources.PendingCpu,
		Memory:  resources.PendingMemory,
		VMCount: resources.TotalReserved.TotalVms,
		AppleVM: resources.TotalReserved.TotalAppleVms,
	}
}

// maxInt64 lets the ledger raise the pending figure a comparison uses without
// rewriting the "current + pending < maximum" shape the count checks already
// had and are tested against.
func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

// queueHostRefresh asks for one host's snapshot to be re-read, instead of the
// blocking full-cluster reload the failure paths used to perform. Under a
// burst — and with several replicas — reloading every host to learn about one
// is a thundering herd at precisely the wrong moment.
func (s *OrchestratorService) queueHostRefresh(hostID string) {
	if s.hwQueue == nil || strings.TrimSpace(hostID) == "" {
		return
	}

	s.hwQueue.Enqueue(hostID)
}
