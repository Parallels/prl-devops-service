# Task 03 — Release on job terminal state

**Phase 2 · depends on: 01, 02**

A claim that is never released starves the cluster, so release must be impossible to miss.
This task wires the one choke point and the backstops.

## Why not the call site

`DispatchCreateVirtualMachine` returns `nil, nil` on success
(`orchestrator/create_host_virtual_machine.go:100-104`) — there is no return value to hang
release on, and the remote path completes minutes later over a WebSocket on a **different**
replica. Terminal job state is the only universal signal.

## Files

| File | Change |
|---|---|
| `src/jobs/main.go` | `SetOnJobTerminal(func(ctx, jobID, outcome))` + invocation from `MarkJobComplete` (`:277`), `MarkJobCompleteWithRecord` (`:307`), `MarkJobError` (`:341`) |
| `src/jobs/main.go` | `SetRefreshHostHook(func(ctx, hostID) error)` — injected, never imported |
| `src/startup/main.go` (or the orchestrator start path) | register both hooks once, after `placement.InitService` |
| `src/placement/service.go` | `ReleaseJob(ctx, jobID, outcome)` implementing the ordering below |

## Required ordering (plan D8)

```
outcome == succeeded:
    best-effort, timeout-bounded: refreshHost(hostID)   // persist real TotalInUse FIRST
    delete reservation
outcome == failed:
    delete reservation                                   // always, even if refresh fails
    best-effort: enqueue refresh
```

Deleting before the refresh would open an instant where neither the reservation nor
`TotalInUse` accounts for the capacity — the same bug relocated. An unreachable host must never
strand capacity, so the delete happens unconditionally and the refresh is best-effort.

## Import rule

`jobs` must **not** import `orchestrator`. Both behaviours are injected as function values;
`jobs` only knows it has optional hooks and calls them when registered.

## Backstop, because hooks are not enough

`src/placement/sweeper.go` (from task 01) must cover what a hook cannot:

- process killed mid-create ⇒ the lease (`expires_at`) frees it;
- a job transitioned by DB-side `DetectStaleJobs` (`database/jobs.go:249-283`) or
  `RecoverOngoingJobs` (`:221-239`) which bypasses the jobs manager ⇒ orphan GC frees it, since
  those rows are terminal or missing by the next tick.

Every sweep is a set-based delete, so two replicas sweeping concurrently is harmless.

## Also fix while here: the silent drop

`HostJobEventHandler` (`orchestrator/handlers/host_job_event_handler.go:85-90`) forwards raw and
**returns** when the in-memory `JobLinkRegistry` has no link. Today that is only a lost UI event;
with reservations it is an orchestrator job that never reaches terminal state, so its capacity
hangs until the lease expires. Task 05 makes links durable; the reconciler covers the gap in the
meantime. Add the regression test here either way.

## Acceptance

- [ ] Success: job `completed` ⇒ host snapshot refreshed **and** reservation gone; assert the
      order by making the refresh observable (record the call before the delete).
- [ ] Failure: job `failed` ⇒ reservation gone within one tick, `resource_reservations` empty.
- [ ] Refresh returns an error / times out ⇒ reservation is **still** deleted.
- [ ] `MarkJobError` twice, or release of an unknown job ⇒ no error, no panic (idempotent).
- [ ] Host event arriving with **no in-memory link** ⇒ capacity still returns (lease or reconcile).
- [ ] Stale-job timeout path (`GHOST_JOB_TIMEOUT_MINUTES`) frees the reservation too.
- [ ] `jobs` still has no dependency on `orchestrator` — `go list -deps` check or build failure test.

## Commands

```bash
cd src && go build -buildvcs=false ./...
GOTMPDIR=$PWD/.gotmp go test -count=1 ./jobs/... ./placement/... ./orchestrator/... ./tests/orchestrator/...
GOTMPDIR=$PWD/.gotmp go test -count=1 -race ./placement/... ./jobs/...
```
