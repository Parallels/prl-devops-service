# Task 01 — Deprecation annotation sweep: 175 `// Deprecated:` markers, one comment-only commit

**PBI** [#426](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** #421 task 18 (final flip) merged · **PR size:** M (mechanical)
**Plan refs:** D-1 (surface — the 175 number + carve-outs) · D-2 (mechanism + marker format + gates) · D-6 (version strings in the message)

## Context

The AC "all public types/functions marked `Deprecated:`" has an empirically-forced implementation:
staticcheck (as bundled in golangci-lint 2.14.0, the version used across this plan's verification) does **not**
propagate a type-level `Deprecated:` to method call sites (D-1 scratch-test evidence) — so per-method markers
are not churn-maximization, they're the minimum that makes the AC's diagnostics story real. The sweep covers
`src/data/*.go` **non-recursively** (subpackages `models`/`interfaces` structurally unreachable — they're LIVE
and must stay marker-free), touches 157 methods + 17 decls + 1 package-doc line = **175**, and is executed by
a throwaway script run from `/tmp` so the repo gains zero codegen files (D-2).

## Objective

One commit, 175 comment-line additions, `go build`/`go test`/`gofmt` all green, diff provably comment-only —
after which `Legacy()` call sites light up SA1019 in every IDE, and #426's remaining tasks have their
foundation.

## Files

| File | Op |
|---|---|
| `src/data/*.go` (19 engine files, non-test) | comment-only edits (per-file marker counts table below) |
| `src/data/main.go` package doc | + the package-level `Deprecated:` paragraph (D-1; import-line SA1019 is intended, D-4 grandfathers every current importer) |

## File size guard

| File | Lines | After | Verdict |
|---|---|---|---|
| `src/data/orchestrator.go` | 933 | 933 + 31 method markers (+ comments above decls — **counts as growth; 964 < 1,000 = ok**) | ok-watch: this PR is the ONLY comment growth the file gets; if any file would cross 1,000 the script skips it and its markers move to a follow-up (none do today — verify per file with `wc -l` before/after in PR) |
| all others | < 800 | +≤40 each | ok |

## Implementation spec

1. **Per-file marker inventory** (methods = `grep -c "func (j \*JsonDatabase) [A-Z]"`; extra decls listed in
   D-1 — regenerate the table into the PR description at task start; expected total 175):
   per engine file: methods count + extras (`main.go`: types `Data`/`JsonDatabase`/`JsonDatabaseConfig`,
   `NewJsonDatabase`, `IsRecordLocked`, `LockRecord`, `UnlockRecord`; `order_layer.go`: `OrderDirection`,
   `Order`, `OrderByProperty`; `index_layer.go`: `GetRecordIndex`; `diff.go`: `Diff`;
   `authorize_layer.go`: 4 decls; `testing_reset.go`: `ResetForTesting` (integration-tagged file — markers
   still apply)).
2. **Carve-out enforcement (script-level, not review-level):** the script matches declaration heads and
   **skips** `Err[A-Z]\w* *=` lines, and the filter-layer symbols `Filter`, `FilterOptions`,
   `ParseFilter`, `FilterByProperty` (D-1 carve-out #2 — still referenced by parallelsdesktop; re-run its
   audit command at task start; any carve-out symbol that became unreferenced by #421/#425 *gets* the
   marker, and still-referenced ones get the dated relocation NOTE comment).
3. **Insertion format (gocritic `deprecatedComment` rules — own paragraph, last):**

```awk
# run from /tmp: awk -i inplace '... (full script written inline in the PR body) ...' src/data/*.go
#   for each match of ^(func \(j \*JsonDatabase\) [A-Z]|type (Data|JsonDatabase|JsonDatabaseConfig) |
#     func NewJsonDatabase|func (IsRecordLocked|LockRecord|UnlockRecord|Parse...)...)
#   insert DIRECTLY ABOVE the decl (breaking any existing doc-comment block is WRONG — insert AFTER the
#   doc block, before the decl):
#   "//\n// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase\n"
#   "// via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0."
```
   (D-2 fixed the exact sentence incl. v1.2.0/v2.0.0 from D-6; the script is embedded in the PR body, run
   against a scratch clone first, diff eyeballed once, then the same diff applied here.)
4. **Package doc** (`main.go` top comment): append the paragraph per D-1 (wording: package-level deprecation
   naming the replacement and removal version).
5. Post-sweep: `gofmt -l src/data/` empty; script's `/tmp` copies deleted (nothing in `scripts/` — repo has
   none and gains none).

## Out of scope

Runtime warns (task 02), depguard/CI (task 03), CHANGELOG/issue (task 04), any non-comment edit, any edit
under `src/data/models/` or `src/data/interfaces/` (forbidden — verify via diff path filter), nolint
annotations (task 03's audit decides those, D-5), test files.

## Validation

- [ ] Comment-only gate (D-2's exact command): `git diff --unified=0 | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | grep -vcE '^\+\s*//|^-\s*//|^$'` → **0**
- [ ] Marker count: `grep -rc "// Deprecated:" src/data --include='*.go' | awk -F: '{s+=$2} END {print s}'` → **175** (or regenerated target ± carve-out adjustments recorded in PR)
- [ ] Carve-out negative check: `grep -B2 "ErrDatabaseNotConnected" src/data/main.go | grep -c "Deprecated"` → 0 (repeat for one sentinel per domain + the 4 filter symbols)
- [ ] `cd src && go build ./... && go test ./... -count=1` (ER-5 — comment-only by construction)
- [ ] `cd src && gofmt -l .` → empty; `wc -l` table of `src/data/*.go` before/after in PR (no file ≥ 1,000)
- [ ] staticcheck spot-proof (local, golangci-lint 2.14.0): a scratch `main` referencing `data.JsonDatabase` **reports SA1019** — evidence the markers work, per D-1's scratch-test protocol

## Rollback

`git revert` — comment-only, nothing can dangle (no code, no behaviour, no API changed).

## PBI acceptance mapping

**Annotation** AC (the 175 + package doc, script-mechanical, carve-outs evidenced) and the **diagnostics**
half of **No new callers** (SA1019 fires in IDE/lint — depguard, task 03, is the CI half); removal-version
strings appear exactly as D-6 fixed them.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `426-01`.** Cut `slice/426-01` from `sql-migration` HEAD
(`.worktrees/slice-426-01`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
