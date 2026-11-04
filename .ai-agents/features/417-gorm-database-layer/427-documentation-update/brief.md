# PBI Brief #427 — Documentation update

| Field | Value |
|---|---|
| PBI | [#427](https://github.com/Parallels/prl-devops-service/issues/427) — `[PBI] Documentation update` |
| Labels | `triage`, `pbi` · Assignee: unassigned |
| Round | Deep-dive round **8** — **last** (the PBI itself constrains: "one of the last merged") |
| Upstream feature docs | [`brief.md`](../brief.md) · [`plan.md`](../plan.md) — required reading, esp. OQ-4 |
| Planning status | **Fully planned** — [plan.md](./plan.md) · tasks [01](./tasks/01-configuration-reference-database-section.md) · [02](./tasks/02-upgrade-guide-page.md) · [03](./tasks/03-postgresql-setup-page.md) · [04](./tasks/04-helm-database-passthrough.md) · [05](./tasks/05-architecture-changelog-readme.md) |

## Description (verbatim from GitHub)

Update all user-facing and operator-facing documentation to reflect the new database layer,
configuration options, migration behaviour, and upgrade guidance. This includes the README,
configuration reference, and any deployment guides.

### User Story

As a **user or operator**, I want up-to-date documentation that clearly explains how to configure
the database backend, what happens on first run after upgrading, and how to set up PostgreSQL for
production use, so that I can deploy and manage the service confidently.

### Acceptance Criteria

- [ ] **Configuration reference updated**: All new env vars and config file keys are documented with their type, default value, and a description.
- [ ] **Upgrade guide**: A dedicated upgrade section explains what existing users should expect when upgrading (automatic migration, no data loss, no manual steps for SQLite users).
- [ ] **PostgreSQL setup guide**: Step-by-step instructions for configuring the service against a PostgreSQL instance, including a sample DSN and recommended PostgreSQL settings.
- [ ] **Architecture note**: A brief explanation of the two-backend model and when to choose each (SQLite for simplicity, PostgreSQL for scale) is added to the README or architecture docs.
- [ ] **Deprecation notice**: The documentation notes that the legacy data layer is deprecated and will be removed in a future version.
- [ ] **Docs reviewed**: Documentation changes are reviewed by at least one team member who is not the author.

### Definition of Done

- [x] Code implemented following best practices.
- [x] Unit tests written and passing.
- [x] Code reviewed and approved.
- [x] Merged into the main branch.
- [x] Documentation updated (if applicable).
- [x] Deployed to staging/production environment.

### Assumptions and Constraints

- **Assumption**: Documentation lives in the repository (e.g. `/docs` or `README.md`) and is version-controlled alongside the code. ✅
- **Constraint**: Docs must be accurate at the time of the feature release — this PBI should be one of the last merged in the sequence.

## Repo context (verified 2026-11-03)

- **Docs site**: `docs/` is a Jekyll site (`docs/_config.yml`, published via GitHub Pages). User
  docs live under `docs/docs/` and `docs/tutorials/` — locate the existing configuration/env-var
  reference with `grep -rl "DATABASE_FOLDER" docs/ --include=*.md` (excluding `_site/`) and extend
  **that** page rather than inventing a new location. Do not add files under `docs/` without front
  matter (`layout: page`) — they'd render broken; and never write into `docs/_site/` (build output).
- **README.md** (repo root) — short; architecture note per AC 4 likely fits in the docs site with
  a README link rather than growing the README.
- **Deployment guides to keep coherent**: `docker-compose.yml` (+override, debug), `helm/` chart
  (env passthrough values — OQ-4: chart should expose `DB_TYPE`/`DB_PATH`/`DB_DSN`/`DB_AUTO_MIGRATE`
  envs; a chart `values.yaml` addition is part of this round's scope if the human approves —
  Helm changes are **not** covered by any other PBI), `docs/charts/*.tgz` are release artifacts
  (do not hand-edit).
- **CHANGELOG.md**: entry for the feature itself (deprecation entry is #426's; don't duplicate —
  cross-reference).
- **Style**: `.markdownlint.json`, `.textlintrc`, `.prettierrc` exist at root — docs PRs should pass
  them (`npx markdownlint` / `npx prettier --check` on touched files only; don't reformat the site).
- **Content sources**: this feature's own docs are internal/agent-facing (`.ai-agents/`) — do **not**
  copy them verbatim; user docs must state behaviour, not implementation. Target config table for
  the config reference is [brief §3](../brief.md#3-configuration-reference-target-state--owned-by-pbi-420).
- Sample DSN for the Postgres guide must show redaction-friendly logging expectations (AC of #420
  "secrets safety") and sane pool defaults shipped by #418 (`max open 25 / idle 5 / lifetime 30m`
  — quote the final values from #418's merged code, not from here).
- **Postgres guide should cover**: role/DB creation, SSL mode, `search_path`, connection-budget
  math for orchestrated replicas (N pods × pool), and the stateless-deployment story (PVC not
  needed with `DB_TYPE=postgres`).

## Engineering-rule triggers

- ER-1 applies loosely here (docs don't follow Go idioms) but the file cap spirit applies: split
  any page past 1,000 lines.
- Accuracy gate: every documented default/behaviour in the docs round must be re-verified against
  merged code (not against these planning docs) — docs describe reality **at merge time**.

## Fit in the feature

- **Depends on**: everything merged (#418…#426) — defaults and behaviours must be final.
- **Deliverables checklist**: config reference · upgrade guide (SQLite auto-migration story) ·
  Postgres setup guide · architecture note · deprecation notice + removal version (from #426's
  changelog/issue) · Helm/compose env examples (OQ-4).
