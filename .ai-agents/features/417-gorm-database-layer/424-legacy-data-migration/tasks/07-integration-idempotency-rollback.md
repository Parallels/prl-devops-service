# Task 07 — Integration suite: golden e2e, idempotency, rollback & exit semantics

**PBI** [#424](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** task 06 (full chain
wired) · **PR size:** M
**Plan refs:** [D-5](../plan.md#d-5--atomicity-one-dbtransaction-around-every-collection--the-marker), [D-9](../plan.md#d-9--dry-run-ac-7--startup-hook--failure-policy), [D-10](../plan.md#d-10--fixtures-committed-tiny-json-everything-generated-at-test-time), brief "definition-of-done link" (`//go:build integration` in CI), feature plan §12

## Context

This is the PBI's proof-of-record: the first-run upgrade story ("user upgrades, keeps all data, no
manual steps") exercised end-to-end through the same code paths production takes, plus the adversarial
cases the unit tests can only simulate — a mid-migration process death and the API-command exit policy
that unit tests must never trigger. Everything lands under `//go:build integration` so the default
`go test ./...` gate (ER-5) is untouched; #425 later folds this into its suite/compose rig.

## Objective

A `//go:build integration` suite over `:memory:` + temp-dir disk SQLite (real #418 factory, zero mocks)
covering: golden migrate→validate→marker, re-run no-op, encrypted store, injected mid-migration failure
⇒ **zero committed rows incl. marker**, and the exit-1-under-`api` policy observed at process level.

## Files

| File | Op |
|---|---|
| `src/database/legacy_migrator/migration_integration_test.go` | create — e2e happy/idempotency/encrypted/dry-run |
| `src/database/legacy_migrator/rollback_integration_test.go` | create — FailAfter + subprocess exit policy |
| `src/database/legacy_migrator/testdata/golden.json` | verify/extend (created in 02/04) |

## File size guard

New ≈ 300 / 220 lines of test code (tests are exempt from D-1's ≤200 guidance, still ≪ ER-2's 1,000
cap). Verdict: **ok**.

## Implementation spec

### `migration_integration_test.go` (`//go:build integration`)

```go
// Run: cd src && go test -tags integration ./database/legacy_migrator/... -count=1
// No external services (SQLite file + :memory: only) — unlike #418/#425's postgres compose tests.
func TestE2E_GoldenUpgrade(t *testing.T)  { /* table of scenarios below */ }
```

| Scenario | Steps | Asserts |
|---|---|---|
| golden upgrade | copy `testdata/golden.json` → temp `data.json`; factory `DatabaseConfig{Path: tmp/prl-devops.db}` (AD-3 coexistence proof: **both files present**); `RunMigration` | all #419 tables populated per fixture; validation zero issues; marker row matches path/size/sha256; **golden file bytes unchanged** (sha before/after) |
| re-run no-op (AC-2) | same handles; `RunMigration` again | `Report.Skipped && SkipReason=="already-migrated"`; counts identical; exactly 1 marker row |
| re-pointed store | move golden to `data2.json` (same bytes, new path/mtime), run against it | hash-step skip (D-4 step 3); info line asserted via log capture |
| changed legacy after migration (frozen, OQ-1) | append a record to `data.json` post-success | skip + frozen-legacy WARN (step 4); still 1 marker row; **no re-migration** |
| encrypted store | reuse task-02's `buildEncryptedLegacy(t, pemKey, goldenBytes)` helper (exported for #425); key via `t.Setenv(ENCRYPTION_PRIVATE_KEY…)` | migrates identically to plaintext |
| encrypted, no key | clear env | error names `ENCRYPTION_PRIVATE_KEY`; DB empty; no marker (R-5) |
| large store chunking | `BuildLargeJobsFixture(t, 5000)` | completes < 30 s; progress line count ≥ 10 (`ChunkSize:500`) |
| dry-run e2e | fresh golden + `DryRun:true` | report counts match fixture; DB `COUNT(*)==0` everywhere **and** `Migrator().HasTable("legacy_migration_marker")` still false or empty; file sha unchanged |

### `rollback_integration_test.go` (`//go:build integration`)

1. **Injected failure ⇒ full rollback (AC-3):**
   `Options.FailAfter: func(name string, done int) error { if name == jobsName { return errors.New("injected crash") } return nil }`
   (jobs = last collection in the plan). Assert `RunMigration` returns the wrapped error, then reopen
   the same SQLite **file** (fresh handle ⇒ proves nothing lived in an open tx): every #419 table
   `COUNT(*)==0`, `legacy_migration_marker` table empty or absent. Re-run without `FailAfter` ⇒ clean
   migration succeeds afterwards (crash-then-retry story closed).
2. **Exit policy under `api` command (process-level):**
   mimic #418 task 05's precedent — build the binary once in `TestMain`
   (`go build -o t.TempDir()/devops-service ..` from `src/`), run it with `API_COMMAND`-selecting args
   (`api`) + `--config` pointing at a fixture config + `DATABASE_FOLDER=<tmp>` + a corrupt `data.json`
   (or `FailAfter` impossible at this level ⇒ corrupt JSON is the realistic mid-boot failure) +
   `DB_TYPE=sqlite`. Assert `cmd.Run()` exit code == 1 and stderr/log contains
   `Legacy data migration failed`. Companion case without the corrupt file ⇒ boot passes migration and
   the test kills the healthy process after seeing a readiness marker line (bounded by
   `context.WithTimeout`, 30 s). If the binary boot pulls in too much platform surface on CI runners, the
   scenario stays build-tagged and the unit-level seam test from task 06 remains the guaranteed signal —
   note this fallback in the PR.

## Out of scope

Postgres variants (need #425's compose service; formulas are backend-agnostic), performance tuning,
real legacy installs from a released binary (issue-level QA).

## Validation

- [ ] `cd src && go build ./...` && `go test ./... -count=1` (default gate ignores integration files)
- [ ] `cd src && go test -tags integration ./database/legacy_migrator/... -count=1 -v` green; paste tail incl. large-store timing
- [ ] Behavioural AC checks (checkbox list in PR from the table above): detection/idempotency/atomicity+rollback/progress/marker/validation/dry-run each mapped to the named test
- [ ] `git diff --stat` shows only the 2–3 test files + fixture growth
- [ ] Verify no `t.Setenv` leakage between tests (suite runs `-count=2` once — order independence, feature plan §12)

## Rollback

Revert PR removes integration files + fixture additions; production code untouched.

## PBI acceptance mapping

Final proof for **AC-1…AC-7** as a system (each row above names its AC), and discharges the PBI-brief
DoD link: "first-run upgrade … exercised in CI (`//go:build integration`)". Hands fixture builders + the
compose hook to **#425**.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `424-07`.** Cut `slice/424-07` from `sql-migration` HEAD
(`.worktrees/slice-424-07`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
