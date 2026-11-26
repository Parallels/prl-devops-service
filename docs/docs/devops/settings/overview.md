---
layout: page
title: Settings
subtitle: Overview
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# Settings Service

The Settings Service lets you change a fixed set of runtime options without
touching config files, environment variables or restarting anything. The values
live in the database, you change them through the API or the UI, and the running
process picks them up straight away. When you run an orchestrator, it keeps the
master copy and pushes each host's configuration down to that host, much like an
MDM pushes a profile to a device.

A few things are worth knowing up front:

- **Only catalogued keys can be set.** The service keeps a list of keys that are
  safe to change while the service is running. Anything the process needs at
  boot, and anything secret, is not on that list and cannot be written through
  the API at all.
- **You can always tell where a value came from.** Every key goes through the
  same precedence chain, and the API tells you which layer provided the answer.
- **A host only sees its own configuration.** The orchestrator sends each host
  exactly its own bundle. A host has no way to read another host's overrides.

> **Availability** — settings are stored in the SQL database (SQLite or
> PostgreSQL). If your service still runs on the legacy JSON database, the
> feature is switched off and every settings endpoint returns `503`. See
> [Upgrading to SQL]({{ site.baseurl }}/docs/devops/getting-started/database-upgrade/).

## How a value is resolved

Every setting on a node is looked up through the same chain, highest priority
first:

| # | Layer | Where it lives | Reported `source` |
|---|-------|----------------|-------------------|
| 1 | Database layers (see below) | settings tables | `orchestrator_db` / `managed_db` / `local_db` |
| 2 | Environment variables | process environment | `env` |
| 3 | CLI flags / configuration file | launch flags, `prldevops_config` file | `config_file` |
| 4 | Catalog default | compiled-in default from the catalog | `default` |

So a value stored in the database beats an environment variable. That is
intentional, otherwise the service would be of little use on a host whose
environment was baked into a container image. Storing an empty string is also
intentional: it counts as "clear this", and it still wins over the environment,
which is handy when you need to unset something that came in from the image.

### The database layers

Within the database layer, each node has its own view:

- **Orchestrator node** — the `orchestrator` scope (its own values) wins over
  its `local` layer. When several orchestrator replicas share this database, a
  value addressed to one of them (`node:<machine-id>`, see below) wins over the
  shared `orchestrator` value for that replica alone.
- **Host node** — the `managed` layer (what the orchestrator pushed) wins over
  the `local` layer (edits made directly on the host through its own console).
  A sync never deletes local edits. They just stop taking effect, and the API
  marks them with `shadowed_by` so the UI can show something like "stored here,
  but the managed value wins".

## Scopes

Every stored value belongs to a scope. The orchestrator's database holds all of
them, while a host only ever receives the rows meant for it:

| Scope | Meaning | Who sets it |
|-------|---------|-------------|
| `orchestrator` | The orchestrator process's own values | Orchestrator console |
| `node:<machine-id>` | One orchestrator replica's own values, wins over `orchestrator` for that machine | Orchestrator console |
| `hosts` | A value for **all** hosts | Orchestrator console |
| `host:<hostId>` | Per-host override, wins over `hosts` | Orchestrator console |
| `local` | A node's own edits (host direct console) | The node itself |

If you write the same key in both `hosts` and `host:h1`, every host gets the
global value and `h1` gets its own. Delete the per-host row and `h1` goes back
to the global value. Delete both and it goes back to the catalog default.

### Per-replica values (`node:<machine-id>`)

Several orchestrator replicas may share one database. That is fine for values
that are the same on every replica, and a trap for values that must not be: a
single shared row would be read by all of them. The `node:` scope is the answer,
keyed by the machine rather than by the process — exactly as a per-host value is
keyed by its host.

Address it with `?node=<machine-id>` on the orchestrator value endpoints:

```bash
# Which machine am I looking at? Its identity and its machine id are both here.
curl -s $BASE/api/v1/config/hardware | jq '{orchestrator_node_id, machine_id}'

# Give that machine a name of its own.
curl -X PUT "$BASE/api/v1/orchestrator/settings/values/ORCHESTRATOR_NODE_ID?node=6E2BF1C4-1B2A" \
  -d '{"value":"orchestrator-1"}'
```

The machine id is the hardware UUID where the platform publishes one, and the
hostname where it does not (Windows has no platform id). It is *stable*: renaming
a replica does not change the key you use to address it, which is what makes a
rename safe to do through the UI.

Two rules keep the scope from undoing the thing it exists for:

- `ORCHESTRATOR_NODE_ID` is refused in the shared `orchestrator` scope. The
  error names the machine you are talking to, so there is no guessing.
