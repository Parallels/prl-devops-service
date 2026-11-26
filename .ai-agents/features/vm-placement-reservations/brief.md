# Feature Brief — Stop the orchestrator overcommitting host CPU/memory on concurrent VM creation

| Field | Value |
|---|---|
| Issue | not opened yet — assign a PBI number at triage |
| Feature folder | `.ai-agents/features/vm-placement-reservations/` |
| Status | Design locked (this folder). Implementation not started. |
| Audience | Back-end, `Parallels/prl-devops-service`, sources under `src/` |
| Depends on | PBI #421 GORM data layer (Postgres support), `MAX_VMS_PER_HOST` host admission pattern |
| Last verified against code | 2026-11-25, branch `feature/host-concurrent-resources-queue` @ `418ec65d` |

---

## 1. The bug

Two simultaneous create requests for the same host are both accepted, and the host ends up
with more virtual CPUs than it has.

Reported reproduction: a host with 10 logical CPUs and 1 system-reserved (9 usable), no VMs.
Two concurrent creates of 6 CPUs each. **Both succeed.** `GET /api/v1/orchestrator/hosts/{id}`
then reports `total_in_use.logical_cpu_count = 12` and `total_available.logical_cpu_count = -3`.

Expected: the second request is refused with `Insufficient host resources` and the detail
`CPU: required 6 logical CPUs, available 3`.

## 2. Root cause

There is exactly one "does it fit?" comparison in the whole system, and it is a **read of a
cached number with no claim attached**. Verified locations:

1. **The host computes numbers, never a verdict.**
   `ParallelsService.GetHardwareUsage` (`serviceprovider/parallelsdesktop/main.go:2834-2917`)
   builds `TotalAvailable = Total − SystemReserved − TotalInUse`
   (`main.go:2896-2897`). `TotalInUse` is the sum over **running** VMs and `TotalReserved` the
   sum over **stopped** VMs — both via `accumulateVmHardwareUsage` (`main.go:2799-2832`).
   Served by `GET /v1/config/hardware` (`controllers/config.go:293-307`).

2. **In-flight work is only ever published as counts.** `PendingVms` / `PendingMacVMs` are
   operation counts; `PendingVmSlots` returns `0` unless `MAX_VMS_PER_HOST` is configured
   (`serviceprovider/parallelsdesktop/vm_capacity.go:201-211`). There is no `PendingCpu` /
   `PendingMemory` field anywhere — `mappers/system.go:20-29` has nowhere to put one.

3. **The orchestrator caches that snapshot** in `orchestrator_hosts.resources` (a JSON blob),
   refreshed on the pull cadence and on host events (`orchestrator/hardware_update_queue.go`,
   `OrchestratorService.UpdateHostResourcesForEvent` at `orchestrator/main.go:551-573`). The
   refresh replaces `host.Resources` wholesale (`orchestrator/main.go:512-527`).

4. **The one comparison** is `checkHostRequirements` (`orchestrator/check_host_requirements.go`):

   ```go
   // We will trust that the host has the reserved cpus setup correctly
   availableCpus := host.Resources.TotalAvailable.LogicalCpuCount   // :104 — stale, never claimed
   ```
   CPU at `:104-116`, memory at `:118-127`. Only the **count** dimensions honour pending
   (`:169-187`, `pending := host.Resources.TotalReserved.TotalVms`). Disk (`:129-163`) is the
   only live network probe, and it too is never claimed.

5. **The host applies whatever it is told.** Requested values go straight to
   `prlctl set <vm> --cpus N` / `--memsize M` (`main.go:2320`, `main.go:2356`). The host's only
   admission gates are `reserveVmSlot` (`vm_capacity.go:89-140`) and `reserveMacVMStart`
   (`mac_vm_capacity.go:36-64`), and both compare `current + pending < max` on **counts**.

The placement loop itself is honest: `getValidHostsForCreate`
(`orchestrator/create_host_virtual_machine.go:187-251`) rejects with
`400 "No host available to create the virtual machine"` when nothing passes (`:232-242`).
The defect is purely the window between read → decide → dispatch, during which nothing is
written down.

## 3. Two failure modes, not one

Concurrency is the loud one, but not the only one:

- **Race** — both requests read the same `TotalAvailable` before either has changed anything.
- **Staleness** — two creates **one second apart** double-book too, because nothing mutates the
  snapshot until the next hardware probe. A create that already succeeded is invisible to the
  next placement for the same reason.

A fix that only adds a mutex around the comparison fixes the first and leaves the second intact.

## 4. The rule being introduced

A **reservation** (also called a pending charge) is a row that says "job J has been promised
`cpu`/`memory`/`vm_count`/`apple_vm` on host H". Rules:

| Rule | Statement |
|---|---|
| Claim atomically | The capacity check and the insert of the reservation happen in **one DB transaction**, serialised per host |
| Charge, don't exclude | A reservation does **not** make a host busy. Concurrent creates keep landing on the best host until it stops fitting |
| Refuse fast | No capacity anywhere → immediate `400`, same message as today. No waiting queue, no deferred state |
| Effective availability | `TotalAvailable − max(Σ live reservations, host-reported pending)` per dimension |
| Always released | Terminal job state deletes the row; a lease TTL and an orphan reconciler are the backstops |
| Release after refresh | On success, persist a fresh host snapshot **before** deleting the row, so no instant exists where neither accounts for the capacity |
| Own table | Never stored inside `host.Resources` — that blob is host-owned and overwritten by every refresh |
| No I/O in the transaction | The admission transaction does two selects, one name probe and one insert — no HTTP calls |
| VM name uniqueness | Per owner, cluster-wide: a name already pending or already existing is refused |

