package constants

// Placement domain (VM placement reservations). Orchestrator-side knobs; the
// host-side gate keys live here too so one operator surface covers both.
// Getter shape: config/placement.go, same tolerant-int posture as
// DbGhostJobTimeoutMinutes (config/database.go:46).
const (
	// PlacementReservationEnabledEnvVar is the orchestrator-side kill switch.
	// "false" restores pre-feature placement exactly, no rebuild required.
	PlacementReservationEnabledEnvVar = "PLACEMENT_RESERVATION_ENABLED"

	// PlacementReservationTTLEnvVar bounds how long an orphaned reservation may
	// survive its owner (crash safety net behind the terminal-state release hook).
	PlacementReservationTTLEnvVar = "PLACEMENT_RESERVATION_TTL_MINUTES"

	// PlacementReconcileIntervalEnvVar is the sweeper cadence in seconds.
	PlacementReconcileIntervalEnvVar = "PLACEMENT_RECONCILE_INTERVAL_SECONDS"

	// OrchestratorNodeIDEnvVar pins this replica's identity. Unset in orchestrator
	// mode falls back to this machine's unique id — the same one telemetry reports
	// as hardware_id — then to the hostname, then to a generated id (helpers.NodeID).
	OrchestratorNodeIDEnvVar = "ORCHESTRATOR_NODE_ID"

	// VMCPUAdmissionEnabledEnvVar opts a host into the CPU/memory admission gate.
	// Off by default: deployments that oversubscribe on purpose must not break on upgrade.
	VMCPUAdmissionEnabledEnvVar = "VM_CPU_ADMISSION_ENABLED"

	// VMCPUOvercommitRatioEnvVar multiplies the capacity the host gate compares
	// against. >1 deliberately re-allows oversubscription.
	VMCPUOvercommitRatioEnvVar = "VM_CPU_OVERCOMMIT_RATIO"
)

const (
	// DefaultPlacementReservationEnabled keeps enforcement on for orchestrators.
	// The failure it prevents is silent — two creates are both told a host has
	// room, and the only trace is a machine later running more VMs than it has
	// CPUs for — so a safeguard whose absence costs nothing until it is needed is
	// a safeguard that should be on. Rollback stays one environment variable.
	DefaultPlacementReservationEnabled = true

	// DefaultPlacementReservationTTLEnMinutes is the last-resort backstop, so it
	// is sized against the slowest legitimate create rather than the fastest dead
	// replica. A create that keeps reporting progress is alive by this repo's own
	// standard (jobs/tracker treats 15 minutes of silence as stale), and a clone
	// plus sysprep plus tools install off a remote catalog runs to the better part
	// of an hour. Expiring the lease before that hands the reserved capacity to
	// someone else while the create is still in flight, which re-opens exactly the
	// over-commit this feature exists to close.
	//
	// A long lease is affordable because it is not the mechanism anything relies
	// on day to day: the terminal hook releases on completion, and the reconciler
	// reclaims a reservation whose job is terminal or gone — the ghost-job
	// timeout (DefaultGhostJobTimeoutMinutes) gets a dead replica's job to
	// terminal in minutes, so the TTL only ever covers the case where even that
	// trail has disappeared.
	DefaultPlacementReservationTTLEnMinutes = 120

	// MaxPlacementReservationTTLEnMinutes is the ceiling past which "eventually
	// reclaimed" means "never". An operator who types 100000 means something other
	// than two months of held capacity.
	MaxPlacementReservationTTLEnMinutes = 24 * 60

	// DefaultPlacementReconcileIntervalSeconds is the sweeper cadence. Not
	// mirrorring the 5-second ghost-job watcher: that one watches for jobs that
	// stopped moving, which is urgent, and this one picks up what the terminal
	// hook already handled — so a miss costs thirty seconds against a lease
	// measured in hours, and thirty seconds is cheaper in queries than ten.
	DefaultPlacementReconcileIntervalSeconds = 30

	// DefaultVMCPUOvercommitRatio means "no oversubscription".
	DefaultVMCPUOvercommitRatio = 1.0
)

// Admission refusal reasons reported by placement.Store.Admit and rendered by
// the pre-flight check payload.
const (
	PlacementReasonCPU     = "cpu"
	PlacementReasonMemory  = "memory"
	PlacementReasonVMCount = "vm_count"
	PlacementReasonAppleVM = "apple_vm"
	PlacementReasonVMName  = "vm_name"
)