- Two machines cannot be given the same identity — the write is refused and the
  message says which machine already holds it. Two replicas answering to one
  name would sweep each other's running jobs at boot and share one reservation
  owner, which is precisely the failure this feature closes.

`ORCHESTRATOR_NODE_ID` applies immediately: the replica adopts the name and
re-stamps its own rows (`jobs.owner_node`, its capacity reservations) so nothing
is left belonging to a name that no longer exists. Deleting the row returns the
replica to the name it derives for itself.

Host-scoped keys are rejected in a `node:` scope: a host pulls its values from
the host scopes, so such a row would sit in the table looking like it had taken
effect.

## The catalog

The catalog is compiled into the binary and written into the database every
time the service boots in SQL mode. Existing rows are left alone, but their
metadata is updated so it matches the release you are running. Each definition
has a `type` (`string` / `int` / `bool`), a `default_value`, an optional
`valid_values` list, an `applies_to` field (`all` / `host` / `orchestrator`) and
a `requires_restart` flag. A write for a key that is not in the catalog, has the
wrong type, or uses a value outside the allowed list is rejected with `400`.

These are the 23 keys currently in the catalog:

| Key | Type | Applies to | Restart required |
|-----|------|-----------|------------------|
| `LOG_LEVEL` | string (`debug,info,warn,error`) | all | yes |
| `ENABLE_CANARY` | bool | all | yes |
| `ENABLE_BETA` | bool | all | yes |
| `ORCHESTRATOR_PULL_FREQUENCY_SECONDS` | int | orchestrator | yes |
| `ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS` | int | orchestrator | yes |
| `MAX_VMS_PER_HOST` | int | host | no |
| `MAX_VMS_PER_HOST_MODE` | string (`any,macos,linux`) | host | no |
| `SYSTEM_RESERVED_CPU` | int | host | no |
| `SYSTEM_RESERVED_MEMORY` | int | host | no |
| `SYSTEM_RESERVED_DISK` | int | host | no |
| `DISABLE_VM_CACHE_REFRESH` | bool | host | no |
| `VM_CACHE_REFRESH_INTERVAL_SECONDS` | int | host | no |
| `ENABLE_PACKER_PLUGIN` | bool | host | no |
| `ENABLE_VAGRANT_PLUGIN` | bool | host | no |
| `USE_ORCHESTRATOR_RESOURCES` | bool | host | no |
| `DISABLE_CATALOG_CACHING` | bool | all | no |
| `CATALOG_CACHE_KEEP_FREE_DISK_SPACE` | int | all | no |
| `CATALOG_CACHE_MAX_SIZE` | int | all | no |
| `CATALOG_CACHE_ALLOW_CACHE_ABOVE_FREE_DISK_SPACE` | bool | all | no |
| `CATALOG_COMPRESS_VM` | bool | orchestrator | no |
| `CATALOG_COMPRESS_VM_RATIO` | string (`best_speed,balanced,best_compression,no_compression,default`) | orchestrator | no |
| `DISABLE_CATALOG_PROVIDER_STREAMING` | bool | all | no |
| `CATALOG_ENABLE_PROVIDER_CREDENTIALS_OBFUSCATION` | bool | all | no |

Two groups of keys are deliberately left out and can never be set this way:

1. **Boot-critical keys** such as `DB_*`, `TLS_*`, `API_PORT`, `MODE`,
   `ENABLED_MODULES`, `BASE_URL` and the log file paths. The process needs
   these before it can even open the database, so storing them in the database
   would not do anything useful. They are simply not in the catalog.
2. **Secrets** such as keys, passwords and HMAC/JWT material. They are never
   catalogued, never stored and never returned.

Some keys are only read once at startup or when a connection is set up. Those
are still in the catalog, but flagged `requires_restart: true`. The value is
stored right away and used on the next restart, and the UI can offer that
restart instead of quietly ignoring the change.

## Live application

Changes take effect without a restart because every runtime read goes through
the configuration resolver (`GetKey`), and the settings service is the first
layer it checks. Change `MAX_VMS_PER_HOST` or one of the cache budgets, for
example, and the running process reads the new value on its next lookup. Keys
flagged `requires_restart` are the exception, as described above.

## The sync engine (orchestrator → hosts)

Hosts do not own their configuration, they receive it. For each host, the
orchestrator builds a **bundle** by taking the `hosts` rows and laying the
`host:<id>` rows on top, then pushes it:

- **Content hash.** The bundle is hashed (SHA-256 over the sorted `key=value`
  set). The host remembers the hash of what it currently holds. If a push
  arrives with the same hash, the host answers `unchanged` and does nothing, so
  a sync is cheap to retry and safe to repeat.
- **Full replace.** The orchestrator always sends the whole bundle. The host
  swaps its entire `managed` layer in one go and drops any keys that are no
  longer in the bundle.
