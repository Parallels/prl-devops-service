package orchestrator

import (
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
)

// hardwareUpdateQueue manages a per-host, deduplicated queue of hardware refresh
// requests. A single background worker drains the queue, holding a per-host
// conceptual lock (only one fetch per host at a time) by processing each host
// to completion before moving to the next.
//
// Concurrency properties:
//   - Enqueue never blocks the caller (no I/O, no wait).
//   - The mutex is held only while mutating pending/inFlight maps, never during
//     the slow HTTP call to /v1/config/hardware.
//   - Deduplication: if a host already has a pending request, a second Enqueue
//     is a no-op. One pending entry is enough because the worker always fetches
//     live data.
//   - If a host is currently in-flight (being processed), new requests are still
//     accepted so the worker re-fetches after the current one completes.
type hardwareUpdateQueue struct {
	mu      sync.Mutex
	cond    *sync.Cond
	pending  map[string]struct{} // hosts with an outstanding refresh request
	inFlight map[string]struct{} // hosts currently being processed by the worker
	stopped  bool
	svc      *OrchestratorService
}

func newHardwareUpdateQueue(svc *OrchestratorService) *hardwareUpdateQueue {
	q := &hardwareUpdateQueue{
		pending:  make(map[string]struct{}),
		inFlight: make(map[string]struct{}),
		svc:      svc,
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Enqueue adds hostID to the pending set. If a request for this host is already
// pending (not yet being processed), the call is a no-op — one pending entry is
// sufficient. Implements interfaces.HardwareEnqueuer.
func (q *hardwareUpdateQueue) Enqueue(hostID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.stopped {
		return
	}

	if _, alreadyPending := q.pending[hostID]; alreadyPending {
		return
	}

	q.pending[hostID] = struct{}{}
	q.cond.Signal()
}

// Start spawns the background worker goroutine. Call once from OrchestratorService.Start.
func (q *hardwareUpdateQueue) Start(ctx basecontext.ApiContext) {
	go q.runWorker(ctx)
}

// Stop signals the worker to exit. The worker finishes any in-progress fetch and
// drains remaining pending requests before returning.
func (q *hardwareUpdateQueue) Stop() {
	q.mu.Lock()
	q.stopped = true
	q.cond.Broadcast()
	q.mu.Unlock()
}

// runWorker is the single background goroutine that drains the queue.
//
// Algorithm:
//  1. Acquire lock.
//  2. While pending is empty and not stopped → cond.Wait (releases lock, sleeps).
//  3. On wakeup: if stopped AND pending is empty → exit.
//  4. Pick one host from pending, move it to inFlight, release lock.
//  5. Fetch hardware (slow HTTP call, no lock held).
//  6. Re-acquire lock, remove from inFlight, release lock.
//  7. Loop back to 1 — immediately processes next if pending is non-empty.
func (q *hardwareUpdateQueue) runWorker(ctx basecontext.ApiContext) {
	for {
		q.mu.Lock()

		for len(q.pending) == 0 && !q.stopped {
			q.cond.Wait()
		}

		if q.stopped && len(q.pending) == 0 {
			q.mu.Unlock()
			return
		}

		// Pick any host from pending (map range gives a pseudo-random first key).
		var hostID string
		for id := range q.pending {
			hostID = id
			break
		}
		delete(q.pending, hostID)
		q.inFlight[hostID] = struct{}{}

		q.mu.Unlock()

		// Slow path: HTTP call to the host. No lock held.
		q.fetchAndPersist(ctx, hostID)

		q.mu.Lock()
		delete(q.inFlight, hostID)
		q.mu.Unlock()
	}
}

// fetchAndPersist calls /v1/config/hardware on the host and atomically updates
// only the Resources field in the DB. Reuses the existing service helpers so
// the mapping logic stays in one place.
func (q *hardwareUpdateQueue) fetchAndPersist(ctx basecontext.ApiContext, hostID string) {
	host, err := q.svc.GetDatabaseHost(ctx, hostID)
	if err != nil || host == nil {
		ctx.LogWarnf("[HardwareUpdateQueue] Host %s not found, skipping hardware update", hostID)
		return
	}

	hardwareInfo, err := q.svc.GetHostHardwareInfo(host)
	if err != nil {
		ctx.LogErrorf("[HardwareUpdateQueue] Error fetching hardware for host %s: %v", hostID, err)
		return
	}

	// updateHostWithHardwareInfo populates host.Resources and several other fields
	// (Architecture, CpuModel, ParallelsDesktopVersion, …). We only persist
	// Resources via the targeted DB method, so the other fields are populated on
	// the local copy and discarded — they will be written on the next full refresh.
	q.svc.updateHostWithHardwareInfo(host, hardwareInfo)

	if err := q.svc.db.UpdateOrchestratorHostResources(ctx, hostID, host.Resources); err != nil {
		ctx.LogErrorf("[HardwareUpdateQueue] Error persisting resources for host %s: %v", hostID, err)
		return
	}

	ctx.LogInfof("[HardwareUpdateQueue] Hardware resources updated for host %s", hostID)
}
