package registry

import "sync"

// JobLink maps a host-side job back to the orchestrator job that spawned it.
type JobLink struct {
	OrchestratorJobID string
	HostID            string
}

// The durable side of the registry. Every replica dials every host's WebSocket,
// so the replica that receives a host's completion event is not necessarily the
// one that dispatched it — and a link that only ever lived in one process's map
// means the orchestrator job that host job belongs to is never told it finished.
// The map stays authoritative for the hot path; these hooks are the fallback
// that survives a restart and crosses replicas.
//
// They are function values rather than a database handle so this package keeps
// its dependencies at zero, and they are optional: with none wired, the
// registry behaves exactly as it always has.
type (
	linkWriter  func(hostJobID string, link JobLink) error
	linkReader  func(hostJobID string) (JobLink, bool, error)
	linkDeleter func(hostJobID string) error
)

// JobLinkRegistry is an in-memory singleton that tracks the 1:1 relationship
// between an orchestrator job and the host job it dispatched.
// Keys are host job IDs; values carry the orchestrator job ID and host ID.
type JobLinkRegistry struct {
	mu      sync.RWMutex
	links   map[string]JobLink
	durable linkWriter
	reader  linkReader
	deleter linkDeleter
}

// SetDurableStore installs the persistence hooks. Any of them may be nil, and
// each is best-effort by contract: a failure is logged by whoever supplied it
// and never propagated to dispatch, which must not fail because bookkeeping
// could not be written — the link's other safety net is the reservation lease.
func (r *JobLinkRegistry) SetDurableStore(write linkWriter, read linkReader, remove linkDeleter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.durable = write
	r.reader = read
	r.deleter = remove
}

var (
	instance     *JobLinkRegistry
	instanceOnce sync.Once
)

// Get returns the singleton registry.
func Get() *JobLinkRegistry {
	instanceOnce.Do(func() {
		instance = &JobLinkRegistry{
			links: make(map[string]JobLink),
		}
	})
	return instance
}

// Register associates hostJobID with the given orchestrator job and host.
func (r *JobLinkRegistry) Register(hostJobID, orchJobID, hostID string) {
	link := JobLink{OrchestratorJobID: orchJobID, HostID: hostID}

	r.mu.Lock()
	r.links[hostJobID] = link
	write := r.durable
	r.mu.Unlock()

	// Outside the lock: the hook touches a database, and no dispatch should
	// queue behind another dispatch's bookkeeping.
	if write != nil {
		_ = write(hostJobID, link)
	}
}

// Lookup returns the JobLink for a host job ID, if one exists.
func (r *JobLinkRegistry) Lookup(hostJobID string) (JobLink, bool) {
	r.mu.RLock()
	link, ok := r.links[hostJobID]
	read := r.reader
	r.mu.RUnlock()

	if ok {
		return link, true
	}
	if read == nil {
		return JobLink{}, false
	}

	// Not this process's link: dispatched by another replica, or by this one
	// before a restart. Ask wherever the links are kept, and remember the answer
	// so the remaining events for this job do not pay for the lookup again.
	found, ok, err := read(hostJobID)
	if err != nil || !ok {
		return JobLink{}, false
	}

	r.mu.Lock()
	r.links[hostJobID] = found
	r.mu.Unlock()

	return found, true
}

// Remove deletes the link for a host job ID (called on terminal state).
func (r *JobLinkRegistry) Remove(hostJobID string) {
	r.mu.Lock()
	delete(r.links, hostJobID)
	remove := r.deleter
	r.mu.Unlock()

	if remove != nil {
		_ = remove(hostJobID)
	}
}