- **When a push happens.** On a scope write (the host is marked pending and the
  push runs in the background), when a host becomes healthy (the pong
  callback), when a host registers, when the service boots, and from a retry
  loop that runs at least every 10 seconds for anything still pending or in
  error.
- **Drift state.** Each host is in one of `never` / `pending` / `synced` /
  `error`, along with the last error and timestamp. You can see this in the
  values listing and in the per-host values view. A host that has never come
  online shows `never`.
- **Unregistering a host** deletes its stored values and its sync state.

On the host side the push lands on one internal endpoint,
`POST /v1/settings/apply`, authenticated with the host's own enrollment API key.
Hosts never poll for configuration.

## The API

Every node that runs the settings domain has two read endpoints. The
orchestrator adds the management endpoints on top. Paths are versioned under
`/v1` and mounted under the service's API prefix.

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/settings/definitions` | The catalog: type, default, allowed values, `applies_to`, `requires_restart` per key (static metadata) |
| GET | `/settings/values` | Stored rows, this node's effective view (value + `source` + `shadowed_by`), drift rows and node posture |
| PUT / DELETE | `/settings/values/{key}` | **Host only** — the host's own `local` layer |
| PUT / DELETE | `/orchestrator/settings/values/{key}` | The orchestrator's own values (`orchestrator` scope). Add `?node=<machine-id>` to address one replica's own layer |
| PUT / DELETE | `/orchestrator/hosts/settings/values/{key}` | The value for all hosts |
| PUT / DELETE | `/orchestrator/hosts/{hostId}/settings/values/{key}` | Per-host override |
| GET | `/orchestrator/hosts/{hostId}/settings/values` | What that host will hold after the next sync (bundle, effective values and drift) |
| POST | `/orchestrator/settings/sync` | Manual push: `{"target":"all"}` or `{"target":"host:<id>"}` |
| POST | `/settings/apply` | Internal: the orchestrator pushes a host's bundle (host API key) |

`GET /settings/values` returns three sections. `definitions` is the static
catalog metadata and does not change when you write. `values` is the stored
rows. `effective` is the resolved answer together with the layer that won. After
you change something, `effective` is the section to look at.

### Security

The endpoints are gated by claims: reads need `LIST_SETTINGS`, writes need
`UPDATE_SETTINGS` and the manual sync needs `SYNC_SETTINGS`. The
`/orchestrator/settings/...` routes are only registered on the orchestrator. A
host running as a service, with the UI connected directly to it, only exposes
its own console and cannot manage another node. `POST /settings/apply` accepts
only the host's own enrollment API key and only writes the managed layer.

### Quick tour

```bash
BASE=http://localhost:3080
TOKEN=$(curl -s -X POST $BASE/api/v1/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"username":"root","password":"***"}' | jq -r .token)
AUTH="Authorization: Bearer $TOKEN"

# Store an override for the orchestrator itself — applied live:
curl -X PUT $BASE/v1/orchestrator/settings/values/LOG_LEVEL \
  -H "$AUTH" -H 'Content-Type: application/json' -d '{"value":"warn"}'

# Where did the answer come from?
curl -s $BASE/v1/settings/values -H "$AUTH" \
  | jq '.effective[] | select(.key=="LOG_LEVEL")'
# => {"key":"LOG_LEVEL","effective":"warn","source":"orchestrator_db"}

# All hosts get canary, h1 does not:
curl -X PUT $BASE/v1/orchestrator/hosts/settings/values/ENABLE_CANARY \
  -H "$AUTH" -H 'Content-Type: application/json' -d '{"value":"true"}'
curl -X PUT $BASE/v1/orchestrator/hosts/h1/settings/values/ENABLE_CANARY \
  -H "$AUTH" -H 'Content-Type: application/json' -d '{"value":"false"}'

# What will h1 actually hold, and is it in sync?
curl -s $BASE/v1/orchestrator/hosts/h1/settings/values -H "$AUTH" | jq

# Remove overrides — the next layer takes over:
curl -X DELETE $BASE/v1/orchestrator/hosts/h1/settings/values/ENABLE_CANARY -H "$AUTH"
curl -X DELETE $BASE/v1/orchestrator/settings/values/LOG_LEVEL -H "$AUTH"
```

## Host running as a service

A host that has the UI connected to it directly, with no orchestrator involved,
gets the same console: `GET /settings/definitions`, `GET /settings/values`, and
`PUT/DELETE /settings/values/{key}` for its `local` layer. Keys that only apply
to the orchestrator (`applies_to: orchestrator`) are rejected there. If that
host later joins an orchestrator, the managed layer takes over from the local
edits, but the local edits stay in the database.
