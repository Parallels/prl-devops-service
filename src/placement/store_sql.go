package placement

// SQL persistence for the reservation ledger. Deliberately does NOT extend the
// frozen datainterfaces.DatabaseService surface (#421 R-1): like the settings
// domain (settings/store.go) it talks to the raw handle from
// serviceprovider.GetSQLDatabase(). Dialect-neutral SQL only — no rowid, no
// ON CONFLICT, no pg_* — because the same statements must run on SQLite and
// Postgres, and because placement_guards is the portable substitute for
// pg_advisory_xact_lock.

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"gorm.io/gorm"
)

// leaseTimeFormat is fixed-width RFC3339 (seconds precision) so SQLite and
// Postgres both compare expiry lexicographically and get chronological order.
// helpers.GetUtcCurrentDateTime() is RFC3339Nano, which trims trailing zeros
// and therefore is display-only here.
const leaseTimeFormat = time.RFC3339

// ErrStoreUnavailable means the domain is inert: legacy JSON DB mode, or the
// SQL handle was never opened. Callers treat it as "cannot measure", which by
// the fail-closed rule means "do not overbook".
var ErrStoreUnavailable = errors.New("placement: sql database unavailable")

type sqlStore struct {
	db *gorm.DB
}

func newSQLStore(db *gorm.DB) *sqlStore {
	return &sqlStore{db: db}
}

func (s *sqlStore) available() bool {
	return s != nil && s.db != nil
}

// Admit performs check-and-claim in a single transaction serialised per host.
//
// The serialisation is the UPDATE on placement_guards, which must come first:
// on Postgres it takes a row-exclusive lock held to COMMIT, so a second
// admission for the same host blocks there and then reads the first one's
// committed insert below; on SQLite the single-writer transaction lock covers
// the whole block. Different hosts never contend.
//
// There is deliberately no network I/O in here (plan D10): the lock is held
// across database statements only, never across a disk probe or a latency ping.
func (s *sqlStore) Admit(ctx basecontext.ApiContext, hostID string, res Reservation,
	hostReportedPending CapacityView) (Admission, error) {
	if !s.available() {
		return Admission{}, ErrStoreUnavailable
	}
	hostID = strings.TrimSpace(hostID)
	if hostID == "" {
		return Admission{}, errors.New("placement: host id is required")
	}
	if strings.TrimSpace(res.JobID) == "" {
		return Admission{}, errors.New("placement: job id is required")
	}

	res.VmName = strings.TrimSpace(res.VmName)
	res.Owner = strings.TrimSpace(res.Owner)
	if res.VMCount == 0 {
		res.VMCount = 1
	}

	var admission Admission
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := lockHost(tx, hostID); err != nil {
			return err
		}

		host, err := loadHostResources(tx, hostID)
		if err != nil {
			return err
		}

		ledger, err := pendingOnHost(tx, hostID)
		if err != nil {
			return err
		}

		pending := CombinePending(ledger, hostReportedPending)
		available := host.sub(pending)
		admission.Available = available

		if reason, message, ok := fits(available, pending, host, res); !ok {
			admission.Reason = reason
			admission.Message = message
			ctx.LogDebugf("[Placement] Refused job=%s host=%s reason=%s %s", res.JobID, hostID, reason, message)
			return nil // a refusal is a valid verdict, not a transaction failure
		}

		taken, err := nameTaken(tx, res.Owner, res.VmName, res.JobID)
		if err != nil {
			return err
		}
		if taken {
			admission.Reason = constants.PlacementReasonVMName
			admission.Message = fmt.Sprintf("VM name %q is already in use by owner %s", res.VmName, res.Owner)
			return nil
		}

		now := helpers.GetUtcCurrentDateTime()
		row := &data_models.ResourceReservation{
			JobID:     res.JobID,
			HostID:    hostID,
			Owner:     res.Owner,
			VmName:    res.VmName,
			CPU:       res.CPU,
			Memory:    res.Memory,
			VMCount:   res.VMCount,
			AppleVM:   res.AppleVM,
			NodeID:    res.NodeID,
			CreatedAt: now,
			UpdatedAt: now,
			ExpiresAt: leaseDeadline(),
		}
		if err := upsertReservation(tx, row); err != nil {
			return err
		}

		admission.Admitted = true
		ctx.LogInfof("[Placement] Admitted job=%s host=%s cpu=%d memory=%.0f pending_cpu=%d available_cpu=%d",
			res.JobID, hostID, res.CPU, res.Memory, pending.CPU, available.CPU)
		return nil
	})
	if err != nil {
		return Admission{}, err
	}

	return admission, nil
}

