# Implementation Plan — VM placement reservations

Companion to [brief.md](brief.md). This file is the **design lock-in** artifact: the decisions
here are settled, with the rejected alternatives recorded so they are not relitigated
mid-implementation. Task cards live in [tasks/](tasks/).

| Field | Value |
|---|---|
| Feature folder | `.ai-agents/features/vm-placement-reservations/` |
| Design status | **Locked** 2026-11-25 |
| Build / test | `cd src && go build -buildvcs=false ./...` · `GOTMPDIR=$PWD/.gotmp go test -count=1 ./...` |

---

## D1 — Reject immediately, no waiting queue

**Decision.** When no host can fit the request, return the existing
`400 Insufficient host resources` aggregate immediately.

**Why.** It is the behaviour the issue reporter expects, it is already the behaviour of
`getValidHostsForCreate` (`create_host_virtual_machine.go:232-242`), and it keeps the create
API synchronous-decisive, which the async job flow depends on.

**Rejected — "queue the request and retry when capacity appears".** A waiting state needs a
durable queue, a requeue trigger, cancellation semantics, and a job state that is neither
pending nor failed. None of that exists. It also hides the real signal from the user: they asked
for a VM that does not fit and were told nothing.

## D2 — Per-host atomic admission, not one shared serial queue

**Decision.** Serialise the check-and-claim **per host**. Different hosts proceed concurrently.

**Why.** Correctness only requires that two claims for the *same* host cannot interleave. The
original design draft serialised every create cluster-wide, which would serialise all creates
behind a placement that itself performs a live disk HTTP probe per host
(`check_host_requirements.go:129-163`) and a 2s-timeout latency ping per host
(`create_host_virtual_machine.go:309-326`). Four hosts would mean seconds of global serial time
per create for no additional safety.

The user's accepted cost model ("checks run in sequence") is preserved where it matters: checks
for a given host are serialised, and a refusal on host x still moves to the next host.

## D3 — A reservation is a charge, not a host lock

**Decision.** Concurrent creates keep landing on the same host while it fits. Two 2-CPU creates
on a host with 10 free CPUs both go to that host.

**Why.** The per-host lock exists only for the microseconds of the check-and-claim, and a
"host busy" flag would turn placement into round-robin, destroying the cache-locality and
latency ordering that `filterAndSortHosts` (`create_host_virtual_machine.go:334-408`) works to
achieve. A host hop must mean *the arithmetic failed*, nothing else.

## D4 — The transaction protocol (portable across SQLite and Postgres)

**Decision.** Inside one transaction, in this order:

```sql
1  INSERT INTO placement_guards (host_id, seq, token, updated_at)
   SELECT ?, 0, ?, ? WHERE NOT EXISTS (SELECT 1 FROM placement_guards WHERE host_id = ?);
2  UPDATE placement_guards SET seq = seq + 1, token = ?, updated_at = ? WHERE host_id = ?;  -- the lock
3  SELECT resources                     FROM orchestrator_hosts    WHERE id = ?;
4  SELECT COALESCE(SUM(cpu),0), COALESCE(SUM(memory),0), COUNT(*), COALESCE(SUM(apple_vm),0)
   FROM resource_reservations WHERE host_id = ?;
5  SELECT 1 FROM resource_reservations WHERE owner = ? AND vm_name = ? AND job_id <> ?;
   SELECT 1 FROM virtual_machines      WHERE user = ? AND name = ?;                          -- name guard
6  INSERT INTO resource_reservations (...) VALUES (...);                                     -- or rollback
```

**Why this shape.**

- Step 2 is the lock. On Postgres the `UPDATE` takes a row-exclusive lock held to `COMMIT`, so
  a second transaction blocks at step 2 and then reads the first transaction's committed rows at
  step 4. On SQLite the single-writer transaction lock covers the whole block.