## 5. Where the numbers come from after the change

`GET /v1/orchestrator/hosts/{id}` keeps reporting host truth unchanged (that is the host's own
inventory and must stay honest). The pending view is additive:

| JSON path | Meaning |
|---|---|
| `detailed_resources.total_pending.logical_cpu_count` | Σ live reservations on this host |
| `detailed_resources.total_pending.memory_size` | Σ live reservations, MiB |
| `detailed_resources.total_pending.total_vms` | number of creates in flight on this host |

`POST /v1/orchestrator/machines/check` keeps being **side-effect free**: it reports the same
`available` the placement will compare against (i.e. reduced by pending), but never claims it.

## 6. Configuration

| Key | Default | Meaning |
|---|---|---|
| `PLACEMENT_RESERVATION_ENABLED` | `true` | Orchestrator-side kill switch. `false` restores today's behaviour exactly |
| `PLACEMENT_RESERVATION_TTL_MINUTES` | `120` | Lease; a row older than this is swept. Sized above the slowest legitimate create, because the fast reclaim paths are event-driven |
| `PLACEMENT_RECONCILE_INTERVAL_SECONDS` | `30` | Sweeper cadence |
| `VM_CPU_ADMISSION_ENABLED` | `false` | Host-side CPU/memory gate. Off by default so existing oversubscribing deployments are unaffected |
| `VM_CPU_OVERCOMMIT_RATIO` | `1.0` | >1 deliberately re-allows oversubscription |
| `ORCHESTRATOR_NODE_ID` | machine unique id | Replica identity; also written to `jobs.owner_node`. Hostname where the platform has no unique id |

Multi-replica requires `DB_TYPE=postgres`; SQLite stays single-node (`config/database.go:73-108`).

## 7. Acceptance criteria

1. The reported repro is fixed: concurrent 6-CPU × 2 on a 10-CPU / 1-reserved host yields one
   running VM and one `400 Insufficient host resources … CPU: required 6 logical CPUs,
   available 3`, and `total_available.logical_cpu_count` is never negative for an accepted create.
2. Sharing still works: 2-CPU × 2 on a host with 10 free admits **both on that host**
   (Σ reserved 4, effective free 6); a 9-CPU request is refused that host.
3. Concurrent creates of the same VM name → exactly one admitted, other refused by a `vm_name` check.
4. After failures and rejections, `resource_reservations` is empty again.
5. A replica killed mid-create returns the capacity within the TTL or one reconcile tick.
6. Two replicas on one Postgres: starting replica B does not fail A's running jobs, and host job
   events arriving at either replica translate and release.
7. `PLACEMENT_RESERVATION_ENABLED=false` reproduces today's behaviour with no rebuild.

## 8. Out of scope

- **Disk reservations.** Pack size is known only mid-create; the disk check stays as is.
- **Orchestrator-side clone placement** — clones have no availability pass today.
- **General `persistHost` atomicity.** Only the loops this feature touches are made targeted.
- **Redis.** The `Store` interface exists precisely so Redis can arrive later; it is not built here.
- **Spreading** VMs across hosts for HA. Placement keeps its cache-locality and latency ordering.

## 9. Related files, by role

| Role | File |
|---|---|
| The one comparison | `src/orchestrator/check_host_requirements.go` |
| The leak window | `src/orchestrator/create_host_virtual_machine.go:187-251` |
| Host arithmetic | `src/serviceprovider/parallelsdesktop/main.go:2799-2917` |
| Count-gate precedent to copy | `src/serviceprovider/parallelsdesktop/vm_capacity.go:89-140` |
| macOS count-gate | `src/serviceprovider/parallelsdesktop/mac_vm_capacity.go:36-86` |
| Terminal-state choke point | `src/jobs/main.go:277,307,341` |
| Host event → job | `src/orchestrator/handlers/host_job_event_handler.go:85-136` |
| Store-outside-the-frozen-interface precedent | `src/settings/store.go`, `src/settings/main.go` |
| Model registry | `src/database/migrate.go:19` `AllGormModels()` |
| Interface-seam precedent | `interfaces.HardwareEnqueuer` (`orchestrator/handlers/pdfm_event_handler.go:19,45`) |
| Host create call sites | `src/controllers/machines.go:826,1808,1837,1958` |

## 10. How to reproduce locally (before the fix)

```bash
# 1 host with 10 logical CPUs, SYSTEM_RESERVED_CPU=1, no VMs. Orchestrator on Postgres.
export TOKEN=$(curl -s -X POST $ORCH/api/v1/auth/token -H 'Content-Type: application/json' \
  -d '{"username":"root","password":"***"}' | jq -r '.token // .access_token')

for i in 1 2; do
  curl -s -o /dev/null -w "%{http_code}\n" -X POST $ORCH/api/v1/orchestrator/machines/async \
    -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
    -d "{\"name\":\"over-$i\",\"architecture\":\"amd64\",\"catalog_manifest\":{\"catalog_id\":\"<id>\",\"machine_name\":\"over-$i\",\"specs\":{\"cpu\":\"6\",\"memory\":\"2048\"}}}" &
done; wait

curl -s -H "Authorization: Bearer $TOKEN" $ORCH/api/v1/orchestrator/hosts/$HOST \
  | jq '.resources.total_available, .resources.total_in_use'   # expect the negative today
```