// Release deletes the claim. Idempotent by contract: an unknown job is a
// success, because this runs from job terminal hooks where a duplicate or a
// late arrival is normal.
func (s *sqlStore) Release(ctx basecontext.ApiContext, jobID string, outcome Outcome) error {
	if !s.available() {
		return ErrStoreUnavailable
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil
	}

	res := s.db.Exec("DELETE FROM resource_reservations WHERE job_id = ?", jobID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		ctx.LogInfof("[Placement] Released job=%s outcome=%s", jobID, outcome)
	}

	return nil
}

// PendingByHost aggregates the ledger for one host. A host with nothing
// pending returns a zero view, never an error.
func (s *sqlStore) PendingByHost(ctx basecontext.ApiContext, hostID string) (CapacityView, error) {
	if !s.available() {
		return CapacityView{}, ErrStoreUnavailable
	}

	return pendingOnHost(s.db, strings.TrimSpace(hostID))
}

// PendingByHosts aggregates the whole ledger in one read so placement can
// subtract pending from every candidate without a query per host.
func (s *sqlStore) PendingByHosts(ctx basecontext.ApiContext) (map[string]CapacityView, error) {
	if !s.available() {
		return nil, ErrStoreUnavailable
	}

	var rows []struct {
		HostID  string
		CPU     int64
		Memory  float64
		VMCount int64
		AppleVM int64
	}
	err := s.db.Table("resource_reservations").
		Select("host_id, COALESCE(SUM(cpu),0) AS cpu, COALESCE(SUM(memory),0) AS memory, " +
			"COUNT(*) AS vm_count, COALESCE(SUM(apple_vm),0) AS apple_vm").
		Group("host_id").Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]CapacityView, len(rows))
	for _, row := range rows {
		result[row.HostID] = CapacityView{CPU: row.CPU, Memory: row.Memory, VMCount: row.VMCount, AppleVM: row.AppleVM}
	}

	return result, nil
}

// HostOf reads the host a claim is against, so Release can refresh that host
// before deleting the row instead of guessing.
func (s *sqlStore) HostOf(ctx basecontext.ApiContext, jobID string) (string, error) {
	if !s.available() {
		return "", ErrStoreUnavailable
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return "", nil
	}

	var hostID string
	err := s.db.Table("resource_reservations").Where("job_id = ?", jobID).Pluck("host_id", &hostID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}

	return hostID, err
}

// NameTaken is the read-only probe used by the pre-flight endpoint; the
// admission path uses the same helper inside its transaction.
func (s *sqlStore) NameTaken(ctx basecontext.ApiContext, owner, vmName, exceptJobID string) (bool, error) {
	if !s.available() {
		return false, ErrStoreUnavailable
	}

	return nameTaken(s.db, strings.TrimSpace(owner), strings.TrimSpace(vmName), strings.TrimSpace(exceptJobID))
}

// ExpireLeases sweeps claims whose owner never released them (crashed replica,
// killed process). Set-based, so concurrent replicas sweeping is harmless.
func (s *sqlStore) ExpireLeases(ctx basecontext.ApiContext, now time.Time) (int64, error) {
	if !s.available() {
		return 0, ErrStoreUnavailable
	}

	cutoff := now.UTC().Format(leaseTimeFormat)
	res := s.db.Exec("DELETE FROM resource_reservations WHERE expires_at < ?", cutoff)
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected > 0 {
		ctx.LogWarnf("[Placement] Expired %d reservation(s) past their lease", res.RowsAffected)
	}

	return res.RowsAffected, nil
}

// ReconcileOrphans sweeps claims whose job is already terminal or gone — the
// path a jobs-manager hook cannot cover, because DetectStaleJobs and
// RecoverOngoingJobs transition rows directly in the database.
// RenameNode retires one of this replica's identities in the ledger.
//
// Nothing filters on node_id today — it records who placed a reservation rather
// than gating anything — so unlike jobs.owner_node this is not a correctness
// repair, it is the difference between a ledger an operator can read after a
// rename and one that attributes live work to a replica that no longer exists.
// That is also why a store that cannot rename must say so loudly rather than
// let the identity move on without its rows.
//
// Idempotent (WHERE node_id = from); empty or unchanged identities are refused,
// because an empty `from` would match rows whose owner was never recorded.
func (s *sqlStore) RenameNode(ctx basecontext.ApiContext, from, to string) (int64, error) {
	if !s.available() {
		return 0, ErrStoreUnavailable
	}

	from, to = strings.TrimSpace(from), strings.TrimSpace(to)
	if from == "" || to == "" || from == to {
		return 0, interfaces.ErrNodeIdentityUnusable
	}

	res := s.db.Exec("UPDATE resource_reservations SET node_id = ? WHERE node_id = ?", to, from)
	if res.Error != nil {
		return 0, res.Error
	}
	if res.RowsAffected > 0 {
		ctx.LogInfof("[Placement] Re-attributed %d reservation(s) from node %q to %q",
			res.RowsAffected, from, to)
	}

	return res.RowsAffected, nil
}

