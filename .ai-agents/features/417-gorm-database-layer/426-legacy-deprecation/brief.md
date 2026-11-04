# PBI Brief #426 — Deprecation of the legacy data layer

| Field | Value |
|---|---|
| PBI | [#426](https://github.com/Parallels/prl-devops-service/issues/426) — `[PBI] Deprecation of the legacy data layer` |
| Labels | `triage`, `pbi` · Assignee: unassigned |
| Round | Deep-dive round **7** (after #421's final flip; can partly overlap #425) |
| Upstream feature docs | [`brief.md`](../brief.md) · [`plan.md`](../plan.md) — required reading, esp. AD-1, OQ-1 |
| Planning status | **Fully planned** — [plan.md](./plan.md) · tasks [01](./tasks/01-annotation-sweep-deprecated-comments.md) · [02](./tasks/02-runtime-deprecation-warns.md) · [03](./tasks/03-depguard-ci-new-caller-guard.md) · [04](./tasks/04-changelog-entry-and-removal-issue.md) |

## Description (verbatim from GitHub)

Formally mark the legacy data layer as deprecated in the codebase, add runtime deprecation warnings
where it is still reachable, and prepare a documented removal plan for a future release.

### User Story

As a **maintainer**, I want the legacy data layer to be clearly marked as deprecated with a
documented removal timeline, so that there is no ambiguity about its future, and contributors do not
accidentally extend it.

### Acceptance Criteria

- [ ] **Code annotation**: All public types, functions, and packages belonging to the legacy layer are annotated with a `// Deprecated:` GoDoc comment pointing to the replacement.
- [ ] **Runtime warning**: If the legacy layer is somehow reached at runtime (e.g. via a feature flag or fallback), a deprecation warning is logged at the `warn` level on startup.
- [ ] **No new callers**: A linter rule or CI check prevents new code from importing the deprecated package.
- [ ] **Removal issue created**: A GitHub issue is created and linked, targeting the next major version, describing the steps to fully remove the legacy layer.
- [ ] **Changelog entry**: The deprecation is documented in `CHANGELOG.md` with the version it takes effect and the planned removal version.

### Definition of Done

- [ ] Code implemented following best practices.
- [ ] Unit tests written and passing.
- [ ] Code reviewed and approved.
- [ ] Merged into the main branch.
- [ ] Documentation updated (if applicable).
- [ ] Deployed to staging/production environment.

### Assumptions and Constraints

- **Assumption**: There are no external consumers of the legacy data layer's public API (it is internal to the service).
- **Constraint**: The legacy layer must remain functional until the removal release to avoid breaking users who are still on older builds.

## Repo context (verified 2026-11-03)

- **Scope of annotation**: package `data` (engine + all 157 `JsonDatabase` methods — a `//
  Deprecated:` on each exported symbol) and the *files* that become legacy-only
  (`src/data/filter_layer.go`, `order_layer.go`, `index_layer.go`, `authorize_layer.go`,
  `backup.go`, `main.go`…). **Do NOT deprecate `src/data/models`** — those structs remain the live
  API wire format AND the GORM models (#419 tags them in place). GoDoc `// Deprecated:` on a type
  propagates to IDEs/linters; one note on `JsonDatabase` + per-file package-level notes beat 157
  mechanical comment PRs — decide granularity in the round.
- **Runtime warning hook exists already by design**: AD-1's serviceprovider selection logs a warn
  when `DB_TYPE` resolves to legacy/fallback — this PBI likely just formalises the message + adds
  the post-migration warning from OQ-1 ("legacy file frozen; serving from SQL").
- **Lint guard**: `.golangci.yml` exists at repo root (v1 config; `.golangci.bck.yml` is a stale
  copy — don't edit that one). `depguard` rule blocking `github.com/Parallels/prl-devops-service/data`
  imports with an allow-list for `src/data/**`, `src/database/legacy_migrator/**` and
  `src/serviceprovider/**` (fallback). Verify golangci-lint version in `.github/` workflows supports
  the config format before writing rules.
- **Escape hatches stay legal callers**: `src/main.go:59` (`SaveAs` panic snapshot) and
  `src/data/backup.go` callers via `Legacy()` — depguard allow-list, not a refactor.
- **Changelog**: `CHANGELOG.md` (45k lines-of-text repo history at root) — follow existing entry
  format when adding the deprecation notice; "version takes effect" = next minor per
  `VERSION` file + release flow; planned removal = next major (name it in the issue).
- **Removal issue**: create via `gh issue create --repo Parallels/prl-devops-service` with body
  checklist derived from AD-1 escape-hatch audit + `grep` enumeration of all importers at merge
  time; link it back to #426 and #417.

## Engineering-rule triggers (feature plan §6)

- ER-1: annotations are comment-only commits where possible (zero behaviour change) — keeps the
  157-method surface compile-identical (risk R-1 assertion stays green).
- ER-2: some `src/data/*.go` files grow slightly from comments — guard-table each; none near cap
  except `orchestrator.go` (933) → comment volume there may trip ER-3; budget the split or scope
  comments tightly (type-level notes only) — decide in the round.

## Fit in the feature

- **Depends on**: #421 final flip (deprecating the *primary* path before it's bypassed would be
  a lie in GoDoc).
- **Feeds**: #427 (docs must carry the deprecation notice + removal version).
- **Explicit non-goal**: no deletion of legacy code, no config option removing legacy — removal is
  the future issue created here.
