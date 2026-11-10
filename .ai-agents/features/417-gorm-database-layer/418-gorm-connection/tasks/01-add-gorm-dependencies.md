# Task 01 — Add GORM dependencies (pure manifest PR)

**PBI** [#418](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** none · **PR size:** S
**Plan refs:** [D-1](./plan.md#d-1--new-package-srcdatabase-one-concern-per-file), [AD-2 (feature)](../plan.md#2-architecture-decision-ad-2--sqlite-driver-must-be-pure-go-deviation-from-418s-literal-ac)

## Context

Every later task imports GORM. `go mod tidy` prunes deps with no importer, so dependency-only PRs
are normally impossible — `doc.go` with blank imports solves that and doubles as a compile smoke
test. Keeping this PR to manifest + one 12-line file makes any toolchain fallout instantly
attributable.

## Objective

`gorm.io/gorm`, `github.com/glebarez/sqlite` and `gorm.io/driver/postgres` are in `src/go.mod`
(and their transitive graphs in `src/go.sum`), pinned by a compiling placeholder package. **No
behaviour of the running service changes.**

## Files

| File | Op |
|---|---|
| `src/go.mod` | modify (via `go get`, never hand-edit) |
| `src/go.sum` | modify (generated) |
| `src/database/doc.go` | **create** |

## File size guard

| File | Lines now | Verdict |
|---|---|---|
| `src/database/doc.go` | new, ~12 | ok |
| `src/go.mod`/`src/go.sum` | n/a — not `.go` files | n/a |

## Implementation spec

1. Create `src/database/doc.go` **first** — `go mod tidy` prunes deps with no importer, so with this
   file in place before the manifest commands, tidy *adds and keeps* the three modules. (Learned
   executing slice 418-01: the original order — tidy before `doc.go` — silently reverted the `go get`s.)
2. In `src/` (the module dir; repo root has `go.work` pointing here):
   ```bash
   go get gorm.io/gorm@latest
   go get github.com/glebarez/sqlite@latest
   go get gorm.io/driver/postgres@latest
   go mod tidy
   ```
3. ~~Create `src/database/doc.go`:~~ *(moved to step 1)* Content:
   ```go
   // Package database implements the GORM-based persistence layer of the service
   // (feature #417). This file is created by the dependency-foundation task; the
   // connection factory, config contract, health checks and logger bridge arrive
   // in the follow-up tasks of the same PBI.
   //
   // Driver note (feature decision AD-2): the SQLite dialect is the pure-Go
   // github.com/glebarez/sqlite (modernc.org/sqlite engine) instead of the CGO
   // gorm.io/driver/sqlite named in issue #418, keeping the project CGO-free for
   // cross-compiled release binaries.
   package database

   import (
       _ "github.com/glebarez/sqlite"
       _ "gorm.io/driver/postgres"
       _ "gorm.io/gorm"
   )
   ```
3. Review the `go.mod` diff: exactly three direct `require` lines added (plus indirect block
   churn). If `go mod tidy` bumped *unrelated* deps, revert those lines and re-run tidy — record
   in the PR why. Record chosen versions in the PR description (they become the baseline for #419).
4. Post the AD-2 deviation comment on issue #418 (feature **OQ-5** — copy the deviation paragraph
   from [PBI brief](./brief.md)) and link the PR. Not optional: the closure audit depends on it.

## Out of scope

No `NewDatabaseConnection`, no config, no serviceprovider edits, no `vendor/` (project doesn't
vendor), no CI changes, no touching `src/sql/`.

## Validation

- [ ] `cd src && go build ./...` — green
- [ ] `cd src && go vet ./database/...` — green
- [ ] `cd src && go list -m gorm.io/gorm github.com/glebarez/sqlite gorm.io/driver/postgres` — all three listed with versions
- [ ] `cd src && go test ./... -count=1` — identical pass/fail set to `git stash`-verified baseline (nothing regressed; nothing new yet)
- [ ] `go env CGO_ENABLED=0 && CGO_ENABLED=0 go build ./database/...` — green **with CGO off** (proves AD-2)
- [ ] PR description lists dep versions + unrelated-churn verdict + link to the #418 comment

## Rollback

Single revert of the merge commit: `doc.go` + manifest lines vanish together. No runtime state,
no consumers.

## PBI acceptance mapping

Moves AC “**GORM added**” (with the AD-2 deviation note) and starts “**No regressions**”.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `418-01`.** Cut `slice/418-01` from `sql-migration` HEAD
(`.worktrees/slice-418-01`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
