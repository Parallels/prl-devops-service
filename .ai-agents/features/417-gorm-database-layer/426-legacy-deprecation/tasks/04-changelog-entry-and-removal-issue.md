# Task 04 — CHANGELOG entry + the v2.0.0 removal issue (the deprecation's forward contract)

**PBI** [#426](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** tasks 01–03 (issue body links to *shipped* facts: real marker strings, real grandfather list) · **PR size:** S
**Plan refs:** D-6 (VERSION=1.1.4 → effective **v1.2.0**, removal **v2.0.0**; CHANGELOG format verified — no Unreleased convention, so this PR *creates* the `## [1.2.0]` heading — OQ-426-3 for the release flow) · D-1/D-4/D-5 (issue checklist contents)

## Context

Two artifacts make the deprecation *operational* rather than aspirational: a CHANGELOG entry that tells
operators what changed in v1.2.0 and when it disappears, and a GitHub **removal issue** targeted at v2.0.0
that carries the complete teardown checklist — the single place the grandfather list, sentinel relocation,
env-var retirement, and depguard flip are tracked to zero. The issue creation itself is a human-approved
action (public side effect — D-6's rule: PRs write files, humans post); task 04 therefore *ships the exact
command and body* and its execution is a checked box, not an assumption.

## Objective

CHANGELOG gains the `## [1.2.0] - Unreleased` deprecation bullet (format matched to existing entries — read
top-of-file first); removal issue exists in `Parallels/prl-devops-service` with the full checklist and
back-links; both reference the same versions (v1.2.0 / v2.0.0 — single source D-6).

## Files

| File | Op |
|---|---|
| `CHANGELOG.md` | modify — new top section, one bullet + context sub-lines |
| GitHub issue (external artifact) | create via the command below |
| `.ai-agents/.../426.../brief.md` links | *(parent maintains; PR body carries the issue URL for cross-linking tasks 03's nolint `TBD-04` if it landed)* |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `CHANGELOG.md` | *(measure at task start — read top 60 lines for format first)* | +≤12 | ok |

## Implementation spec

1. **CHANGELOG entry** (Keep-a-Changelog shape per existing sections; date line omitted → the release PR
   fills it, OQ-426-3):

```markdown
## [1.2.0] - Unreleased

### Deprecated

- **Legacy JSON data layer (`src/data` engine + `JsonDatabase`) is deprecated** and scheduled for removal in
  v2.0.0 (#426, feature #417). Default storage is now GORM-backed: SQLite (`DB_TYPE=sqlite`, file
  `prl-devops.db`) or PostgreSQL (`DB_TYPE=postgres` + `DB_DSN`). Existing `data.json` files are migrated
  automatically on first start (#424) and become frozen source-of-past. `DB_TYPE=legacy` remains available
  until v2.0.0 as a documented escape hatch; using it logs a deprecation warning at startup. See the
  configuration & upgrade guides (#427 docs) for details.
```

   (Placement/format per file's real convention — if entries are `## [x.y.z]` with different punctuation,
   mirror it; wording fixed here.)
2. **Removal issue** — command (run by human, output URL pasted back into the PR):

```bash
gh issue create --repo Parallels/prl-devops-service \
  --title "Remove legacy JSON data layer (v2.0.0 breaking removal — #426/#417)" \
  --label "breaking-change,tech-debt" --body-file /tmp/removal-issue-body.md
```

   (`--label` names must be verified against the repo's real labels first — `gh label list`; adjust or omit.)
3. **Body file content** (shipped in the PR description as the authoritative copy — the file itself lives in
   `/tmp`, nothing committed):

```markdown
**Target release: v2.0.0 (breaking).** Follows deprecation from v1.2.0 (#426). Part of feature #417.

## Teardown checklist
- [ ] Regenerate importer list: `grep -rln '"github.com/Parallels/prl-devops-service/data"' src --include='*.go' | grep -v '^src/data/'` → must be empty except the self-package; each survivor blocked on its own removal PR
- [ ] Relocate shared vocabulary OUT of `data` before deletion (#426 D-1 carve-outs): `Err*` sentinels → new `src/dataerrs` (or `interfaces`) package; `Filter/FilterOptions/ParseFilter/FilterByProperty` → neutral helper package (paralledesktop consumers re-point)
- [ ] Delete engine + engine-only decls (`JsonDatabase`, `Data`, layers, `Diff`, `ResetForTesting`…) — `src/data/` retains only `models/` + `interfaces/` (+ relocated vocabulary)
- [ ] Retire file-engine settings + docs: `DATABASE_FOLDER` (superseded by `DB_PATH` — #420/#427), `DATABASE_NUMBER_BACKUP_FILES`, `DATABASE_SAVE_INTERVAL_MINUTES`, `SYSTEM_AUTO_RECOVER_DATABASE` (AD-4: legacy `startup/migrations` migrators retire in the same release — data.json is frozen since #424)
- [ ] `DB_TYPE=legacy` branch + legacy goroutines (save-queue/backup/stale-job legacy side) + `Legacy()` escape hatch + provider legacy fields + `selectDatabaseService` legacy arm; drop the grandfathered `//nolint:staticcheck // #426` lines (grep `#426` for the full set)
- [ ] depguard: flip `.golangci.yml` — deny `github.com/Parallels/prl-devops-service/data` **with** subpaths (models/interfaces move to their own allow or path-split — decide then), delete `#426 grandfather` exclusions, keep `go-legacy-guard.yml` as permanent import-ratchet precedent (extend pattern for future deprecations)
- [ ] CHANGELOG v2.0.0 breaking-change entry; docs: remove legacy sections (#427 pages), README pointers
- [ ] Freeze-window guard: `legacy_migration_marker`-present deployments are the migration-complete signal — removal postdates their adoption metrics (operator decision, PO)

## Links
Deprecation shipped in: #426 (annotations `// Deprecated:` ×175, warns, depguard CI) · Feature: #417 ·
Migration: #424 · Tests: #425 · Docs: #427
```

4. **Back-links:** comment on #426 (and #417) with the new issue URL + CHANGELOG commit SHA; if task 03's
   nolints carry `TBD-04`, that PR (or a micro-PR from this one) replaces them with the real number.

## Out of scope

Issue *labels* creation (only use existing), CHANGELOG release mechanics (dates, version bump — release
flow owns; VERSION file untouched at 1.1.4 per D-6), any code change.

## Validation

- [ ] `git diff CHANGELOG.md` — section added above the current top entry, format-consistent (screenshot/paste in PR)
- [ ] Issue URL pasted; issue body renders (checklist boxes real); title/links verified
- [ ] Version-string sweep: `grep -rn "v1.2.0\|v2.0.0" CHANGELOG.md` ↔ marker strings in task 01's commit ↔ constants in task 02 — three-way identical (paste the grep)
- [ ] `cd src && go build ./... && go test ./... -count=1` still green (docs-only PR — ER-5 trivially)

## Rollback

Revert CHANGELOG hunk + close the issue with a comment (closing ≠ deleting; the plan text survives in this
folder regardless — the specs folder is the durable record).

## PBI acceptance mapping

**Removal issue created** AC — the deliverable itself; CHANGELOG completes the operator-facing deprecation
contract (AC's deprecation-notice spirit), and the issue's checklist encodes every carve-out D-1 promised so
nothing gets "cleaned up" accidentally at v2.0.0.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `426-04`.** Cut `slice/426-04` from `sql-migration` HEAD
(`.worktrees/slice-426-04`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
