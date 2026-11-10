# Task 05 — Post-migration validation counters (warn-only)

**PBI** [#424](../brief.md) · **Feature** [#417](../../brief.md) · **Depends-on:** tasks 03, 04 (formulas
live in `CollectionSpec.Validate`, transcribed from #419's mapping doc) · **PR size:** S
**Plan ref:** [D-8](../plan.md#d-8--validation-ac-6-legacy-counts-in-memory-vs-count-after-commit-warn-only)

## Context

AC-6: "A post-migration validation step counts records in the legacy store vs. the new store and logs a
warning if counts differ." Legacy side needs no query — `len(Data.<Collection>)` in the already-loaded
struct; the new side is `COUNT(*)` with the **join-table-aware formulas from #419's mapping table**
(the `user_roles`-style expectations, e.g. `Σ len(u.Roles)`). Warn-only is literal AC wording: a
mismatch must **never** fail startup — migration already committed, and bricking boot over a counter
would be worse than the bug it reports. Runs post-commit: `COUNT` then sees exactly the committed truth.

## Objective

`ValidateCounts(ctx, db, legacy, plan)` emits one info line per collection and a
`ctx.LogWarnf("…legacy=%d sql=%d…")` on mismatch, wired into `RunMigration`'s success path; a
deliberately broken formula cannot make `RunMigration` return an error.

## Files

| File | Op |
|---|---|
| `src/database/legacy_migrator/validation.go` | create |
| `src/database/legacy_migrator/validation_test.go` | create |
| `src/database/legacy_migrator/migrator.go` | modify — call after `db.Transaction` success, skip on dry-run |

## File size guard

New ≈ 110 / 150 lines; `migrator.go` 180 → ~190 (all created in tasks 03/04 — no existing repo file
touched). Verdict: **ok**.

## Implementation spec

```go
type ValidationIssue struct { Collection string; Expected, Actual int64 }

// ValidateCounts compares in-memory legacy counts against SQL counts via the plan's formulas.
// WARN-ONLY by AC-6 — errors from Count are logged (LogWarnf) and recorded as Actual=-1, never fatal.
func ValidateCounts(ctx basecontext.ApiContext, db *gorm.DB, legacy *data.Data, plan []CollectionSpec) []ValidationIssue {
        var issues []ValidationIssue
        for _, s := range plan {
                expected, actual, err := s.Validate(db, legacy)
                if err != nil {
                        ctx.LogWarnf("legacy migration validation: %s count query failed: %v", s.Name, err)
                        issues = append(issues, ValidationIssue{s.Name, expected, -1}); continue
                }
                want := s.LegacyCount(legacy)
                if want != int(expected) { // defensive: formula vs accessor drift is itself a finding
                        ctx.LogWarnf("legacy migration validation: %s accessor=%d formula=%d — mapping doc drift?", s.Name, want, expected)
                }
                if want != int(actual) {
                        ctx.LogWarnf("legacy migration validation MISMATCH %s: legacy=%d new=%d — investigate (migration NOT failed by validation)", s.Name, want, actual)
                        issues = append(issues, ValidationIssue{s.Name, int64(want), actual})
                        continue
                }
                ctx.LogInfof("legacy migration validation ok: %s %d/%d", s.Name, actual, want)
        }
        if len(issues) > 0 {
                ctx.LogWarnf("legacy migration validation finished with %d mismatched collection(s)", len(issues))
        } else {
                ctx.LogInfof("legacy migration validation finished: all collections matched")
        }
        return issues
}
```

`migrator.go` hook (after successful `db.Transaction`, before `return report`):

```go
report.Validation = ValidateCounts(ctx, gdb, &legacy, plan) // Report gains []ValidationIssue (task-05 edit)
```

## Tests

| Test | Assert |
|---|---|
| happy golden | zero issues; per-collection `validation ok` info lines |
| seeded mismatch | insert one extra row into a scalar table post-migration ⇒ issue recorded + MISMATCH warn captured; `RunMigration` still returned nil error (re-run in dry-run style or unit-level on ValidateCounts) |
| count query error (drop a table mid-test) | warn logged, `Actual=-1`, no panic, no error |
| formula vs accessor drift | both values surfaced, warn text includes "mapping doc drift" |
| dry-run path | no validation lines (nothing committed) |

## Out of scope

Failing/healing on mismatch (AC says warn); checksums/field-level diffs (beyond AC-6); #425's
differential-engine tests.

## Validation

- [ ] `cd src && go build ./...`
- [ ] `cd src && go test ./database/legacy_migrator/... -count=1` green
- [ ] `cd src && go test ./... -count=1` regression green
- [ ] Manual behavioural check: run migrator against `testdata/partial.json` with a hand-crafted wrong
      formula in a scratch branch — boot path still returns success (paste log excerpt in PR)

## Rollback

Revert PR removes validation + the one-line hook; migration itself unaffected.

## PBI acceptance mapping

Completes **AC-6 Data integrity** (counts + warn-only) and adds the operator-visible
`validation ok` lines supporting **AC-4**.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `424-05`.** Cut `slice/424-05` from `sql-migration` HEAD
(`.worktrees/slice-424-05`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
