package registry

import "sync"

// JobLink maps a host-side job back to the orchestrator job that spawned it.
type JobLink struct {
	OrchestratorJobID string
	HostID            string
}

// JobLinkRegistry is an in-memory singleton that tracks the 1:1 relationship
// between an orchestrator job and the host job it dispatched.
// Keys are host job IDs; values carry the orchestrator job ID and host ID.
type JobLinkRegistry struct {
	mu    sync.RWMutex
	links map[string]JobLink
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
	r.mu.Lock()
	defer r.mu.Unlock()
	r.links[hostJobID] = JobLink{OrchestratorJobID: orchJobID, HostID: hostID}
}

// Lookup returns the JobLink for a host job ID, if one exists.
func (r *JobLinkRegistry) Lookup(hostJobID string) (JobLink, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	link, ok := r.links[hostJobID]
	return link, ok
}

// Remove deletes the link for a host job ID (called on terminal state).
func (r *JobLinkRegistry) Remove(hostJobID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.links, hostJobID)
}
