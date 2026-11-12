---
layout: page
title: Getting Started
subtitle: Upgrading to the SQL Database
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

The service now stores its data in a SQL database: SQLite by default (a
single file next to your existing data), or PostgreSQL for orchestrated
deployments (backend enablement in progress):

[PostgreSQL Setup]({{ site.baseurl }}/docs/devops/getting-started/postgresql/)

The legacy JSON store (`data.json`) is deprecated; see
[Deprecation](#deprecation) below.

## What happens on the first start after upgrading

1. **Detection** — the service looks for the legacy `data.json` in the
   database folder (`DATABASE_FOLDER`; if unset: `/etc/parallels-devops`
   when running as root, otherwise `~/.parallels-devops`). If the file is
   absent, unreadable, or holds no data, nothing happens — fresh installs
   just start.
2. **Schema** — the SQL schema is created and kept current on every start
   by `DB_AUTO_MIGRATE` (default `true`).
3. **Data migration** — all records are copied into the SQL database
   inside one single transaction. If any collection fails, the whole
   migration rolls back, the service logs the error and (for the API
   command) exits; your `data.json` is never modified.
4. **Marker** — a completion marker row is written **in the new
   database** (table `legacy_migration_marker`: source file path, size,
   modification time, content SHA-256, records total, completion time).
   Every later start sees the marker and skips the migration — even if
   `data.json` changed in the meantime.
5. **Serve** — the service runs against the SQL database.

## What you will see in the log

| Stage                              | Log line (abridged)                                                                                                  |
| ---------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Fresh install / nothing to migrate | `legacy migration skipped: <reason>` (debug level)                                                                   |
| Per-collection progress            | `migrated 12/12 accounts records (total so far 12/57, 3ms)`                                                          |
| Record counts verified             | `legacy migration validation ok: accounts 12/12`                                                                     |
| Count mismatch (non-fatal)         | `legacy migration validation MISMATCH accounts: legacy=12 new=11 — investigate (migration NOT failed by validation)` |
| Failure and rollback               | `legacy migration rolled back: <detail>` and `Legacy data migration failed (all changes rolled back): <detail>`      |
| Every later start                  | `legacy store <file> already migrated (unchanged since migration)`                                                   |

Count mismatches are logged but never fail the migration; check the
service's support data if you see one.

## How long does the migration take?

Seconds on typical installations; the duration is proportional to the
size of your `data.json`. The service does not serve requests while the
startup migration runs. On Kubernetes, size the startup probe
accordingly.

## Can I try it without committing? (dry run)

`DB_MIGRATE_DRY_RUN=true` (or `--migrate-dry-run=true`) prints what would
be migrated and writes nothing:

```text
legacy migration (dry-run): would migrate 12 accounts records
legacy migration (dry-run): 57 records across 11 collections — nothing written
```

The real migration then runs on the next normal start.

## Backups after upgrading

The built-in file backups (`DATABASE_NUMBER_BACKUP_FILES`,
`DATABASE_BACKUP_INTERVAL_MINUTES`) were a feature of the legacy JSON
engine and only ever covered `data.json`. With SQLite (the default) your
backup routine must include the SQLite file — copy it with
`sqlite3 <file> ".backup <dest>"` or snapshot it while the service is
stopped. No SQL backup is built in today.

## Rollback

A rollback works only **before you create new data**, otherwise changes
made in the SQL database are lost. Your `data.json` was not modified by
the migration, so the older release can read it again.

To roll back:

1. Stop the service.
2. Start an older release against the same `DATABASE_FOLDER`. If that
   release predates the SQL deprecation warnings it will not print them —
   that is expected. A current release also keeps serving from the old
   engine when you set `DB_TYPE=legacy`; it then logs, on every start:

   | Legacy-mode startup warning (verbatim)                                                                                                                                                                                     |
   | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
   | `[Database] DEPRECATED: legacy JSON data layer selected (DB_TYPE=legacy). The legacy layer is deprecated since v1.2.0 and will be removed in v1.3.0 (issues #426/#417). Plan migration to DB_TYPE=sqlite — see CHANGELOG.` |

3. Data created or changed in the SQL database while it was live is not
   in `data.json`.
4. If you start in SQL mode again later, the migration does **not**
   re-run: the marker says it happened, and the SQL data (possibly
   stale) is kept. Removing the marker row from the
   `legacy_migration_marker` table forces a re-run — an advanced
   operation, not recommended.

This page targets releases that already contain the SQL layer. If you
are coming from older releases, upgrade directly to a release from this
one onward — the JSON migration is identical on every later release;
it is never re-run on already-migrated databases.

## Deprecation

The legacy JSON data layer (`data.json` and the `DB_TYPE=legacy` mode) is
deprecated (#426, feature #417). The deprecation is in effect from the
[CHANGELOG entry](https://github.com/Parallels/prl-devops-service/blob/main/CHANGELOG.md).
The changelog is the authoritative source for the exact removal version.
Removal progress is tracked in
[issue #426](https://github.com/Parallels/prl-devops-service/issues/426).

When the removal happens, all of the following stop working:

- `DB_TYPE=legacy` is no longer accepted (startup rejects the value),
- the `data.json` file is no longer read or migrated — the automatic
  first-start migration goes away with the engine,
- the legacy-only settings are removed.

Never re-introduce `data.json` after a successful migration: the SQL
database is the live record and the legacy file goes stale immediately.
A deployment that runs the legacy engine on a store that was already
migrated gets a loud warning naming the frozen file on every start:

| Frozen-store warning (verbatim)                                                                                                                                                                                   |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `[Database] DEPRECATED: legacy data file <file> was already migrated and is now FROZEN/stale — this instance serves stale data; writes here never reach the SQL database. Switch to DB_TYPE=sqlite (#424, #426).` |

## FAQ

**Does the migration delete my `data.json`?**
No. The file is never modified or deleted by the migration. After a
successful migration it is a frozen snapshot of the data at migration
time; you may archive — or delete — it yourself.

**Can I move from SQLite to PostgreSQL later?**
There is no supported conversion path in this release; the automatic
migration only covers the legacy JSON store. Plan a fresh PostgreSQL
setup, or export what you need via the API before switching. Conversion
tooling between database backends is future work.

**Why is there no backup for my SQLite file?**
The built-in backups were a JSON-engine feature; SQL backups are the
operator's tooling — see [Backups after upgrading](#backups-after-upgrading).
