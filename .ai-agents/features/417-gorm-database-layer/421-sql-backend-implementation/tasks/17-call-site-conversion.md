# Task 17 — S2 call-site conversion: `GetDatabaseService` returns the interface

**PBI** [#421](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** tasks 04–16 **all ported** · **PR size:** M (D-10 exception applies)
**Plan refs:** D-6 step S2 · D-7 (lifecycle no-ops / seeds row) · D-10 (ER-3 exception — **needs reviewer approval, surfaced in PR**) · D-11 unchanged

## Context

After task 16 every method on `GormDatabase` answers from SQL, but nobody outside the provider *can* use it:
`GetDatabaseService` still returns `*data.JsonDatabase`. S2 retypes the **consumer-facing** surface to
`interfaces.DatabaseService` — a compile-only change with **zero runtime behavior change** (the returned
value is still the legacy instance until S3). The blast radius was measured, not estimated:
`grep -rn "\*data\.JsonDatabase" src --include=*.go | grep -v _test | grep -v '^src/data/'` yields **14
declarations in 10 files**, and the majority intentionally keep the concrete type.

## Objective

Interface-typed consumers everywhere it is safe; concrete type retained exactly where legacy-only machinery
lives (each retention a named, justified row); ~143 method-call sites verified to compile untouched; app
behaves bit-identically (still legacy engine by default — ER-5 posture preserved).

## Files (the measured table — re-run the grep at task start; new call sites get rows here)

| Declaration | Verdict |
|---|---|
| `src/serviceprovider/data.go:8` `GetDatabaseService` **return** | **retype** → `interfaces.DatabaseService` (the S2 moment; body: `Connect` guard retained) |
| `src/catalog/import_vm.go:335, :364` (`importClaims`/`importRoles` params) | **retype** → interface (they only call CRUD methods) |
| `src/startup/seeds/packertemplates/*.go` ×3 params (`ubuntu_23_04.go:11`, `kali_linux_2023_3.go:11`, `macos_14_0.go:11`) | **retype** → interface (D-7 seeds row: existence-checks + SaveNow become no-op-safe) |
| `src/main.go:120` `cleanup(ctx, db *data.JsonDatabase)` | **retype** → interface (`SaveNow`/`SaveAs` are interface members — D-2 lifecycle; panic-snapshot keeps working via GormDatabase's D-7 semantics *once S3*; under S2 the value is legacy, unchanged) |
| `src/jobs/main.go:20` field `db` | **retype** → interface |
| `src/orchestrator/main.go:34` field `db` | **retype** → interface |
| `src/reverse_proxy/main.go:79` field `db` (+ constructor param) | **retype** — the D-10 file: **ER-3 exception required** (1,302 lines > cap; type-only 2-3-line edit; alternative = same-PR pure-move split — reviewer chooses, default proposal = exception, OQ-421-1) |
| `src/serviceprovider/main.go:41` field | already retyped by task 01 (S1) — verify, no action |
| `src/serviceprovider/parallelsdesktop/main.go:103` field + `:167` `SetDatabaseService` param | **keep concrete** (D-6 ripple rule: ParallelsService machinery uses legacy-only surface — OQ-421-2; injection passes `globalProvider.legacyDatabase`) |
| `src/startup/migrations/service.go:11` `DbService` | **keep concrete** (schema migrators are legacy-store-only by AD-4/D-7; SQL mode nil-guards per task 18) |

**Retype = change the declared type only.** No logic edits in this PR beyond type plumbing.

## File size guard

| File | Lines | After | Verdict |
|---|---|---|---|
| `src/reverse_proxy/main.go` | 1,302 (D-10 measured) | +0 | **D-10 exception** or split — PR description carries the decision |
| `src/serviceprovider/data.go` | 25 (measured) | ~25 | ok |
| others in table | all <700 except seeds (small) | ±0 | ok |

## Implementation spec

1. Apply the verdicts top-down; each retype compiles against the interface **without new methods** — if a
   consumer needs a method *not* on `DatabaseService`, that's a stop-sign: the method is legacy-only ⇒
   consumer stays concrete (add row with justification), do **not** widen the interface.
2. Locals (`db := serviceprovider.Get().JsonDatabase` in controllers — verified locals-only at
   `controllers/orchestrator.go:835,3934,3968,4456`, `machines.go:2116` per D-6) adopt the interface
   automatically; no controller file is edited.
3. Seeds (D-7): after retype, `SaveNow` calls in seed flow are interface no-ops under SQL mode — verify no
   seed *reads* assume file-engine error strings; fix message-string assertions only (`-count=1` full suite
   catches).
4. `cleanup`/panic-snapshot: prove `SaveAs` compiles via interface (it is a lifecycle member, task 01);
   runtime behaviour unchanged in S2 (value is legacy).
5. Grep-audit artifact (PR body): the §Files grep, *after* this PR, must return exactly the 4 retained rows
   (+ their constructors/uses). Any other survivor = mistake.

## Out of scope

Any behavior change; provider selection (S3 = task 18); the retained-concrete files' internals; interface
widening; controller bodies.

## Validation

- [ ] `cd src && go build ./...` — the interface is the type-checker: build green ⇒ no hidden method need
- [ ] `cd src && go vet ./... && go test ./... -count=1` — full suite green, **behavior-identical evidence**:
      no test file edits allowed in this PR (if a test edits, behavior changed — stop)
- [ ] `grep -rn "\*data\.JsonDatabase" src --include=*.go | grep -v _test | grep -v '^src/data/'` → exactly
      the retained rows pasted in PR
- [ ] Boot smoke (API command, default env): identical startup log sequence to pre-PR binary (S2 is invisible)

## Rollback

Single revert; consumers re-grab concrete type; no data or schema surface touched.

## PBI acceptance mapping

**Legacy layer bypassed** (half 1 of 2): after this PR *no consumer can reach legacy-only methods without an
explicit escape hatch* — the type system enforces the bypass; **file path / backend selection** untouched
(ER-5: legacy still the running engine); **D-10 decision recorded** for the feature's ER-3 ledger.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `421-17`.** Cut `slice/421-17` from `sql-migration` HEAD
(`.worktrees/slice-421-17`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
