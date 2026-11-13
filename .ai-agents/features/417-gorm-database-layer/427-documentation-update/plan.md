# PBI Plan #427 — Documentation update

> Reading order: [feature brief](../brief.md) → [feature plan](../plan.md) (esp. AD-3, OQ-1, OQ-4) →
> [PBI brief](./brief.md) → this plan → your task spec.
>
> **This is the last PBI of the feature. Every number, log line and behaviour the docs will state
> must be quoted from MERGED code at task-execution time — never from these planning docs.** This
> plan deliberately sketches content and marks `[QUOTE-AT-MERGE: …]` where the writer must open the
> merged source, copy the exact string/value, and record `file:line` evidence in the PR
> description (D-2).

## Scope

**In:** user-facing docs site pages — configuration reference extension (new `Database` section in
`docs/docs/devops/getting-started/configuration.md`), a new **Database Upgrade** guide page, a new
**PostgreSQL Setup** page, both registered in the docs menu; the Helm chart's minimal DB env
passthrough (OQ-4) + its values documentation; architecture note section (`architecture.md`);
deprecation notice (placement per D-5, wording linked to #426's changelog entry); CHANGELOG feature
entry (cross-referencing #426, never restating its version); README → docs link (README stays
tiny); formatting/lint/render validation for every touched file.
**Out:** any `.go` change; any change under `src/`; hand-editing `docs/charts/*.tgz` (release
artifacts); writing anything into `docs/_site/` (gitignored build output — builds for validation go
to a temp destination, D-6); reformatting existing prose/tables beyond the additions; changes to
`helm/values-prd.yaml`/`values-stg.yaml`/`values.local.staging.yaml` (maintainer environments);
editing the root `docker-compose.yml`/`docker-compose.override.yml`/`docker-compose.debug.yml`
(compose guidance lives in the docs instead — D-4); duplicating #426's changelog entry.

## Prerequisite (merge gate — hard)

Every task in this PBI is **blocked until its cited code PBIs are merged** (#418, #420, #421, #424,
#426 — per [feature plan §8](../plan.md#8-delivery-order--dependency-graph) merge order). Docs
describe reality at merge time; starting a task against unmerged plans is a process violation for
this PBI specifically (AC constraint: "one of the last merged").

## Design decisions

### D-1 — Page inventory and nav placement (decided from the real docs structure)

Verified 2026-11-03 against `docs/` (Jekyll 4.3, `bulma-clean-theme`, `permalink: pretty` —
`docs/_config.yml:13`; nav is driven by `_data` menus, not per-page files):

| Doc surface | Decision |
|---|---|
| Config/env-var reference | **Extend** `docs/docs/devops/getting-started/configuration.md` — it is the only page mentioning `DATABASE_FOLDER` (line 40) and defines the site's env-var-table format (Flag / Description / Default Value). Add one `### Database` section; do not relocate or restructure. |
| Upgrade guide | **New page** `docs/docs/devops/getting-started/database-upgrade.md` → URL `/docs/devops/getting-started/database-upgrade/` (pretty permalink). |
| PostgreSQL setup | **New page** `docs/docs/devops/getting-started/postgresql.md` → URL `/docs/devops/getting-started/postgresql/`. |
| Nav registration (both new pages) | **Exact file: `docs/_data/docs_devops_menu.yml`** — `Getting Started` group (lines 1–14), insert after the `Configuration` item (line 9–10): YAML in task 02. Verified it is the sole registration point (`docs/docs/devops/index.md` has no page list; menu groups = 8, parses as valid YAML). `menubar: docs_devops_menu` in front matter activates it. |
| Architecture note (AC 4) | New `## Data Persistence` section in the existing `docs/docs/devops/getting-started/architecture.md` (55 lines, no DB content today) + one README link (D-5). |
| Helm values doc | `docs/docs/devops/getting-started/kubernetes/helm-chart.md` (154 lines) gains a short Database values section (task 04). |
| README | Stays a pointer: one bullet in the existing `## Documentation` section (`README.md:34-36`), no prose growth. |

**Front matter for both new pages — copy verbatim from the sibling
`configuration.md:1-8` (read it at task time), changing only `subtitle`:**

```yaml
---
layout: page
title: Getting Started
subtitle: <Upgrading to the SQL Database | PostgreSQL Setup>
menubar: docs_devops_menu
show_sidebar: false
toc: true
---
```

Internal links use the site's Liquid absolute form as in `docker.md:31`, e.g.
`[Configuration]({{ site.url }}{{ site.baseurl }}/docs/devops/getting-started/configuration/)`.
ER-2 spirit: pages are drafted ~120–200 lines — nowhere near the 1,000-line cap; split if a page
ever grows past it.

### D-2 — Content-accuracy protocol (binding on every task)

1. Each task carries an **Evidence table**: every user-visible claim (default, log line, file name,
   flag, env-var name, redaction output, pool number) → the merged source file to quote → where the
   claim lands.
2. At execution time the writer opens the merged file, copies the **exact** string/value, and
   records `file:line` for each claim in the PR description. Reviewers spot-check; a claim without
   evidence is cut, not softened.
3. Sources (paths known at plan time; **names inside them are `[QUOTE-AT-MERGE]` until proven**):
   - backend env keys + flat config keys + defaults + validation messages → merged `DB_*` constants
     in `src/constants/main.go` and getters in `src/config/` (#420); `src/database/config.go`
     (#418/#420) for `DatabaseConfig` defaults (`AutoMigrate: true`, filename
     `prl-devops.db` per AD-3).
   - SQLite pragmas/WAL + Postgres pool defaults → `src/database/sqlite.go`, `postgres.go`,
     `config.go` (#418). Pool numbers currently planned as 25/5/30m — **quote the merged values**.
   - DSN redaction form → `src/database/redact.go` + its tests (#418) — quote a real redacted
     output, not this plan's sketch.
   - migration steps, log lines, marker table, dry-run flag name → `src/database/legacy_migrator/`
     (#424).
   - frozen-`data.json` guarantee + legacy-selection warn message → merged `#426` code and its
     OQ-1 decision; deprecation/removal wording → **#426's CHANGELOG.md entry: link it, never
     restate the version number**.
4. If a `[QUOTE-AT-MERGE]` placeholder cannot be resolved from merged code, the sentence is
   deleted or rewritten to state only what the code proves (see "say the gap honestly" items in
   tasks 02/03 — new-DB backups, pool tunability, rollback loss).

### D-3 — Config-file keys are FLAT `environment:` keys; the issue's dotted notation is not documented

The repo config loader (custom, `Config.GetKey`: config file → flag → env wins; page already says
keys are lowercase flag names under `environment:`, e.g. `api_port` — `configuration.md:26-32`).
The issue's `database.type`-style keys were **not implementable** without inventing a second
hierarchy; #420's round records the deviation (`#420 plan.md D-decision` — cite that plan for the
deviation text in the PR description, not in the docs). **Docs therefore describe the FLAT keys
only** — `db_type`, `db_path`, `db_dsn`, `db_auto_migrate` — and the writer confirms those exact
names from merged #420 constants before typing the table. The docs never mention `database.type`;
users must not learn about a form that doesn't work.

### D-4 — Deployment surfaces: Helm gains a minimal gated passthrough; compose files stay untouched (OQ-4 resolution)

Verified how env vars reach the container today:
`helm/templates/deployment.yaml:68-259` is an **explicit env list**, each entry gated by a value and
sourced from the chart ConfigMap (`config-map.yaml`, which already sets `DATABASE_FOLDER` from
`.Values.storage.databasePath`) or Secret (`secrets.yaml`); `.Values.envFrom`
(`helm/values.yaml:34-37`, wired at `deployment.yaml:256-259`) is a generic passthrough — an
operator can already ship `DB_DSN` today via `envFrom: - secretRef:` with **no chart change**.
The PVC story is already stateless-ready: volume + mount are gated on
`.Values.storage.databasePath` (`deployment.yaml:46-50, 282-285`), so `storage.databasePath: ""`
removes persistent storage entirely.
**Decision (OQ-4: minimal, additive, backward-compatible):** add a `database:` values block
(`type`/`path`/`autoMigrate` non-secret → ConfigMap; `dsn`/`dsnSecret` secret → Secret/chart
Secret), every template addition gated by `{{- if }}` with empty defaults, so
`helm template` output for default values is **byte-identical to today** (checked in task 04).
The `docker-compose*.yml` files are not edited — the Postgres guide shows the 2–3 env lines to add
to their copy (matches "keep deployment guides coherent" without touching maintainer files).

### D-5 — Deprecation notice placement; README is a link

AC 5 has one authoritative home: the **upgrade guide page** (task 02) — "what happens to
`data.json`" is where users read it. A one-line pointer sits in the `configuration.md` Database
section (task 01) and a one-paragraph note in the architecture section (task 05). All removal-version
mentions are links to #426's CHANGELOG entry/issue — the number is never restated in body text.
README change = one bullet in `## Documentation` linking config/upgrade/postgres pages.
CHANGELOG gets the **feature** entry in the release format (`## [X.Y.Z] - YYYY-MM-DD`, newest on
top; VERSION file currently `1.1.4`) cross-referencing #426's deprecation entry; the exact version
heading is the maintainer's release-flow call at merge time (task 05 step).

### D-6 — Toolchain reality (verified 2026-11-03 — drives the Validation sections)

| Fact (measured) | Consequence for tasks |
|---|---|
| `~/.npm` has root-owned files → plain `npx` dies with EPERM (no sudo here) | Every lint command is prefixed `export npm_config_cache=/tmp/pbi427-npmcache`. |
| `markdownlint-cli@0.41` + root `.markdownlint.json` (line-length 80, tables exempt): `configuration.md` has **pre-existing** MD013 @18/20/22/32 and MD056 @98 | New files must be **clean**; pre-existing files are judged **zero-new-violations vs the recorded baseline** (brief says: don't reformat the site). Prose in new/added text wraps ≤ 80 chars; tables may run long. |
| `prettier@3.6 --check` **fails today** on `configuration.md` and `CHANGELOG.md` (passes `README.md`) | `prettier --check` must pass on **new** files; pre-existing files keep their (failing) status — never `--write` them. |
| `docs/Gemfile` = Jekyll 4.3 + `bulma-clean-theme` (remote theme, several plugins); `bundle check` = **gems missing** | Render check: `cd docs && bundle install && bundle exec jekyll build --destination /tmp/pbi427-site` (temp dest — never `docs/_site`, which is gitignored anyway). If bundler/network infeasible on the machine, tasks carry a fallback checklist (front-matter parity, menu-YAML parse, link resolution) — validation is "build **if feasible**, else checklist". |

## Task index (each task = one PR; merge strictly in order)

| Task | Goal | Depends on | Touches | Size |
|---|---|---|---|---|
| [01](./tasks/01-configuration-reference-database-section.md) | `Database` section in the config reference (env table + flat config keys + deprecation pointer) | #420 merged | `docs/docs/devops/getting-started/configuration.md` | S |
| [02](./tasks/02-upgrade-guide-page.md) | New upgrade guide page + nav entry | 01; #424/#426 merged | new `database-upgrade.md`, `docs/_data/docs_devops_menu.yml` | M |
| [03](./tasks/03-postgresql-setup-page.md) | New PostgreSQL setup page + nav entry (role/db SQL, DSNs, sslmode, pool math, stateless, k8s secret, compose snippet) | 02 (menu-file conflict → sequential); #418/#420/#421 merged | new `postgresql.md`, menu yml (append) | M |
| [04](./tasks/04-helm-database-passthrough.md) | Helm `database:` values passthrough (OQ-4) + helm-chart.md values doc + paragraph appended to postgres page | 03; #420 merged | `helm/values.yaml`, `helm/templates/{config-map,secrets,deployment}.yaml`, `kubernetes/helm-chart.md`, `postgresql.md` (append) | S/M |
| [05](./tasks/05-architecture-changelog-readme.md) | Architecture `## Data Persistence` section, CHANGELOG feature entry (cross-ref #426), README pointer | 02, 03, 04 (all links must exist) | `architecture.md`, `CHANGELOG.md`, `README.md` | S |

Merge order rationale: **01 → 02 → 03 → 04 → 05** — strictly sequential: 02 and 03 share the menu
file; 04 edits the page 03 creates; 05 links to everything. No two unfinished tasks touch one file.
No code PRs from other PBIs should be merging between them without a re-verification pass (D-2.1)
of evidence tables whose sources moved.

## AC-coverage table (every checkbox from [brief](./brief.md))

| AC checkbox | Covered by | How |
|---|---|---|
| Configuration reference updated (env vars + config keys, type/default/description) | 01 (primary), 04 (helm values cross-ref) | `### Database` section: 4-row Flag/Description/Default table + flat-key YAML example, all defaults quoted from merged code |
| Upgrade guide (automatic migration, no data loss, no manual steps) | 02 (primary), 01 (pointer) | New dedicated page: first-start sequence, quoted log lines, timing, idempotency, truthful rollback, SQLite/WAL ops, "no manual steps" promise |
| PostgreSQL setup guide (steps, sample DSN, recommended settings) | 03 (primary), 04 (values) | New page: SQL role/db/schema blocks, URL + keyword DSN samples, explicit-sslmode table, pool/connection-budget math, stateless notes, k8s Secret + compose examples |
| Architecture note (two-backend model, when to choose each) | 05 (primary) | `## Data Persistence` section in architecture.md + README bullet linking it |
| Deprecation notice (legacy layer deprecated, removal in future version) | 02 (authoritative section), 01 + 05 (pointers) | Notice wording links #426's changelog entry/issue for the removal version — never restated; frozen-`data.json` + warn behaviour quoted from merged #424/#426 |
| Docs reviewed by ≥ 1 non-author team member | **all tasks** | Each PR's Validation checklist: `PR reviewed + approved by ≥ 1 reviewer who is not the author` + reviewer receives the `file:line` evidence table (D-2) as the review artifact |

DoD rows (code/tests/review/merged/deployed) are process-level at feature scope; docs tasks carry
them via the same PR gates (CI + review) — nothing to test beyond the lint/build checks in each
task's Validation.

## Risks

| Risk | Mitigation |
|---|---|
| Planning docs drift ahead of merged code (this plan's numbers are stale by design) | `[QUOTE-AT-MERGE]` markers + evidence tables + merge gate; reviewer checks evidence, not this plan |
| Writers copy `.ai-agents/` wording into user docs | Brief already forbids it; PR checklist item "no implementation vocabulary in user docs" (D-2.4 honesty items) |
| Menu/page links break on rename or permalink change | Validation builds the site to a temp dest and greps generated URLs; fallback checklist verifies permalink↔file mapping |
| Helm change accidentally alters current deployments | Task 04 validation: `helm template` before/after diff on default values must be empty (plus `helm lint`) |
| markdownlint/prettier baseline confusion (existing files fail today) | D-6 records the exact baseline; pass criteria differ for new vs pre-existing files |
| CHANGELOG version-heading conflict with release flow | Task 05 defers the number to merge time; format + #426 cross-reference fixed |

## Open questions (for the human at plan review)

| ID | Question | Impact |
|---|---|---|
| OQ-4 (feature) | **Resolved by D-4** — Helm gets the minimal gated `database:` block; `envFrom` already works. Needs human approval to touch `helm/` (no other PBI owns chart changes). | task 04 |
| OQ-427-1 | New-DB backup ownership: `src/data/backup.go` today copies only `data.json` (`*.save.bak.*` pattern) — nothing will cover `prl-devops.db`. Docs will state the gap + interim `VACUUM INTO` guidance. Should this PBI also open the follow-up issue, or does one exist? | task 02 wording |
| OQ-427-2 | CHANGELOG heading version for the feature entry (VERSION=1.1.4 now; #426 plans next-minor) — release manager's call at merge. | task 05 |
| OQ-427-3 | Postgres floor (guide says "currently supported stable") and driver `sslmode` default — guide mandates explicit `sslmode` either way; confirm driver default from merged dep only if it changes the recommendation. | task 03 |
| OQ-427-4 | Did #420 ship pool overrides (`DB_PG_*`)? Connection-budget paragraph in 03 branches on the merged answer; if absent, it states pool sizes are fixed defaults in this release. | task 03 |

## Execution model — worktree slices + human gate (feature plan §14, ER-6/ER-7)

Every task here executes as an isolated worktree slice `slice/<pbi>-<nn>` cut from **`sql-migration`**,
passes its spec's Validation inside the worktree, closes with a **test plan + results entry** in
[`../VERIFICATION.md`](../VERIFICATION.md) and the **human verdict** — only then merges `--no-ff` into
`sql-migration` and the next slice may start. ER-5's merge boundary = the merge into `sql-migration`.