- **No `pg_advisory_xact_lock`** — it is Postgres-only and would fork the SQL. The repo's rule
  for this layer is dialect-neutral statements (see `src/settings/store.go` and the #421 rules).
- **No `ON CONFLICT`**, no `rowid` — both dialect-specific.
- All capacity arithmetic stays in Go: `orchestrator_hosts.resources` is a JSON column and
  cannot be aggregated in portable SQL.
- Step 5's `virtual_machines` read uses the existing indexed columns `user` and `name`
  (`data/models/virtual_machine.go:17,23`).

**Rejected — optimistic row version / retry loop.** More code, livelock risk under exactly the
burst this feature exists for, and it still needs the same guard row to be fair.

**Rejected — serialising on a single global guard row.** Correct but globally serial, see D2.

## D5 — Reservations live in their own table

**Decision.** New `resource_reservations` table. Never inside `host.Resources`.

**Why.** `host.Resources` is host-owned: `updateHostWithHardwareInfo` (`main.go:512-527`) and
`UpdateOrchestratorHostResources` overwrite it wholesale on every probe. A reservation stored
there would be erased by the next refresh — which is precisely the timing that produces the bug.

**Primary key = orchestrator job ID.** One row per job, so re-planning from host A to host B
updates `host_id` instead of double-charging, and a job can never leak a second row. Rows are
hard-deleted on release, so "rows present = pending demand" and the aggregate is a single `SUM`
with no status column to reason about. Mirrors the `DeleteExpiredEnrollmentTokens` posture.

**Rejected — a `status` column with soft release.** Adds a state machine and a GC problem to a
table whose entire job is "outstanding demand".

## D6 — Store behind an interface, Postgres/SQLite now, Redis later

**Decision.** `placement.Store` is the only seam. `store_sql.go` is the first implementation.

**Why.** The user's explicit requirement: the queue must be an interface so a Redis backend can
be merged later without redesigning the feature. Callers depend on `Store`, never on GORM.

```go
type Store interface {
    Admit(ctx basecontext.ApiContext, hostID string, res Reservation, extra CapacityView) (Admission, error)
    Release(ctx basecontext.ApiContext, jobID string, outcome Outcome) error
    PendingByHost(ctx basecontext.ApiContext, hostID string) (CapacityView, error)
    PendingByHosts(ctx basecontext.ApiContext) (map[string]CapacityView, error)
    NameTaken(ctx basecontext.ApiContext, owner, vmName, exceptJobID string) (bool, error)
    ExpireLeases(ctx basecontext.ApiContext, now time.Time) (int, error)
    ReconcileOrphans(ctx basecontext.ApiContext) (int, error)
}
```

Deliberately **not** on `datainterfaces.DatabaseService` — that surface is frozen at 158 methods
(#421 R-1). The store talks to `serviceprovider.GetSQLDatabase()` directly, exactly like the
`settings` domain (`settings/store.go:5`).

**Fail closed.** Any store error is a refusal, mirroring `reserveVmSlot`'s posture
(`vm_capacity.go:89-140`) and `reserveMacVMStart`'s 503 when the inventory is unreadable:
a limit that cannot be measured is not silently unlimited.

**Legacy JSON DB mode.** `GetSQLDatabase()` is unavailable → store is nil, one boot warning,
placement behaves as today. Same degradation shape as `settings.ErrStoreUnavailable`. No JSON
fallback store.

## D7 — Pending is combined with `max`, never summed

**Decision.** `effective_pending = max(Σ orchestrator reservations, host-reported pending)` per
dimension.

**Why.** Once the host publishes `PendingCpu`/`PendingMemory` (D9), a single create is known to
both sides. Summing would charge it twice and reject fits that actually work; taking the max
charges it once and still covers host-side work the orchestrator never placed (a create started
directly against the host).

## D8 — Release hangs off job terminal state, not off the dispatch call

**Decision.** One hook on the jobs manager, `SetOnJobTerminal(func(ctx, jobID, outcome))`,
registered at startup; it calls `placement.ReleaseJob`.

**Why.** `DispatchCreateVirtualMachine` returns `nil, nil` on success
(`create_host_virtual_machine.go:100-104`) — there is no return value to hang release on.
Terminal state is written only by `jobs.MarkJobComplete` (`jobs/main.go:277`),
`MarkJobCompleteWithRecord` (`:307`), `MarkJobError` (`:341`), plus the DB-side
`DetectStaleJobs` (`database/jobs.go:249-283`) and `RecoverOngoingJobs` (`:221-239`). Hooking
the choke point covers sync create, async dispatch, the WebSocket path and the stale-job path
with one registration.

**Import rule.** `jobs` must not import `orchestrator`. The "refresh the host before releasing
on success" step is injected as a second settable hook, so the dependency stays one-way.

**Ordering that matters.** Success ⇒ `UpdateHostResourcesForEvent(host)` first (bounded by a
timeout, best-effort), **then** delete the row. Releasing first would open an instant where
neither the reservation nor `TotalInUse` accounts for the capacity — the same bug in a new place.
Failure ⇒ delete immediately, refresh best-effort afterwards. Release always happens, even if
the refresh fails: an unreachable host must never strand capacity.

**Backstop.** A periodic reconciler deletes rows whose job is terminal or missing, and rows past
their lease. Crash-safe by construction, and every sweep is a set-based delete so two replicas
sweeping at once is harmless.

## D9 — Host-side CPU/memory gate, opt-in

**Decision.** `ReserveVmCapacity(ctx, operationID, cpus, memoryMiB) (func(), error)` in
`serviceprovider/parallelsdesktop/vm_capacity.go`, cloned from `reserveVmSlot`, plus
`PendingCpu`/`PendingMemory` on `SystemUsageResponse`.

**Why.** Without it, the sequential double-book (brief §3) survives whenever the orchestrator's
view is stale, and creates submitted straight to a host bypass placement entirely.

**Off by default** (`VM_CPU_ADMISSION_ENABLED=false`): deployments that oversubscribe on
purpose must not break on upgrade. `VM_CPU_OVERCOMMIT_RATIO` gives them the knob back.

Charge exactly what the create provisions: request `Specs`
(`models/create_catalog_virtual_machine.go:22`) → `MinimumSpecRequirements`
(`catalog/models/virtual_machine_manifest.go:50`). If neither resolves, admit `count-only` and
log — matching the existing "older host without the endpoint stays eligible" posture rather
than blocking creates on a spec we cannot read.

## D10 — No network I/O inside the admission transaction

**Decision.** Hard constraint, pinned by test.

**Why.** The guard row is held to `COMMIT`. A slow host's disk probe or latency ping inside the
transaction would make every create for that host queue behind a network round trip — turning a
correctness fix into a throughput incident.

## D11 — VM name uniqueness, per owner, cluster-wide

**Decision.** The admission transaction refuses a create whose `vm_name` is already pending for
that owner on **any** host, or already exists as a VM of that owner. New reported check
`vm_name` in `checkHostRequirements` so pre-flight shows it too.

**Why.** There is no name guard in the placement path today — `checkNameCpu`, `checkNameMemory`
etc. (`check_host_requirements.go:20-27`) are check *labels*, not name checks. The only name
checks are check-then-act on the host: `machine %v already exists` in the catalog pull
(`catalog/pull.go:338-353`) and `VM with name %s already exists` in `RegisterVm`
(`main.go:1551-1557`). Two concurrent same-name creates today produce either a mid-create
failure after a full pack download, or — if they land on different hosts — two identical names
with nothing detecting it.

Same-job re-placement is exempt (`job_id <> ?`) so host hopping does not collide with itself.
Comparison is exact on `strings.TrimSpace`ed input, aligned with the host's own comparison
during implementation.

## D12 — Multi-replica is in scope

Because the store is per-cluster state, these pre-existing single-node assumptions become bugs
the moment `replicaCount > 1`, and are fixed here rather than left behind:

| Gap | Evidence | Fix |
|---|---|---|
| No replica identity exists | `os.Hostname` appears nowhere in `src` | `helpers.NodeID()`: `ORCHESTRATOR_NODE_ID` → machine unique id (the telemetry `hardware_id`; installed by startup in orchestrator mode) → hostname → generated id |
| Booting any replica fails **every** running job cluster-wide | `database/jobs.go:221-239` | `jobs.owner_node` + `RecoverOngoingJobsForNode(ctx, nodeID)`, reached by a local interface assertion (the `StartStaleJobWatcher` precedent) |
| Host job events untranslatable on the replica that did not dispatch | `JobLinkRegistry` is in-process memory; `host_job_event_handler.go:85-90` silently forwards raw and returns | Persist `job_links`; registry falls back to the table on a miss |
| ~30 `Order("rowid")` sites break on Postgres | `database/testutil/engines.go:115-130` T17 probe skips those legs | `database/roworder.go` dialect helper, then un-skip the probe |
| N replicas duplicate refresh loops and clobber host rows | `persistHost` read-modify-write | Claim ticks with the same guard-row idiom; targeted `UpdateOrchestratorHostResources` |
| Blocking `s.Refresh()` on failure paths | `create_host_virtual_machine.go:302`, `:163` | Replace with targeted `hwQueue.Enqueue(hostID)` |

---

## Package layout

```
src/placement/
  interfaces.go   Store, CapacityView, Reservation, Admission, Outcome, VMCPUConfig + the seam
  store_sql.go    GORM Store. Dialect-neutral SQL only. Fail-closed on any error.
  service.go      InitService(ctx)/Get() singleton over serviceprovider.GetSQLDatabase(),
                  New(store) for tests, Reserve / ReleaseJob / PendingFor
  sweeper.go      lease expiry + orphan reconcile, orchestrator-only ticker
  *_test.go
src/data/models/placement.go   ResourceReservation / PlacementGuard / JobLink (+ AllGormModels)
```

Every file stays under 500 lines (repo rule).

## Phase order

| Phase | Task card | Deliverable |
|---|---|---|
| 0 | this folder | design locked |
| 1 | [tasks/01-store.md](tasks/01-store.md) | models, `Store`, SQL store, service, sweeper, migration registry |
| 2 | [tasks/02-admission-wiring.md](tasks/02-admission-wiring.md) | pending-aware checks, admission hook, observability |
| 2 | [tasks/03-release-hooks.md](tasks/03-release-hooks.md) | jobs terminal hook, refresh-then-release, reconciler wiring |
| 3 | [tasks/04-host-capacity-gate.md](tasks/04-host-capacity-gate.md) | `ReserveVmCapacity`, pending cpu/mem publication, opt-in config |
| 4 | [tasks/05-multi-replica.md](tasks/05-multi-replica.md) | node identity, node-scoped recovery, durable links, rowid, poll claiming |
| 1 | [tasks/06-name-uniqueness.md](tasks/06-name-uniqueness.md) | cluster-wide name guard + `vm_name` check |

Phase 1 is independently mergeable and ships nothing user-visible; phases 2 and 3 are where the
bug dies. Anything that can be split lands split.

## Test plan

| Layer | Test |
|---|---|
| Store unit | exact-fit admitted / 1-CPU-over refused; concurrent burst keeps Σ ≤ capacity (shape of `TestReserveVmSlotConcurrentBurstRespectsLimit`); sharing: two fits land on one host, a third that does not fit hops; lease expiry; orphan GC; release idempotency; store-unavailable fails closed |
| Transaction purity | counting `http.RoundTripper` injected; `Admit` issues zero HTTP calls |
| Orchestrator | issue repro verbatim in `check_host_requirements_test.go` / `resource_failure_test.go`, asserting `CPU: required 6 logical CPUs, available 3` |
| Race | two goroutines into `DispatchCreateVirtualMachine` on one in-memory DB (shape of `data/concurrency_test.go`), `-race` |
| Parity | new tables over `RunOnEngines` in `src/database/parity/`; legacy excluded in the `parity/doc.go` ledger |
| Postgres | `POSTGRES_TEST_DSN` leg; T17 probe un-skipped and asserted |
| Host gate | burst, overcommit ratio, disabled no-op, fail-closed |
| Event path | completion and failure release, **including when the in-memory link is absent** |

Gates before "done": `go build -buildvcs=false ./...`, `go test -count=1 ./...`,
`go test -count=1 -race ./...`, `POSTGRES_TEST_DSN=... go test -tags integration ./database/...`,
`gofmt -l` clean, `go vet -tags release ./...`.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| Leaked reservations starve the cluster | lease TTL + orphan reconciler + the invariant "release always happens, refresh failure included" |
| Throughput regression from the lock | lock held only across DB statements (D10, test-pinned); per-host key so hosts do not contend |
| Oversubscription breaks for existing users | host gate off by default; `PLACEMENT_RESERVATION_ENABLED` is a one-variable rollback |
| `Order("rowid")` changes documented read ordering | land only with parity green on both engines; note in `VERIFICATION.md` |
| Frozen data interface violated by mistake | store is package-private over the raw handle, as `settings` does; no `DatabaseService` method added anywhere |
| Double counting after the host reports pending | `max()` rule (D7), unit-tested |

## Explicitly deferred

Disk reservations · orchestrator-side clone placement · general `persistHost` atomicity ·
Redis `Store` · spreading VMs across hosts for HA · queueing/waiting for capacity.
