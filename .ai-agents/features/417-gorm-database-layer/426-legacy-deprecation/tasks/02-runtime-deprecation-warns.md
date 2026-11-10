# Task 02 — Runtime deprecation warnings: selection warn + frozen-legacy warn (OQ-1)

**PBI** [#426](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 01 · **PR size:** S
**Plan refs:** D-3 (both warn strings verbatim, test pattern, extraction rule) · OQ-1 · #424's marker API (read, never re-implement)

## Context

AC "a deprecation warning is logged when the legacy data layer is used" is formalized here from ad-hoc
#421/#418 warns into **two pinned constant strings** at the selection site (`src/serviceprovider/data.go`):
(1) fires every startup that selects legacy; (2) the OQ-1 **frozen-legacy** warn — legacy selected *and*
#424's `legacy_migration_marker` says the store was already migrated — meaning the instance is serving
stale, never-written-through data, the single worst failure mode of the transition period. Tests are
included despite the PBI DoD's unchecked unit-tests box: **the AC is unprovable without them, and ACs
outrank DoD defaults** (D-3's recorded stance — PR description carries this rationale).

## Objective

Both warns emit exactly their D-3 strings (asserted, not eyeballed); marker-read failure degrades to a
single warn + continue (legacy stays functional — PBI constraint); selection code is test-reachable.

## Files

| File | Op |
|---|---|
| `src/serviceprovider/data.go` | modify — the two constants + emission points (+ `selectDatabaseService` extraction if #421 left selection inlined — D-3) |
| `src/serviceprovider/data_test.go` | create/extend — recording-fake tests |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `src/serviceprovider/data.go` | 25 (measured) | ~85 (constants + guard + marker probe + extraction) | ok |
| `src/serviceprovider/data_test.go` | new | ~140 | ok |

## Implementation spec

1. **Constants (verbatim from D-3, versions from D-6):**

```go
const (
    legacyDeprecationWarnMsg = "[Database] DEPRECATED: legacy JSON data layer selected (DB_TYPE=%s). " +
        "The legacy layer is deprecated since v1.2.0 and will be removed in v2.0.0 (issues #426/#417). " +
        "Plan migration to DB_TYPE=sqlite — see CHANGELOG."
    legacyFrozenWarnMsg = "[Database] DEPRECATED: legacy data file %s was already migrated and is now " +
        "FROZEN/stale — this instance serves stale data; writes here never reach the SQL database. " +
        "Switch to DB_TYPE=sqlite (#424, #426)."
)
```

   (`DB_TYPE=%s` value = the legacy-mode literal shipped by #420/#421 — re-grep the constant, don't hardcode
   "legacy"; frozen warn `%s` = the legacy filename via `Filename()`.)
2. **Emission:** (1) unconditionally at legacy selection, `ctx.LogWarnf`; (2) after (1), probe marker via
   `legacy_migrator`'s **exported** reader (signature at implementation time — #424 task 01's marker.go;
   if it exposes `Migrated(ctx) (bool, error)`-shaped API use it; if only the *gated* startup path exposes
   nothing yet, PR note + use the SQL connection handle + raw row-exists query — never copy marker logic).
   Marker probe errors ⇒ one `LogWarnf("could not verify migration status: %v")` and proceed — **never
   fatal**; a missing SQL connection (pure-legacy deployment, never migrated) ⇒ warn (2) simply doesn't
   apply, skip silently (the probe presupposes a reachable SQL DB — state this boundary in code comment).
3. **Testability extraction** (if needed): `selectDatabaseService(ctx basecontext.ApiContext, …)` covering
   the branch; behaviour-identical; existing call sites forward.
4. **Tests** (recording fake implementing the 14-method `basecontext.ApiContext` — copy the #418 task 05
   pattern, capture `LogWarnf(format, args...)` into a slice):

| Case | Setup | Assert |
|---|---|---|
| a | `DB_TYPE=legacy`-equivalent, no SQL DB | exactly 1 warn == `legacyDeprecationWarnMsg` formatted |
| b | legacy selected + SQL reachable + marker row present (insert via #424 API into memory DB) | both warns, in order, exact strings |
| c | legacy + SQL reachable + marker absent | deprecation warn only |
| d | SQL mode (`DB_TYPE=sqlite`) | **zero** legacy warns (positive-space proof the warn is selection-scoped) |
| e | marker probe forced error (corrupt/detached DB handle) | deprecation warn + single probe-fail warn, **no fatal**, selection still succeeds |

   Postgres legs unneeded (marker read is engine-agnostic via GORM); memory DBs via `testutil` (task 01
   of #425 — if #425 hasn't merged, plain `t.TempDir()` SQLite file DBs; note the cross-PBI soft edge).

## Out of scope

Selection *logic* changes (none — this PR is warn-plumbing), marker schema/migration (#424), depguard
(task 03), changelog (04), touching the 3,079-line parallelsdesktop file's own warns (OQ-421-2 territory).

## Validation

- [ ] `cd src && go build ./... && go test ./serviceprovider/... -count=1 -v` (rows a–e green, string-exact)
- [ ] `cd src && go test ./... -count=1` (ER-5)
- [ ] String-drift gate: `grep -c "v1.2.0\|v2.0.0" src/serviceprovider/data.go` matches D-6 exactly; constants are the **only** warn-string source (grep no inline duplicates at call sites)
- [ ] Manual boot: `DB_TYPE=legacy` startup output line pasted (matches constant formatting — catches `%s` slips tests' fake can't)

## Rollback

Revert — warns vanish, selection logic (extracted, if extracted) is behaviour-identical so nothing else moves.

## PBI acceptance mapping

**Runtime warning** AC, fully — both the base AC sentence ("logged on startup") and OQ-1's frozen-store
hazard; the strings double as #427's doc source (D-3 → docs handover) and #426's changelog cross-link
subject (04).

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `426-02`.** Cut `slice/426-02` from `sql-migration` HEAD
(`.worktrees/slice-426-02`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
