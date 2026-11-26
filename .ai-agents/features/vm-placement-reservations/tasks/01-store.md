# Task 01 — Reservation store

**Phase 1 · depends on: nothing · blocks: 02, 03, 06**

Create the `placement` bounded context and its durable store. Ships nothing user-visible:
after this task the tables exist and are exercised only by tests.

## Files

| File | Contents |
|---|---|
| `src/data/models/placement.go` | `ResourceReservation`, `PlacementGuard`, `JobLink` |
| `src/database/migrate.go` | append the three models to `AllGormModels()` (`:19`) |
| `src/placement/interfaces.go` | `Store` + value types — the Redis seam |
| `src/placement/store_sql.go` | GORM implementation over the raw handle |
| `src/placement/service.go` | singleton + injectable constructor |
| `src/placement/sweeper.go` | lease expiry + orphan reconcile |
| `src/placement/*_test.go` | the tests listed below |

## Models

`ResourceReservation` (`resource_reservations`)

| Column | Go | Notes |
|---|---|---|
| `job_id` | `JobID string` | **primaryKey** = orchestrator job ID. One row per job |
| `host_id` | `HostID string` | `index`; updated on re-placement, never a second row |
| `owner`, `vm_name` | `Owner`, `VmName string` | composite `index` for the name guard (task 06) |
| `cpu` | `Cpu int64` | logical CPUs |
| `memory` | `Memory float64` | MiB, matches `HostResourceItem.MemorySize` |
| `vm_count`, `apple_vm` | `int64` | count charges |
| `node_id` | `NodeID string` | replica that placed it (task 05) |
| `created_at`, `updated_at`, `expires_at` | `string` | string timestamps via `helpers.GetUtcCurrentDateTime()`, matching the `settings`/`jobs` convention |

`PlacementGuard` (`placement_guards`): `HostID` primaryKey, `Seq int64`, `Token string`,
`UpdatedAt string`. `JobLink` (`job_links`): `HostJobID` primaryKey, `OrchestratorJobID`,
`HostID`, `CreatedAt` — written here, used by task 05.

Pin `TableName()` on all three and follow `src/data/models/settings.go` tag conventions
(`gorm:"primaryKey"`, `*DbRecord json:"db_record" gorm:"-:all"` if a record wrapper is used).

## Store contract

```go
type CapacityView struct {
	CPU int64; Memory float64; VmCount, AppleVm, MaxVms int64; MaxVmsMode string
}
type Reservation struct {
	JobID, HostID, VmName, Owner, NodeID string
	CPU int64; Memory float64; VmCount, AppleVm int64
}
type Admission struct { Admitted bool; Available CapacityView; Reason string } // ""|cpu|memory|vm_count|apple_vm|vm_name
type Outcome string // "succeeded" | "failed"
```

`Admit` implements the plan's D4 transaction exactly, in that statement order. `Release` is a
delete by `job_id` and must be **idempotent** (deleting an absent row is success).
`PendingByHost` returns a zero view when nothing is pending — never an error.
`ExpireLeases` / `ReconcileOrphans` are set-based deletes returning the row count.

## Rules that are not negotiable

1. **Dialect-neutral SQL only.** No `rowid`, no `ON CONFLICT`, no `pg_*`. `INSERT ... SELECT ...
   WHERE NOT EXISTS` for the guard-row ensure. Must pass on both SQLite and Postgres.
2. **Fail closed.** Any store error returns an error, and the caller treats it as a refusal. No
   "best effort, admit anyway".
3. **Out of the frozen interface.** Use `serviceprovider.GetSQLDatabase()` the way
   `settings/store.go:5` does. **No method is added to `datainterfaces.DatabaseService`** (#421 R-1).
   When `GetSQLDatabase()` is unavailable (legacy JSON mode) the store is nil and
   `service.Get()` reports unavailable — one boot warning, placement stays disabled.
4. **No HTTP calls** anywhere inside `Admit` (plan D10).
5. Every file under 500 lines.

## Service

Mirror `settings/main.go`: `InitService(ctx)` builds once from the raw handle, `Get()` returns
the singleton, `New(store Store)` exists for tests, and `SetStore(Store)` allows swapping in a
fake or a future Redis implementation. `Reserve(ctx, hostID, res, extra)` and
`ReleaseJob(ctx, jobID, outcome)` are the only calls placement makes.

## Sweeper

Ticker on the `constants.GhostJobCheckIntervalSeconds` cadence (`constants/jobs.go:45`),
started only on orchestrator nodes, following the `StartStaleJobWatcher` /
`StopStaleJobWatcher` shape in `database/jobs_watcher.go`. Cadence from
`PLACEMENT_RECONCILE_INTERVAL_SECONDS` (default 30 — the sweeper picks up what the terminal hook
already handled, so it needs neither the 5-second urgency of the ghost-job watcher nor the cost of
running at it). Two sweeps racing must be harmless.

## Acceptance

- [ ] `AllGormModels()` lists the three models; `RunAutoMigrate` creates them on both engines.
- [ ] Concurrent `Admit` burst keeps `Σ cpu ≤ capacity` (mirrors
      `TestReserveVmSlotConcurrentBurstRespectsLimit` in `vm_capacity_test.go`).
- [ ] Two creates that both fit land on the **same** host; the first that does not fit is refused
      with `Reason: "cpu"` while a second host accepts it.
- [ ] `Release` twice = no error; `PendingByHost` on an idle host = zero view, no error.
- [ ] A row past `expires_at` disappears on the sweep; a row whose job is `completed`/`failed`/
      missing disappears on reconcile; a row whose job is `running` survives.
- [ ] With the store nil, `Reserve` returns an error and the caller refuses (fail closed).
- [ ] Parity scenarios for the three tables run under `RunOnEngines`; the legacy exclusion is
      recorded in `src/database/parity/doc.go`.

## Commands

```bash
cd src && go build -buildvcs=false ./...
mkdir -p .gotmp && GOTMPDIR=$PWD/.gotmp go test -count=1 ./placement/... ./database/...
GOTMPDIR=$PWD/.gotmp go test -count=1 -race ./placement/...
POSTGRES_TEST_DSN='postgres://postgres:postgres@127.0.0.1:5432/prl_test?sslmode=disable' \
  GOTMPDIR=$PWD/.gotmp go test -tags integration -count=1 ./database/...
gofmt -l placement data/models database
```
