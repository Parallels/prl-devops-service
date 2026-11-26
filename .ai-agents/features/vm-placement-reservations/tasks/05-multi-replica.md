# Task 05 — Multi-replica readiness

**Phase 4 · depends on: 01-03**

The reservation table is per-cluster state, which turns several long-standing single-node
assumptions into real bugs the moment `replicaCount > 1`. They are fixed here, not deferred,
because a reservation that leaks on a second replica is worse than no reservation.
Multi-replica requires `DB_TYPE=postgres` (`config/database.go:73-108`); SQLite stays single-node.

## 5.1 Node identity

Nothing identifies a replica today — `os.Hostname` appears nowhere in `src`.

`src/helpers/node_id.go`: `NodeID()` = `ORCHESTRATOR_NODE_ID` → machine unique id → hostname →
generated id, computed once, logged at boot, stamped on reservation rows and jobs. The machine id
arrives through a resolver installed by startup in orchestrator mode (`SetNodeIDSource`) because
`serviceprovider` imports `helpers` and a lower layer cannot reach up for its own identity; it is
the same value telemetry sends as `hardware_id`. The generated id is random hex, not a ULID — the
claim this card originally made about it is corrected in 5.4.

## 5.2 Node-scoped job recovery

`GormDatabase.RecoverOngoingJobs` (`src/database/jobs.go:221-239`) unconditionally fails **every**
`running`/`pending` job in the table. With two replicas, a rolling restart or a scale-up kills
the other replica's in-flight creates — and strands their reservations behind a lease wait.

