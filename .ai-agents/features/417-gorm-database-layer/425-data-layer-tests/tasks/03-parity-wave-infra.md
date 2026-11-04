# Task 03 — Engine-parity wave: **infra** (catalog, reverse proxy, snapshots, packer, enrollment)

**PBI** [#425](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 02 · **PR size:** M
**Plan refs:** [D-1](./plan.md) Tier-2 · [D-2](./plan.md) harness API · [D-5](./plan.md) scenario mechanics · task index row 03

## Context

Wave 2 of the Tier-2 promotion: the infra domains #421 ported in its tasks 10/11/12/14/16 get permanent
parity scenarios. Highest-risk behaviours to pin (each has a #421 spec section): serialised array columns
(`CatalogManifest.Tags/Claims/RequiredRoles`) survive read-modify-write, **route reorder invariant**
(reverse_proxy:342 — concurrent reorder), catalog ordering call site (`UpdatedAt` desc — order_layer tie
semantics), snapshot set/get by VM id, enrollment TTL edges. Task 01/02 already established `Scenario.Run`
+ `testutil.OpenMemoryDB`; this task adds files, edits none.

## Objective

Each infra domain has a parity file: one fixture dataset, both engines (`RunOnEngines`), behavioural
assertions incl. the invariant/error edges above; ledger rows added to the `parity/doc.go` table (ledger
*content* finalizes in task 04 — here only the wave's own rows).

## Files

| File | Op |
|---|---|
| `src/database/parity/catalog_parity_test.go` | create (catalog + catalog_managers — one file, both domains share fixtures) |
| `src/database/parity/reverse_proxy_parity_test.go` | create |
| `src/database/parity/snapshots_parity_test.go` | create (`Set/GetListVMSnapshotsByVMId` + host snapshot records) |
| `src/database/parity/packer_enrollment_parity_test.go` | create (the two S-domains; TTL edges need clock control — §4) |

## File size guard

| File | Lines | After | Verdict |
|---|---|---|---|
| `src/database/parity/*` (4 new) | — | ~140–190 each | ok |
| `src/database/parity/` (whole dir after task 03) | 01+02 output | ~900 | ok-watch (split files stay ≤200; dir budget is task 07's concern only) |

## Implementation spec

1. **Catalog fixture:** 2 manifests (one tainted-with-roles/claims arrays non-empty, one pristine), 1 manager,
   tags with ordering-sensitive values (D-5 tie rule); scenarios: create/read/update round-trip **JSON fidelity
   assert** (`assert.Equal(expectedManifest, got)` — serialised columns must not reorder array elements —
   pin if legacy does; log-and-pin if not), delete cascades per #421 task 11's tx rows (delete manifest ⇒
   dependent rows gone; delete manager guard), ordering scenario: insert out-of-order timestamps via direct
   GORM seeding (string dates — #419 D-1 keeps `UpdatedAt` string; lexicographic tie == legacy's
   input-order ties? assert **equality between engines**, not "correctness" — D-5).
2. **Reverse proxy:** 1 config + 2 hosts (tls/cors/tcp set on one; 3 http routes on the other); enable/disable
   toggle read-back; the invariant scenario mirrors #421 task 12's `TestReorderInvariant_Concurrent` but
   **through `DatabaseService` methods only** (no direct GORM) — engine-agnostic by construction; route-order
   compacting delete (order gaps never observable).
3. **Snapshots:** global collection set/get per VM id (replace-per-vm semantics), host-level snapshot map
   upsert (`SetHostVMSnapshots` idempotence: set twice, read once, deep-equal).
4. **Packer + enrollment:** CRUD + dup-name (packer: 400 sentinel per D-4); enrollment TTL: scenarios with
   `expires_at` seeded in the past/future **relative to the engines' clock** (±2 s margin — injected `nowFn`
   is #421-internal; the suite tests *observable* outcomes: valid/expired/used tri-state through the public
   API; sweep scenario via `DeleteExpiredEnrollmentTokens` return-count parity).
5. **Error assertions:** by D-4 sentinel identity inline (each scenario's failure step asserts the same
   sentinel both engines) — the dedicated matrix file lands in task 04.

## Out of scope

Orchestrator/jobs/config (task 04), the D-4 matrix file (04), Postgres legs (05), migration fixtures (06),
any edit to `testutil` (read-only since 01 — additive only).

## Validation

- [ ] `cd src && go build ./... && go test ./database/parity/... -count=1` (both legs green)
- [ ] `cd src && go test ./database/parity/... -count=5 -race` (leak/isolation posture holds across 4 new files)
- [ ] `cd src && go test ./... -count=1` (ER-5 — Tier 1 legacy suite untouched: `git diff --stat src/data/` empty)
- [ ] `go test ./... -count=1` **wall time logged in PR** vs task 01's baseline entry (D-9 early-warning, not final evidence — 07 owns that)

## Rollback

Delete the four files + the wave's ledger rows; harness and wave-auth untouched.

## PBI acceptance mapping

**Existing tests pass → promoted coverage**: infra legacy tests (`catalog_test.go`, `reverse_proxy` coverage
in `orchestrator_test.go`, `vm_snapshots_test.go`, `packer_template_test.go`) gain dual-engine equivalents;
**In-memory SQLite** (all new legs via unique-DSN harness); **Error paths** (inline D-4 identities);
constraint budget tracked per run.
