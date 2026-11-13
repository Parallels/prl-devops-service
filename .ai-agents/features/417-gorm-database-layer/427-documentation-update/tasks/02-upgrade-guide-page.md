# Task 02 — Upgrade guide page + nav registration

**PBI** [#427](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 01; #424 + #426 merged (migration + deprecation behaviours final) · **PR size:** M
**Plan refs:** [D-1](./plan.md#d-1--page-inventory-and-nav-placement-decided-from-the-real-docs-structure), [D-2](./plan.md#d-2--content-accuracy-protocol-binding-on-every-task), [D-5](./plan.md#d-5--deprecation-notice-placement-readme-is-a-link)

## Context

AC 2 needs a dedicated page telling upgraders what happens on first start: detection → schema
migration → one-time data migration → completion marker → serve, with **real** log lines and the
truthful rollback/backup story. It is also the authoritative home of the deprecation notice (D-5).
Every behaviour statement is `[QUOTE-AT-MERGE]` from `src/database/legacy_migrator/` (#424) and
#426's merged code — the draft below marks each spot.

## Objective

`docs/docs/devops/getting-started/database-upgrade.md` published at
`/docs/devops/getting-started/database-upgrade/`, reachable from the docs menu, passing
markdownlint/prettier from zero (new file), content fully evidenced.

## Files

| File | Op | Current lines |
|---|---|---|
| `docs/docs/devops/getting-started/database-upgrade.md` | create | 0 → ≈180 |
| `docs/_data/docs_devops_menu.yml` | modify — insert 2 lines inside the `Getting Started` group, after the `Configuration` item (`docs_devops_menu.yml:9-10`) | 74 → 76 |

## File size guard

Both files ≪ 1,000 lines. **ok**. Split trigger: if the page would exceed 1,000 lines (it won't),
move the FAQ to a sibling page and register it too.

## Implementation spec

### 1. Menu insert (exact YAML, matching sibling indentation)

```yaml
    - name: Upgrading to SQL
      link: /docs/devops/getting-started/database-upgrade/
```

Permalink `pretty` (`docs/_config.yml:13`) makes the file→URL mapping above exact. Task 03 appends
its own entry to the same group later — this PR does not reserve space for it.

### 2. Page draft (front matter copied verbatim from the sibling `configuration.md:1-8`, only
`subtitle` changes — verified 2026-11-03):

```markdown
---
layout: page
title: Getting Started
subtitle: Upgrading to the SQL Database
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# Upgrading to the SQL Database

Starting with the [release that introduces the SQL data layer — link the
CHANGELOG feature entry here], the service stores its data in a SQL
database: SQLite by default (a single file next to your existing data), or
PostgreSQL for orchestrated deployments
([PostgreSQL Setup](.../postgresql/)). The legacy JSON store (`data.json`)
is deprecated; see [Deprecation](#deprecation) below.

## What happens on the first start after upgrading

1. **Detection** — the service looks for the legacy `data.json` in the
   database folder (`DATABASE_FOLDER`, default `<home>/.parallels-devops`
   or `/etc/parallels-devops` for root). Not found or empty → nothing
   happens, fresh installs just start. `[QUOTE-AT-MERGE: detector.go]`
2. **Schema** — the SQL schema is created (`DB_AUTO_MIGRATE`, default
   true).
3. **Data migration** — all records are copied into the SQL database in
   one transaction per collection batch. A failure rolls the whole
   migration back and the service exits with an error; `data.json` is
   never modified. `[QUOTE-AT-MERGE: migrator.go fatal path]`
4. **Marker** — a completion marker row is written **in the new database**
   (`[QUOTE-AT-MERGE: exact table name + columns from marker.go]`). Every
   later start sees the marker and skips migration: re-runs are no-ops.
5. **Serve** — the API starts as usual.

**There are no manual steps for SQLite users.** Do not move, rename or
edit `data.json`; the migration never writes to it, and it stays in place
untouched (frozen) as your rollback copy. `[QUOTE-AT-MERGE: #424 code/test
proving data.json is never modified + #426 frozen-store warn message text]`

## What you will see in the logs

Migration progress is logged at `info` between startup and "server up".
Quote the real lines — copy each string from merged
`src/database/legacy_migrator/` into the table below and show it verbatim:

| Stage | Example log line (verbatim from merged code) |
|---|---|
| Legacy store detected | `[QUOTE-AT-MERGE]` |
| Per-collection progress | `[QUOTE-AT-MERGE]` |
| Integrity check result (info/warn) | `[QUOTE-AT-MERGE]` |
| Completion | `[QUOTE-AT-MERGE]` |

A warning-level line from the count check means record counts differed
between the two stores; open an issue including that line (counts are
per-table, join-aware — `[QUOTE-AT-MERGE: validation.go]`).

## How long does it take?

It runs during startup, before the API serves requests. Typical stores
migrate in seconds; size roughly scales with `data.json`.
`[QUOTE-AT-MERGE: any measured guidance from #424/#425 test evidence, else
state "seconds to minutes for large stores" — no invented numbers]`
On Kubernetes, raise `startupProbe`/startup timings above your observed
migration window for large stores (readiness must not kill a migrating
pod).

### Previewing without changing anything

`[QUOTE-AT-MERGE: exact flag + env name from #424 — planned
`--migrate-dry-run` / `DB_MIGRATE_DRY_RUN`]` runs the whole flow and
reports what it *would* migrate without writing data.

## After the upgrade (SQLite operations)

- Your data now lives in `<database folder>/prl-devops.db`.
  `[QUOTE-AT-MERGE: filename from merged config.go constant]`
- SQLite runs in WAL mode (`[QUOTE-AT-MERGE: pragma in sqlite.go]`): you
  may see `prl-devops.db-wal` and `prl-devops.db-shm` next to it. They are
  part of the database — back up all files together, never delete the
  sidecars while the service runs.
- **SQLite is single-host.** Do not place the file on a shared/network
  filesystem or run multiple replicas against it; that topology is what
  PostgreSQL is for ([PostgreSQL Setup](.../postgresql/)).
- **Backups:** the automatic backup settings
  (`DATABASE_NUMBER_BACKUP_FILES`, `DATABASE_BACKUP_INTERVAL_MINUTES`)
  still back up the legacy `data.json` only — nothing built into this
  release backs up `prl-devops.db` automatically
  `[RE-VERIFY at merge: src/data/backup.go still legacy-only?]`. Until
  backup support for the new database ships, take consistent backups with
  standard SQLite tooling (for example
  `VACUUM INTO '/backup/prl-devops-$(date +%F).db'`, or a file copy of
  `prl-devops.db*` with the service stopped). Backup tooling for the new
  database is tracked outside this release
  ([OQ-427-1 — link the issue here once filed]).

## Rolling back

`data.json` keeps its pre-upgrade content, so rollback = stop the new
binary, start the previous release pointed at the same folder
(`DATABASE_FOLDER`): it reads `data.json` exactly as before. The
deprecation warning it now logs is expected
`[QUOTE-AT-MERGE: exact warn line from #426]`.

State this plainly: **any change made while the SQL version was running is
not in `data.json`** — rolling back loses writes made after the upgrade
unless you export them first (via the API). Rollback support lasts only
until the legacy layer is removed — see [Deprecation](#deprecation).
`[RE-VERIFY against merged #424/#426 that the previous release really can
open the folder again — adjust wording to merged truth]`

## Deprecation {#deprecation}

The legacy JSON data layer (`data.json`) is **deprecated**. It still ships
and still runs (with a startup warning) so upgrades can roll back, but it
receives no new features, and it will be removed in a future version — see
the deprecation entry in the
[project changelog](https://github.com/Parallels/prl-devops-service/blob/main/CHANGELOG.md)`[+#426 entry anchor]`
and the removal issue `[QUOTE-AT-MERGE: link from #426's changelog entry]`.
**Never restate a version number in prose** — link.

## FAQ

- **Is starting the service twice mid-migration dangerous?** No — a failed
  run rolls back; the marker makes completed runs no-ops.
- **Can I point two installs at one SQLite file?** No — see
  [After the upgrade](#after-the-upgrade-sqlite-operations).
- **How do I move from SQLite to PostgreSQL later?**
  `[QUOTE-AT-MERGE: does #424 support migrating into an empty postgres
  target? If unmerged/unsupported, state: take a backup and open the
  issue — do not promise a path that does not exist]`
- **Can I delete data.json after upgrading?** Not yet — it is the rollback
  copy and is needed until the legacy layer's removal (see
  [Deprecation](#deprecation)).
```

(Replace the `...` link prefixes with the site's `{{ site.url }}{{ site.baseurl }}` Liquid form, as
`docker.md:31` does. Prose wraps ≤ 80 cols.)

### 3. Evidence table (PR description artifact — D-2)

| Claim in the page | Merged source to quote |
|---|---|
| detection condition (file present + non-empty; no side effects) | `src/database/legacy_migrator/detector.go` |
| the 4 verbatim log lines + integrity-check warning | `legacy_migrator/*.go` `ctx.Log*` calls |
| marker table name/columns; skip-on-marker behaviour | `legacy_migrator/marker.go` |
| transactional failure → exit + error text | `legacy_migrator/migrator.go` + startup wiring |
| dry-run flag/env names | #424 merged flags (`src/constants`/startup) |
| `prl-devops.db` filename; WAL pragma | `src/database/config.go`, `sqlite.go` |
| frozen `data.json` guarantee; legacy-start warn text; removal issue link | #424/#426 merged code + #426's CHANGELOG entry |
| backups cover only `data.json` | `src/data/backup.go` (still legacy-only at merge?) |

### 4. Formatting compliance

New files start clean: `markdownlint` (prose ≤ 80; tables exempt) and `prettier@3 --check` both
must report **zero** findings on the new page. Commands in Validation.

## Out of scope

Editing `configuration.md` (01) or `architecture.md`/README/CHANGELOG (05); the Postgres page
(03); `docs/_site`; anything under `src/`.

## Validation

- [ ] `export npm_config_cache=/tmp/pbi427-npmcache`
- [ ] `npx --yes markdownlint-cli@0.41 docs/docs/devops/getting-started/database-upgrade.md --config .markdownlint.json` → 0 findings
- [ ] `npx --yes prettier@3.6 --check docs/docs/devops/getting-started/database-upgrade.md` → pass
- [ ] Menu YAML still parses: `ruby -ryaml -e "YAML.load_file('docs/_data/docs_devops_menu.yml')"`
- [ ] Render check (preferred): `cd docs && bundle install && bundle exec jekyll build --destination /tmp/pbi427-site` → page HTML exists at `/tmp/pbi427-site/docs/devops/getting-started/database-upgrade/index.html` **with TOC**; the new menu item appears in the rendered sidebar/navbar. Fallback checklist if build infeasible: front-matter keys identical to `configuration.md:1-8` (only `subtitle` differs); permalink↔file mapping verified; every internal link target exists as a file or is a site-absolute URL used elsewhere; menu link equals the page path + trailing `/`
- [ ] No `docs/_site` changes (`git status` clean; it is gitignored anyway); `git diff --stat` shows exactly the two files above
- [ ] PR description contains the evidence table, each row with `file:line`, and zero `[QUOTE-AT-MERGE]` markers remain in the committed page (grep it): `grep -c 'QUOTE-AT-MERGE' docs/docs/devops/getting-started/database-upgrade.md` → 0
- [ ] **AC 6:** approved by ≥ 1 non-author reviewer (they re-check ≥ 3 evidence rows)

## Rollback

Revert commit (page + menu lines vanish together — single commit). A dead nav link cannot outlive
the PR because both changes are in it.

## PBI acceptance mapping

AC 2 (the whole upgrade spine); AC 5 (authoritative deprecation section); AC 6 (review gate).

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `427-02`.** Cut `slice/427-02` from `sql-migration` HEAD
(`.worktrees/slice-427-02`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
