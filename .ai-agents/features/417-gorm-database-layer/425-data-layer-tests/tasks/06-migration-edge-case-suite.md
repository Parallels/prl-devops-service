# Task 06 — Migration edge-case suite (against #424's exported API)

**PBI** [#425](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 01 (recording ctx, memory DB) + **#424 merged** (its API + `testdata/` fixtures are the SUT — this task tests it from the outside) · **PR size:** M
**Plan refs:** [D-7](./plan.md) (THE scenario matrix — every row is a named test) · [D-2](./plan.md) isolation · #424 ACs (each matrix row maps back)

## Context

#424 ships `src/database/legacy_migrator/` with its own unit tests; this suite is **behavioural coverage
through the package's public API only** — the AC-4 deliverable ("migration tests covering empty, partial,
encrypted stores") plus the trust-critical properties: idempotence, re-entry after a crash, warn-not-fail
validation, dry-run inertness, and the **source-never-modified** invariant (OQ-1's frozen-source stance:
`data.json` hash invariant across every scenario — the property that makes the whole migration strategy
reversible). Static fixtures cover states JSON can express; transitions (crash-mid-run, marker
short-circuit) are *constructed* per D-7 — the one place white-box access is allowed (`package
legacy_migrator_test` + #424's exported seams; if a seam is missing, that's a #424 PR, not a copy of its
internals).

## Objective

Every D-7 row is a green test; every test asserts the source-hash invariant; the suite runs in the default
`go test ./...` (in-memory targets; disk usage only for the *fixture* store copies — the sanctioned
exception, AC-2 note).

## Files

| File | Op |
|---|---|
| `src/database/legacy_migrator/behaviour_test.go` | create (package `legacy_migrator_test` — the matrix) |
| `src/database/legacy_migrator/fixture_gap_test.go` + `testdata/` additions | create **only for genuine gaps** — `corrupt.json` etc. (audit #424's shipped `testdata/` first; reuse is the rule — list what was reused in PR) |

## File size guard

| File | Lines | After | Verdict |
|---|---|---|---|
| `behaviour_test.go` | new | ~220 (9 scenarios + shared harness helpers) | ok — split `_2` file if >240 |
| `src/database/legacy_migrator/*` (#424-owned) | — | untouched (test-only PR) | ok |

## Implementation spec

Shared harness per scenario (mirrors D-2 + recording ctx):

```go
type migCase struct {
    fixture   string            // testdata file ("" = no file at all)
    configure func(t *testing.T, store *fixture.Store) // mutate fixture copy pre-run (partial/corrupt)
    // assert receives: recording ctx + opened target (testutil memory DB) + sha256 of source before/after
}
```

Matrix rows → tests (D-7 verbatim): `TestMigration_EmptyStoreSilentSkip` (absent + zero-byte; trap guard:
detector used, file **not created** — `os.Stat` error still `IsNotExist` afterwards);
`TestMigration_HappyPathAllCollections` (16 collections incl. nil-able `Configuration`/`ReverseProxy`
pointers and embedded `User.Roles/Claims` — row counts + spot field checks via GORM queries; **no** warn
line in recording ctx); `TestMigration_EncryptedStore` (key fixture is #424's — header comment "synthetic
test-only key" per Risk-8); `TestMigration_CorruptJSON` (err non-nil, target count 0, no panic);
`TestMigration_RerunMarkerShortCircuit` (second run emits **zero** per-collection info lines via recording
ctx — the cheap no-op proof) + `TestMigration_CrashReEntryNoDuplicates` (marker row deleted / partial batch
commit via seam ⇒ re-run completes, per-table `COUNT(DISTINCT id) == COUNT(*)`);
`TestMigration_ValidationMismatchWarnsNotFails` (seam-injected count skew ⇒ `err == nil` + warn line naming
the collection + data committed); `TestMigration_DryRunInert` (target 0 rows, no marker, report lines at
info, source hash identical); every case ends `require.Equal(t, srcHashBefore, srcHashAfter, "source must
never be modified")`.

## Out of scope

Migrator internals/unit tests (#424), AutoMigrate schema tests (#419 task 06), provider wiring (#424
task 06), performance of large stores (a 10k-VM perf smoke is a *follow-up candidate* — ledger note, not
this PR).

## Validation

- [ ] `cd src && go test ./database/legacy_migrator/... -count=1 -v` all matrix rows green
- [ ] `cd src && go test ./database/... -count=5 -shuffle=on` (migration fixtures can't leak state — memory targets)
- [ ] `git diff --stat src/database/legacy_migrator/ -- ':(exclude)*_test.go' ':(exclude)*/testdata/*'` → empty (test-only PR)
- [ ] `cd src && go test ./... -count=1` (ER-5) + wall-time line in PR
- [ ] Fixture-reuse list in PR description (what came from #424 vs new)

## Rollback

Delete test files + any new fixtures (fixtures are additions; deleting changes no behaviour). The suite
tests #424 from outside ⇒ zero coupling to its internals beyond exported API — no cascade.

## PBI acceptance mapping

**Migration tests** (AC 4) — the nine D-7 rows, each traceable to a #424 AC (comment headers carry
`#424 AC<n>` tags); **No flaky tests** — `-shuffle=on -count=5` gate here; **≤+30 s** — memory targets keep
it AutoMigrate-bound (measured gate 01→07); **Existing tests pass** — legacy `main_test.go` file-engine
paths stay the Tier-1 home of file behaviour (excluded list unchanged).
