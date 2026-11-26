# Task 02 — Admission wiring in placement

**Phase 2 · depends on: 01, 06 · this is where the reported bug dies**

Turn the store into enforcement: placement subtracts pending when it judges a host, and claims
capacity when it picks one.

## Files

| File | Change |
|---|---|
| `src/orchestrator/check_host_requirements.go` | `checkHostRequirements` gains a `pending placement.CapacityView` argument; CPU (`:104-116`) and memory (`:118-127`) subtract it; new `vm_name` check (task 06) |
| `src/orchestrator/validate_host.go` | thread `pending` through, unchanged decision semantics |
| `src/orchestrator/create_host_virtual_machine.go` | admission inside the per-host loop at `:217-230`; drop candidates the store refuses |
| `src/orchestrator/get_valid_hosts_for_create.go` | no signature churn needed — the loop above is the only caller |
| `src/models/orchestrator_host_resource.go` | `TotalPending HostResourceItem \`json:"total_pending"\`` on `HostResourceOverviewResponse` |
| `src/controllers/orchestrator.go` | populate it at response assembly (`:1064-1090`) from `placement.Get().PendingByHost` |

## The two rules this task exists to encode

**1. Judge with pending subtracted.**

```
available = host.Resources.TotalAvailable  −  max(Σ live reservations, host-reported pending)
```

The `max` (plan D7), never a sum — once the host publishes `PendingCpu`/`PendingMemory` (task 04)
a create is known to both sides and must be charged once. Keep the message wording exactly as
the issue expects: `CPU: required 6 logical CPUs, available 3`, and follow the existing
`VM limit: current %d, pending %d, maximum %d` style by adding `pending %d` to the CPU/memory
messages. `isBlockingRequirementCheck` (`:403`) semantics are unchanged.

**2. Claim only on the host you pick.** Inside the candidate loop, after `validateHost` passes:

```go
admission, err := placement.Get().Reserve(ctx, host.ID, res, extra)
if err != nil || !admission.Admitted {   // fail closed: an error is a refusal
    rejections = append(rejections, ...)
    continue                              // next candidate, exactly as a validation failure does today
}
```

First admitted host wins and is dispatched. All refused ⇒ the existing aggregate `400`
(`:232-242`) still fires, now carrying the real reasons.

**Charge the same numbers the create will actually apply** — resolve `specs` once
(`getSpecsFromRequest`, `:264`) and charge from it, so the placement charge and the host-side
gate (task 04) can never disagree.

## Hard constraints

- `validateHost` / `checkHostRequirements` stay **pure reads**. The pre-flight endpoint
  (`controllers/orchestrator_check_machine.go:21`) must report pending-reduced availability
  without claiming a single row — that is how QA distinguishes "would fit" from "is mine".
- All four entry points reserve: `CreateVirtualMachine`, `DispatchCreateVirtualMachine`,
  `CreateHosVirtualMachine` (`:117-141`), `DispatchCreateHosVirtualMachine` (`:143-185`). The
  explicit-host ones bypass selection but must still count and still honour the name guard.
- Replace the blocking `s.Refresh()` calls on failure paths (`:302`, `:163`) with targeted
  `s.hwQueue.Enqueue(hostID)` — under a burst, and with N replicas, a full cluster reload is a
  thundering herd at exactly the wrong moment. `interfaces.HardwareEnqueuer`
  (`orchestrator/handlers/pdfm_event_handler.go:19,45`) is the seam precedent.
- No new `DatabaseService` method; no network I/O inside the store transaction (plan D10).

## Acceptance

- [ ] Issue repro as a unit test: 10 CPUs, 1 system-reserved, two 6-CPU requests → one admitted,
      one `400` whose message contains `CPU: required 6 logical CPUs, available 3`.
- [ ] Sharing: two 2-CPU requests on a 10-free-CPU host both admitted **on that host**, Σ reserved 4.
- [ ] A 9-CPU request after those two is refused that host and accepted by the next candidate.
- [ ] `POST /v1/orchestrator/machines/check` ×10 leaves `resource_reservations` untouched.
- [ ] Explicit-host async create with `cpu = available + 1` is refused (the bypass path counts).
- [ ] `GET /api/v1/orchestrator/hosts/{id}` shows `total_pending` while a create is in flight and
      zeroes after release.
- [ ] Existing suites (`orchestrator`, `tests/orchestrator`, `tests/controllers`) stay green.

## Commands

```bash
cd src && go build -buildvcs=false ./...
GOTMPDIR=$PWD/.gotmp go test -count=1 ./orchestrator/... ./tests/orchestrator/... ./controllers/...
GOTMPDIR=$PWD/.gotmp go test -count=1 -race ./orchestrator/...
```
