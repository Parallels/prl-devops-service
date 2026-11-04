# Task 13 — Orchestrator hosts on GORM (CRUD, aggregations, field-scoped updates)

**PBI** [#421](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 11 (plan index — "VM dep order noted"; VM half is task 14) · **PR size:** M
**Plan refs:** D-9 (host rows `orchestrator.go:121-146`) · D-11 · D-12 · #419 D-7 (never default-preload VMs)

## Context

Legacy `src/data/orchestrator.go` (933 lines — **untouched**, ER-2 watch per feature plan: this PBI must not
grow it) splits cleanly: 15 host-domain methods here, 16 VM/snapshot/view methods in task 14 (inventory
partitioned below, line numbers verified by grep). The landmines are **aggregation reads** (five
resource-summing methods iterate hosts and sum `HostResources` — serialised JSON per #419, so sum Go-side
over loaded rows) and **field-scoped updates** (`Timestamp`, `WebsocketStatus`): with `VirtualMachines` a
serialised column, a whole-row `Save` would clobber VM arrays written concurrently by task-14 paths —
field-scoped `Update`/`Select` is mandatory here, not stylistic.

## Objective

Host CRUD + aggregations + heartbeat-status writes served by GORM, parity vs legacy, with a concurrency test
proving heartbeat updates never clobber VM writes (cross-task table).

## Files

| File | Op |
|---|---|
| `src/database/orchestrator.go` | create — **hosts part only**; task 14 appends VM methods to sibling `orchestrator_vm.go` (1:1 rule, plan index) |
| `src/database/delegate_orchestrator.go` | trim to the 16 task-14 stubs (05→06 precedent: delete ported, never edit survivors) |
| `src/database/orchestrator_hosts_parity_test.go` | create |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `src/data/orchestrator.go` | 933 (measured) | 933 (untouched — this PBI's #1 growth-hazard) | ok-watch |
| `src/database/orchestrator.go` | new | ~200 | ok; if >230 split aggregations to `orchestrator_resources.go` |

## Implementation spec

**Method inventory — hosts (15):** `GetOrchestratorHosts` :22 · `GetActiveOrchestratorHosts` :58 ·
`GetOrchestratorHost` :93 · `CreateOrchestratorHost` :121 · `DeleteOrchestratorHost` :147 ·
`UpdateOrchestratorHost` :200 · `UpdateOrchestratorHostDetails` :281 · `UpdateOrchestratorHostTimestamp` :341 ·
`UpdateOrchestratorHostWebsocketStatus` :372 · `GetOrchestratorAvailableResources` :398 ·
`GetOrchestratorTotalResources` :425 · `GetOrchestratorInUseResources` :452 · `GetOrchestratorReservedResources` :479 ·
`GetOrchestratorSystemReservedResources` :507 · `GetOrchestratorHostResources` :533 · `UpdateOrchestratorHostResources` :820 ·
`UpdateOrchestratorHostCacheConfig` :846. (Unexported `getOrchestratorHostsLocked` :37 / `getOrchestratorHostUnsafe` :104
are the legacy's lock helpers — no interface counterparts; GORM tx/field-scope replaces their role.)
*17 rows listed; the interface has 31 total — recount belongs to the implementer's task-01 file: 2 rows
(`GetOrchestratorHostResources`/`UpdateOrchestratorHostResources`) may be judged resource-part — keep
partition by *ownership of the row being written*, not by name prefix; task 14's list is the source of truth
for what's left.*

1. **Reads.** `GetOrchestratorHosts`: `Find` all, `Preload` nothing (VMs serialised; legacy returns hosts
   **without** VM bodies — **implementer to confirm legacy VM-strip behaviour (:22-58)** before deciding the
   column is even loaded — `Omit("virtual_machines")` saves a JSON decode per row).
   `GetActiveOrchestratorHosts` :58: `WHERE state = ?` — active-state enum from legacy; golden corpus:
   hosts in every state.
2. **Aggregations (:398-533, :820).** Load active hosts (`Omit` VM payload columns except resources),
   sum `HostResources` **Go-side** (D-12: serialised column — no SQL SUM of JSON). Zero-host edge: legacy
   returns a zero-value struct, not error — pin in test.
3. **Writes.** `Create` :121: dup-UUID/name pre-check per legacy (:121-146 — D-9 host+VM nesting note applies
   ONLY when `CreateOrchestratorHost` arrives with VMs payload: **tx create host + VM rows together** even
   though VM method family lands in task 14 — host-create writes both tables; import the serialised VM rows
   via the #419 child-table/join decision, and note this is the ONE overlap: task 14 tests exercise it, this
   task implements it). `DeleteOrchestratorHost` :147: **tx** — delete host + its VM rows (cascade per #419 FK
   mapping; SQLite `foreign_keys=ON` from #418 makes DB-cascade *possible* — prefer explicit deletes for
   Postgres parity).
4. **Field-scoped discipline (the rule):** any method that touches ONE field uses
   `db.Model(&h).Select("<col>").Updates(...)` — `Timestamp` (:341), `WebsocketStatus` (:372), `Details`
   (:281, multi-col Select of the detail subset), `CacheConfig` (:846), `Resources` (:820), `TlsConfig`-style
   helpers. Whole-row `Save` ONLY in `UpdateOrchestratorHost` (:200) where legacy replaces the record —
   **implementer to confirm :200-281 field-replacement set** so Select-list matches (empty columns preserved,
   legacy semantics).
5. **Concurrency proof:**

```go
func TestHeartbeatDoesNotClobberVMs(t *testing.T) {
    // seed host with 2 VMs via task-14 path once landed (or direct GORM insert now);
    // goroutine A: UpdateOrchestratorHostTimestamp loop; goroutine B: UpsertOrchestratorHostVM loop;
    // after: VM count/column intact AND timestamp advanced. Runs green on both engines.
}
```

## Out of scope

VM/snapshot methods + RP **views** (`GetOrchestratorReverseProxy*`) — task 14; enrollment tokens — task 16;
provider flip — 17/18; growing `src/data/orchestrator.go` — forbidden.

## Validation

- [ ] `cd src && go build ./... && go vet ./database/...`
- [ ] `cd src && go test ./database/... -run OrchestratorHost -count=1 -race`
- [ ] `cd src && go test ./... -count=1` (ER-5)
- [ ] Aggregation golden corpus green (empty/1/N hosts × full/zero resources)
- [ ] `wc -l src/database/orchestrator*.go` ≤ guard; `git diff --stat src/data/orchestrator.go` → empty

## Rollback

Revert commit restores delegate stubs (all 31 again); no schema change outlives the revert; task 14 specs a
dependency note but only *tests* this file's paths — no rollback ordering problem.

## PBI acceptance mapping

**CRUD via GORM** — 15 host methods; **Transactions** — create-with-VMs + delete-cascade (D-9
`orchestrator.go:121-146` row); **Error handling** — D-11; **Concurrent access** — heartbeat/VM-write
interleaving test (the domain's sharpest race, proven before WAL verification in 18).
