# Task 01 — Configuration reference: Database section

**PBI** [#427](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** #420 merged (env vars + flat config keys shipped) · **PR size:** S
**Plan refs:** [D-1](./plan.md#d-1--page-inventory-and-nav-placement-decided-from-the-real-docs-structure), [D-2](./plan.md#d-2--content-accuracy-protocol-binding-on-every-task), [D-3](./plan.md#d-3--config-file-keys-are-flat-environment-keys-the-issues-dotted-notation-is-not-documented)

## Context

`docs/docs/devops/getting-started/configuration.md` is the site's single env-var/config reference
(the only page mentioning `DATABASE_FOLDER`, line 40) and defines the Flag/Description/Default
Value table format. The feature adds four settings that must appear there in the page's own
convention — extend, never relocate. The config-file section (lines 26-32) already teaches the FLAT
`environment:` snake_case key convention; the docs keep it that way and never mention the issue's
`database.type` dotted form (not implementable — deviation rationale lives in **#420's plan.md**,
cite it in the PR description only).

## Objective

A `### Database` section in the config reference documents `DB_TYPE`, `DB_PATH`, `DB_DSN`,
`DB_AUTO_MIGRATE` with type/default/description quoted from merged code, shows the flat config-file
keys, and carries the one-line deprecation pointer. Zero new lint findings; nav unchanged.

## Files

| File | Op | Current lines |
|---|---|---|
| `docs/docs/devops/getting-started/configuration.md` | modify — insert one section after the `### System` table (i.e. before `### Rest API`, line 62) | 133 |
| `docs/_data/docs_devops_menu.yml` | **read-only sanity check only** (page already registered at lines 9-10) | 74 |

## File size guard

`configuration.md` 133 → ≈165 after insert. ER-2 spirit: **ok** (cap 1,000). No new files.

## Implementation spec

1. **Evidence first (D-2).** Before writing, collect `file:line` from merged code for: exact env
   var names + accepted values + defaults (`src/constants/main.go` `DB_*` constants and the #420
   getters; `src/database/config.go` for `AutoMigrate: true` and the default SQLite file name —
   planned as `prl-devops.db` under the `DATABASE_FOLDER`-honoured folder per AD-3), and the exact
   flat config-file key names #420 shipped (`db_type`, `db_path`, `db_dsn`, `db_auto_migrate` —
   **copy the merged names**, do not trust this line). Record the table in the PR description.

2. **Insert draft** (adjust bracketed values ONLY from evidence; prose wraps ≤ 80 cols; table rows
   match the existing 3-column format — `line-length` exempts tables per `.markdownlint.json`):

   ```markdown
   ### Database

   The service stores its state either in a local SQLite database file (the
   default, zero configuration) or in a PostgreSQL server. These settings
   select and configure the backend. The legacy JSON data store
   (`data.json`) is deprecated; see the *Upgrading to the SQL database*
   page for what happens on the first start after an upgrade, and the
   project changelog for the planned removal version.

   | Flag | Description | Default Value |
   | ---- | ----------- | ------------- |
   | DB_TYPE | The database backend to use. One of `sqlite` or `postgres` | sqlite |
   | DB_PATH | Full file path of the SQLite database file. Only used when DB_TYPE is `sqlite`; ignored when DB_TYPE is `postgres` | `<DATABASE_FOLDER>/prl-devops.db` |
   | DB_DSN | The PostgreSQL connection string (URL or keyword/value form). Required when DB_TYPE is `postgres`. Passwords are redacted in logs | |
   | DB_AUTO_MIGRATE | Specifies whether the service creates or updates the database schema automatically on startup | true |

   These settings can also be set in the configuration file under the
   `environment` object, using the same keys in lower case:

   ```yaml
   environment:
     db_type: postgres
     db_dsn: postgres://devops:change-me@db.example.com:5432/devops?sslmode=require
     db_auto_migrate: true
   ```

   Environment variables always take precedence over configuration file
   values, as described above.
   ```

   Bracketed/quoted facts to re-verify against merged code: the default path string (AD-3 file
   name), accepted `DB_TYPE` values, `DB_AUTO_MIGRATE` default. The DSN sample must stay
   obviously-fake (`change-me`). If merged #420 named anything differently (getter/constant
   disagreement included), **the merged code wins** and this task's diff follows it.
   **Forward-link rule (plan merge order 01→…→05):** the pointer to *Upgrading to the SQL
   database* is deliberately **plain text** — that page does not exist yet; task 02's PR upgrades
   this exact sentence to a live link as its "backlink upgrade" step. No link added by this task
   may target a page created by a later task.

3. **Nav sanity (no edit):** `Configuration` entry exists in
   `docs/_data/docs_devops_menu.yml:9-10`; the `menubar: docs_devops_menu` front matter in
   `configuration.md:1-8` is untouched.

4. **Formatting compliance:** prose ≤ 80 cols; tables may be long (exempt); code fences tagged
   (`yaml`); no HTML. Lint commands + baseline criteria in Validation.

## Out of scope

Any other section of the page; menu file edits; the two new pages (tasks 02/03); Helm/compose
(task 04); README/CHANGELOG (task 05); any reformatting of pre-existing prose/tables (the page has
known MD013 @18/20/22/32 and MD056 @98 and **keeps** them).

## Validation

- [ ] `export npm_config_cache=/tmp/pbi427-npmcache` (root-owned `~/.npm` workaround, [D-6](./plan.md#d-6--toolchain-reality-verified-2026-11-03--drives-the-validation-sections))
- [ ] `npx --yes markdownlint-cli@0.41 docs/docs/devops/getting-started/configuration.md --config .markdownlint.json` → findings **identical in rule+count** to the recorded baseline (4×MD013, 1×MD056); zero other rules
- [ ] Prettier on the added block only (pre-existing file fails `--check` by baseline — never `--write` it): extract the section to `/tmp/db-section.md` and run `npx --yes prettier@3.6 --check /tmp/db-section.md` → pass
- [ ] Render check (feasible path): `cd docs && bundle install && bundle exec jekyll build --destination /tmp/pbi427-site` → succeeds; the generated `database section` page HTML contains the new table; `docs/_site` untouched. Fallback checklist if bundler/network infeasible: front matter untouched; menu link ↔ file mapping unchanged; **every link added by this diff targets a page that already exists at HEAD** (forward-link rule above); internal site links use the `{{ site.url }}{{ site.baseurl }}` form
- [ ] `git diff --stat` touches only `docs/docs/devops/getting-started/configuration.md`
- [ ] PR description embeds the D-2 evidence table (`claim → file:line` from merged code); every bracketed value in the diff traces to a row
- [ ] **AC 6:** PR reviewed and approved by ≥ 1 team member who is not the author

## Rollback

Revert the single commit; the page returns to its prior state; no nav or cross-file coupling to
undo (tasks 02/05 own their own links).

## PBI acceptance mapping

AC 1 (primary — all four settings with type/default/description + config-file parity); AC 5
(pointer sentence, authoritative notice is task 02); AC 6 (review gate above).

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `427-01`.** Cut `slice/427-01` from `sql-migration` HEAD
(`.worktrees/slice-427-01`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
