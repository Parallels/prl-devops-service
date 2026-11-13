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
| Host has insufficient CPU | Available logical CPU count is below the VM's requirement |
| Host has insufficient memory | Available memory is below the VM's requirement |
| Host has insufficient disk space | See [Disk Space Check](#disk-space-check) below |
| Apple VM limit reached | `macvm` type VMs are capped at a maximum per host |

Hosts that pass all checks are added to the **valid hosts** list. If no hosts pass, the request fails with HTTP 400.

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
│  Stage 5: Dispatch      │  try each host in order, stop at first HTTP 202
└─────────────────────────┘
```
