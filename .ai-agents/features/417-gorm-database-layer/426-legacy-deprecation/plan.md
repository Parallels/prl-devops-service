# PBI Plan #426 — Deprecation of the legacy data layer

> Reading order: [feature brief](../brief.md) → [feature plan](../plan.md) (AD-1, AD-4, OQ-1, ER-1…ER-5) →
> [PBI brief](./brief.md) → this plan → your task spec.
>
> Prerequisite: **#421's final flip has merged** — `GetDatabaseService()` returns
> `interfaces.DatabaseService`, the legacy engine is bypassed-by-default, `Legacy()` exists. Without
> that, deprecating in GoDoc would be a lie (PBI brief, "Depends on").

All tool-behaviour claims in this plan were **verified empirically on 2026-11-03** against
`golangci-lint 2.14.0` (staticcheck built-in, gocritic built-in) in a scratch module mirroring this
repo's import paths; the verification recipes are re-embedded in the tasks so they can be re-proved
at implementation time.

## Scope

**In:** `// Deprecated:` annotation of the legacy engine (package `data` doc + `JsonDatabase` + its
157 exported methods + engine-level package decls) via one scripted mechanical sweep · formalized
runtime deprecation warn + the OQ-1 "frozen legacy" warn in the serviceprovider selection path (with
unit tests) · `depguard` rule in `.golangci.yml` + a dedicated CI guard workflow · CHANGELOG entry ·
creation of the removal GitHub issue.
**Out:** any deletion of legacy code · moving/renaming symbols (sentinel relocation and helper
extraction are **removal-time** work, listed in the removal issue) · `src/data/models` and
`src/data/interfaces` (LIVE — wire format + GORM models + the interface) · ER-3 splits of
out-of-scope files (`serviceprovider/parallelsdesktop/main.go` etc.) · docs pages (#427 carries the
notice) · the `#417` feature-level docs.

## Design decisions

### D-1 — Deprecation surface: engine-only, script-driven, with two empirically-forced consequences

**What gets `// Deprecated:`** (measured 2026-11-03, `src/data/`):
- the package doc comment on `src/data/main.go` ("Package data … Deprecated: …");
- `type JsonDatabase`, `type Data`, `type JsonDatabaseConfig`, `func NewJsonDatabase`
  (`main.go:33,53,67,75`), plus engine free functions `IsRecordLocked`, `LockRecord`,
  `UnlockRecord` (`main.go:399,406,419`);
- **all 157 exported methods** `func (j *JsonDatabase) [A-Z]` (verified count: 157 across the 19
  engine files; per-file table in task 01);
- engine layer decls `OrderDirection`, `Order`, `OrderByProperty` (`order_layer.go`),
  `GetRecordIndex` (`index_layer.go`), `Diff` (`diff.go`),
  `AuthorizedRecord`, `GetAuthorizedRecords`, `IsRootUser`, `IsAuthorized` (`authorize_layer.go`),
  `ResetForTesting` (`testing_reset.go`, integration-tagged). No cross-package live callers for
  these (audit in D-5);
- total markers: 157 + 17 + 1 (package doc) = **175**.

**Carve-outs (must NOT carry `Deprecated:`)** — each justified because the symbol survives the
legacy engine and belongs to the live path, exactly like `data/models` / `data/interfaces`:
1. **`Err*` sentinel vars** (`ErrDatabaseNotConnected` etc., `main.go:21-24`, `roles.go:13-20`,
   `claims.go:13-18`, more). They are the repo's shared error vocabulary — `#421`'s `GormDatabase`
   returns the same sentinels (`#421` brief: "keep `ErrXxx` sentinels"), and cross-package
   consumers exist today (`src/catalog/*`, `src/cmd/root_password.go`, `src/startup/seeds/*`,
   `src/startup/migrations/add_api_key_own_claims.go`). Deprecating them would poison the *new*
   layer's error handling. Removal-time: relocate them out of `data` (removal-issue checklist).
2. **`Filter`, `FilterOptions`, `ParseFilter`, `FilterByProperty`** (`filter_layer.go`) — still
   referenced by `src/serviceprovider/parallelsdesktop/main.go` (`ParseFilter` ×2,
   `FilterByProperty` ×2 — measured); these are generic in-memory list helpers independent of the
   engine. Re-run the D-5 audit at implementation time: symbols that became unreferenced by then get
   the marker too; still-referenced ones stay carve-outs with a dated
   `// NOTE(#426): relocated out of package data at removal (see removal issue).`
