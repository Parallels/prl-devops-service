package helpers

// Node identity for multi-replica deployments (orchestrator capacity
// reservations, node-scoped job recovery).
//
// Nothing identified a replica before this: RecoverOngoingJobs failed every
// running job in the table on boot, which is only safe while exactly one
// orchestrator exists.
//
// Two different things are resolved here and must not be confused:
//
//   - MachineID is what this *piece of hardware* is. It comes from the
//     platform (the hardware UUID telemetry also reports as hardware_id),
//     falls back to the hostname, and never changes when an operator renames
//     the node. It is the key an operator uses to address this replica in the
//     settings service, so it has to be stable and it has to be visible.
//
//   - NodeID is what this *replica is called* in the ledger. Resolution order
//     is the operator's setting for this machine, then the ORCHESTRATOR_NODE_ID
//     pin, then the machine id, then the hostname, then a generated id — the
//     last so that a container with no stable hostname still gets an identity
//     unique to the process rather than the empty string that would put every
//     replica in one bucket.
//
// Settings outrank the pin because that is the rule every other key follows
// (the DB layer beats env); the pin remains the escape hatch for a machine
// whose settings row has been mis-edited, and for a process that starts before
// the settings service exists.
//
// Both values come from injected resolvers rather than imports: serviceprovider
// and settings both sit above this package, so reaching up for either would be
// a cycle, and a lower layer must not choose its own identity.

import (
	"os"
	"strings"
	"sync"

	"github.com/Parallels/prl-devops-service/constants"
)

// maxIdentityBytes bounds an identity or machine id. The platform values are
// far shorter; the ceiling exists so a mistyped setting cannot smuggle a
// paragraph into an indexed column that is joined against in queries.
const maxIdentityBytes = 128

var (
	nodeIDOnce     sync.Once
	nodeIDCached   string
	nodeIDResolved bool
	nodeIDMu       sync.Mutex
	nodeIDSource   func() (string, bool)

	machineIDOnce   sync.Once
	machineIDCached string
	machineIDMu     sync.Mutex
	machineIDSource func() (string, bool)
)

// NodeID returns this process's replica identity, computed once.
func NodeID() string {
	nodeIDOnce.Do(func() {
		resolved := resolveNodeID()

		nodeIDMu.Lock()
		nodeIDCached = resolved
		nodeIDResolved = true
		nodeIDMu.Unlock()
	})

	return NodeIDCached()
}

// NodeIDCached reports the resolved identity without triggering resolution. It
// exists for callers that must not be the first to ask — a resolver installed
// after the first question is ignored, and a caller that can log about that is
// in a better position to say so than this package.
func NodeIDCached() string {
	nodeIDMu.Lock()
	defer nodeIDMu.Unlock()

	return nodeIDCached
}

// SetNodeIDSource installs the resolver for the operator-configured identity —
// the settings value scoped to this machine. It reports false when the identity
// has already been resolved and the resolver therefore cannot take effect, so
// the caller (which has a logger, unlike this package) can say so.
//
// The resolver runs once, lazily, on the first NodeID call: it may hit the
// database, and boot should not pay for an identity nobody asks for.
func SetNodeIDSource(source func() (string, bool)) bool {
	nodeIDMu.Lock()
	defer nodeIDMu.Unlock()

	if nodeIDResolved {
		return false
	}

	nodeIDSource = source

	return true
}

// SetMachineIDSource installs the resolver for the hardware identifier (the
// platform unique id). Like SetNodeIDSource it reports false once the machine
// id has been observed, because installing a provider afterwards would be
// silently ignored.
//
// It may shell out, so it is called lazily and its result cached — telemetry
// asks for the same value and would otherwise pay for it twice.
func SetMachineIDSource(source func() (string, bool)) bool {
	machineIDMu.Lock()
	defer machineIDMu.Unlock()

	if machineIDCached != "" {
		return false
	}

	machineIDSource = source

	return true
}

// MachineID returns this machine's stable identifier: the hardware UUID where
// the platform publishes one, the hostname where it does not, and the empty
// string if neither is available.
//
// The hostname fallback is what makes this usable on Windows, where the
// platform id is not implemented — at the cost of inheriting hostname
// collisions from containers, which is why the settings layer refuses to map
// two machines onto one identity.
func MachineID() string {
	machineIDOnce.Do(func() {
		resolved := resolveMachineID()

		machineIDMu.Lock()
		machineIDCached = resolved
		machineIDMu.Unlock()
	})

	machineIDMu.Lock()
	defer machineIDMu.Unlock()

	return machineIDCached
}

func resolveMachineID() string {
	machineIDMu.Lock()
	source := machineIDSource
	machineIDMu.Unlock()

	if source != nil {
		if id, ok := source(); ok {
			if normalized := normalizeIdentity(id); normalized != "" {
				return normalized
			}
		}
	}

	return hostname()
}

