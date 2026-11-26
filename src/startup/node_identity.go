package startup

// Replica identity wiring.
//
// helpers cannot reach the machine's unique id or the settings service on its own
// — both sit above it, and a lower layer reaching up for its own identity would
// be a cycle as well as a bad idea — so startup hands it both, and does so before
// anything can stamp a row with an identity or recover one by it. The order is the
// substance of this file: helpers caches the identity on first use, so resolving
// it after the housekeeping block in main.go would have boot-time job recovery
// match on a fallback name the process is about to outgrow, which is the same
// stranding the node-scoped recovery exists to prevent.

import (
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/placement"
	"github.com/Parallels/prl-devops-service/settings"
)

// installNodeIdentity hands helpers its two external sources and settles the
// identity for this boot: the operator's value for this machine, else the
// ORCHESTRATOR_NODE_ID pin, else the machine's unique id, else the hostname.
//
// The machine id is worth the shell-out: it is what telemetry already reports as
// hardware_id, so it survives a container being re-created where a hostname does
// not. Two replicas resolving to one identity is not cosmetic — each one recovers
// (fails) the other's running jobs on boot and they share one reservation owner.
//
// The platform id arrives as a function rather than the service itself: it carries
// a mutex, and copying it into a helper is a bug vet will find for us.
func installNodeIdentity(ctx basecontext.ApiContext, uniqueID func() (string, error)) {
	if installed := helpers.SetMachineIDSource(func() (string, bool) {
		id, err := uniqueID()
		if err != nil {
			// Windows has no platform unique id implemented; the hostname takes
			// over, collisions and all. Saying so beats a node id the operator
			// cannot explain.
			ctx.LogWarnf("No machine unique id for node identity (%v); using the hostname", err)

			return "", false
		}

		return id, true
	}); !installed {
		ctx.LogWarnf("Machine id was already observed before startup finished; " +
			"node identity keeps the value derived without the platform id")
	}

	if svc := settings.Get(); svc != nil {
		if installed := helpers.SetNodeIDSource(func() (string, bool) {
			return svc.NodeScopedValue(constants.OrchestratorNodeIDEnvVar)
		}); !installed {
			ctx.LogWarnf("Node identity was already resolved before startup finished; " +
				"the per-machine settings override will not be applied until restart")
		}
	}

	// Resolve here rather than lazily: settings are loaded, and from this line on
	// every row this process writes is attributable.
	ctx.LogInfof("[Orchestrator] Node identity %q (machine %q)", helpers.NodeID(), helpers.MachineID())
}

// jobOwnerRenamer is the jobs-side half of a rename, reached by a LOCAL interface
// assertion — R-1 keeps DatabaseService closed, the same shape StartStaleJobWatcher
// and RecoverOngoingJobsForNode use.
type jobOwnerRenamer interface {
	RenameJobOwner(ctx basecontext.ApiContext, from, to string) error
}

// registerIdentityRenameHook makes a rename through the settings service live
// rather than pending: the process adopts the new identity and its own rows follow
// it. Without the row update the rename is half-done — recovery matches on
// owner_node, so jobs written under the retired name belong to a node that no
// longer exists, and nobody will ever finish them.
//
// Ordering is deliberate: the in-process identity flips FIRST, so a claim created
// during the window already carries the new name and is not missed by an update
// keyed on the old one. Both updates are idempotent (WHERE = old name), which is
// what makes a rename interrupted between them safe to simply run again.
func registerIdentityRenameHook(ctx basecontext.ApiContext, renamer jobOwnerRenamer) {
	settings.SetIdentityApplyHook(identityRenameHook(ctx, renamer))
}

// identityRenameHook is split out from registration so the ordering and the
// refusals can be tested without a settings service behind them.
func identityRenameHook(ctx basecontext.ApiContext, renamer jobOwnerRenamer) settings.IdentityApplyHook {
	return func(previous, current string) error {
		target := strings.TrimSpace(current)
		if target == "" {
			// The override was deleted: fall back to the same chain boot uses,
			// not to whatever name was in force a moment ago.
			target = helpers.DeriveNodeID()
		}

		from := strings.TrimSpace(previous)
		if from == "" {
			// No previous row — the process is already answering to something.
			from = helpers.NodeIDCached()
		}

		if from == target {
			return nil
		}

		if !helpers.ApplyNodeID(target) {
			return fmt.Errorf("identity %q is not usable; it was stored but this replica keeps %q until restart",
				target, from)
		}

		var failures []string
		if renamer != nil {
			if err := renamer.RenameJobOwner(ctx, from, target); err != nil {
				failures = append(failures, fmt.Sprintf("jobs: %v", err))
			}
		}
		// Resolved inside the call: placement is initialized by the orchestrator
		// service, which starts after this hook is registered.
		if err := placement.Get().RenameNodeIdentity(ctx, from, target); err != nil {
			failures = append(failures, fmt.Sprintf("reservations: %v", err))
		}

		ctx.LogInfof("[Orchestrator] Node identity renamed %q -> %q", from, target)

		if len(failures) > 0 {
			// The value IS stored and the identity IS adopted; what failed is the
			// bookkeeping behind it. Reported as a partial failure rather than
			// dressed up as a rollback — and both updates are keyed on the old
			// name, so writing the same rename again repairs what did not move.
			return fmt.Errorf("identity adopted, but its rows did not all follow (%s); re-save the same value to retry",
				strings.Join(failures, ", "))
		}

		return nil
	}
}
