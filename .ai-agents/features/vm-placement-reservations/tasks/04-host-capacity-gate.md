# Task 04 — Host-side CPU/memory admission gate

**Phase 3 · depends on: 02 · opt-in, default off**

The orchestrator-side reservation fixes placement, but two gaps stay open on the host itself:
creates submitted **directly to a host** bypass placement entirely, and the host's published
snapshot changes only when someone probes it. This task closes both, without breaking anyone who
oversubscribes on purpose.

## The precedent being cloned

`src/serviceprovider/parallelsdesktop/vm_capacity.go` already does exactly this shape for VM
**counts** — `reserveVmSlot` (`:89-140`) reads a live inventory, takes `vmAdmissionMu`, compares
`current + pending >= max`, records the slot, and returns an idempotent `sync.Once` release,
`409` on refusal and `503` when the inventory cannot be read. `reserveMacVMStart`
(`mac_vm_capacity.go:36-64`) is the same again for macOS slots. There is no CPU or memory
analogue: today `prlctl set <vm> --cpus N` (`main.go:2320`) and `--memsize M` (`main.go:2356`)
are applied unconditionally.

## Files

| File | Change |
|---|---|
| `src/serviceprovider/parallelsdesktop/vm_capacity.go` | `ReserveVmCapacity(ctx, operationID, cpus int64, memoryMiB float64) (func(), error)` + `capacityCharge` type |
| `src/serviceprovider/parallelsdesktop/main.go` | `vmCapacitySlots map[string]capacityCharge` field; publish `PendingCpu`/`PendingMemory` in `GetHardwareUsage` (`:2834-2917`) |
| `src/models/system_usage.go` | `PendingCpu int64`, `PendingMemory float64` on `SystemUsageResponse` |
| `src/mappers/system.go` | map them (`:8-32` region); orchestrator-side combine stays in task 02's `max()` rule |
| `src/data/models/host_resources.go` | pending fields where the orchestrator caches the published values — **not** inside the persisted `resources` blob |
| `src/controllers/machines.go` | call the gate beside the existing `ReserveVmSlot` at `:826`, `:1808`, `:1837`, `:1958` |
| `src/config/*` + `src/constants/main.go` | `VM_CPU_ADMISSION_ENABLED` (default `false`), `VM_CPU_OVERCOMMIT_RATIO` (default `1.0`) |

## Behaviour

- **Disabled by default.** `VM_CPU_ADMISSION_ENABLED=false` ⇒ no-op returning a no-op release, so
  upgrade changes nothing for existing deployments. `VM_CPU_OVERCOMMIT_RATIO` multiplies the
  capacity, giving oversubscribing users their knob back explicitly.
- **Charge what the create will actually apply**: request `Specs`
  (`models/create_catalog_virtual_machine.go:22`) → `MinimumSpecRequirements`
  (`catalog/models/virtual_machine_manifest.go:50`). Note `createCatalogMachine` (`machines.go:1958`)
  reserves **before** the catalog pull, so the specs must be resolved before the charge or
  admitted `count-only` with a log line — never blocked on a number we cannot read.
- Refusal is `409`, phrased like the existing
  `host VM limit reached: current %d, pending %d, maximum %d (mode %s)`.
- Unreadable inventory is `503` **fail-closed**, same as `vm_capacity.go:100-106`: a limit that
  cannot be measured is not silently unlimited.
- `GetHardwareUsage` publishes the in-flight charge so the orchestrator's cached view is honest
  between probes. Combined by `max(Σ reservations, host pending)`, never summed (plan D7),
  otherwise one create known to both sides is charged twice and rejects fits that work.

## Acceptance

- [ ] Gate disabled ⇒ identical behaviour to today, byte for byte in the payload except the new
      `pending_*` fields being absent/zero.
- [ ] Gate enabled, host with 4 free CPUs, two concurrent 4-CPU creates ⇒ exactly one admitted,
      the other `409` naming required/available; VM list shows one VM.
- [ ] `VM_CPU_OVERCOMMIT_RATIO=2` ⇒ both admitted.
- [ ] Memory refusal works the same way and reports MiB.
- [ ] Inventory read fails ⇒ `503`, no create (fail closed), error text matches the existing
      `cannot verify host VM capacity` posture.
- [ ] Release is idempotent: calling it twice does not under-count.
- [ ] `GET /v1/config/hardware` reports `pending_cpu`/`pending_memory` during an in-flight create
      and zeroes it after release.
- [ ] Orchestrator refuses a second create using host-published pending **alone** (no orchestrator
      reservation), proving the sequential double-book is closed independently.

## Commands

```bash
cd src && go build -buildvcs=false ./...
GOTMPDIR=$PWD/.gotmp go test -count=1 ./serviceprovider/parallelsdesktop/... ./mappers/... ./controllers/...
GOTMPDIR=$PWD/.gotmp go test -count=1 -race ./serviceprovider/parallelsdesktop/...
```
