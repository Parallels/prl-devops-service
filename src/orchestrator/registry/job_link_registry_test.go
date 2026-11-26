package registry

import (
	"errors"
	"testing"
)

// The registry's whole job is to survive the gap between the replica that
// dispatched a create and the replica that hears about it finishing — while
// behaving exactly as it did before when nothing is wired underneath it.

func newRegistry() *JobLinkRegistry {
	return &JobLinkRegistry{links: make(map[string]JobLink)}
}

func TestJobLinkRegistryWithoutDurableStoreIsUnchanged(t *testing.T) {
	r := newRegistry()

	r.Register("h1", "o1", "host-1")
	link, ok := r.Lookup("h1")
	if !ok || link.OrchestratorJobID != "o1" || link.HostID != "host-1" {
		t.Fatalf("in-process link lost: %+v %v", link, ok)
	}

	if _, ok := r.Lookup("unknown"); ok {
		t.Fatal("unknown host job reported as linked")
	}

	r.Remove("h1")
	if _, ok := r.Lookup("h1"); ok {
		t.Fatal("link survived Remove")
	}
}

func TestJobLinkRegistryFallsBackToTheDurableStore(t *testing.T) {
	r := newRegistry()
	table := map[string]JobLink{"h1": {OrchestratorJobID: "o1", HostID: "host-1"}}
	var removed []string
	reads := 0

	r.SetDurableStore(
		func(hostJobID string, link JobLink) error {
			table[hostJobID] = link
			return nil
		},
		func(hostJobID string) (JobLink, bool, error) {
			reads++
			link, ok := table[hostJobID]
			return link, ok, nil
		},
		func(hostJobID string) error {
			delete(table, hostJobID)
			removed = append(removed, hostJobID)
			return nil
		},
	)

	// The event arrives here, but the dispatch happened somewhere else: the map
	// has nothing, the table has everything.
	link, ok := r.Lookup("h1")
	if !ok || link.OrchestratorJobID != "o1" {
		t.Fatalf("durable link not found: %+v %v", link, ok)
	}

	// And only once: the remaining events for the same host job must not each
	// pay for a query the map can already answer.
	afterFirst := reads
	if _, ok := r.Lookup("h1"); !ok || reads != afterFirst {
		t.Fatalf("fallback re-queried the store on every event (reads %d)", reads)
	}

	// Register writes both halves, so a link cannot exist in one and not the
	// other — which is the difference between a job that completes and one that
	// waits out its lease.
	r.Register("h2", "o2", "host-2")
	if stored, ok := table["h2"]; !ok || stored.OrchestratorJobID != "o2" {
		t.Fatalf("register did not persist: %+v", table)
	}

	// Remove clears both, or a finished job's link is resurrected by the next
	// lookup of an id the host may hand to someone else.
	r.Remove("h1")
	if len(removed) != 1 || removed[0] != "h1" {
		t.Fatalf("remove did not reach the store: %v", removed)
	}
	if _, ok := r.Lookup("h1"); ok {
		t.Fatal("link came back after Remove")
	}
}

func TestJobLinkRegistryToleratesAStoreThatCannotAnswer(t *testing.T) {
	r := newRegistry()
	r.SetDurableStore(nil, func(string) (JobLink, bool, error) {
		return JobLink{}, false, errors.New("database is down")
	}, nil)

	// A store that cannot answer is a miss, never a crash and never a false
	// positive: forwarding a host event to the wrong orchestrator job would
	// complete somebody else's work.
	if _, ok := r.Lookup("h1"); ok {
		t.Fatal("failed lookup reported as a hit")
	}

	// A nil writer is not a registration failure either — dispatch must not
	// notice whether bookkeeping is available.
	r.Register("h2", "o2", "host-2")
	if _, ok := r.Lookup("h2"); !ok {
		t.Fatal("in-process registration lost when no writer is wired")
	}
}