3. Everything under `src/data/models/` and `src/data/interfaces/` — **untouched** (LIVE). The sweep
   script globs `src/data/*.go` non-recursively — subpackages are structurally unreachable by it.

**Why per-method markers (full sweep = option (a) from the brief) and not type-level only** —
empirical finding: with `staticcheck` as bundled in golangci-lint 2.14.0, marking only the *type*
does **not** make `SA1019` fire at call sites of its methods (scratch test: `SaveNow` with no own
marker → not reported; `SaveAs` with own marker → reported). Type-level-only would leave the AC's
diagnostics story empty; the full 175-marker sweep is one mechanical comment-only commit and is the
honest reading of "All public types, functions … annotated". The churn concern is answered by
mechanical generation + whole-diff review (D-2), not by weaker annotations.

**Package-level `Deprecated:` has a verified side effect**: staticcheck flags the **import line** of
a package whose doc comment is deprecated (`SA1019: github.com/…/data is deprecated: …`). That
lights up *every* importer — which is fine and desired: all current importers are grandfathered in
the shared exclusions (D-4), and any *new* importer is simultaneously caught by depguard (CI) and
SA1019 (IDE / full lint). Double-lock, one exclusion list.

### D-2 — Annotation mechanism: throwaway script, one comment-only commit, nothing committed but code

The sweep is a **one-shot awk script embedded in task 01's spec** and run from `/tmp` on the
implementer's machine — nothing lands in `scripts/`, `src/`, or the repo (the repo has no `scripts/`
dir today; creating one for a single-use codegen file violates "no unnecessary files"). Diff review
works because every added line is a comment: the validation gate asserts
`git diff --unified=0 | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | grep -vcE '^\+\s*//|^-\s*//|^$'`
== 0 added/changed non-comment lines, plus `go build ./...`/`go test ./...` green and `gofmt -l`
empty. Marker format is **fixed by gocritic's `deprecatedComment` check** (enabled via the repo's
`.golangci.yml` `gocritic.enabled-checks` — and note golangci warns "already enabled" because
current gocritic always includes it): the marker must be **its own paragraph, last in the doc
comment** — verified violation messages. So the script inserts, directly above the declaration:
`//` + `// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v2.0.0.`
(1 line when the symbol has no existing doc comment — the paragraph rule only applies *after* text).
Version numbers come from D-6 (`v1.2.0` effective, `v2.0.0` removal) and appear in the message.

### D-3 — Runtime warnings: formalize the selection warn, add the OQ-1 frozen-legacy warn

