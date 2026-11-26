---
layout: page
title: Orchestrator
subtitle: Host Selection for VM Creation
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# Host Selection for VM Creation

When a virtual machine creation request arrives at the orchestrator, it goes through a multi-stage pipeline to select the best available host. This document describes each stage in order.

---

## Stage 0: Hard Tag Constraint (`target_tags`)

If the request includes `target_tags`, placement is restricted to hosts that carry **every** one of the requested host tags (case-insensitive exact match; see [Tag format](#tag-format)). This is the first filter applied: tag-mismatched hosts are rejected before any capacity evaluation or disk-space probe, so they cost neither scheduling effort nor network calls.

- `pool=build` + `os=sonoma` selects hosts tagged with **both** labels.
- Absent or empty `target_tags` imposes no constraint (all hosts remain eligible).
- If no host carries all the tags, the request fails with HTTP 400 naming the missing tags per host.
- The pre-flight check (`POST /v1/orchestrator/machines/check`) accepts the same `target_tags` field and reports a `host_tags` check with the missing tags for each evaluated host.
- The explicit-host endpoints (`POST /v1/orchestrator/hosts/{id}/machines[/async]`) also enforce the constraint: naming a host that lacks a required tag fails with HTTP 400.
- Only **host** tags participate. Tags attached to catalog VM manifests never influence placement.

### Tag format

Host tags are free-form labels managed by the operator: set at registration (`--tags=pool=build,os=sonoma`) or via `PUT /v1/orchestrator/hosts` (`tags` field — absent leaves tags untouched, an empty array clears them). `key=value` is a naming convention, not a parsed structure: `pool=build` matches only the identical tag (ignoring case), there is no key-only or wildcard matching. Tags cannot contain commas, because the registration flag splits the list on `,`.

---

## Stage 1: Basic Host Validation

Every registered host is evaluated against a set of baseline criteria. A host is skipped if any of the following is true:

| Check | Reason |
|-------|--------|
| Host is not enabled | Administratively disabled hosts are never considered |
| Host state is not `healthy` | Unhealthy hosts cannot reliably run workloads |
| Host has no resource information | Cannot make capacity decisions without resource data |
| Host architecture does not match request | e.g. `arm64` request cannot be fulfilled by an `x86_64` host |
| Host has insufficient CPU | Available logical CPU count is below the VM's requirement — see [What "available" means](#what-available-means) |
| Host has insufficient memory | Available memory is below the VM's requirement — see [What "available" means](#what-available-means) |
| Host has insufficient disk space | See [Disk Space Check](#disk-space-check) below |
| Apple VM limit reached | `macvm` type VMs are capped at a maximum per host |
| VM name already in use | The owner already has a VM, created or in flight, under that name |

Hosts that pass all checks are added to the **valid hosts** list. If no hosts pass, the request fails with HTTP 400.

### What "available" means

A host reports its own capacity periodically, and that snapshot is what the orchestrator caches. On its own it is a moment in the past: between two reports, several creations can each be told the same free capacity and all of them be accepted — which is how a host ends up running more VMs than it has CPUs for.

So the decision consults two sources, and charges the larger of the two per dimension (CPU, memory, VM count, Apple VMs), never the sum, because both can be describing the same work in flight:

| Source | Knows about | Blind to |
|--------|--------------|----------|
| Host-reported pending | what this host was asked to do, including creates submitted straight against it | what other replicas of the orchestrator promised |
| Reservation ledger | what placement has promised and the host has not applied yet | a create that never went through placement |

The ledger is the **capacity reservation**: a per-host row recording what a create has been charged, written in the same database transaction that re-checks the host's numbers, and released when the job running that create completes or fails. Two concurrent requests for the same host are therefore serialised where it counts — one of them is refused with the shortage spelled out (`CPU: required 6 logical CPUs, available 3`) instead of both being accepted.

Two properties are worth stating explicitly, because they are what makes the ledger safe to keep:

- **It charges, it does not reserve a seat.** Several creates are expected to keep landing on the same good host while it has room. A reservation never takes a host out of rotation.
- **Reading it is free of side effects.** The pre-flight check (`POST /v1/orchestrator/machines/check`) and the host resources endpoint report the same figures a create will be measured against, without claiming anything. Only dispatching to a host claims.

Reservations are released the moment their job finishes, and a reconciler reclaims one whose job has already gone away — so a crashed replica cannot leak capacity for long. The lease itself is the backstop for the case where even the job record has disappeared, and it is deliberately longer than the slowest create: a create that keeps reporting progress is alive, and freeing its capacity early would let a second create claim the same CPUs. Configuration:

| Variable | Default | Purpose |
|----------|---------|---------|
| `PLACEMENT_RESERVATION_ENABLED` | `true` | Set to `false` to switch the whole mechanism off — the documented rollback |
| `PLACEMENT_RESERVATION_TTL_MINUTES` | `120` | Lease after which an unreleased reservation is reclaimable |
| `PLACEMENT_RECONCILE_INTERVAL_SECONDS` | `30` | How often expired leases and orphaned reservations are swept |
| `ORCHESTRATOR_NODE_ID` | machine unique id | Identifies the replica that holds a reservation; falls back to the hostname where the platform has no unique id |

Placement requires the SQL data layer. On the legacy JSON data layer the ledger cannot exist, and placement reports itself disabled rather than pretending to count.

The host can enforce the same arithmetic on its own side, which also catches creates submitted directly to it rather than through placement. It is opt-in and off by default, so upgrading changes nothing for deployments that oversubscribe deliberately:

| Variable | Default | Purpose |
|----------|---------|---------|
| `VM_CPU_ADMISSION_ENABLED` | `false` | Admit or refuse creates on this host by CPU and memory |
| `VM_CPU_OVERCOMMIT_RATIO` | `1.0` | Multiplies the capacity left for VMs after the host keeps its reserved share |

While that gate is on, the host publishes what it is holding for creates that do not exist yet as `pending_cpu` / `pending_memory` on `GET /v1/orchestrator/overview/{id}/resources` — and that endpoint reports the larger of the host's figure and the orchestrator's ledger, so what you read there is exactly what a create will be measured against.

### Configuring placement

All six variables above are also settable through the [settings service](../settings/overview.md), which takes precedence over the environment. The two host-side ones (`VM_CPU_ADMISSION_ENABLED`, `VM_CPU_OVERCOMMIT_RATIO`) are host-scoped and take effect on the next sync; the four orchestrator-side ones are orchestrator-scoped.

`PLACEMENT_RESERVATION_TTL_MINUTES` and `PLACEMENT_RECONCILE_INTERVAL_SECONDS` are read once when placement starts, so the settings API marks them `requires_restart` and the UI asks for a restart after editing them. The two booleans are read per request and apply immediately.

To confirm what a running replica is actually doing, ask it:

```bash
curl -s $BASE/api/v1/config/hardware | jq '{enabled_modules, orchestrator_node_id, machine_id}'
```

`enabled_modules` lists **`placement`** when the mechanism is guarding dispatch — which means the switch is on *and* the SQL data layer is live. A replica on the legacy JSON data layer leaves it out even with the switch on, because in that mode nothing is counting and the list would be describing an intention rather than a behaviour. It is reported, never configured: it is not a member of `ENABLED_MODULES`.

`orchestrator_node_id` is the name this replica stamps on its reservations and on the jobs it owns, and `machine_id` is the key that sets it — `PUT /v1/orchestrator/settings/values/ORCHESTRATOR_NODE_ID?node=<machine_id>`. With several replicas on one database, that per-replica layer is the only safe place for the value; see [per-replica values](../settings/overview.md#per-replica-values-node-machine-id).

### Disk Space Check

Disk space is only checked when the catalog manifest provides a known VM size. The required free space depends on whether the Parallels home directory and the cache folder share the same volume:

- **Same volume**: `3 × VM size` — space is needed to download the pack, copy it to the cache, and expand the VM.
- **Different volumes**: `2 × VM size` — the download and the VM expansion happen on separate volumes.

Hosts running an older API version that does not expose the disk-space endpoint are not penalised — the check is skipped and the host remains eligible.

---

## Stage 2: Selection Tag Filter

If the request includes `selection_tags`, only hosts whose tag list contains **at least one** of the requested tags (case-insensitive) are kept. If tags are specified but no host matches, the request fails with HTTP 400.

`selection_tags` is the legacy OR-style filter; prefer `target_tags` (Stage 0) for hard placement rules. The two can be combined: a host must satisfy the `target_tags` AND-constraint **and** match at least one `selection_tags` entry.

---

## Stage 3: Cache Locality

When the request references a catalog manifest, the orchestrator checks whether any valid host already has that manifest version cached locally (matched on `catalog_id`, `version`, and `architecture`). If one or more hosts have the cache, only those hosts advance to the next stage — hosts without the cache are dropped. This avoids unnecessary downloads from the catalog and speeds up VM creation.

If **no** host has the cache, all valid hosts are kept and the chosen host will download the manifest at creation time.

---

## Stage 4: Latency Sorting

The remaining hosts are pinged (HTTP GET `/api/v1/config/health`, 2-second timeout) and sorted by round-trip time, shortest first. This favours lower-latency hosts, which generally results in faster creation and better reliability. Hosts that time out or return a non-200 response are assigned a penalty of 10 seconds and sorted to the back.

---

## Stage 5: Dispatch

The orchestrator iterates through the sorted host list and attempts to dispatch the creation request to each host in turn, calling the host's async endpoint (`POST /machines/async`). The first host to accept (HTTP 202) wins. The resulting host job ID is registered in the job registry so that completion events (success or failure) are forwarded back to the original orchestrator job.

Before a host is dispatched to — and only then — the create **claims** its capacity against that host. The claim re-reads the host's numbers and the ledger inside one per-host transaction and refuses the host if the arithmetic no longer fits, in which case the next candidate is tried. That is the moment two simultaneous requests are separated: the evaluation in Stage 1 only shortlisted hosts, the claim is what commits them. A host that was eligible a moment ago can be refused here, and that is the intended behaviour — someone else took the room.

If every host in the list rejects the request, the orchestrator returns an error.

---

## Related placement rules

| Need | Mechanism |
|------|-----------|
| Hard tag-based placement | `target_tags` on the create request (Stage 0) |
| Check feasibility before creating | `target_tags` on `POST /v1/orchestrator/machines/check` — reported as the `host_tags` check |
| Pin to a specific host | Explicit-host endpoints `POST /v1/orchestrator/hosts/{id}/machines[/async]` (still enforce `target_tags`) |
| Preferred hosts / soft affinity / anti-affinity | Not supported today |

---

## Summary

```
All registered hosts
        │
        ▼
┌─────────────────────────┐
│  Stage 0: Target Tags   │  target_tags (optional) — host must carry ALL requested tags
└─────────────────────────┘
        │ tag-eligible hosts
        ▼
┌─────────────────────────┐
│  Stage 1: Validation    │  enabled, healthy, architecture, CPU, memory, disk, Apple VM limit
└─────────────────────────┘
        │ valid hosts
        ▼
┌─────────────────────────┐
│  Stage 2: Tag Filter    │  selection_tags (optional)
└─────────────────────────┘
        │
        ▼
┌─────────────────────────┐
│  Stage 3: Cache Locality│  prefer hosts with catalog manifest already cached
└─────────────────────────┘
        │
        ▼
┌─────────────────────────┐
│  Stage 4: Latency Sort  │  ping each host, sort by round-trip time
└─────────────────────────┘
        │ ordered list
        ▼
┌─────────────────────────┐
│  Stage 5: Dispatch      │  claim capacity per host, try each in order, stop at first HTTP 202
└─────────────────────────┘
```