- Add indexed `OwnerNode string` (`owner_node`) to `data/models/job.go`.
- Stamp it in `JobManagerService.CreateNewJob` (`jobs/main.go:82`) and `CreateOrchestratorJob` (`:101`).
- Add `RecoverOngoingJobsForNode(ctx, nodeIDs []string)` scoped to `owner_node = ?` for any of the
  caller's identities, reached by startup through a **local interface assertion** — the
  `StartStaleJobWatcher` precedent (`database/jobs_watcher.go`). **No method added to
  `datainterfaces.DatabaseService`** (#421 R-1). An empty list recovers nothing rather than
  sweeping the cluster.
- Legacy JSON DB keeps the current global behaviour; record the exclusion in
  `src/database/parity/doc.go`.
- A job with an empty `owner_node` (row written before the field) stays recoverable by whoever
  picks it up. And a row carrying an identity this machine has since stopped using is still its
  own: `helpers.NodeIDAliases()` hands recovery the current id plus the hostname, so switching
  identity from hostname to machine id — 5.1 — does not leave a replica's own earlier rows
  running with nobody to finish them. They would otherwise sit until the stale-job watcher reaped
  them, which is a silent regression, not an edge case.

## 5.3 Durable job links

`JobLinkRegistry` (`orchestrator/registry/job_link_registry.go:35-52`) is an in-process map while
**every** replica dials **every** host's WebSocket. So the replica that receives a host's
`JOB_COMPLETED` may not own the link: `host_job_event_handler.go:85-90` forwards raw and
**returns**, the orchestrator job never reaches terminal state, and its reservation hangs until
the lease expires.

- Persist links in `job_links` (model from task 01), written at dispatch
  (`create_host_virtual_machine.go:95`, `:176`).
- `Lookup` falls back to the table on an in-memory miss; `Remove` deletes from both.
- Keeps the map: the durable table is a fallback, not a replacement, so no hot-path latency change.

## 5.4 `rowid` portability — **done, and it was not deferrable**

`Order("rowid")` appeared in ~30 sites (`database/{jobs,users,roles,claims,catalog,orchestrator,
snapshots,configuration,packer_template,reverse_proxy,catalog_managers,orchestrator_vm,
enrollment_token,schema,api_key,user_config}.go`) and Postgres has no `rowid`.

This card deferred it on the reasoning that SQLite — the supported backend — was unaffected.
That reasoning was wrong in one important way: the sites are not only list-order cosmetics, they
are on the boot path. A fresh Postgres install reached `SELECT * FROM "configurations" ORDER BY
rowid LIMIT 1` inside `SeedDefaultConfig` and exited 1. A defect that prevents the engine from
starting is not deferrable behind "we do not support Postgres yet" while someone is installing
Postgres.

`src/database/roworder.go`: one dialect-aware helper, `insertionOrder(db, table)`, routed through
`dialectOf` so the repo keeps a single place that asks the driver what it is. `rowid` on SQLite —
byte-identical to before, which the parity suite re-checks — and the qualified primary key
(`<table>.id`, `schema_metadata.version`) on Postgres. Applied at every site, including the
`Preload` closures, which order the *related* table rather than the parent's.

**What the card got wrong, and what the fix therefore is not:** it claimed "ULID ids sort by
creation time, so `ORDER BY id` preserves insertion order". They do not — `helpers.GenerateId()`
is 32 random hex bytes. So the Postgres leg is a *stable, deterministic* order, not insertion
order:

- list position relative to SQLite changes (same rows, different order);
- a role/claim rename rewrites the id, which on SQLite leaves the rowid and so the position alone
  — on Postgres the row moves. Preserving that needs a real insertion-sequence column, a schema
  change deliberately not smuggled in here;
- the singleton tables are unaffected in practice (one row by invariant).

Un-skipping the T17 probe needed no test change: it already skips only when the probe error
mentions `rowid`, so it resumes on its own now that the queries are portable. Its legs then assert
the ordering contracts against a live Postgres — which is the check this environment cannot run,
and the reason the bullet below stays open.

## 5.5 Stop N replicas hammering the same hosts

Each replica runs `runFullRefreshLoop`, the VM refetch loop
(`orchestrator/vm_refetch_loop.go`) and the hardware queue, all doing read-modify-write on the
same host rows via `persistHost` (whole-record overwrite).

- Claim the tick with the same guard-row idiom: `UPDATE poll_guards SET next_attempt_at = ?
  WHERE host_id = ? AND next_attempt_at <= now` — one winner, everyone else skips.
- Use the targeted `UpdateOrchestratorHostResources` in those loops instead of whole-record
  `persistHost`.
- The cheap health-ping loop may stay duplicated; note it, do not gold-plate it.

## Acceptance

Implemented here: 5.1, 5.2, 5.3, 5.4. Deferred: 5.5 — see the note at the bottom.

- [x] Booting replica B does **not** fail replica A's running jobs; `jobs.owner_node` shows `qa-A`
      on A's rows and B logs that it recovered zero foreign jobs.
      (`TestSQLJobs_RecoverByNode` — node-scoped, ownerless rows included, an empty node id
      recovers nothing rather than sweeping the cluster.)
- [x] With B killed mid-create, A still receives the host event and releases (durable link).
      (`job_links` written at dispatch; `Lookup` falls back to the table and caches.)
- [x] A host event arriving at the replica that did not dispatch still translates and releases.
      (Same mechanism: the link, not the process, decides which orchestrator job is told.)
- [ ] The T17 probe runs un-skipped on Postgres and passes. — **needs a live Postgres**; the probe
      itself is un-blocked (it skips only on a `rowid` error, which no query produces now)
- [ ] `POSTGRES_TEST_DSN=... go test -tags integration ./database/...` green. — **CI only**
      (no Postgres in this environment; `placement/store_postgres_test.go` skips without the DSN)
- [ ] Two replicas against one host set produce one refresh per interval, not two (log evidence).
      — **deferred with 5.5**
- [x] Legacy `DB_TYPE=legacy` boots, warns once about placement, and keeps working.
      (Store nil ⇒ placement disabled; job links stay in-process with one warning; startup falls
      back to cluster-wide `RecoverOngoingJobs`.)

## Deferred: 5.5

**5.5 poll-tick claiming** duplicates read-modify-write refreshes between replicas. It wastes work
and can briefly regress a host row, but it cannot over-commit CPU: admission is guarded by the
per-host `placement_guards` row, which is what the bug report is about. Land it with a Postgres
harness, where the "one refresh per interval" acceptance line can actually be observed.

## Found while booting Postgres: duplicate rows in one batch

Not a multi-replica issue, but it surfaced on the same empty-database run and it belongs on this
record because it is the same class of defect — a Postgres constraint exposing a SQLite-only
assumption the feature had inherited:

```
legacy migration failed in VirtualMachines: ERROR: ON CONFLICT DO UPDATE command cannot
affect row a second time (SQLSTATE 21000)
```

`vmRows` flattens every host's embedded VM slice, so a VM stored under two hosts is handed to one
`CreateInBatches` twice. SQLite's `ON CONFLICT DO UPDATE` silently let the later copy overwrite the
earlier one — including its `host_id`, so which host a VM ended up belonging to depended on file
order. Postgres refused the statement, which aborted the migration and rolled the install back.

Fixed in `upsertAll` as a backstop (`dedupeByPK`, first-seen wins, the rule the typed `rows()`
functions already follow) rather than only in `vmRows`, because every collection sharing that
writer carried the same latent failure and only one of them had ever been observed doing it. The
Postgres leg is not verifiable here; it rests on the unit test that no batch can contain a
duplicate key, which is precisely the condition Postgres rejects.

## Commands

```bash
cd src && go build -buildvcs=false ./...
GOTMPDIR=$PWD/.gotmp go test -count=1 ./database/... ./jobs/... ./orchestrator/...
GOTMPDIR=$PWD/.gotmp POSTGRES_TEST_DSN='postgres://postgres:postgres@127.0.0.1:5432/prl_test?sslmode=disable' \
  go test -tags integration -count=1 ./database/...
```