// resolveNodeID is split out so tests can drive every fallback without
// fighting the once.
//
// The pin is read straight from the environment rather than through
// config.GetKey: helpers sits below config (config imports helpers), and the
// central-settings layer would in any case be wrong for this knob if it were
// *shared* — every replica reads one database, so a single shared node id
// would hand the same identity to all of them. That is precisely why the
// settings resolver handed to SetNodeIDSource is scoped to this machine.
func resolveNodeID() string {
	nodeIDMu.Lock()
	source := nodeIDSource
	nodeIDMu.Unlock()

	if source != nil {
		if id, ok := source(); ok {
			if normalized := normalizeIdentity(id); normalized != "" {
				return normalized
			}
		}
	}

	return DeriveNodeID()
}

// DeriveNodeID is what this machine calls itself when nobody has told it
// otherwise: the pin, then the machine id, then the hostname, then a generated
// id. It is the same chain boot falls back to, exposed so the settings rename
// path can return an identity to its default when an operator deletes the
// override — without it, "unset the name" would have no defined answer.
//
// It deliberately does not consult the settings resolver: an override being
// removed must not read back as still set.
func DeriveNodeID() string {
	if pinned := strings.TrimSpace(os.Getenv(constants.OrchestratorNodeIDEnvVar)); pinned != "" {
		if normalized := normalizeIdentity(pinned); normalized != "" {
			return normalized
		}
	}

	if machine := MachineID(); machine != "" {
		return machine
	}

	if host := hostname(); host != "" {
		return host
	}

	return "node-" + GenerateId()
}

// NodeIDAliases returns every identity this machine may have stamped on its own
// work: the current one, its machine id, and its hostname.
//
// It exists for boot-time recovery. Rows written before a rename, or before
// identity moved from the hostname to the machine id, carry an identity this
// replica no longer answers to by default — and a row nobody owns is a row
// nobody will ever finish, since the terminal hook fires for the process that
// wrote it, not for one that has since been renamed. Every one of these values
// identifies this machine to itself, so claiming them costs nothing.
func NodeIDAliases() []string {
	candidates := []string{NodeID(), MachineID(), hostname()}

	seen := make(map[string]bool, len(candidates))
	aliases := make([]string, 0, len(candidates))

	for _, candidate := range candidates {
		if candidate == "" || seen[candidate] {
			continue
		}

		seen[candidate] = true
		aliases = append(aliases, candidate)
	}

	return aliases
}

// ApplyNodeID sets the replica identity after it has been resolved — the
// settings rename path. It returns false if the value was not usable, leaving
// the previous identity in place rather than silently dropping to an empty
// owner, which would re-merge every replica's jobs into one bucket.
func ApplyNodeID(id string) bool {
	normalized := normalizeIdentity(id)
	if normalized == "" {
		return false
	}

	nodeIDOnce.Do(func() {}) // consume the once so a later NodeID() keeps this value

	nodeIDMu.Lock()
	nodeIDCached = normalized
	nodeIDResolved = true
	nodeIDMu.Unlock()

	return true
}

// normalizeIdentity strips the quoting the platform adds (ioreg emits
// IOPlatformUUID quoted, and telemetry strips the same quotes) and rejects
// anything unusable rather than half-accepting it: control characters could
// break the log lines and SQL literals an identity ends up in, and a value
// past the ceiling is a mistype, not a name.
func normalizeIdentity(id string) string {
	trimmed := strings.Trim(strings.TrimSpace(id), `"`)
	if trimmed == "" || len(trimmed) > maxIdentityBytes {
		return ""
	}

	for _, r := range trimmed {
		if r < 0x20 || r == 0x7f {
			return ""
		}
	}

	return trimmed
}

func hostname() string {
	host, err := os.Hostname()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(host)
}

// ResetNodeIdentityForTests forgets the resolved identities and the installed
// sources so a test can drive the resolution chain from a clean slate. Production
// resolves once and never forgets; nothing but a test may call this.
func ResetNodeIdentityForTests() {
	nodeIDMu.Lock()
	nodeIDCached, nodeIDResolved, nodeIDSource = "", false, nil
	nodeIDMu.Unlock()

	machineIDMu.Lock()
	machineIDCached, machineIDSource = "", nil
	machineIDMu.Unlock()

	// Re-arming by assignment, not by copying a used lock.
	nodeIDOnce = sync.Once{}
	machineIDOnce = sync.Once{}
}

// SetNodeIDForTests pins the identity for a test and returns the previous
// value. Production code must not call it — production uses ApplyNodeID, which
// refuses an unusable value instead of taking whatever it is handed.
func SetNodeIDForTests(id string) string {
	nodeIDOnce.Do(func() {}) // ensure the once is consumed so Do will not overwrite

	var previous string
	nodeIDMu.Lock()
	previous = nodeIDCached
	nodeIDCached = id
	nodeIDResolved = true // a forced identity counts as resolved: a resolver installed after it is refused
	nodeIDMu.Unlock()

	return previous
}
