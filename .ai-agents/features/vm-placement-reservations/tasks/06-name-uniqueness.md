# Task 06 — Cluster-wide VM name uniqueness

**Phase 1 · depends on: 01 · co-owned with 02 (the `vm_name` check renders there)**

Closes the concurrent same-name create. Today there is no name guard in the placement path at
all: `checkNameCpu` / `checkNameMemory` etc. (`orchestrator/check_host_requirements.go:20-27`)
are check *labels*, not name checks.

## Today's behaviour (verified)

- Placement path: no name check whatsoever.
- Only guards are check-then-act on the host: `machine %v already exists`
  (`catalog/pull.go:338-353`) and `VM with name %s already exists` in `RegisterVm`
  (`serviceprovider/parallelsdesktop/main.go:1551-1557`).
- Consequence: same host → one create fails late, after a full pack download; **different hosts
  → two identical names, undetected**.

## Scope

| File | Change |
|---|---|
| `src/placement/store_sql.go` | `NameTaken(ctx, owner, vmName, exceptJobID)` — pending rows for that owner/name on **any** host, plus the existing-VM read |
| `src/placement/interfaces.go` | already declared in task 01 |
| `src/orchestrator/check_host_requirements.go` | new constant `checkNameVmName = "vm_name"` and a check entry so pre-flight reports it |
| `src/orchestrator/create_host_virtual_machine.go` | `Admit` refuses when the name is taken |

Existing-VM lookup reuses `GetOrchestratorVirtualMachines(ctx, filter)` filtered on `user` +
`name` — both indexed in `data/models/virtual_machine.go:17,23`. **No new `DatabaseService`
method** (#421 R-1).

## Rules

- Enforced **inside** the admission transaction, so the check and the claim commit together and
  two racing creates cannot both pass.
- Scope is **per owner, cluster-wide** — pending on any host counts, not just the candidate host.
- Same job re-placing hosts is exempt (`job_id <> ?`), otherwise host hopping collides with itself.
- Comparison: exact match on `strings.TrimSpace`ed input. Align casing behaviour with the
  host-side `findVmInCacheAndSystem` comparison during implementation and note the choice in a
  code comment, so orchestrator and host never disagree on what "the same name" means.
- Reported `Reason: "vm_name"`, surfaced to the user as part of the existing
  `Insufficient host resources`-style aggregate failure path (a taken name is a 400, not a 500).
- Pre-flight (`POST /v1/orchestrator/machines/check`) includes the check but **never claims**.

## Acceptance

- [ ] Two concurrent creates with the same name, generous capacity, one host → exactly one
      admitted; the other refused with the `vm_name` reason. Verified under `-race`.
- [ ] Two concurrent same-name creates where only one host fits each → still exactly one
      admitted (proves the cluster-wide scope, not per-host).
- [ ] Creating a name that already exists as a VM of that owner → refused, and the pre-flight
      response lists `{"name":"vm_name","passed":false,...}`.
- [ ] A single job that fails on host A and is re-placed to host B is **not** blocked by its own
      pending row.
- [ ] Deleting the VM frees the name for a new create.

## Commands

```bash
cd src && GOTMPDIR=$PWD/.gotmp go test -count=1 -race ./placement/... ./orchestrator/...
GOTMPDIR=$PWD/.gotmp go test -count=1 ./tests/orchestrator/...
```
