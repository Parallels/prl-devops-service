package models

// Placement domain: pending capacity reservations that close the
// read-decide-dispatch window in orchestrator VM placement.
//
// Reservations live in their own tables and NEVER inside
// orchestrator_hosts.resources: that blob is host-owned and is overwritten
// wholesale by every hardware refresh (orchestrator/main.go updateHostWithHardwareInfo),
// which is exactly the timing that produces the overcommit they prevent.
//
// Rows are hard-deleted on release, so "rows present == pending demand" and
// the aggregate an admission decision needs is a single SUM with no status
// column to reason about.

// ResourceReservation is one outstanding capacity claim, keyed by the
// orchestrator job that made it. One row per job: re-placing a create from
// host A to host B updates host_id instead of adding a second charge, and a
// job can never leak an extra row.
type ResourceReservation struct {
	// JobID is the orchestrator job ID that owns this claim. Primary key, which
	// is what makes "at most one charge per create" structural.
	JobID string `json:"job_id" gorm:"primaryKey"`
	// HostID is the host the claim is against. Indexed: every admission reads
	// the per-host aggregate.
	HostID string `json:"host_id" gorm:"index"`
	// Owner + VmName carry the cluster-wide name guard. The unique index makes
	// "one pending create per name per owner" a database invariant rather than
	// only a check inside the transaction; the transaction check still runs
	// first so the user gets a readable refusal instead of a constraint error.
	Owner     string  `json:"owner" gorm:"uniqueIndex:ux_reservation_owner_vmname"`
	VmName    string  `json:"vm_name" gorm:"uniqueIndex:ux_reservation_owner_vmname"`
	CPU       int64   `json:"cpu"`
	Memory    float64 `json:"memory"` // MiB, matches HostResourceItem.MemorySize
	VMCount   int64   `json:"vm_count"`
	AppleVM   int64   `json:"apple_vm"`
	NodeID    string  `json:"node_id" gorm:"index"` // replica that placed it
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	// ExpiresAt is the lease deadline, written in fixed-width RFC3339
	// (time.RFC3339, seconds precision) so SQL string comparison is also
	// chronological. GetUtcCurrentDateTime's RFC3339Nano trims trailing zeros,
	// which is fine for display and not fine for comparison.
	ExpiresAt string `json:"expires_at" gorm:"index"`
	*DbRecord `json:"db_record" gorm:"-:all"`
}

// TableName pins the name so the store's hand-written SQL and AutoMigrate
// cannot drift apart.
func (ResourceReservation) TableName() string { return "resource_reservations" }

// PlacementGuard is the per-host serialisation row. The admission transaction
// takes the lock by updating this row: on Postgres the UPDATE holds a
// row-exclusive lock until COMMIT, on SQLite the single-writer transaction
// lock covers the block. Deliberately not pg_advisory_xact_lock — that is
// Postgres-only and the data layer must stay dialect-neutral.
type PlacementGuard struct {
	HostID    string `json:"host_id" gorm:"primaryKey"`
	Seq       int64  `json:"seq"` // monotonic counter; the value is irrelevant, the lock is the point
	Token     string `json:"token"`
	UpdatedAt string `json:"updated_at"`
	*DbRecord `json:"db_record" gorm:"-:all"`
}

// TableName pins the guard table name.
func (PlacementGuard) TableName() string { return "placement_guards" }

// JobLink is the durable copy of the in-process JobLinkRegistry. Every replica
// dials every host's WebSocket, so the replica that receives a host job event
// is not necessarily the one that dispatched it; without a shared table the
// event is dropped (orchestrator/handlers/host_job_event_handler.go) and the
// orchestrator job — and its reservation — hangs until the lease expires.
type JobLink struct {
	HostJobID         string `json:"host_job_id" gorm:"primaryKey"`
	OrchestratorJobID string `json:"orchestrator_job_id" gorm:"index"`
	HostID            string `json:"host_id" gorm:"index"`
	CreatedAt         string `json:"created_at"`
	*DbRecord         `json:"db_record" gorm:"-:all"`
}

// TableName pins the job links table name.
func (JobLink) TableName() string { return "job_links" }
