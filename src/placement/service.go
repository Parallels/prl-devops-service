package placement

// The service is what the orchestrator talks to. It owns three decisions the
// callers must not re-implement: whether enforcement is on, the ordering that
// makes release safe, and the fail-closed posture.

import (
	"strings"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

// Service fronts a Store. The store is an interface field, not a concrete
// type, so a Redis (or in-memory test) backend drops in without touching a
// single caller.
type Service struct {
	ctx   basecontext.ApiContext
	store Store

	// refreshHost is injected by the orchestrator at startup. It must exist:
	// jobs (and therefore this package) cannot import orchestrator without a
	// cycle, and the release ordering needs a host refresh.
	refreshHost func(ctx basecontext.ApiContext, hostID string) error
	refreshMu   sync.RWMutex

	// Reconciler bookkeeping, same chan-bool idiom as database.GormDatabase's
	// stale-job watcher.
	startMu        sync.Mutex
	stopReconciler chan bool
}

var (
	globalService *Service
	globalMu      sync.Mutex
)

// InitService opens the reservation ledger against the live SQL handle.
// A nil store is not an error here: legacy JSON DB mode has no SQL backend,
// and placement degrades to disabled with one warning instead of failing boot
// — the same shape as settings.InitService.
func InitService(ctx basecontext.ApiContext) (*Service, error) {
	db, err := serviceprovider.GetSQLDatabase(ctx)
	if err != nil {
		ctx.LogWarnf("[Placement] SQL database unavailable, capacity reservations disabled: %v", err)
		return register(&Service{ctx: ctx}), nil
	}

	svc := &Service{ctx: ctx, store: newSQLStore(db)}
	register(svc)
	if svc.Enabled() {
		ctx.LogInfof("[Placement] Capacity reservations enabled (node=%s, ttl=%dm)",
			helpers.NodeID(), config.Get().PlacementReservationTTLEnMinutes())
	}

	return svc, nil
}

func register(svc *Service) *Service {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalService = svc

	return svc
}

// Get returns the initialized service, or a disabled one when the domain was
// never initialized, so callers never nil-check.
func Get() *Service {
	globalMu.Lock()
	defer globalMu.Unlock()
	if globalService == nil {
		return &Service{}
	}

	return globalService
}

// New builds a service over an arbitrary store. Tests inject a fake; a future
// Redis backend is wired the same way.
func New(ctx basecontext.ApiContext, store Store) *Service {
	return register(&Service{ctx: ctx, store: store})
}

// Shutdown drops the singleton (tests, graceful stop).
func Shutdown() {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalService = nil
}

// SetRefreshHostHook registers the function that re-reads a host's real
// inventory. Injected rather than imported to keep the dependency one-way.
func (s *Service) SetRefreshHostHook(fn func(ctx basecontext.ApiContext, hostID string) error) {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	s.refreshHost = fn
}

// Enabled reports whether enforcement is on: the operator switch plus the
// availability of a store. Flipping PLACEMENT_RESERVATION_ENABLED=false is the
// documented rollback and needs no rebuild.
func (s *Service) Enabled() bool {
	if s == nil || s.store == nil {
		return false
	}

	return config.Get().PlacementReservationEnabled()
}

// Reserve claims capacity for one create on one host. Fail-closed by design:
// a store error is returned to the caller, which must treat it as a refusal
// rather than as "unknown, dispatch anyway" — an unmeasurable limit is not an
// unlimited one.
//
// When enforcement is off it returns an admitted, bypassed verdict so the
// create proceeds exactly as it did before this feature existed.
func (s *Service) Reserve(ctx basecontext.ApiContext, hostID string, res Reservation) (Admission, error) {
	if !s.Enabled() {
		return Admission{Admitted: true, Bypassed: true}, nil
	}
	if strings.TrimSpace(res.JobID) == "" {
		return Admission{}, ErrStoreUnavailable
	}
	if res.NodeID == "" {
		res.NodeID = helpers.NodeID()
	}

	return s.store.Admit(ctx, hostID, res, CapacityView{})
}

// ReleaseJob frees a claim. Ordering is the substance of this method (plan D8):
//
//   - success: refresh the host snapshot FIRST, then delete. Deleting first
//     would open an instant where neither the reservation nor the host's
//     TotalInUse accounts for the capacity — the original bug relocated.
//   - failure: delete immediately; the refresh is best effort afterwards.
//
// A refresh failure never blocks the delete. An unreachable host must not
// strand capacity, and the lease is the backstop for everything else.
func (s *Service) ReleaseJob(ctx basecontext.ApiContext, jobID string, outcome Outcome) error {
	if s == nil || s.store == nil {
		return nil
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil
	}

	hostID, err := s.store.HostOf(ctx, jobID)
	if err != nil {
		return err
	}

	if outcome == OutcomeSucceeded && hostID != "" {
		s.refresh(ctx, hostID)
	}

	if err := s.store.Release(ctx, jobID, outcome); err != nil {
		return err
	}

	if outcome != OutcomeSucceeded && hostID != "" {
		s.refresh(ctx, hostID)
	}

	return nil
}

func (s *Service) refresh(ctx basecontext.ApiContext, hostID string) {
	s.refreshMu.RLock()
	refresh := s.refreshHost
	s.refreshMu.RUnlock()
	if refresh == nil {
		return
	}

	if err := refresh(ctx, hostID); err != nil {
		ctx.LogWarnf("[Placement] Host %s refresh on release failed (best effort): %v", hostID, err)
	}
}

// PendingForHost is the advisory view used by the read-only requirement checks
// and the host API. Unlike Reserve it never propagates an error into a create
// decision: the authoritative gate is the atomic Admit, and a ledger that
// cannot be read must not fail a create that the gate will still judge.
func (s *Service) PendingForHost(ctx basecontext.ApiContext, hostID string) CapacityView {
	if s == nil || s.store == nil || !s.Enabled() {
		return CapacityView{}
	}

	pending, err := s.store.PendingByHost(ctx, hostID)
	if err != nil {
		ctx.LogWarnf("[Placement] Pending lookup failed for host %s: %v", hostID, err)
		return CapacityView{}
	}

	return pending
}

// PendingForAllHosts is the batched form, one read for the whole candidate set.
// It returns an empty map rather than an error for the same reason as
// PendingForHost.
func (s *Service) PendingForAllHosts(ctx basecontext.ApiContext) map[string]CapacityView {
	if s == nil || s.store == nil || !s.Enabled() {
		return map[string]CapacityView{}
	}

	pending, err := s.store.PendingByHosts(ctx)
	if err != nil {
		ctx.LogWarnf("[Placement] Pending lookup failed: %v", err)
		return map[string]CapacityView{}
	}

	return pending
}

// NameTaken is the read-only probe behind the vm_name requirement check.
func (s *Service) NameTaken(ctx basecontext.ApiContext, owner, vmName, exceptJobID string) bool {
	if s == nil || s.store == nil || !s.Enabled() {
		return false
	}

	taken, err := s.store.NameTaken(ctx, owner, vmName, exceptJobID)
	if err != nil {
		ctx.LogWarnf("[Placement] Name check failed for %s/%s: %v", owner, vmName, err)
		return false
	}

	return taken
}

// RenameNodeIdentity re-attributes this replica's outstanding claims after an
// operator renames it through the settings service. The jobs table is renamed by
// its own domain from the same hook, in the same order: identity first, rows
// after, so a claim created during the window already carries the new name and is
// not missed by an update keyed on the old one.
//
// Not gated on Enabled(): rows written while enforcement was on still exist after
// it is switched off, and they need the rename just as much.
func (s *Service) RenameNodeIdentity(ctx basecontext.ApiContext, from, to string) error {
	if s == nil || s.store == nil {
		// No ledger, so nothing to strand — the same reading as ReleaseJob.
		return nil
	}

	_, err := s.store.RenameNode(ctx, from, to)

	return err
}

func leaseTTL() time.Duration {
	return time.Duration(config.Get().PlacementReservationTTLEnMinutes()) * time.Minute
}
