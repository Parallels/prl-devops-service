# Task 04 — Engine-parity wave: **orchestrator + jobs/config** · error matrix · ledger closure

**PBI** [#425](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 03 · **PR size:** M
**Plan refs:** [D-4](./plan.md) (THE error matrix — implement it verbatim, quirks included) · [D-5](./plan.md) · [D-1](./plan.md) exclusion ledger · #421 tasks 13/14/15/08/09 specs

## Context

Wave 3 (last parity wave): orchestrator (the 933-line domain #421 split across tasks 13/14), jobs (incl.
timeout behaviour — the engine-timing corner), config + user_config singletons. Then the two closures that
make Tier 2 *auditable*: the **D-4 matrix file** (the AC-5 deliverable in one place) and the **finalized
exclusion ledger** in `parity/doc.go` (every legacy test file → converted scenarios or a one-line rationale,
e.g. `TestSaveAs*` = file-engine-only). This task also stress-pins #421's trickiest semantics:
`ReplaceOrchestratorHostVMs` full-sync and the heartbeat/VM-write field-scope rule (#421 task 13 §4).

## Objective

Orchestrator/jobs/config scenarios green on both engines; every D-4 row implemented exactly once in
`errors_parity_test.go`; ledger complete: **every scenario in the suite traces to a legacy test or
call-site behaviour** and vice versa.

## Files

| File | Op |
|---|---|
| `src/database/parity/orchestrator_parity_test.go` | create (hosts + VMs + views + resources aggregations) |
| `src/database/parity/jobs_config_parity_test.go` | create (jobs CRUD/states/delete-by-state/timeout shape + config singleton + user_config owner scoping) |
| `src/database/parity/errors_parity_test.go` | create (D-4 matrix, all rows, table-driven) |
| `src/database/parity/doc.go` | edit (ledger finalize — file stays ≤120 lines; ledger may move to `LEDGER.md` in the package dir if growth demands, note in PR) |

## File size guard

| File | Lines | After | Verdict |
|---|---|---|---|
| new test files | — | ~180 + ~150 + ~160 | ok |
| `parity/doc.go` | task-02 stub | ≤120 (or split per §Files note) | ok |

## Implementation spec

1. **Orchestrator fixture:** 2 active hosts + 1 inactive (`HostResources` full/zero variants — aggregation
   golden values computed once, asserted both engines), nested VMs seeded via **engine APIs only**
   (Upsert×3 + one Replace-with-different-set to exercise full-sync — assert post-state by ID set, #421
   task 14 §2); RP views scenario: hosts exist + proxy config disabled ⇒ legacy-identical empties (#421
   task 14 §5); delete-host cascade (VM rows gone — #421 task 13 §3); the heartbeat-vs-VM-write
   interleaving is *already* covered by #421's own concurrency test — here assert only **result equality**,
   not race behaviour (unit-suite discipline, D-3).
2. **Jobs:** create→running→complete state walk through the public API; `DeleteJobsByState` count parity;
   timeout *shape*: pre-`UpdatedAt`-stale job + the domain's observable timeout outcome (per #421 task 15
   §3-4 semantics — if the engine's `SetJobTimeoutHandler` firing is #421-internal, the suite asserts the
   state transition, not the goroutine; ledger notes the split). Config: singleton create-on-read +
   `UpdateSchemaVersion` row semantics (#421 task 08 — distinct from #424's marker table — assert BOTH
   tables exist post-`schema_metadata` write: cheap AD-4 guard). User configs: owner-scoping leakage probe
   (user A must not read B; #421 task 09 mandatory fixture reused).
3. **`errors_parity_test.go` = D-4 as code:** one table, rows exactly as D-4's four classes;
   not-found × 16 sentinels ⇒ loop over `{domain, create-fn, get-fn, sentinel, wantCode}`; the no-code
   `ErrCatalogManagerNotFound` quirk asserts *identity + zero-value code*, never panics; quirk duplicates
   (500/404/500) each carry `t.Log("quirk preserved, see plan D-4 — do not fix")`; unreachable row per
   D-4 (loopback `:1` DSN, `errors.Is`, redaction check); constraint rows assert sentinels, never driver
   strings (SQLite forces now; Postgres legs extend the same table in task 05's file).
4. **Ledger closure:** walk `src/data/*_test.go` files (list in brief) — each maps to: parity scenarios
   (name them), or exclusion + rationale. Format: one row per legacy test file. Any row that says "TODO"
   fails this task.

## Out of scope

Postgres legs (05), migration suite (06), Makefile/CI (05/07), any legacy-test conversion — Tier 1 stays
byte-identical (AC interpretation D-1).

## Validation

- [ ] `cd src && go test ./database/parity/... -count=1` full matrix green
- [ ] `cd src && go test ./database/parity/... -run Errors -count=5 -shuffle=on` (D-3 posture + matrix order-independence)
- [ ] `cd src && go test ./... -count=1` (ER-5)
- [ ] Ledger completeness gate: `grep -c "TODO\|TBD" src/database/parity/doc.go` → 0 (or LEDGER.md)
- [ ] Wall time logged in PR

## Rollback

Revert: two wave files + matrix + ledger delta (ledger rows revert cleanly — 03's rows are separate table
lines). Nothing structural.

## PBI acceptance mapping

**Error path coverage** (AC 5) — the D-4 matrix file IS the deliverable; **Existing tests pass / promoted**
— orchestrator/jobs/config Tier-2 coverage + the ledger makes the claim checkable by a human in one screen;
**Backend coverage** — matrix structured so task 05 appends Postgres rows without edits; budget logged.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `425-04`.** Cut `slice/425-04` from `sql-migration` HEAD
(`.worktrees/slice-425-04`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
