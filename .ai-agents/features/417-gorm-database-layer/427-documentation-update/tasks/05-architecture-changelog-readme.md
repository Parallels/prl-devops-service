# Task 05 — Architecture section · CHANGELOG feature entry · README pointer (the linking close)

**PBI** [#427](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** tasks 02, 03, 04 (every link it writes must exist) · **PR size:** S
**Plan refs:** D-5 (placement decisions — architecture paragraph + README one bullet + CHANGELOG cross-ref) · D-6 (toolchain) · D-2 (evidence protocol)

## Context

The closing PR of the feature's documentation. Three small files, one rule: **it links, never restates**
(D-5: removal versions live in #426's CHANGELOG entry — a link; config specifics live in task 01's page —
links). The architecture doc gains a `## Data Persistence` section stating today's truth: SQLite file /
PostgreSQL via GORM as default engine, with the JSON file as frozen-legacy input; the CHANGELOG gets the
**feature** entry (distinct from #426's deprecation entry — cross-referenced, not duplicated); README stays
2 KB of pointers (one bullet in its `## Documentation` section — read it, mirror its bullet format).

## Objective

A newcomer's path works with zero tribal knowledge: README → configuration page (what are the settings) →
upgrade guide (what happens at 03:00 tonight) → Postgres page (how to choose the better engine) — plus the
architecture doc no longer describes a system that doesn't exist after the flip.

## Files

| File | Op |
|---|---|
| `docs/docs/devops/getting-started/architecture.md` | modify — `## Data Persistence` section (~20 lines: engine diagram sentence, SQLite/Postgres paths, migration + marker behavior **linked** to upgrade guide, legacy-frozen one-paragraph note per D-5, backup posture gap honestly noted per #421 OQ-421-3 → points to operator guidance section of postgres page / file-copy for SQLite) |
| `CHANGELOG.md` | modify — feature entry under the topmost release heading convention (`## [X.Y.Z] - …`, newest on top, date/version = maintainer call at merge — leave `X.Y.Z` decision flagged in PR, propose next release after 1.1.4; **cross-ref #426's deprecation entry**, don't restate v2.0.0 text beyond the link per D-5) |
| `README.md` | modify — one bullet under `## Documentation`: links to the three pages (config / database-upgrade / postgresql) |

## File size guard

| File | Lines now | After | Verdict |
|---|---|---|---|
| `README.md` | ~60 (measured earlier — 2,088 B) | +3 | ok — README stays a pointer, no content migration |
| `architecture.md` | *(measure)* | +≤25 | ok |
| `CHANGELOG.md` | (04 of #426 may have touched it — rebase on that hunk if #426's task 04 merged first; feature entry sits **above** the deprecation entry only if version order says so) | +≤10 | ok |

## Implementation spec

1. **Architecture section** — accuracy protocol binding (D-2): every behaviour sentence cites merged-code
   evidence in an HTML comment (`engine default ← src/serviceprovider selection switch post-#421`; `prl-devops.db`
   ← `src/constants`/config code; marker/frozen semantics ← `legacy_migrator/marker.go` + #426 warn strings).
   Deprecation note paragraph: exactly D-5's one-paragraph scope, removal version **only as a link** to the
   #426 CHANGELOG anchor / removal issue.
2. **CHANGELOG feature entry** — bullets: GORM layer (`DB_TYPE` sqlite/postgres), config settings + flat
   config-file keys (one-line mention + link to configuration page), automatic first-start migration (#424),
   deprecation cross-ref line (`See also: Deprecated entry for v1.2.0 (#426)`), docs links. Format: mirror
   the two newest existing entries **exactly** (heading punctuation, bullet style — read before writing).
3. **README bullet** — one line, existing bullet syntax, three links (relative docs paths verified).
4. **Link integrity sweep** (the task's own gate, commands in Validation) across **all five pages** this PBI
   created/edited (01–04 outputs + these): internal anchors exist, relative paths resolve from the file's
   location, no link to `docs/_site`, no `database.type` dotted keys anywhere (D-3 re-check — grep).

## Out of scope

Restating other tasks' content (links only — D-5), site nav/menu (02/03 own it; verify not edit), version
numbers invention (maintainer call, PR-flagged), reformatting anything (D-6: zero-new-violations regime),
`docs/_site`.

## Validation

- [ ] `export npm_config_cache=/tmp/pbi427-npmcache; npx markdownlint-cli <the three files>` — zero **new** violations vs baseline (CHANGELOG/README pre-existing failures per D-6 table left untouched; record the two baselines in PR)
- [ ] `npx prettier@3.6 --check README.md` passes (it did pre-change — must stay true); architecture/CHANGELOG judged vs their pre-existing status (never `--write`)
- [ ] jekyll build-if-feasible (D-6): all 01–05 pages render; link sweep from `/tmp/pbi427-site` output (grep for `href` targets resolving; report zero dead links touching PBI-427 pages)
- [ ] D-3 negative grep: `grep -rn "database\.type\|database\.path\|database\.dsn\|database\.auto_migrate" docs/docs/ README.md CHANGELOG.md` → **empty**
- [ ] Link targets exist in-tree: `ls docs/docs/devops/getting-started/configuration.md …/database-upgrade.md …/postgresql.md` + CHANGELOG anchor + README bullets' relative paths from `README.md` context
- [ ] `git diff --stat` → exactly the three files

## Rollback

Revert — three-file docs change; links it owns were terminal (nothing links *into* this PR's additions except
each other within the PBI).

## PBI acceptance mapping

**Architecture doc + CHANGELOG + README** AC boxes close here; D-5's single-source-of-truth deprecation
placement completes (page 02 = home, pointer in 01, paragraph here, versions never restated); the newcomer
path closes. This PR = #427 done = feature #417 documentation contract satisfied.
