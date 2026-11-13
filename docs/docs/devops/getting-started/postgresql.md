---
layout: page
title: Getting Started
subtitle: PostgreSQL Setup
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

> **Status:** PostgreSQL enablement is in progress. The service's
> data-access layer still relies on SQLite-specific row-ordering
> behavior, so running the service end-to-end against PostgreSQL is
> not yet supported; the gap is tracked in
> [#417](https://github.com/Parallels/prl-devops-service/issues/417).
> SQLite (the default) is the supported backend today. This page is
> the setup reference for when enablement completes — re-check the
> issue before planning a production deployment.

With `DB_TYPE=postgres` the service is stateless with respect to its
data: no local database file — the SQLite setting (`DB_PATH`) and the
legacy `data.json` play no role. Multiple replicas sharing one
database is the target deployment once the backend is enabled
([Kubernetes setup guide]({{ site.baseurl }}/docs/devops/getting-started/kubernetes/)).

## Prerequisites

- A PostgreSQL server. The service's CI suite exercises PostgreSQL 16;
  newer stable versions are expected to work (the driver is libpq
  compatible).
- Network reachability from every service replica. Credentials below.

## 1. Create the role, database and schema

Run as a PostgreSQL administrator:

```sql
CREATE ROLE devops LOGIN PASSWORD 'change-me';
CREATE DATABASE devops OWNER devops;

\c devops
CREATE SCHEMA devops_service AUTHORIZATION devops;
ALTER ROLE devops IN DATABASE devops
  SET search_path = devops_service, public;
```

The service creates its own tables on startup (`DB_AUTO_MIGRATE`,
default true) — you do not run schema scripts. The dedicated schema
keeps the service's objects isolated and makes `GRANT`s and dumps
scoped.

## 2. Build the DSN and set the environment variables

`DB_DSN` is passed through to the PostgreSQL driver unchanged, so both
standard libpq forms work. The URL form is what the service's CI
exercises:

| DSN form      | Example (password always fake)                                                               |
| ------------- | -------------------------------------------------------------------------------------------- |
| URL           | `postgres://<db-user>:<db-password>@db.example.com:5432/devops?sslmode=require`                     |
| keyword/value | `host=db.example.com port=5432 user=<db-user> password=<db-password> dbname=devops sslmode=require` |

If you created a dedicated schema, add `search_path=devops_service` to
the DSN: as `&search_path=devops_service` in URL form, or as the
`search_path=devops_service` keyword in keyword/value form.

```bash
export DB_TYPE=postgres
export DB_DSN='postgres://<db-user>:<db-password>@db:5432/devops?sslmode=require'
export DB_AUTO_MIGRATE=true   # default; set false once the schema exists
```

### SSL mode

Set `sslmode` **explicitly**. The service never rewrites the DSN; with
`sslmode` unset the driver default is `prefer` (TLS if the server
offers it, plaintext otherwise), which is not a security posture:

| sslmode                         | Use it when                                      |
| ------------------------------- | ------------------------------------------------ |
| `disable`                       | private/trusted network only                     |
| `require`                       | encryption needed, server cert not pinned        |
| `verify-full` + `sslrootcert=…` | production default: encryption + server identity |

### Passwords in logs

The service redacts credentials from every DSN it logs, fail-closed:
an unrecognised string collapses to `[redacted-dsn]` rather than being
echoed. URL usernames are deliberately kept for triage; passwords,
`sslkey` paths and `sslpasswords` never survive. A redacted line looks
like `postgres://devops:***@db:5432/devops?sslmode=require`.

### Connection budget (pool math)

Each replica caps its connections at: **25 max open**, **5 max idle**,
**30 minutes** max lifetime (`src/database/config.go`, applied via
`sql.DB.Set*`). Plan against the server's `max_connections`:

> replicas × max_open ≤ max_connections − headroom (monitoring, admin,
> other clients)

With the shipped defaults that is roughly **3 replicas** against a
stock server (`max_connections` 100, before superuser reservations).
Pool sizes are fixed defaults in this release — there is no environment
override. Scale replicas against `max_connections`, or run a
transaction pooler (PgBouncer) in front; with transaction pooling,
keep server-side `statement_timeout` low and avoid session-level
features (prepared statements, advisory locks) that assume sticky
connections.

## 3. Verify

Start the service. Startup and connection logs use the redacted DSN
only. A wrong or missing configuration fails at startup with wording
that names the setting, never the secret value:

| Startup failure (verbatim)                                                                                     |
| -------------------------------------------------------------------------------------------------------------- |
| `unsupported database type, expected 'sqlite' or 'postgres'`                                                   |
| `database type 'postgres' requires a connection string (DSN) — set DB_DSN (env) or db_dsn (config key db_dsn)` |

Then confirm the schema is populated: connect with `psql`, run
`SET search_path TO devops_service;` and
`SELECT count(*) FROM users;` inside your schema.

## Kubernetes (no PVC needed)

With a PostgreSQL backend the deployment needs no persistent volume
for the database: leave the chart's `storage.databasePath` empty — the
PVC and its mount are skipped entirely (the template gates both on
that value). Other local state is unaffected by the database choice:
file logging (`LOG_TO_FILE`) and the catalog cache still write to
disk.

Pass `DB_DSN` as a Kubernetes Secret so it never lands in a values
file:

```bash
kubectl create secret generic devops-db \
  --from-literal=DB_DSN='postgres://<db-user>:<db-password>@db:5432/devops?sslmode=require'
```

and inject it with the chart's generic environment passthrough
(`envFrom` in `values.yaml`):

```yaml
envFrom:
  - secretRef:
      name: devops-db
```

See the
[Kubernetes setup guide]({{ site.baseurl }}/docs/devops/getting-started/kubernetes/)
for the chart walkthrough.

### Helm values

The chart also accepts first-class database values, keeping the DSN in
an existing Secret you manage:

```yaml
database:
  type: postgres
  dsnSecret:
    name: devops-db
    key: dsn
```

These render to the same `DB_*` environment variables as the manual
setup above; the full block lives in the chart's `values.yaml`.
With a SQL backend set `storage.databasePath: ""` — the PVC and its
mount are skipped entirely.

<!-- gated in helm/templates/deployment.yaml:46-50 and 282-285 -->

## Docker Compose

Add to the `environment:` block of your `docker-compose.yml` (the
shipped file keeps `DATABASE_FOLDER` for its volume mount; the
database variables are yours to add):

```yaml
environment:
  DB_TYPE: "postgres"
  DB_DSN: "postgres://<db-user>:<db-password>@db:5432/devops?sslmode=require"
```

Example sidecar shape:

```yaml
db:
  image: postgres:16
  environment:
    POSTGRES_USER: devops
    POSTGRES_PASSWORD: change-me
    POSTGRES_DB: devops
  volumes:
    - db-data:/var/lib/postgresql/data
```

In real deployments use secrets or a secret manager, not compose
literals.

## Recommended PostgreSQL settings

- Right-size `max_connections` for the formula above; a transaction
  pooler in front for large fleets.
- Keep `DB_AUTO_MIGRATE=true` only where the service may own schema
  changes; managed or production databases often prefer one migrating
  instance, then `DB_AUTO_MIGRATE=false` everywhere else.
- Back up per your PostgreSQL operations runbook; the service's
  built-in backup settings target the legacy file only. See the
  [upgrade guide]({{ site.baseurl }}/docs/devops/getting-started/database-upgrade/).
