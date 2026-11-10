# Task 04 — Relational collections (users/roles/claims joins, catalog, orchestrator, snapshots)

**PBI** [#424](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 03 **and #419's
merged mapping-table document** (`../419-gorm-model-tags/plan.md`) — hard blocker: every join table,
FK column, and JSON-serialised field below is transcribed from that document at implementation time;
**this spec deliberately contains no column names** (D-7) · **PR size:** M
**Plan refs:** [D-6](../plan.md#d-6--write-path-419-tagged-models--clauseonconflict-upserts-on-the-418-handle), [D-7](../plan.md#d-7--mappinggo_consumes_419s_mapping_document_verbatim_nothing_invented)

## Context

The legacy store embeds relationships inside documents: `User.Roles`/`User.Claims`
(`src/data/models/user.go:11-12`), `Role.Claims` (`role.go:8-9`), `CatalogManifest`
`VirtualMachineContents`/`PackContents` (`catalog_manifest.go:28-29`), `OrchestratorHost`
`VirtualMachines`/`CacheItems` (`orchestrator_host.go:36,43`), `HostsVMSnapshotsRecord.VMSnapshots
map[string][]VMSnapshot` (`orchestrator_vm_snapshots.go:6`), `VMSnapshots.VMSnapshot []VMSnapshot`
(`VM_Snapshots.go:14`), `Job.Steps []JobStep` (`job.go:16`). #419 decided per field whether it becomes a
join table, a child table, or a serialised JSON column ("no data loss on round-trip" is its binding
AC). This task turns those decisions into the `relationalPlan()` half of `collectionPlan()`, replacing
task 03's placeholders.

## Objective

Every one of the 16 `Data` collections migrates end-to-end from `testdata/golden.json`, relationships
inclusive, inside the same single-transaction/upsert engine; FK-safe deterministic order; a user's role
and claim edges survive a re-run without duplicates.

## Files

| File | Op |
|---|---|
| `src/database/legacy_migrator/mapping_relational.go` | create (ER-2 split from `mapping.go` per task-03 note if needed; same package) |
| `src/database/legacy_migrator/mapping.go` | modify — `collectionPlan()` returns full ordered list |
| `src/database/legacy_migrator/mapping_relational_test.go` | create |
| `src/database/legacy_migrator/testdata/golden.json` | modify — enrich: users with multiple roles/claims overlapping across users, manifest with nested contents, orchestrator host with VMs, snapshot map |

## File size guard

Real `wc -l` at plan time: `mapping.go` (new in 03) planned 150; relational half planned ≈ 260 — total
≈ 410 ⇒ **split preemptively** into `mapping_relational.go` (D-1 split plan; stays ≪ ER-2's 1,000,
keeps files reviewable). Others new ≈ 260 / 220 test / fixture < 200. Verdict: **ok** (with split).

## Implementation spec

Order rule (deterministic, D-7): parents before children by #419's own table list —
`Claims`/`Roles` (top-level collections) → `Users` (+ edges) → `ApiKeys` (+ any key-claims edges #419
defines) → `Configuration` → `PackerTemplates` → `ManifestsCatalog` (+contents) → `OrchestratorHosts`
(+VMs/cache) → `HostsVMSnapshots`/`VMSnapshots` → `ReverseProxy`/`ReverseProxyHosts` →
`CatalogManagers` → `Jobs` (+steps) → `EnrollmentTokens` → `UserConfigs`. Where #419 serialises a slice
as a JSON column, the spec row degrades to the scalar path (fewer rows) — the doc decides; the code
transcribes.

Per-relationship pattern (sketched for users↔roles↔claims; same engine everywhere):

```go
// inside one CollectionSpec.Migrate, running on the task-03 tx:
// 1) upsert the parent rows (Users) — engine from 03 (upsertAll on the PK from #419 doc).
// 2) dedupe child entities: legacy embeds FULL Role/Claim documents inside each User AND keeps
//    top-level Roles/Claims collections (src/data/main.go:37-38). Canonical entity rows come from
//    the top-level collections; embedded copies contribute EDGES only — keyed by the child ID per
//    #419 doc. Embedded children whose ID is absent from the top-level collection are ALSO upserted
//    as entities (orphan-preserving; "no data loss" wins over dedupe cleverness).
// 3) edges: build []user_roles / []user_claims-style row structs (#419 names them) and
//    tx.Clauses(clause.OnConflict{Columns: {user_id, role_id}, Ignore: true}).CreateInBatches(...)
//    — ignore-on-conflict, NEVER delete-then-insert (feature rule + idempotency).
// 4) LegacyCount for edge collections = Σ len(u.Roles) over users — the #419 validation formula column.
```

Edge semantics note: `Role.Users`/`Claim.Users` back-references are `json:"-"`
(`role.go:9`, `claim.go:11`) — they never exist in `data.json`; the migrator must derive edges only
from the forward direction.

`map[string][]VMSnapshot` (`orchestrator_vm_snapshots.go:6`): flatten to the #419-designed rows with
the map key preserved exactly (it is data — a VM uuid); flatten order = sorted keys so runs are
byte-deterministic.

Tests (`mapping_relational_test.go`, `:memory:`):

| Test | Assert |
|---|---|
| golden relational e2e | per-#419-table `COUNT(*)` equals formulas from its validation column |
| shared role across 2 users | ONE role entity row; TWO edge rows |
| re-run after success (AC-2) | counts identical; no duplicate edges (the OnConflict-Ignore path exercised) |
| orphan embedded role | role exists even though absent from top-level `Roles` |
| json:"-" backrefs | no panic; no phantom edges |
| snapshot map flatten | key preserved as FK/column per doc; sorted deterministic order (run twice, compare ordered row dumps) |
| FK-safe order | permute `collectionPlan()` deliberately wrong in a unit test ⇒ FK violation error surfaces (guards that the plan order is load-bearing, not luck) |

## Out of scope

Anything about *where* the schema comes from (#419/#421), validation reporting (05), startup (06),
legacy-side semantics changes, `Data.Schema` (AD-4 — never migrated).

## Validation

- [ ] `cd src && go build ./...` && `go vet ./database/...`
- [ ] `cd src && go test ./database/legacy_migrator/... -count=1` green
- [ ] `cd src && go test ./... -count=1` regression green
- [ ] PR description table: every new table/column string ↔ row in #419's mapping doc (reviewer gate; if a needed row is missing from that doc, STOP — raise on the #419 doc, do not invent)
- [ ] Golden fixture re-inspected for synthetic values only (no real hostnames/usernames)

## Rollback

Revert PR; task-03 `collectionPlan()` shrinks back to scalars.

## PBI acceptance mapping

Completes the "converts **all** existing data" clause of the AC-1/DoD scope, feeds **AC-6** formulas,
and proves **AC-2** under the join/edge case (the classic duplicate-data trap).

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `424-04`.** Cut `slice/424-04` from `sql-migration` HEAD
(`.worktrees/slice-424-04`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
