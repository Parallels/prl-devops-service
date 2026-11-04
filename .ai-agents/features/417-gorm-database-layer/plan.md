# Feature Plan — Database Layer Migration to GORM (#417)

> Companion to [`brief.md`](./brief.md). This document locks the architecture decisions for the
> whole feature, the engineering rules every task must obey, the delivery order, and the loop used
> to deep-dive each PBI. **Changes here require human approval**; PBIs may not reinterpret them.

Status: **locked** 2026-11-03. Decision records are append-only — superseding one adds `Superseded by AD-n`, never deletes.

---

## 1. Architecture Decision AD-1 — Strangler fig via a `DatabaseService` interface

**Context.** `serviceprovider.GetDatabaseService()` returns the concrete `*data.JsonDatabase`
(`src/serviceprovider/data.go:8`). ~143 non-test call sites bind to that concrete type, invoking
its 157 public methods. PBI #426 explicitly assumes the legacy layer stays "reachable via a
feature flag or fallback" and logs a warn — implying runtime coexistence, not a hard cutover.

**Decision.**
1. Introduce `DatabaseService` — an interface capturing the 157-method surface, split by domain
   into one file per domain under `src/data/interfaces/` (e.g. `users.go`, `roles_claims.go`,
   `api_keys.go`, `catalog.go`, `catalog_managers.go`, `packer_templates.go`, `orchestrator.go`,
   `snapshots.go`, `jobs.go`, `reverse_proxy.go`, `user_config.go`, `enrollment_tokens.go`)
   composed into one aggregate interface (`database_service.go`). The existing unused
   `DatabaseContext` stub in `interfaces/database_context.go` stays untouched (no references).
2. New package `src/database/` holds `GormDatabase` — implements `DatabaseService` via GORM.
   Import direction: `database → data/interfaces → data/models`; `data` and `database` never
   import each other's engine code; `serviceprovider` imports both. No cycles.
3. `JsonDatabase` continues to satisfy the same interface unchanged (methods already match), and
   is marked deprecated in PBI #426.
4. `serviceprovider` selects the implementation at init from the `DB_TYPE` configuration
   (PBI #420 supplies the config; until then the default `sqlite` only selects GORM once #421
   flips the switch). `GetDatabaseService` return type becomes `DatabaseService`.
5. Call sites only *invoke methods* on the returned value, so they compile unchanged. The escape
   hatches (`SaveAs`, `SaveNow`, `Backup` used in `src/main.go:59` and `src/data/backup.go`) stay
   reachable through a `Legacy() *JsonDatabase` accessor on the provider — **task list of #421
   must enumerate all concrete-type usages before flipping the switch.**

**Why chosen over alternatives.** Hard cutover (rewrite `JsonDatabase` internals in place) has no
runtime fallback and forces one giant risky PR. Parallel-package domain-by-domain rewrite touches
~143 call sites ~once per domain and diverges from the issue text. Strangler gives small PRs,
a runtime fallback, and matches #426's acceptance criteria literally.

## 2. Architecture Decision AD-2 — SQLite driver must be pure Go (deviation from #418's literal AC)

#421's constraint demands the pure-Go `modernc.org/sqlite` (no CGO). `gorm.io/driver/sqlite` is
CGO (`mattn/go-sqlite3`) despite being named in #418's AC text. **Decision:** use
`github.com/glebarez/sqlite` — the maintained GORM dialect built on `modernc.org/sqlite`.
PostgreSQL uses the canonical `gorm.io/driver/postgres`.
**Required:** every PR touching this states the deviation in its description, and a comment is
posted on #418 so the literal AC box is checked with the known deviation. Rationale: the project
ships static multi-platform binaries (Dockerfile, Helm charts); CGO would break cross-compilation.

## 3. Architecture Decision AD-3 — SQLite default path, preserving the file-path convention