func (s *sqlStore) ReconcileOrphans(ctx basecontext.ApiContext) (int64, error) {
	if !s.available() {
		return 0, ErrStoreUnavailable
	}

	var jobIDs []string
	if err := s.db.Table("resource_reservations").Pluck("job_id", &jobIDs).Error; err != nil {
		return 0, err
	}
	if len(jobIDs) == 0 {
		return 0, nil
	}

	live := make(map[string]bool)
	var liveIDs []string
	err := s.db.Model(&data_models.Job{}).
		Where("id IN ? AND state IN ?", jobIDs,
			[]constants.JobState{constants.JobStateInit, constants.JobStatePending, constants.JobStateRunning}).
		Pluck("id", &liveIDs).Error
	if err != nil {
		return 0, err
	}
	for _, id := range liveIDs {
		live[id] = true
	}

	orphan := make([]string, 0, len(jobIDs))
	for _, id := range jobIDs {
		if !live[id] {
			orphan = append(orphan, id)
		}
	}
	if len(orphan) == 0 {
		return 0, nil
	}

	res := s.db.Exec("DELETE FROM resource_reservations WHERE job_id IN ?", orphan)
	if res.Error != nil {
		return 0, res.Error
	}
	ctx.LogWarnf("[Placement] Reconciled %d orphan reservation(s)", res.RowsAffected)

	return res.RowsAffected, nil
}

// ── transaction helpers ─────────────────────────────────────────────────────

// lockHost serialises admissions for one host. The INSERT ... WHERE NOT EXISTS
// seed is the race-closed idiom settings/store.go uses for its singleton; the
// UPDATE is what actually takes the lock (and bumps a counter purely so the
// statement has something to write).
func lockHost(tx *gorm.DB, hostID string) error {
	now := helpers.GetUtcCurrentDateTime()
	seed := tx.Exec(
		"INSERT INTO placement_guards (host_id, seq, token, updated_at) "+
			"SELECT ?, 0, ?, ? WHERE NOT EXISTS (SELECT 1 FROM placement_guards WHERE host_id = ?)",
		hostID, now, now, hostID)
	if seed.Error != nil {
		return seed.Error
	}

	return tx.Exec("UPDATE placement_guards SET seq = seq + 1, token = ?, updated_at = ? WHERE host_id = ?",
		now, now, hostID).Error
}

// hostCapacity is the host's own view of what it has, taken from the cached
// snapshot. Reading it inside the transaction is the point: the value cannot
// change between the decision and the insert.
type hostCapacity struct {
	available  CapacityView
	totalVms   int64
	maxVms     int64
	maxVmsMode string
}

func (h hostCapacity) sub(pending CapacityView) CapacityView {
	return h.available.Sub(pending)
}

func loadHostResources(tx *gorm.DB, hostID string) (hostCapacity, error) {
	var host data_models.OrchestratorHost
	if err := tx.Where("id = ?", hostID).Take(&host).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return hostCapacity{}, fmt.Errorf("%w: host %s not found", ErrStoreUnavailable, hostID)
		}

		return hostCapacity{}, err
	}

	capacity := hostCapacity{}
	if host.Resources != nil {
		capacity.available = CapacityView{
			CPU:     host.Resources.TotalAvailable.LogicalCpuCount,
			Memory:  host.Resources.TotalAvailable.MemorySize,
			VMCount: host.Resources.TotalAvailable.TotalVms,
		}
		capacity.totalVms = host.Resources.Total.TotalVms
		capacity.maxVms = host.Resources.MaxVmsPerHost
		capacity.maxVmsMode = host.Resources.MaxVmsPerHostMode
		if strings.EqualFold(capacity.maxVmsMode, constants.MAX_VMS_PER_HOST_MODE_RUNNING) {
			capacity.totalVms = host.Resources.TotalInUse.TotalVms
		}
	}

	return capacity, nil
}

