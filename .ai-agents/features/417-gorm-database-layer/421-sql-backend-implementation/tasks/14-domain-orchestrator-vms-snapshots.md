# Task 14 — Orchestrator VMs, host snapshots, and RP views on GORM

**PBI** [#421](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 13 · **PR size:** M
**Plan refs:** D-9 (`ReplaceOrchestratorHostVMs` :871-888 full-sync; `SetHostVMSnapshots` :907-933 upsert) · D-11 · D-12 · #419 D-3 (VM child-table/serialised decision — consume, don't re-decide)

## Context

Second half of the orchestrator file's 31 interface methods: 14 in `orchestrator.go` + 2 in
`vm_snapshots.go` (counts verified by grep: `:548 :582 :608 :170 :630 :656 :685 :704 :732 :760 :786 :871 :890
:907` + `vm_snapshots.go:7 :25`). The interesting semantics: `ReplaceOrchestratorHostVMs` is a **full-sync**
(legacy swaps the host's whole VM slice — the SQL port is delete-where-host + insert-all in one tx),
`Upsert/RemoveOrchestratorHostVM` are the single-row twins, and the **RP views** (`GetOrchestratorReverseProxy*`)
are read-only projections that *join across domains* (hosts × reverse-proxy config — both GORM-side after
task 12, so the view composes two queries, never the legacy delegate).

## Objective

All 16 methods on GORM with parity; full-sync tx provably leaves no orphan/duplicate VM rows under the
harness; snapshot upsert provably insert-then-update.

## Files

| File | Op |
|---|---|
| `src/database/orchestrator_vm.go` | create (VM methods + RP views) |
| `src/database/snapshots.go` | create (`HostsVMSnapshotsRecord` + the 2 `vm_snapshots.go` methods) |
| `src/database/delegate_orchestrator.go` | delete remainder (all 31 now ported) |
| `src/database/delegate_vm_snapshots.go` | delete |
| `src/database/orchestrator_vm_parity_test.go`, `snapshots_parity_test.go` | create |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `src/data/orchestrator.go` / `src/data/vm_snapshots.go` | 933 / 40 (measured) | untouched | ok-watch / ok |
| `src/database/orchestrator_vm.go` | new | ~190 | ok; views split to `orchestrator_views.go` if >230 |
| `src/database/snapshots.go` | new | ~90 | ok |

## Implementation spec

**Inventory (16):** `GetOrchestratorVirtualMachines` :548 · `GetOrchestratorHostVirtualMachines` :582 ·
`GetOrchestratorHostVirtualMachine` :608 · `DeleteOrchestratorVirtualMachine` :170 · `UpsertOrchestratorHostVM`
:704 · `RemoveOrchestratorHostVM` :732 · `UpdateOrchestratorHostVMState` :760 · `UpdateOrchestratorHostVMUptime`
:786 · `ReplaceOrchestratorHostVMs` :871 · `GetHostVMSnapshots` :890 · `SetHostVMSnapshots` :907 ·
`GetOrchestratorReverseProxyHosts` :630 · `GetOrchestratorReverseProxyHost` :656 ·
`GetOrchestratorReverseProxyConfig` :685 · `SetListVMSnapshotsByVMId` (vm_snapshots.go:7) ·
`GetListVMSnapshotsByVMId` (:25).

1. **Per-VM storage shape follows #419** (D-3): if VMs are child-table rows → these are direct CRUD; if the
   decision was host-serialised column → every mutation is a tx'd read-modify-write of that column and
   **must** coordinate with task 13's field-scope rule (the heartbeat test from 13 is the regression guard).
   Implementer reads #419 mapping rows for `OrchestratorHost.VirtualMachines` first; spec both shapes'
   skeletons in the PR, implement the chosen one, delete the other.
2. **Full-sync (`ReplaceOrchestratorHostVMs` :871-888):**

```go
err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Where("host_id = ?", hostID).Delete(&models.OrchestratorVirtualMachine{}).Error; err != nil { return err }
    for i := range vms { vms[i].HostID = hostID }          // legacy stamps host ownership on swap
    return tx.CreateInBatches(&vms, 50).Error               // empty vms ⇒ delete-all only (legacy parity)
})
```

   Parity asserts post-state sets (by VM ID) equal to legacy on: replace-with-3, replace-with-0, replace
   overlapping-1.

3. **Single-VM twins:** `Upsert` :704 ⇒ `clause.OnConflict{Columns: vm-id, UpdateAll: true}` (idempotence
   contract per feature ER; **implementer to confirm legacy match key (:704-732)** — ID vs name-in-host);
   `Remove` :732 ⇒ delete by (host, vm) with not-found mapping; `State`/`Uptime` :760/:786 ⇒ field-scoped
   `Update` (same rule as 13 §4). `DeleteOrchestratorVirtualMachine` :170 — cross-host delete by VM id
   (legacy scans all hosts): SQL `WHERE id = ?`; multi-host duplicate ids impossible post-#419 unique index —
   **implementer to confirm #419 pinned vm-id uniqueness** (if not, delete-by-(host,vm) after host resolution).
4. **Snapshots:** `HostsVMSnapshotsRecord` per #419 (`map[string][]VMSnapshot` serialised vs normalised —
   consume mapping); `SetHostVMSnapshots` :907-933 upsert ⇒ `OnConflict{host_id}` UpdateAll (D-9's "insert-or-
   update" row); the two `vm_snapshots.go` methods are the global `VMSnapshots` collection — storage per #419
   (normalised table `vm_snapshots`, keyed VM id; `Set` = replace-per-VM idiom inside tx, same pattern as
   §2 mini-form).
5. **RP views (:630/:656/:685):** no new tables — compose task-12 host/config queries + active-host filter
   into the view structs (field mapping from legacy bodies **implementer to confirm :630-704**; projection
   is pure Go). Views must return `[]`/zero-config exactly like legacy when hosts exist but proxy is
   disabled.

## Out of scope

Host-row CRUD/aggregations (13), enrollment tokens (16), provider flip (17/18), legacy growth (forbidden).

## Validation

- [ ] `cd src && go build ./... && go vet ./database/...`
- [ ] `cd src && go test ./database/... -run "OrchestratorVM|Snapshot" -count=1 -race`
- [ ] Full-sync + upsert golden sets green on both engines; view fixtures with proxy disabled/enabled
- [ ] `cd src && go test ./... -count=1` (ER-5) — note: with task 04's delegates fully deleted for
      orchestrator+snapshots, the parity harness covers them; `git diff --stat src/data/` → empty
- [ ] `wc -l src/database/orchestrator_vm.go src/database/snapshots.go` ≤ guard

## Rollback

Revert restores the two delegate files' orchestrator/snapshot stubs; task 13's tests still pass (they
exercise host rows only — 13 §3 create-with-VMs overlap means its parity fixture stays, harmless).

## PBI acceptance mapping

**CRUD via GORM** — 16 methods incl. the two largest write flows in the PBI (full-sync, upsert);
**Transactions** — D-9 rows `:871-888`, `:907-933`; **Error handling** — D-11; **Legacy bypassed** — after
this PR the orchestrator domain contains zero delegate paths (structural evidence for 18's audit).

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-14`.** Cut `slice/421-14` from `sql-migration` HEAD
(`.worktrees/slice-421-14`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