The legacy convention (brief §5) is folder-based: root → `/etc/parallels-devops/`, else
`~/.parallels-devops/`, overridable by `DATABASE_FOLDER`; legacy file `<folder>/data.json`.
**Decision:** GORM SQLite default = **`<same folder>/prl-devops.db`** — same folder, distinct
filename, so legacy coexistence, backups and user muscle-memory all survive the upgrade.
`:memory:` is accepted as a special path value for tests (#425 depends on it). The factory may
resolve the folder itself (uid/home rules) without the `System` service; #420 owns the final
`DB_PATH` resolution and may refine this helper — do not duplicate the logic in a second place.

## 4. Architecture Decision AD-4 — Schema management coexists with legacy schema migrations

`src/startup/migrations/` is a versioned migration framework **for the JSON store** (marker lives
in `Data.Schema`). GORM `AutoMigrate` manages the **new** schema only, gated by
`DB_AUTO_MIGRATE` (default true). #424's completion marker is a row in a new table
(`schema_migrations` or equivalent) **in the new database** — never in `data.json`. On a legacy
fallback deployment, GORM AutoMigrate and the legacy migrators each only ever touch their own
store. Ordering at startup: legacy schema migrations (unchanged) → GORM AutoMigrate → legacy data
migration (#424) → service up.

## 5. Architecture Decision AD-5 — Migration source reads the legacy store through its own loader

The legacy file may be encrypted (`EncryptionPrivateKey`). #424 must **reuse** `JsonDatabase`'s
load path (`Load`/`loadFromFile`, possibly via a read-only constructor) instead of duplicating
decrypt/repair logic. Migration reads the in-memory `Data` struct and inserts through GORM
transactions; it never parses `data.json` by hand.

## 6. Engineering rules (ER) — binding on every task in every PBI

- **ER-1 — Follow the existing code structure.** New code lives in the established domain layout
  (`src/<domain>/`) and reuses project idioms: `basecontext.ApiContext` as first parameter,
  `errors.NewWithCode` for domain errors, `config.Get().GetKey` for config reads,
  `ctx.Log*` for logging, testify for test assertions, `//go:build integration` for tests
  needing external services. No new frameworks, no unrelated re-plumbing, no drive-by renames.
- **ER-2 — Hard cap: no `.go` file is created or left above 1,000 lines.** This rule was set by
  the product owner for this feature and **supersedes** the 500-line guideline in `CLAUDE.md`
  where they conflict (user instruction wins; the 500-line number stays as "aim for less").
- **ER-3 — Refactor-on-touch.** If a task would edit a file already ≥ 1,000 lines, or the edit
  would push it past 1,000, the task must **first** split that file into cohesive same-package
  files as a *pure code move* (zero behaviour change, build+tests green after the split), then do
  the feature work on top. Split commit and feature commit stay in the same PR but must be
  separate commits so review sees the move cleanly.
- **ER-4 — Every task spec carries a `File size guard` table**: each file it creates/edits with
  `wc -l` at plan time and a verdict `ok` or `refactor-first`.
- **ER-5 — One task = one PR = the repo builds and all tests pass at the merge boundary.**
  Behaviour-changing flips (e.g. #421 switching a domain to GORM) ship behind the config default
  (`sqlite`/legacy selection) only after the fallback path is proven.

### Pre-seeded ER-3 trigger list (measured 2026-11-03 — re-run `wc -l` before implementing)

| File | Lines | Trigger |
|---|---|---|
| `src/config/main.go` | 1,013 | **Already over cap.** Any #420 edit ⇒ split first (suggested: `database.go`, `modules.go`, `logging.go`, keep `main.go` core). Owned by #420's task chain. |
| `src/config/models.go` | **72** (corrected) | **NOT a trigger** — pre-seeded figure (1,085) was wrong; corrected by #420's agent measurement (2026-11-03). No split needed in this feature. |
| `src/data/models/virtual_machine.go` | 1,018 | Already over cap. #419 adds tags ⇒ split first. |
| `src/data/orchestrator.go` | 933 | Near cap. #421's orchestrator-domain methods will breach ⇒ split as part of that domain task. |
| `src/data/catalog.go` | 714 | Watch. #421 catalog methods added to the *new* package, not here — stays safe. |
| `src/serviceprovider/parallelsdesktop/main.go` | 3,079 | Over cap but **out of this feature's scope** — do not touch; not our debt to pay here. |
| `src/serviceprovider/system/main.go` | 1,076 | Same — out of scope. |

New `src/database/` files are designed small from the start (≤ ~200 lines each). The 157-method
interface is split per-domain from day one (AD-1) so it can never breach the cap.

## 7. Target package layout (end state)

```
src/
├── data/                     # legacy — deprecated at the end (#426), still default until #421 flips
│   ├── interfaces/           # NEW: DatabaseService (aggregate) + per-domain interface files
│   └── models/               # gains gorm tags (#419); virtual_machine.go split first (ER-3)
├── database/                 # NEW package — the GORM implementation
│   ├── config.go             #   DatabaseConfig contract (#418; fed by #420)
│   ├── connection.go         #   factory + Close (#418)
│   ├── sqlite.go             #   SQLite branch, WAL/busy_timeout pragmas (#418; #421 may extend)
│   ├── postgres.go           #   Postgres branch + pool (#418)
│   ├── redact.go             #   DSN redaction (#418, AC #420 "secrets safety" support)
│   ├── health.go             #   ping with retry (#418)
│   ├── logger.go             #   gorm log adapter onto basecontext (#418)
│   ├── legacy_migrator/      #   #424: detection, mapping, validation, dry-run
│   ├── gorm_database.go      #   #421: GormDatabase type + bootstrap (stays small)
│   └── <domain>.go           #   #421: one file per domain implementing that domain's interface
└── sql/                      # unused mysql stub — leave untouched
```

## 8. Delivery order & dependency graph

```
#418 connection ──┬─→ #420 configuration ──→ (feeds 419/421/424)
                  ├─→ #419 models/tags  ──→ AutoMigrate schema exists
                  └─→ #421 SQL backend (needs 418+419+420; largest PBI, many PRs)
                          └─→ #424 legacy migration (needs 419+421 + a stable schema)
                                  └─→ #425 test suite completion (runs throughout too)
#426 deprecation ── after #421 ships (annotations + lint + changelog)
#427 docs ────────── last (#427's own constraint: "one of the last merged")
```

Merge order: **418 → 420 → 419 → 421 → 424 → 425 → 426 → 427.**
Rationale: 420 before 419/421 because factory & models consume config; 419 before 421 because
queries need tagged models; 424 after 421 because mapping needs the query layer; each earlier PBI
still ships its own tests (ER-5), #425 completes suite-level parity and edge cases.
#421 is expected to decompose into per-domain task chains (users → roles/claims → api keys →
config → catalog → orchestrator → jobs → proxies → snapshots) with the interface + filter/order/
index/authorize layers as its foundation tasks.

## 9. Risk register

| # | Risk | Mitigation |
|---|---|---|
| R-1 | Interface capture drift: a `JsonDatabase` method changes signature between #421 foundation and domain PRs | Generate interface in one PR; CI compile-time assertion `var _ DatabaseService = (*JsonDatabase)(nil)` |
| R-2 | Embedded slices (`User.Roles`, `User.Claims`, role→claims…) → normalised join tables; round-trip fidelity risk | #419 designs schema + round-trip tests; #424 validation step counts records both sides |
| R-3 | Mongo-style `ParseFilter` semantics must survive → SQL WHERE translation parity | #421 foundation task ports filter layer with differential tests (same fixtures, both engines) |
| R-4 | Global singleton (`memoryDatabase`, provider) makes tests order-dependent | New package holds no globals; #425 keeps per-test in-memory DBs |
| R-5 | SQLite write concurrency under load | WAL + busy_timeout pragmas (#418); documented single-writer guidance (#427) |
| R-6 | Encrypted legacy file unreadable during migration | AD-5: reuse loader; #424 dry-run + rollback (transaction) + never delete legacy file |
| R-7 | Startup cost on large legacy stores blocks orchestrated liveness | #424 progress logging; readiness before serving; documented timeout guidance |
| R-8 | Postgres behaviour differences (types, constraint errors) | #425 differential tests behind `//go:build integration` + docker-compose `postgres` service |
| R-9 | Hidden >1,000-line refactors inflate small tasks | ER-3 + trigger list above; guard table in every task spec |

## 10. The per-PBI deep-dive loop (repeat per PBI, in order)

1. Re-read: this `plan.md` + the PBI's `brief.md` (already contains verbatim ACs from GitHub).
2. Re-verify repo anchors cited in the brief (file paths, `wc -l`) against current HEAD; update the
   brief if stale.
3. Design: write the PBI's `plan.md` (sections: Scope, Design decisions D-n with rationale, Target
   files, Data/contract sketches, Risks, Open questions, Task index table).
4. Decompose into `tasks/NN-<slug>.md` using the template in §11 — one PR per task, each
   self-validating (ER-5), file size guard completed (ER-4).
5. Sanity pass: tasks ordered so partial completion never breaks main; no two unfinished tasks
   edit the same file.
6. Stop and present to the human for review before the next PBI starts. Implementation of a PBI
   may begin task-by-task as its tasks get approved — but never ahead of the sequence.

## 11. Document & task templates (canonical)

**PBI `brief.md`**: header table (number/URL/labels/assignee/round) · verbatim Description, User
Story, Acceptance Criteria (checkboxes), Definition of Done, Assumptions & Constraints,
Dependencies · Repo context (verified anchors) · Fit in the feature (upstream/downstream PBIs) ·
Engineering-rule triggers for this PBI.

**PBI `plan.md`**: Scope (in/out) · Design decisions `D-n` · Target files & layout · Contract
sketches · Risks · Open questions · Task index table (file → goal → depends-on → touched files →
PR hint).

**Task spec `tasks/NN-<slug>.md`**:
```
# Task NN — <title>            PBI #NNN · Feature #417 · Depends-on: task(s)/PR(s) · PR size: S/M
## Context                     1 paragraph — why this exists
## Objective                   single observable outcome
## Files                       create/modify list, exact paths
## File size guard             wc -l per touched file → ok | refactor-first (+ split plan)
## Implementation spec         signatures, code sketches, edge cases, error handling, idioms (ER-1)
## Out of scope                explicit exclusions
## Validation                  numbered commands + behavioural checks (checkboxes)
## Rollback                    revert surface / feature-flag off
## PBI acceptance mapping      which AC checkboxes this task moves
```

## 12. Test strategy (summary — #425 owns the details)

- Unit tests: in-memory SQLite (`:memory:` / `file::memory:?cache=shared`) through the real
  factory — no mocks of GORM; disk-free, order-independent, bounded (+≤30 s suite budget from #425).
- Postgres paths: unit-testable parts offline (DSN validation, pool settings, redaction); real
  behaviour under `//go:build integration` + a `DB_TYPE=postgres` docker-compose service.
- Migration tests: synthetic `data.json` fixtures (empty, partial, encrypted, re-run after
  success) driven through #424's package API.
- Differential tests: same dataset in legacy vs GORM engines asserting identical list/filter
  output (#421/#425).

## 13. Open questions (resolve in each PBI's round, keep human in the loop)

| ID | Question | Blocks |
|---|---|---|
| OQ-1 | SQLite `data.json` → does the *legacy file* become read-only after migration or keep syncing while the fallback exists? (proposal: frozen, warn on fallback start) | #424, #426 |
| OQ-2 | Jobs collection retention policy when moving to SQL (legacy keeps growing all jobs forever?) | #421 |
| OQ-3 | Table naming: keep GORM defaults (snake_case plural) or add `TableName()` overrides mirroring legacy JSON collection names? (proposal: GORM defaults; document mapping) | #419 |
| OQ-4 | Docker/Helm: does the chart need a `DB_TYPE`/`DB_DSN` values surface in this feature? (proposal: yes, minimal, #427 docs + helm values passthrough env) | #427 |
| OQ-5 | #418 AC names `gorm.io/driver/sqlite` literally — AD-2 deviation needs a comment on the issue so closure is honest. Post it during task 01's PR. | #418 |
