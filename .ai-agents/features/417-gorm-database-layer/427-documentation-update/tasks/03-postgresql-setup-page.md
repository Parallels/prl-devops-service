# Task 03 — PostgreSQL setup page + nav registration

**PBI** [#427](./brief.md) · **Feature** [#417](../brief.md) · **Depends-on:** task 02 (menu-file conflict → strictly sequential merge); #418/#420/#421 merged (DSN handling, pool, redaction final) · **PR size:** M
**Plan refs:** [D-1](./plan.md#d-1--page-inventory-and-nav-placement-decided-from-the-real-docs-structure), [D-2](./plan.md#d-2--content-accuracy-protocol-binding-on-every-task), OQ-427-3/-4

## Context

AC 3 wants a step-by-step Postgres guide: role/db creation SQL, both DSN forms, sslmode, connection
budget math (pods × pool — quote merged defaults), stateless-deployment story, k8s Secret pattern,
compose example. The Helm chart's first-class `database.*` values land in task 04 **after** this
page — so this page documents only what works **at its merge time** (explicit env vars / `envFrom`
secretRef, verified available today: `helm/templates/deployment.yaml:256-259` + `helm/values.yaml:34`).
Task 04 appends the chart-values paragraph.

## Objective

`docs/docs/devops/getting-started/postgresql.md` published at `/docs/devops/getting-started/postgresql/`,
menu-registered, lint-clean from zero, every default/behaviour evidenced from merged code.

## Files

| File | Op | Current lines |
|---|---|---|
| `docs/docs/devops/getting-started/postgresql.md` | create | 0 → ≈190 |
| `docs/_data/docs_devops_menu.yml` | modify — append inside `Getting Started` group, after the entry task 02 added | 76 → 78 |

## File size guard

Both ≪ 1,000. **ok**.

## Implementation spec

### 1. Menu insert (exact YAML, after task 02's lines)

```yaml
    - name: PostgreSQL Setup
      link: /docs/devops/getting-started/postgresql/
```

### 2. Page draft (front matter copied verbatim from sibling `configuration.md:1-8`, only
`subtitle` changes)

```markdown
---
layout: page
title: Getting Started
subtitle: PostgreSQL Setup
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# PostgreSQL Setup

Set `DB_TYPE=postgres` to run the service statelessly against PostgreSQL —
required for multiple replicas; recommended for production
([architecture note](.../architecture/)). With PostgreSQL the service
needs no local database file: the SQLite settings (`DB_PATH`) and the
legacy `data.json` play no role.

## Prerequisites

- A PostgreSQL server at a [currently supported stable
  version]`[OQ-427-3: name the floor only if #417's PBIs or CI evidence
  tested a specific one — otherwise keep the general guidance]`.
- Network reachability from every service replica, credentials as below.

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

The service creates its own tables on startup (`DB_AUTO_MIGRATE`, default
true) — you do not run schema scripts. The dedicated schema keeps the
service's objects isolated and makes `GRANT`s and dumps scoped.

## 2. Build the DSN and set the environment variables

Both standard forms work (`[QUOTE-AT-MERGE: confirm both accepted by the
merged postgres driver wiring in src/database/postgres.go — driver is
pgx-based per #418; if only one form is proven in tests, show that one
prominently and keep the other as "also supported (URL form)" if proven]`):

```text
# URL form
postgres://devops:change-me@db.example.com:5432/devops?sslmode=require&search_path=devops_service

# keyword/value form
host=db.example.com port=5432 user=devops password=change-me dbname=devops sslmode=require search_path=devops_service
```

```bash
export DB_TYPE=postgres
export DB_DSN='postgres://devops:change-me@db.example.com:5432/devops?sslmode=require&search_path=devops_service'
export DB_AUTO_MIGRATE=true   # default; set false once the schema exists
```

Passwords in `DB_DSN` are redacted in every log/error line the service
emits (`show the redacted output actually produced by merged redaction
code/tests` `[QUOTE-AT-MERGE: src/database/redact.go + *_test.go]`).

### SSL mode

Set `sslmode` **explicitly** — do not rely on a driver default
`[OQ-427-3: if the merged driver's default differs from 'disable', say so
in one sentence]`:

| sslmode | Use it when |
| --- | --- |
| `disable` | private/trusted network only |
| `require` | encryption needed, server cert not pinned (common default choice) |
| `verify-full` + `sslrootcert=…` | production default: encryption + server identity |

### Connection budget (pool math)

The service caps each replica's PostgreSQL connections at
`max open [QUOTE-AT-MERGE: merged DefaultPostgresMaxOpenConns — planned
25]`, `max idle […  planned 5]`, connection lifetime [… planned 30m]
(`src/database/config.go`, applied via `sql.DB.Set*` in `postgres.go`).
Plan against the server's `max_connections` (default 100, of which
superuser reservations apply):

> replicas × max_open ≤ max_connections − headroom (monitoring, admin,
> other clients)

At the shipped defaults that is roughly **3 replicas** against a stock
server. `[Branch on OQ-427-4: if #420 shipped pool env overrides, name
them here and replace the "fixed" sentence; if not, state: pool sizes are
fixed defaults in this release — scale replicas against max_connections
or run a transaction pooler (PgBouncer) and document
statement-timeout/prepare pitfalls.]`

## 3. Verify

Start the service; GORM logs the migration/connect at info (`LOG_LEVEL=info`)
using the **redacted** DSN. A wrong/missing `DB_DSN` fails at startup with
an actionable error naming the setting, not the secret value
(`[QUOTE-AT-MERGE: startup validation error strings from merged #420]`).
Then confirm rows: `\c devops` → `SET search_path…` context → `SELECT
count(*) FROM users;` inside your schema.

## Kubernetes (no PVC needed)

With `DB_TYPE=postgres` the deployment needs no persistent volume for the
database: leave the chart's `storage.databasePath` empty — the PVC and its
mount are skipped entirely (`helm/templates/deployment.yaml` gates both on
that value). Keep in mind other local state (logs if `LOG_TO_FILE`, the
catalog cache folder) is unaffected by the database choice.
`[RE-VERIFY at merge: what still writes to disk — check merged
serviceprovider path resolution; state only the truth.]`

Pass `DB_DSN` as a Kubernetes Secret so it never lands in values files:

```bash
kubectl create secret generic devops-db \
  --from-literal=DB_DSN='postgres://devops:...@db:5432/devops?sslmode=require'
```

and reference it via the chart's existing generic env passthrough
(`helm/values.yaml`, `envFrom`):

```yaml
envFrom:
  - secretRef:
      name: devops-db
```

(First-class `database.*` chart values are added on top of this — see the
[Helm chart page](.../kubernetes/helm-chart/)`[paragraph added by task 04;
keep this forward-link only if task 04 is already merged when this PR
merges — per merge order it is not, so OMIT this sentence in the initial
PR; task 04 appends its own paragraph]`.)

## Docker Compose

Add to the `environment:` block of your `docker-compose.yml` (the shipped
file keeps `DATABASE_FOLDER` for the volume mount; the DB envs are yours
to add):

```yaml
    environment:
      DB_TYPE: 'postgres'
      DB_DSN: 'postgres://devops:change-me@db:5432/devops?sslmode=require'
```

Example sidecar shape:

```yaml
  db:
    image: postgres:17
    environment:
      POSTGRES_USER: devops
      POSTGRES_PASSWORD: change-me
      POSTGRES_DB: devops
    volumes:
      - db-data:/var/lib/postgresql/data
```

(Passwords here are placeholders; in real deployments use secrets/secret
managers, not compose literals.)

## Recommended PostgreSQL settings

- Right-size `max_connections` for `replicas × max_open + headroom`
  (formula above); a transaction pooler in front for large fleets.
- Keep `DB_AUTO_MIGRATE=true` only where the service may own schema
  changes; managed/production DBs may prefer running one migrating
  instance then setting `DB_AUTO_MIGRATE=false` on the rest.
- Back up per your PostgreSQL ops (the service's built-in backup settings
  target the legacy file only — see
  [upgrade guide](.../database-upgrade/)).
```

(`...` prefixes → the site's `{{ site.url }}{{ site.baseurl }}` Liquid
form; prose ≤ 80 cols; tables exempt from line-length.)

### 3. Evidence table (PR artifact)

| Claim | Merged source |
|---|---|
| both DSN forms accepted (or which one) | `src/database/postgres.go` + its tests (pgx parse) |
| pool numbers (25/5/30m or final) | `src/database/config.go` constants |
| redacted-DSN sample | `src/database/redact.go` + tests |
| startup validation error wording | #420 merged errors |
| envFrom passthrough + PVC gating facts | `helm/values.yaml:34-37`, `helm/templates/deployment.yaml:46-50,256-259,282-285` (verify lines still true at merge) |
| DB_AUTO_MIGRATE default; accepted DB_TYPE values | merged #420 |
| "what still writes to disk with postgres" | merged serviceprovider path resolution |
| pool override existence (OQ-427-4) | merged `src/constants`/`src/config` |

### 4. Formatting compliance

New file must report **zero** markdownlint + prettier findings (commands
below).

## Out of scope

Helm template/values edits + their docs (task 04); upgrade/rollback
narrative (task 02); changelog/README (task 05); maintainer values files
(`values-prd.yaml` etc.); root compose files.

## Validation

- [ ] `export npm_config_cache=/tmp/pbi427-npmcache`
- [ ] `npx --yes markdownlint-cli@0.41 docs/docs/devops/getting-started/postgresql.md --config .markdownlint.json` → 0 findings
- [ ] `npx --yes prettier@3.6 --check docs/docs/devops/getting-started/postgresql.md` → pass
- [ ] `ruby -ryaml -e "YAML.load_file('docs/_data/docs_devops_menu.yml')"` succeeds
- [ ] Render check (preferred): `cd docs && bundle exec jekyll build --destination /tmp/pbi427-site` → `…/postgresql/index.html` exists, TOC renders, SQL/bash/yaml fences highlighted, both new menu items present. Fallback checklist as in task 02 (front-matter parity, permalink mapping, link targets exist)
- [ ] `git diff --stat` → exactly the two files above
- [ ] PR description: evidence table with `file:line` rows; `grep -c 'QUOTE-AT-MERGE' docs/docs/devops/getting-started/postgresql.md` → 0; secrets in samples obviously fake (`change-me`)
- [ ] **AC 6:** approved by ≥ 1 non-author reviewer

## Rollback

Revert commit (page + menu line together). No cross-page edits were made
in this PR — task 02's page links to this URL, and reverting 03 after 03
merged leaves that link dangling, so rollback after merge must be paired
with a follow-up PR removing task 02's forward link (note in revert PR).

## PBI acceptance mapping

AC 3 (primary, whole spine); AC 6 (review gate). Supports AC 4 via the
architecture link (note owned by 05).