func pendingOnHost(tx *gorm.DB, hostID string) (CapacityView, error) {
	var rows []struct {
		CPU     int64
		Memory  float64
		VMCount int64
		AppleVM int64
	}
	err := tx.Table("resource_reservations").
		Select("COALESCE(SUM(cpu),0) AS cpu, COALESCE(SUM(memory),0) AS memory, "+
			"COUNT(*) AS vm_count, COALESCE(SUM(apple_vm),0) AS apple_vm").
		Where("host_id = ?", hostID).Scan(&rows).Error
	if err != nil {
		return CapacityView{}, err
	}
	if len(rows) == 0 {
		return CapacityView{}, nil
	}

	return CapacityView{CPU: rows[0].CPU, Memory: rows[0].Memory, VMCount: rows[0].VMCount, AppleVM: rows[0].AppleVM}, nil
}

// fits compares the request against what is left. Reasons match the
// constants.PlacementReason* set so the caller can render the same user-facing
// check payload the read-only path produces.
//
// The VM-count limit deliberately mirrors check_host_requirements
// (check_host_requirements.go:169-187): current VMs from Total (or TotalInUse
// in "running" mode) plus pending against the host's own MAX_VMS_PER_HOST,
// rather than inventing a second policy on top of TotalAvailable.
func fits(available, pending CapacityView, host hostCapacity, res Reservation) (string, string, bool) {
	if res.CPU > 0 && available.CPU < res.CPU {
		return constants.PlacementReasonCPU,
			fmt.Sprintf("CPU: required %d logical CPUs, available %d", res.CPU, available.CPU), false
	}
	if res.Memory > 0 && available.Memory < res.Memory {
		return constants.PlacementReasonMemory,
			fmt.Sprintf("Memory: required %.0f MiB, available %.0f MiB", res.Memory, available.Memory), false
	}
	if host.maxVms > 0 && host.totalVms+pending.VMCount+res.VMCount > host.maxVms {
		return constants.PlacementReasonVMCount,
			fmt.Sprintf("VM limit: current %d, pending %d, maximum %d",
				host.totalVms, pending.VMCount+res.VMCount, host.maxVms), false
	}

	return "", "", true
}

// nameTaken checks the two places a name can be spoken for: another owner's
// pending create on any host (cluster-wide, so twins landing on different
// hosts are caught) and an already-registered VM. exceptJobID exempts a job
// re-placing itself onto a different host.
//
// The virtual_machines lookup goes through the model with a map condition so
// GORM quotes the "user" column per dialect — unquoted it is a reserved word
// in Postgres and would silently compare against current_user.
func nameTaken(tx *gorm.DB, owner, vmName, exceptJobID string) (bool, error) {
	if owner == "" || vmName == "" {
		return false, nil
	}

	var pending int64
	err := tx.Table("resource_reservations").
		Where(map[string]interface{}{"owner": owner, "vm_name": vmName}).
		Where("job_id <> ?", exceptJobID).Count(&pending).Error
	if err != nil {
		return false, err
	}
	if pending > 0 {
		return true, nil
	}

	var existing int64
	err = tx.Model(&data_models.VirtualMachine{}).
		Where(map[string]interface{}{"user": owner, "name": vmName}).Count(&existing).Error
	if err != nil {
		return false, err
	}

	return existing > 0, nil
}

// upsertReservation keeps "one charge per job" true when a create is re-placed
// from one host to another: the primary key is the job id, so a second Admit
// for the same job moves the charge instead of adding one.
func upsertReservation(tx *gorm.DB, row *data_models.ResourceReservation) error {
	var existing int64
	err := tx.Table("resource_reservations").Where("job_id = ?", row.JobID).Count(&existing).Error
	if err != nil {
		return err
	}
	if existing == 0 {
		return tx.Create(row).Error
	}

	return tx.Model(&data_models.ResourceReservation{}).Where("job_id = ?", row.JobID).
		Updates(map[string]interface{}{
			"host_id": row.HostID, "cpu": row.CPU, "memory": row.Memory,
			"vm_count": row.VMCount, "apple_vm": row.AppleVM,
			"updated_at": row.UpdatedAt, "expires_at": row.ExpiresAt,
		}).Error
}

func leaseDeadline() string {
	return time.Now().UTC().Add(leaseTTL()).Format(leaseTimeFormat)
}
