// Package placement owns the pending capacity claims that close the
// read-decide-dispatch window in orchestrator VM placement.
//
// The bug it fixes: checkHostRequirements reads a cached host.Resources
// snapshot, nothing is written between the decision and the host actually
// creating the VM, so two concurrent creates both read "N CPUs free" and both
// are accepted — see .ai-agents/features/vm-placement-reservations/brief.md.
//
// Design is settled in .ai-agents/features/vm-placement-reservations/plan.md.
// The Store interface below is the seam: the SQL implementation is the first
// one, and a Redis-backed implementation is expected to slot in behind it
// without any caller changing.
package placement

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
)

// CapacityView is a per-dimension capacity number. It is deliberately a value
// type: the admission decision combines three of these (host snapshot,
// reservations ledger, host-reported pending) and must stay readable.
type CapacityView struct {
	// CPU is logical CPUs.
	CPU int64
	// Memory is MiB, matching HostResourceItem.MemorySize.
	Memory float64
	// VMCount is the number of create operations.
	VMCount int64
	// AppleVM is the number of macOS create operations.
	AppleVM int64
}

// Add returns the element-wise sum. Used to fold host truth and ledger
// together where addition is the right semantic (never for pending, which
// combines with CombinePending — see the max() note there).
func (c CapacityView) Add(other CapacityView) CapacityView {
	return CapacityView{
		CPU:     c.CPU + other.CPU,
		Memory:  c.Memory + other.Memory,
		VMCount: c.VMCount + other.VMCount,
		AppleVM: c.AppleVM + other.AppleVM,
	}
}

// Sub returns the element-wise difference, clamped at zero: a host already
// reporting negative availability must not be offered as if it had headroom.
func (c CapacityView) Sub(other CapacityView) CapacityView {
	return CapacityView{
		CPU:     maxInt64(0, c.CPU-other.CPU),
		Memory:  maxFloat(0, c.Memory-other.Memory),
		VMCount: maxInt64(0, c.VMCount-other.VMCount),
		AppleVM: maxInt64(0, c.AppleVM-other.AppleVM),
	}
}

// CombinePending merges the two independent sources of in-flight demand: the
// orchestrator's own reservation ledger and whatever the host itself publishes
// as pending. It takes the MAX per dimension, never the sum: once the host
// publishes PendingCpu/PendingMemory, a create placed by this orchestrator is
// known to both sides, and summing would charge it twice — rejecting creates
// that genuinely fit.
func CombinePending(ledger, hostReported CapacityView) CapacityView {
	return CapacityView{
		CPU:     maxInt64(ledger.CPU, hostReported.CPU),
		Memory:  maxFloat(ledger.Memory, hostReported.Memory),
		VMCount: maxInt64(ledger.VMCount, hostReported.VMCount),
		AppleVM: maxInt64(ledger.AppleVM, hostReported.AppleVM),
	}
}

// Reservation describes what one create intends to consume.
type Reservation struct {
	// JobID is the orchestrator job ID and the primary key of the ledger row.
	JobID  string
	HostID string
	// VmName and Owner carry the cluster-wide name guard. They are required:
	// an unnameable create cannot be guarded against a concurrent twin.
	VmName string
	Owner  string
	CPU    int64
	Memory float64
	// VMCount is normally 1 for a create; AppleVM is 1 for a macvm create.
	VMCount int64
	AppleVM int64
	// NodeID is the replica that placed it (multi-replica diagnostics).
	NodeID string
}

// Admission is the verdict of an atomic check-and-claim.
type Admission struct {
	Admitted bool
	// Bypassed marks a verdict produced while enforcement is off (operator
	// switch or no SQL backend). It is Admitted, but it claims nothing: logs
	// and tests use it to tell "enforced and passed" apart from "not enforced".
	Bypassed bool
	// Available is what the host had free for this request after subtracting
	// pending, so the caller can build the same user-facing message the
	// read-only checks produce.
	Available CapacityView
	// Reason is one of the constants.PlacementReason* values, "" when admitted.
	Reason string
	// Message is the human-readable refusal, shaped like the existing
	// checkHostRequirements output ("CPU: required 6 logical CPUs, available 3").
	Message string
}

// Outcome distinguishes the two release paths: success refreshes the host
// snapshot before deleting the row, failure deletes immediately.
type Outcome string

const (
	// OutcomeSucceeded releases after the host has been re-read.
	OutcomeSucceeded Outcome = "succeeded"
	// OutcomeFailed releases straight away; a refresh is best effort.
	OutcomeFailed Outcome = "failed"
)

// Store is the persistence seam for pending capacity claims. Every method must
// fail closed: an error is not "unknown, carry on", it is "cannot be measured,
// do not overbook" — the posture the host-side count gates already take in
// serviceprovider/parallelsdesktop/vm_capacity.go.
//
// Implementations are safe for concurrent use and must behave correctly when
// several orchestrator replicas call them against the same database.
type Store interface {
	// Admit atomically checks capacity and records the claim, serialised per
	// host. It performs no network I/O: the per-host lock is held only across
	// database statements (plan D10).
	Admit(ctx basecontext.ApiContext, hostID string, res Reservation, hostReportedPending CapacityView) (Admission, error)

	// Release deletes a claim. It is idempotent: releasing an unknown job is
	// success, never an error, because it runs from job terminal hooks where a
	// duplicate call is normal.
	Release(ctx basecontext.ApiContext, jobID string, outcome Outcome) error

	// PendingByHost aggregates outstanding claims for one host. A host with
	// nothing pending returns a zero view and no error.
	PendingByHost(ctx basecontext.ApiContext, hostID string) (CapacityView, error)

	// PendingByHosts aggregates outstanding claims for every host in one read,
	// so placement does not issue one query per candidate.
	PendingByHosts(ctx basecontext.ApiContext) (map[string]CapacityView, error)

	// HostOf returns the host a claim is against ("" when the job has none).
	// Release needs it to refresh the right host before deleting the row.
	HostOf(ctx basecontext.ApiContext, jobID string) (string, error)

	// NameTaken reports whether owner/vmName is already claimed by another
	// job's pending reservation or already exists as that owner's VM.
	// exceptJobID exempts a job re-placing itself onto another host.
	NameTaken(ctx basecontext.ApiContext, owner, vmName, exceptJobID string) (bool, error)

	// ExpireLeases deletes claims past their lease and returns how many went.
	ExpireLeases(ctx basecontext.ApiContext, now time.Time) (int64, error)

	// RenameNode re-attributes this replica's claims from one identity to
	// another and returns how many moved. It runs when an operator renames a
	// replica through the settings service, so that the ledger keeps naming a
	// replica that exists. Part of the store contract rather than a SQL-only
	// extra: a backend that cannot re-key its rows would strand them exactly
	// the same way, and a Redis store must answer for it too.
	RenameNode(ctx basecontext.ApiContext, from, to string) (int64, error)

	// ReconcileOrphans deletes claims whose job is terminal or gone.
	ReconcileOrphans(ctx basecontext.ApiContext) (int64, error)
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}

	return b
}
