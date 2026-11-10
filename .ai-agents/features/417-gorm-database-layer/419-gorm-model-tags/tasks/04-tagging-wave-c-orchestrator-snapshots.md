# Task 04 — Tagging wave C: orchestrator + snapshots + virtual machine (nested-value policy)

## Context
Plan mapping rows 14–24. Runs after task 03 merges (sequential waves; `orchestrator_host.go`
references reverse-proxy types tagged in 03). Applies: the RELATION decisions on
`OrchestratorHost` (D-3), snapshot **normalisation** (wrappers → no table; `VMSnapshot` owns
`vm_snapshots` — the default-name collision from D-2), and the 25-column SERIALISE policy on
`VirtualMachine`'s nested config structs.

## Objective
`orchestrator_host.go`, `VM_Snapshots.go`, and post-split `virtual_machine.go` tagged per mapping;
golden tests green; no table is created for the two snapshot wrappers or any `VirtualMachine*`
nested type; build+tests green.

## Files
- Modify: `src/data/models/orchestrator_host.go`, `src/data/models/VM_Snapshots.go`,
  `src/data/models/virtual_machine.go` (post-task-01 main file),
  `src/data/models/orchestrator_vm_snapshots.go` (doc comment recording the no-table decision only)
- Modify: `models_gorm_test.go`, `testdata/golden/` (`orchestrator_host.json`, `vm_snapshot.json`,
  `vm_snapshots_wrapper.json`, `virtual_machine.json`, `virtual_machine_minimal.json`)

## File size guard (ER-4)
| File | Now (measured / post-01) | After (proj.) | Verdict |
|---|---|---|---|
| `orchestrator_host.go` | 315 | ~380 | ok |
| `VM_Snapshots.go` | 15 | ~28 | ok |
| `orchestrator_vm_snapshots.go` | 7 | 9 (comment) | ok |
| `virtual_machine.go` | ~230 (post-split; was 1,018) | ~270 | ok |
| `virtual_machine_{hardware,config,network}.go` | ~270/~470/~80 | unchanged | ok |

## Implementation spec
`orchestrator_host.go` (excerpt — mind exact Go field spellings: child VM uses `HostId`,
proxy children use `HostID`):
```go
Authentication *OrchestratorHostAuthentication `json:"authentication,omitempty" gorm:"serializer:json"`
Resources      *HostResources                  `json:"resources,omitempty" gorm:"serializer:json"`
Tags           []string                        `json:"tags,omitempty" gorm:"serializer:json"`
EnabledModules []string                        `json:"enabled_modules,omitempty" gorm:"serializer:json"`
RequiredClaims []string                        `json:"required_claims,omitempty" gorm:"serializer:json"`
RequiredRoles  []string                        `json:"required_roles,omitempty" gorm:"serializer:json"`
CacheConfig    *models.CatalogCacheConfig      `json:"cache_config,omitempty" gorm:"serializer:json"`
CacheItems     []models.HostCatalogCacheItem   `json:"cache_items,omitempty" gorm:"serializer:json"`
HealthCheck    *models.ApiHealthCheck          `json:"-" gorm:"-:all"`
VirtualMachines   []VirtualMachine      `json:"virtual_machines,omitempty" gorm:"foreignKey:HostId;references:ID;constraint:-"`
ReverseProxy      *ReverseProxy         `json:"reverse_proxy,omitempty" gorm:"foreignKey:HostID;references:ID"`
ReverseProxyHosts []*ReverseProxyHost   `json:"reverse_proxy_hosts,omitempty" gorm:"foreignKey:HostID;references:ID;constraint:-"`
```
`ID` PK, `Host gorm:"index"` (unique deferred — see plan risk list). `LastUnhealthy` etc. keep
identity snake columns; wire keys (`last_seen`) untouched.

`VM_Snapshots.go` — normalisation carriers (D-3/D-2). Two **DB-only** fields added, wire-safe via
`json:"-"`; `VMSnapshots` and `HostsVMSnapshotsRecord` remain untagged wrappers:
```go
type VMSnapshot struct {
	ID      string `json:"id,omitempty" gorm:"primaryKey"`
	Name    string `json:"name"`
	Date    string `json:"date"`
	State   string `json:"state"`
	Current bool   `json:"current"`
	Parent  string `json:"parent"`
	HostID  string `json:"-" gorm:"column:host_id;index:idx_vmsnapshots_host_vm"` // from HostsVMSnapshotsRecord key
	VMID    string `json:"-" gorm:"column:vm_id;index:idx_vmsnapshots_host_vm"`   // from map key / wrapper VMId
}
```
Migration note (row-count formula for #424): `rows(vm_snapshots) = Σ_hosts Σ_vm |VMSnapshots[vm_id]|`;
empty `ID` filled by `helpers.GenerateId` at migration time.

`virtual_machine.go`: `ID string json:"ID,omitempty"` PK (JSON key stays capitalised `ID`);
`HostId gorm:"column:host_id;index"`; `User gorm:"index"`; all 25 nested/map fields get
`gorm:"serializer:json"` per mapping row 22 (guest_tools … network_information); the 31 nested
`VirtualMachine*` types stay **untagged** — comment header states "serialised into
virtual_machines TEXT columns; not AutoMigrate roots".

## Out of scope
Join-table schema tests (05), AutoMigrate matrix + snapshot (06), `virtual_machines.go`
`ParallelsVM` family (untagged DTOs — mapping row 24).

## Validation
1. `cd src && go build ./... && go test ./... -count=1` green.
2. `go test ./data/models/... -run 'JSONGolden' -count=1` green incl. VM + snapshot fixtures
   (byte-exact: proves `json:"-"` carriers leak nothing into the wire).
3. `grep -c 'serializer:json' src/data/models/virtual_machine.go` = 25; `grep -c 'gorm:"' VM_Snapshots.go` covers 8 fields of `VMSnapshot` only.
4. `git diff testdata/` empty in the tag commit.

## Rollback
Revert PR. Added `json:"-"` fields are additive struct fields — revert cleanly, no consumers yet.

## PBI acceptance mapping
AC1 (rows 14, 17–24… 14/20/22), AC2 (has-many ×2 + has-one declared), AC4 (VM/snapshot golden bytes;
normalisation loses nothing: every wrapper key lands in a column), ER-2/ER-3 closure of the
1,018-line trigger.

## Slice gate — execution protocol (feature plan §14 · ER-6/ER-7)

**Slice id: `419-04`.** Cut `slice/419-04` from `sql-migration` HEAD
(`.worktrees/slice-419-04`) → implement this spec only → pass `## Validation` inside the
worktree → append the slice entry (diff summary + **human test plan** with expected outputs + observed
results + deviations) to [`VERIFICATION.md`](../../VERIFICATION.md) → **STOP for the human gate**.
Approved ⇒ `git merge --no-ff` into `sql-migration`, remove worktree + branch, next slice.
Rejected ⇒ fix in-worktree, re-report. `sql-migration` never takes a red merge (ER-5).