The serviceprovider selection branch that picks `*data.JsonDatabase` (introduced by #421's switch in
`src/serviceprovider/data.go`) already logs a warn (per #418/#421 wiring). This PBI pins the exact
strings (constant `legacyDeprecationWarnMsg` + `legacyFrozenWarnMsg` next to the selection code —
one source of truth, testable):

1. **Selection warn** — emitted at `LogWarnf` on every startup where the legacy engine is selected:
   `[Database] DEPRECATED: legacy JSON data layer selected (DB_TYPE=%s). The legacy layer is deprecated since v1.2.0 and will be removed in v2.0.0 (issues #426/#417). Plan migration to DB_TYPE=sqlite — see CHANGELOG.`
   (`%s` = the legacy DB_TYPE value — **#421/#420 own the literal**; re-grep the selection constant
   at implementation time.)
2. **Frozen-legacy warn (OQ-1)** — legacy selected **AND** #424's completion marker
   (`legacy_migration_marker`, row in the *new* database's migrations table per AD-4, owned by
   `src/database/legacy_migrator/marker.go`):
   `[Database] DEPRECATED: legacy data file %s was already migrated and is now FROZEN/stale — this instance serves stale data; writes here never reach the SQL database. Switch to DB_TYPE=sqlite (#424, #426).`
   Read the marker through the legacy_migrator's exported reader (align to its real signature at
   implementation time — the table/API is #424's, never re-implemented here); marker read failure →
   warn once and continue (never fatal; legacy path must stay functional per the PBI constraint).

**Tests despite the unchecked "unit tests" DoD box**: the AC "a deprecation warning is logged at the
warn level on startup" is unprovable without them — the PBI's own AC outranks the loose DoD default.
Pattern: recording fake implementing `basecontext.ApiContext` (the 14-method interface at
`src/basecontext/api_context.go:10`; capture `LogWarnf` calls) per #418 task 05's precedent;
`src/basecontext/test/mock_base_context.go` (`MockBaseContext.On`) is the alternative. Selection
must be reachable as a plain function for the test — if #421 left it inlined in an init path,
extract `selectDatabaseService(ctx, …)` in the same PR (additive, behaviour-identical).

### D-4 — No-new-callers guard: depguard rule (verified YAML) + dedicated CI job, because the existing lint workflow is dead

**Grounds (verified):** `.golangci.yml` (root, 107 lines) is **v2 dialect** (`version: "2"`,
`default: none`, staticcheck + gocritic enabled). `.github/workflows/linter.yml` is
`if: false` + super-linter with `VALIDATE_GO: false` + Go 1.21 + root `go.sum` path — it cannot run
golangci-lint v2 or anything useful. So "a linter rule **or CI check**" needs a live CI lane: the
rule goes into the canonical `.golangci.yml` (devs + IDEs + future full lint inherit it) and a
new minimal `.github/workflows/go-legacy-guard.yml` runs exactly it:
`cd src && golangci-lint run -c ../.golangci.yml --enable-only=depguard ./...`
with `golangci-lint` **pinned v2.14.0** (the version used for all verification in this plan).
`--enable-only` was verified to suppress staticcheck/gocritic noise → CI fails *only* on new
imports, never on the repo's pre-existing lint debt.

**Rule (empirically proven behaviour):** deny exactly one import path —
`pkg: github.com/Parallels/prl-devops-service/data$`. Verified on v2.14.0: the `$` suffix matches
the engine package only; `…/data/models` and `…/data/interfaces` imports are **not** matched (also
by the `data/...` form — verified, and irrelevant since we use `$`); test files are covered
(`run.tests: true` stays). Exclusions (same config, `exclusions.rules`, one entry) grandfather the
importers that legitimately exist when #426 lands; path regexes match relative to the run dir
(`src/`) and are unanchored substring matches — verified (`(data|provider)/.*` excluded both
`provider/…` and `src/provider/…` forms). Grandfather set is **generated, not hand-written**:
task 03's audit command enumerates every non-self importer of the exact package; at today's HEAD the
baseline is: `src/data/**` (self) + `serviceprovider/**` (selection, `Legacy()`, parallelsdesktop
helpers), `main.go` (panic-snapshot escape hatch), `catalog/`, `cmd/`, `jobs/`, `orchestrator/`,
`reverse_proxy/`, `startup/migrations/**`, `startup/seeds/**`, `controllers`/`restapi`/`tests`
legacy test files, and (post-#421) `database/legacy_migrator/**`. Each entry carries the
`#426 grandfather` comment; the *removal issue* owns shrinking this list to zero. The list must be
regenerated at implementation time (after #421's flip + #425's test migration it will be much
shorter).

**AC proof gates (in task 03, both directions, exit-code-checked):** on main →
`exit=0, 0 issues`; with a scratch importer file (added, run, deleted) → `exit=1` with the
rule's `desc` text in the message. `github.com/Parallels/prl-devops-service/data/...` wildcard
entries are NOT used (see above); depguard version drift is caught by the same two gates.

### D-5 — Escape-hatch audit: targeted `//nolint:staticcheck` only where a non-grandfathered file references the deprecated engine

Enumerate at implementation time with
`grep -rn '"github.com/Parallels/prl-devops-service/data"' src --include='*.go' | grep -v '^src/data/'`
then per-symbol `grep -o 'data\.[A-Z][A-Za-z0-9_]*' <file> | sort -u`. Today's HEAD baseline (all
re-verified post-#421 flip; most `*data.JsonDatabase` parameter types become `DatabaseService` and
drop out):
| Site | Symbols used | Disposition |
|---|---|---|
| `src/serviceprovider/data.go` | `JsonDatabase`, `ErrDatabaseNotConnected` (+ post-flip `Legacy()`) | grandfather path; `Legacy()` body may keep the marker-free field access |
| `src/serviceprovider/main.go` ×8 | `NewJsonDatabase` (fallback construction branches) | grandfather path |
| `src/main.go:52-63,120` | `sp.JsonDatabase` recover-snapshot (`SaveNow`/`SaveAs`), `cleanup(…, *data.JsonDatabase)` | grandfather (`main.go` in exclusion list) |
| `src/serviceprovider/parallelsdesktop/main.go` | `ParseFilter`, `FilterByProperty` (live helpers → carve-outs, no diagnostic), `JsonDatabase` (param, #421 refactors to interface) | grandfather; **no** nolint sprinkle in this 3,079-line file (out of ER-3 scope) |
| `src/startup/migrations/**` | `JsonDatabase` params, 1×`ErrUserAlreadyContainsClaim` | grandfather (AD-4: legacy migrators keep operating on the JSON store) |
| `src/startup/seeds/**`, `src/catalog/*`, `src/cmd/root_password.go` | `Err*` sentinels only (carve-outs → no diagnostic) | grandfather for the import line (package-level SA1019, D-1) |
| `src/jobs`, `src/orchestrator`, `src/reverse_proxy` mains | `*data.JsonDatabase` struct fields/params | refactored by #421; if any survive → grandfather entry |
| `src/restapi/*_test.go`, `src/controllers/*_test.go`, `src/tests/**` | `NewJsonDatabase`, `IsAuthorized` | grandfather (legacy tests; #425 migrates them) |
| `src/integration_test/**` | `NewJsonDatabase`, `ResetForTesting` | build-tagged `integration` → not linted by default; nolint on use lines when touched |
Any surviving deprecated-symbol reference in a file **not** in the grandfather list gets
`//nolint:staticcheck // #426: legacy escape hatch — remove with the legacy layer (removal issue #…)`
on the exact line. Never bulk-nolint a file.

### D-6 — Versions, CHANGELOG, removal-issue targets

`VERSION` = **1.1.4** (verified). Deprecation is semver-minor → **effective in v1.2.0** (next minor;
release flow bumps at release — CHANGELOG has no Unreleased convention, entries appear under a
version heading at release time, verified top-60-lines). **Planned removal: v2.0.0** (next major —
matches "breaking removal" and every marker string in D-1/D-3). CHANGELOG addition: new top section
`## [1.2.0] - Unreleased` with the deprecation bullet (exact text in task 04; the release PR fills
the date — flagged as OQ-426-3). Removal issue: created with `gh issue create --repo
Parallels/prl-devops-service` (gh 2.82.0 verified locally) — title, full body file content, target
v2.0.0, removal checklist (importer-enumeration greps, sentinel/helper relocation,
save-path/backup semantics `DATABASE_NUMBER_BACKUP_FILES`/`DATABASE_SAVE_INTERVAL_MINUTES`/
`SYSTEM_AUTO_RECOVER_DATABASE` retirement, `startup/migrations` fate per AD-4, depguard rule flip +
exclusion removal, doc links), cross-links to #426 and #417. Body content is authored in task 04's
spec (fenced, copy-paste ready); creation itself is a human-approved action (it's a public side
effect, not a file write).

## Target files & layout

```
src/data/*.go                     # task 01 — comment-only markers (175 total, engine files only)
src/serviceprovider/data.go       # task 02 — warn constants + frozen-legacy check (+ extract-select seam)
src/serviceprovider/data_test.go  # task 02 — recording-fake tests (new or extended)
.golangci.yml                     # task 03 — depguard rule + grandfather exclusions (107 → ~140 lines)
.github/workflows/go-legacy-guard.yml   # task 03 — new, ~35 lines, pinned golangci-lint v2.14.0
CHANGELOG.md                      # task 04 — one new top section (928 → ~936 lines)
# created outside git via gh: the v2.0.0 removal issue (task 04)
```

## Risks

| # | Risk | Mitigation |
|---|---|---|
| R1 | 175-marker diff churn | one mechanical commit, comment-only diff gate, counts asserted per file before merge |
| R2 | `orchestrator.go` breaches ER-2 cap | verified strict budget: 933 + 2×31 = **995 ≤ 1,000** (realistic 974: 10 of 31 already have doc comments → some inserts cost 2 lines, most 1). Guard table + mandatory post-sweep `wc -l ≤ 1000` check; fallback: ER-3 pure-move split (separate commit) if a re-measure moves ≥ 935 |
| R3 | SA1019 flood in IDEs / future full lint | intentional signal; CI runs depguard only; grandfather exclusions shared with depguard if full lint ever turns on |
| R4 | Package-level Deprecated flags every import line (verified) | grandfather list covers all legitimate importers; new ones are exactly what we want red |
| R5 | depguard dialect drift in future golangci versions | both-direction gates (scratch importer must fail; main must pass) run locally in the PR and in CI |
| R6 | #421/#424 code shape differs from the anchors above | every task carries re-grep directives + explicit prerequisites; no line numbers trusted |
| R7 | 1.2.0 numbering vs release process | OQ-426-3; CHANGELOG section text is date-"Unreleased" and release PR renames |

## Open questions

| ID | Question | Blocks / owner |
|---|---|---|
| OQ-426-1 | The `DB_TYPE` literal that selects the legacy engine (e.g. `legacy`/`json`) — #421's selection switch owns it | #421; message templates interpolate the constant |
| OQ-426-2 | `legacy_migrator` marker reader API (name/signature; table per #424 plan) | #424; task 02 aligns, never duplicates |
| OQ-426-3 | Confirm v1.2.0 is the next release train for this repo's cadence (release-manager) | task 04 |
| OQ-426-4 | How much of the legacy-test grandfather list #425 already removed by merge time | task 03 (regenerate, don't guess) |

## Task index (each task = one PR, merge in order)

| Task | Goal | Depends on | Touches | Size |
|---|---|---|---|---|
| [01](./tasks/01-annotation-sweep-deprecated-comments.md) | 175 `// Deprecated:` markers, script-driven, comment-only | #421 final flip | `src/data/*.go` (non-test, engine files) | M |
| [02](./tasks/02-runtime-deprecation-warns.md) | Formalized selection warn + OQ-1 frozen-legacy warn + tests | 01 | `serviceprovider/data.go`, `serviceprovider/data_test.go` | S |
| [03](./tasks/03-depguard-ci-new-caller-guard.md) | depguard rule + grandfather exclusions + CI job + both-direction gates | 01 | `.golangci.yml`, `.github/workflows/go-legacy-guard.yml` | M |
| [04](./tasks/04-changelog-entry-and-removal-issue.md) | CHANGELOG entry + removal issue via `gh` + back-links | 01–03 | `CHANGELOG.md`, GitHub issue | S |

No file overlap between tasks; 02 and 03 may run in parallel after 01; 04 last (links to shipped
facts). Verification for every task ends with `cd src && go build ./... && go test ./... -count=1`
green (ER-5).

## AC coverage

| AC (verbatim checkbox) | Task | Proof executed |
|---|---|---|
| **Code annotation** — all public legacy symbols carry `// Deprecated:` | 01 | marker count == 175 + per-file table; comment-only diff gate; carve-outs asserted grep-clean (`Err`, `ParseFilter`, `FilterByProperty`, models, interfaces); `gofmt -l` empty; build+suite green |
| **Runtime warning** — warn on startup when legacy reached | 02 | unit tests with recording fake: legacy selection → exact message at warn ×1; frozen marker → frozen warn ×1; sqlite selection → zero warns |
| **No new callers** — linter/CI blocks new imports | 03 | CI job in PR runs; local: scratch importer → exit 1 with rule message; main → 0 issues exit 0 |
| **Removal issue created** — issue targeting next major, linked | 04 | `gh issue create` output (issue URL) pasted in PR; body links #426/#417; back-link comment on #426 |
| **Changelog entry** — version effective + planned removal | 04 | `## [1.2.0]` section names v1.2.0 + v2.0.0 + issue links |
| DoD "unit tests" (unchecked in #426 DoD) | 02 | included deliberately — AC 2 is otherwise unverifiable (D-3 rationale); tests listed in task 02 |
| DoD "Documentation updated (if applicable)" | 04 + #427 | CHANGELOG is the in-repo doc; user docs ride #427 |

## Handover to #427

- Deprecation notice copy = D-3 messages + D-6 CHANGELOG bullet; removal version **v2.0.0** must
  appear in README/docs; `DB_TYPE` legacy-value wording per OQ-426-1.
