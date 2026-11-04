# PBI #419 Plan — Data model conversion with GORM struct tags

Fact base: one grep over `src/data/models/*.go` (`^type |json:`) + `wc -l` (23 files, 2,462 lines),
`user.go`/`role.go`/`db_common.go` anchors, `src/data/users.go:29-59` re-hydration, feature plan §6 ER rules.
**95 structs** inventoried. **20 SQL tables** produced (17 entity + 3 join); 75 structs are
serialised/nested or transient and correctly get **no table**.

## Scope

- Add `gorm:"..."` tags **in place** to every persisted struct in `src/data/models/` (ER-3 split of
  `virtual_machine.go` lands first, task 01). No JSON tag changes — these structs are REST wire
  formats; golden-JSON byte tests guard every tagged struct.
- Declare relationships (M2M `user_roles`/`user_claims`/`role_claims`; has-many
  `reverse_proxy_hosts`, `vm_snapshots`, `virtual_machines` via `host_id`; has-one
  `reverse_proxies.host_id`).
- Produce the **field→column mapping table** below — #424's migration spec derives from it verbatim.
- Verify `AutoMigrate` against fresh SQLite via #418 factory (`database.DefaultDatabaseConfig(ctx)`
  post-#420) + schema snapshot test.
- Out: query methods (#421), data migration (#424), `src/models` package structs
  (`ApiHealthCheck`, `CatalogCacheConfig`, `HostCatalogCacheItem` — foreign, handled via `gorm:"-:all"`/serializer).

## Decisions

### D-1 — Keep string date columns in v1
All `CreatedAt/UpdatedAt/ExpiresAt/BlockedSince/…` stay `string` (grep: every date field is `string`
in all 23 files). `time.Time` would change JSON encoding (RFC3339 vs stored string) → breaking wire
change + migration cast risk. Revisit post-migration behind a separate PBI if typing improves anything.

### D-2 — Table naming: GORM defaults + minimal `TableName()`
Default NamingStrategy (snake_case plural) used everywhere. Collisions/overrides found in grep:
- `VMSnapshot` → `vm_snapshots` and `VMSnapshots` → `vm_snapshots` collide under default pluralising.
  Resolved: `VMSnapshot` owns `vm_snapshots`; the two wrappers (`VMSnapshots`,
  `HostsVMSnapshotsRecord`) get **no table** (normalised, D-3) — no `TableName()` needed, collision gone.
- Default M2M join names would be `role_users`/`claim_users` (alphabetical) → explicit
  `many2many:user_roles`, `many2many:user_claims` for intent-revealing legacy-compatible names.
- `RoleClaims` join model keeps its own table name `role_claims` (matches legacy model exactly).

### D-3 — Embedded-collection policy, per field (walk of every slice/map/nested struct in grep)
Three policies: **NORMALISE** (own table/join), **SERIALISE** (`gorm:"serializer:json"` TEXT column —
GORM v2 stores canonical `json.Marshal` of the field value; deterministic field order + sorted map keys
⇒ byte-stable round-trip), **DROP** (`gorm:"-:all"`).

| Field | Decision | Rationale | Read-shape constraint |
|---|---|---|---|
| `User.Roles []Role` | NORMALISE join `user_roles(user_id,role_id)` | queryable M2M | `users.go:44-56` refills each Role's live Claims ⇒ reads must `Preload("Roles.Claims")` |
| `User.Claims []Claim` | NORMALISE join `user_claims(user_id,claim_id)` | queryable M2M | `Preload("Claims")` in same query path |
| `Claim.Users []User` / `Role.Users []User` (`json:"-"`) | inverse M2M refs only | hidden in JSON already | no preload; keeps `json:"-"` shape untouched |
| `Role.Claims []Claim` | NORMALISE — reuse legacy `RoleClaims{RoleID,ClaimID}` as explicit join model (`joinForeignKey:RoleID;joinReferences:ClaimID`) | link table already exists in grep | feeds `Roles.Claims` nested preload |
| `OrchestratorHost.VirtualMachines []VirtualMachine` | RELATION `has-many` via existing `VirtualMachine.HostId` — **not** a host column | VMs are their own legacy collection; double-storage would desync | fill via `HostId` query/preload; never cascade-save (D-7) |
| `OrchestratorHost.ReverseProxyHosts []*ReverseProxyHost` | RELATION has-many via child's existing `HostID` | child already carries FK field | preload only on proxy endpoints |
| `OrchestratorHost.ReverseProxy *ReverseProxy` | RELATION has-one via `reverse_proxies.host_id` | struct carries `HostID` | preload on demand |
| `OrchestratorHost.Tags/EnabledModules/RequiredClaims/RequiredRoles []string` | SERIALISE | filter-only-in-app today; no cross-row query | — |
| `OrchestratorHost.Authentication/Resources/CacheConfig/CacheItems` | SERIALISE (foreign `models.*` types can't be tagged here) | 1:1 blobs, loaded as unit | — |
| `OrchestratorHost.HealthCheck` (`json:"-"`) | DROP `gorm:"-:all"` | runtime-only | — |
| `HostsVMSnapshotsRecord.VMSnapshots map[string][]VMSnapshot` | NORMALISE into `vm_snapshots` rows; parent keys become `host_id`/`vm_id` columns | snapshots queried per (host,vm); map column would be unqueryable | wrapper struct stays untagged |
| `VMSnapshots.VMSnapshot []VMSnapshot` | NORMALISE same table (`vm_id` from wrapper) | ditto | wrapper untagged |
| `CatalogManifest.Provider/MinimumSpecRequirements` | SERIALISE | nested 1:1 config | — |
| `CatalogManifest.VirtualMachineContents/PackContents []CatalogManifestContentItem` | SERIALISE | file inventory, never row-queried | — |
| `CatalogManifest.Tags/RequiredRoles/RequiredClaims` | SERIALISE | same as host tags | — |
| `CatalogManager.RequiredClaims` | SERIALISE | ditto | — |
| `Job.Steps []JobStep` | SERIALISE | ordered progress blob, rewritten whole | — |
| `PackerTemplate.Variables/Specs/Defaults map, Addons/RequiredRoles/RequiredClaims` | SERIALISE | config maps | — |
| `ReverseProxyHost.Tls/Cors/HttpRoutes/TcpRoute` | SERIALISE | loaded as unit; route `RegexpPattern *regexp.Regexp` is `json:"-"` → stays out (serialiser uses JSON tags ⇒ already excluded); child IDs inside preserve `Order` | — |
| `HostResourceItem.MacVmsRunning []string`, `HostReverseProxy.Hosts` | inside SERIALISED `HostResources`/`Host.ReverseProxy` host blob | host inventory snapshot | — |
| `VirtualMachine` nested `VirtualMachine*` structs (31) + `HostSharedFolders map[string]interface{}` | SERIALISE, one TEXT column per field | 1:1 VM config mirrors; only ever read/written whole with wire JSON | nested types stay untagged |
| `VirtualMachineNetworkInformation.IPAddresses []…IPAddress` | inside SERIALISED `NetworkInformation` column | ditto | — |
| `DbRecord.{IsLocked,LockedBy,LockedAt}` | DROP — `gorm:"-:all"` on the `*DbRecord` embed | D-4 | `"db_record"` JSON key still emitted (tag untouched) |

### D-4 — `DbRecord` lock metadata dropped from SQL
`DbRecord` is transient lock bookkeeping for the legacy engine (`LockRecord/UnlockRecord`,
`src/data/main.go:399-426` per brief). SQL row locks/transactions replace it. Embedded `*DbRecord`
stays in the struct with `gorm:"-:all"` so `json:"db_record"` wire output is unchanged; no columns.

### D-5 — Soft delete: **none qualifies — AC5 skipped with evidence**
Grep evidence across all 23 files: no `IsDeleted`/`Deleted`/`Archived` flag exists on any persisted
struct. The only delete-flag-ish fields are `ApiKey.Revoked+RevokedAt` and
`CatalogManifest.Revoked+RevokedBy/RevokedAt` — business **statuses** exposed in wire JSON (with
`*By` audit columns `gorm.DeletedAt` cannot express), and `CatalogManifestContentItem.DeletedAt`
(string) which is data inside a serialised inventory list, not a row lifetime. `gorm.DeletedAt` is
therefore **not applied anywhere**; adding it would silently filter list queries and break #421 parity.
AC5 is closed by this evidence note (task 06 records it in the schema-snapshot test header).

### D-6 — Tags added in place; no parallel DTO copies
Same structs keep serving REST responses (feature AD-1). Only `gorm:"-:all"` for API-only/runtime
fields. Any new Go types are join-model-adjacent only if GORM requires them (it doesn't — `RoleClaims`
already exists).

### D-7 — Relationship loading (pointer for #421)
- Users: `Preload("Claims").Preload("Roles.Claims")` — mandatory for API parity with
  `GetUsers` re-hydration (`users.go:44-56`); nested preload = one JOIN-batch per level, not N+1.
- Hosts list: do **not** preload `VirtualMachines`/`ReverseProxyHosts` (N+1 + payload bloat);
  load per endpoint that needs them. `VirtualMachines` must never be association-saved — VM rows are
  written only via VM-domain methods (`constraint:OnDelete:SET NULL` avoided; use `constraint:-`).
- Snapshot reads: query `vm_snapshots WHERE host_id=? AND vm_id=?`, re-group in Go to legacy map shape.

## Mapping table (#424 consumes verbatim — no TBD rows)

Default rules unless overridden per row: **PK** = `ID string → "<table>.id" TEXT PRIMARY KEY`
(migration fills empties via `helpers.GenerateId`); every non-serialised, non-dropped scalar keeps its
legacy name snake_cased (GORM default) and Go type (string→TEXT, bool→integer, int→integer,
int64→integer, float64→real). "Row count" = expected SQL rows per legacy collection.

| # | Struct(s) | Table (PK) | Unique / Index | Serialised → TEXT columns | Dropped (gorm:"-:all") | Relationships (join + FKs) | Row-count formula |
|---|---|---|---|---|---|---|---|
| 1 | `User` | `users` (`id`) | `username` uniq; idx `email` | `roles`,`claims` → **join tables not cols** (NORMALISE) | — | M2M→`user_roles(user_id,role_id)`, `user_claims(user_id,claim_id)` | 1 per legacy user |
| 2 | `Role` | `roles` (`id`) | idx `name` | — | — | M2M→`role_claims`(explicit model, `RoleID`/`ClaimID`), `user_roles` inverse | 1 per legacy role |
| 3 | `Claim` | `claims` (`id`) | idx `name` | — | — | M2M inverses via `user_claims`, `role_claims` | 1 per legacy claim |
| 4 | `RoleClaims` | `role_claims` (`role_id`+`claim_id` composite) | uniq composite | — | — | the join model itself | Σ_roles \|claims\| |
| 5 | `ApiKey` | `api_keys` (`id`) | `key` uniq; idx `user_id` | — | `*DbRecord` | belongs-to `User` via `user_id` (no cascade) | 1 per legacy key |
| 6 | `CatalogManager` | `catalog_managers` (`id`) | idx `owner_id` | `required_claims` | — | — | 1 per legacy manager |
| 7 | `CatalogManifest` | `catalog_manifests` (`id`) | idx `catalog_id` | `provider`,`minimum_requirements`,`virtual_machine_contents`,`pack_contents`,`required_roles`,`required_claims`,`tags` | — | — | 1 per legacy manifest |
| 8 | `CatalogManifestContentItem`,`CatalogManifestProvider`,`MinimumSpecRequirement` | — (no tables) | — | live inside row-7 TEXT columns | — | — | 0 (0 is the spec) |
| 9 | `Job` | `jobs` (`id`) | idx `owner`,`result_record_id` | `steps` | `*DbRecord` | — | 1 per legacy job |
| 10 | `JobStep` | — | — | inside `jobs.steps` | — | — | 0 |
| 11 | `PackerTemplate` | `packer_templates` (`id`) | idx `owner` | `variables`,`specs`,`defaults`,`addons`,`required_roles`,`required_claims` | — | — | 1 per legacy template |
| 12 | `UserConfig` | `user_configs` (`id`) | uniq (`user_id`,`slug`); idx `user_id` | — | `*DbRecord` | belongs-to `User` via `user_id` | 1 per legacy config |
| 13 | `OrchestratorEnrollmentToken` | `orchestrator_enrollment_tokens` (`id`) | `token` uniq | — | `*DbRecord` | — | 1 per legacy token |
| 14 | `OrchestratorHost` | `orchestrator_hosts` (`id`) | idx `host` | `tags`,`enabled_modules`,`required_claims`,`required_roles`,`authentication`,`resources`,`cache_config`,`cache_items` | `health_check` (`json:"-"`) | has-many `virtual_machines` (`virtual_machines.host_id`→`id`, `constraint:-`), has-many `reverse_proxy_hosts` (`host_id`), has-one `reverse_proxies` (`host_id`) | 1 per legacy host |
| 15 | `OrchestratorHostAuthentication` | — | — | inside `orchestrator_hosts.authentication` | — | — | 0 |
| 16 | `HostResources`,`HostResourceItem`,`HostReverseProxy`,`HostResourceOverviewResponseItem` | — (never AutoMigrate roots; DTO/aggregate shapes) | — | persist only via host `resources` blob | — | — | 0 |
| 17 | `ReverseProxy` | `reverse_proxies` (`id`) | idx `host_id` | — | — | belongs-to `OrchestratorHost` via `host_id` | Σ_hosts [1 if reverse_proxy≠nil] |
| 18 | `ReverseProxyHost` | `reverse_proxy_hosts` (`id`) | idx `host_id` | `tls`,`cors`,`http_routes`,`tcp_route` | — | belongs-to `OrchestratorHost` via `host_id` | Σ_hosts \|reverse_proxy_hosts\| |
| 19 | `ReverseProxyHost{Tls,Cors,HttpRoute,TcpRoute}` | — | — | inside row-18 TEXT cols (`RegexpPattern` stays `json:"-"`, never persisted) | — | — | 0 |
| 20 | `VMSnapshot` | `vm_snapshots` (`id`) | idx (`host_id`,`vm_id`) | — | — | belongs-to host (plain `host_id` col, no FK constraint) | Σ_hosts Σ_vm \|snapshots\| |
| 21 | `VMSnapshots`,`HostsVMSnapshotsRecord` | — (normalised wrappers; untagged) | — | keys become `vm_snapshots.vm_id`/`host_id` | — | — | 0 |
| 22 | `VirtualMachine` | `virtual_machines` (`id`) | idx `host_id`,`user` | `guest_tools`,`mouse_and_keyboard`,`usb_and_bluetooth`,`startup_and_shutdown`,`optimization`,`travel_mode`,`security`,`smart_guard`,`modality`,`full_screen`,`coherence`,`time_synchronization`,`expiration`,`smbios_settings`,`hardware`,`host_shared_folders`,`shared_profile`,`shared_applications`,`smart_mount`,`miscellaneous_sharing`,`advanced`,`print _management`,`guest_shared_folders`,`network_information` (JSON-key typos like `print _management` are wire-preserved by serializer — column name is snake of Go field) | — | belongs-to host via existing `host_id` string col | 1 per legacy VM record |
| 23 | `VirtualMachine*` nested (31: `Advanced Cdrom0 Coherence CPU Expiration Fullscreen GuestSharedFolders GuestTools Hdd0 Hardware Memory MemoryQuota MiscellaneousSharing Modality MouseAndKeyboard Net0 NetworkInformation NetworkInformationBound NetworkInformationIPAddress Optimization PrintManagement SMBIOSSettings Security SharedApplications SmartMount Sound0 StartupAndShutdown TimeSynchronization TravelMode USBAndBluetooth Video`) | — | — | inside row-22 TEXT cols; types stay **untagged** | — | — | 0 |
| 24 | `ParallelsVM` + 28 untagged nested types (`GuestTools MouseAndKeyboard USBAndBluetooth StartupAndShutdown Optimization TravelMode Security Expiration Modality Fullscreen Coherence TimeSynchronization BootOrder-class SMBIOSSettings Hardware CPU Cdrom0 Hdd0 Memory MemoryQuota Net0 Sound0 Video MiscellaneousSharing SharedApplications SmartMount Advanced PrintManagement GuestSharedFolders SharedProfile`) | — runtime prlctl DTOs, never AutoMigrate roots, remain untagged | — | n/a | n/a | n/a | 0 |
| 25 | `Configuration` | `configurations` (`id`) | single-row | — | — | — | ≤1 (legacy meta record) |
| 26 | `DatabaseSchema` | `database_schemas` (`version`) | single-row | — | — | — | ≤1 (#424 may redirect to `migration_meta`) |
| 27 | `RepositoryImage` | — (4-line empty struct; zero fields ⇒ zero columns) | — | — | — | — | 0 |
| 28 | `DbRecord` | — (D-4) | — | — | all 3 fields | — | 0 |

**Join-table totals:** `user_roles`, `user_claims` (implicit via `many2many` tags), `role_claims`
(explicit model). **AutoMigrate list (20 models):** User, Role, Claim, RoleClaims, ApiKey,
CatalogManager, CatalogManifest, Job, PackerTemplate, UserConfig, OrchestratorEnrollmentToken,
OrchestratorHost, ReverseProxy, ReverseProxyHost, VMSnapshot, VirtualMachine, Configuration,
DatabaseSchema. Rows: **28 blocks / 95 structs mapped, 0 TBD.**

## Target files

| File | Action | `wc -l` now → after |
|---|---|---|
| `virtual_machine.go` | split (01) then tag main struct (04) | 1,018 → ~230 → **~270 (proj.)** |
| `virtual_machine_hardware.go` / `_config.go` / `_network.go` | new (01), tag columns wired only via main struct | new → ~270/~470/~80 (proj.) |
| `user.go`,`role.go`,`claim.go`,`role_claims.go`,`api_key.go` | tag (02) | 17→~45, 10→~25, 12→~30, 10→~25, 20→~50 (proj.) |
| `catalog_manifest.go`,`catalog_manager.go`,`packer_template.go`,`config.go`,`user_config.go`,`job.go`,`reverse_proxy.go`,`orchestrator_enrollment_token.go` | tag (03) | 144→~200, 18→~45, 54→~85, 5→~15, 22→~45, 43→~70, 299→~340, 11→~30 (proj.) |
| `orchestrator_host.go`,`VM_Snapshots.go`,`orchestrator_vm_snapshots.go` | tag (04; wrappers stay untagged, doc-comment added) | 315→~380 (proj.) |
| `models_gorm_test.go` (+`testdata/golden/*.json`) | new tests (02→06) | new ≤~350 total across waves (proj.) |
| `db_common.go`,`host_resources.go`,`repository.go`,`schema.go`,`virtual_machines.go` | untouched (0 tags needed — no AutoMigrate roots) | 7/144/4/5/261 unchanged |

All post-tag files ≤ 500 « 1,000 ⇒ **ER-2 satisfied**; only ER-3 trigger is `virtual_machine.go`
(1,018, confirmed by this round's `wc -l`).

## Risks

- **Wire compat**: serializer round-trips nil-slice vs `[]` and map `null` vs `{}` — golden tests
  pin exact bytes per struct (task 02 skeleton); JSON-key typos (`"removable drives,omitempty"`,
  `"print _management"`) must survive ⇒ no tag renames allowed anywhere.
- **N+1 / over-eager cascade**: `Save(host)` with associations populated upserts children — hard rule
  for #421: write children via their own domain methods; `constraint:-` on host has-manys.
- **virtual_machine split regressions**: method↔type clusters must move together; task 01 gates on
  green tests **before** any tag exists.
- **Unique indexes vs legacy dupes**: uniq on `users.username`,`api_keys.key`,`tokens.token`,
  (`user_configs.user_id`,`slug`) ⇒ #424 must dedupe or migration fails loudly; deferred decision:
  `users.email` intentionally **not** unique (empty-string collisions plausible in legacy data).

## Task index

| Task | PR | Depends | Summary |
|---|---|---|---|
| [01](tasks/01-split-virtual-machine.md) | 1 | — | Pure-move split of `virtual_machine.go` (ER-3), tests green pre-tagging |
| [02](tasks/02-tagging-wave-a-auth.md) | 2 | 01 | Wave A: user/role/claim/role_claims/api_key tags + joins + golden-JSON harness |
| [03](tasks/03-tagging-wave-b-infra.md) | 3 | 02 (sequential — same package files) | Wave B: catalog*, packer, config, user_config, job, reverse-proxy family, enrollment token |
| [04](tasks/04-tagging-wave-c-orchestrator-snapshots.md) | 4 | 03 | Wave C: orchestrator_host, snapshot normalisation, virtual_machine* nested policy |
| [05](tasks/05-relationships-embed-fidelity.md) | 5 | 04 | Join-table schema + embed-fidelity parity tests vs legacy-engine read shape |
| [06](tasks/06-automigrate-verification-mapping-final.md) | 6 | 05 | AutoMigrate + schema snapshot + mapping-table finalisation pass |

Waves 02→03→04 are sequential by rule (shared-package edits forbidden in parallel).
Validation gate on **every** PR: `cd src && go build ./... && go test ./... -count=1` (ER-5).

## AC coverage

| AC | Tasks |
|---|---|
| AC1 all models annotated (incl. every row of mapping table) | 01 (enabler), 02, 03, 04, 06 (gap pass) |
| AC2 relationships declared (M2M×3, has-many×2, has-one) | 02, 04, 05 |
| AC3 AutoMigrate verified on fresh SQLite | 06 |
| AC4 no data loss — round-trip all fields | 02–05 golden/parity tests, 06 schema snapshot |
| AC5 soft deletes where appropriate | D-5 evidence (none qualify) — documented in 06; **deliberately skipped, evidence-linked** |

## Conventions binding every wave

1. **Capture-first protocol**: golden fixtures are generated + committed in a pre-tag commit inside
   the same PR; the tag commit may not touch `testdata/golden/` — a non-empty `git diff testdata/`
   on the tag commit is the "tags changed the wire format" alarm.
2. Tag style: one `gorm:"..."` per field needing a directive; scalars with identity mapping get no
   tag (GORM derives column names); `gorm:"-"` variants used are `gorm:"-:all"` (field fully
   ignored by GORM, JSON untouched).
3. Commit shape (ER-3/ER-5): `split (if any)` → `golden capture` → `tags` → `tests`, each green.
4. Never add, rename, or reorder a `json:"..."` key — including the typos (`print _management`,
   `removable drives,omitempty`, `opacity_(percentage)`, capital `"ID"` on VMs): they are shipped
   API surface.

## Appendix A — struct census by file (from the one grep; **95 total**, 23 files, 2,462 lines)

| File (`wc -l`) | structs | | File (`wc -l`) | structs |
|---|---|---|---|---|
| api_key.go (20) | 1 | | packer_template.go (54) | 1 |
| catalog_manager.go (18) | 1 | | repository.go (4) | 1 (`RepositoryImage` — empty) |
| catalog_manifest.go (144) | 4 | | reverse_proxy.go (299) | 6 |
| claim.go (12) | 1 | | role.go (10) | 1 |
| config.go (5) | 1 | | role_claims.go (10) | 1 |
| db_common.go (7) | 1 | | schema.go (5) | 1 |
| host_resources.go (144) | 4 | | user.go (17) | 1 |
| job.go (43) | 2 | | user_config.go (22) | 1 |
| orchestrator_enrollment_token.go (11) | 1 | | virtual_machine.go (1,018) | 32 (1 root + 31 nested) |
| orchestrator_host.go (315) | 1 | | virtual_machines.go (261) | 29 (`ParallelsVM` + 28 DTO) |
| orchestrator_host_authentication.go (21) | 1 | | VM_Snapshots.go (15) | 2 |
| orchestrator_vm_snapshots.go (7) | 1 | | **total** | **95** |

## Appendix B — AutoMigrate model list (single source of truth; shared helper in tasks 05/06)

```go
func allAutoMigrateModels() []any { return []any{ // 18 roots → 20 tables
	&User{}, &Role{}, &Claim{}, &RoleClaims{}, &ApiKey{}, &CatalogManager{}, &CatalogManifest{},
	&Job{}, &PackerTemplate{}, &UserConfig{}, &OrchestratorEnrollmentToken{},
	&OrchestratorHost{}, &ReverseProxy{}, &ReverseProxyHost{}, &VMSnapshot{},
	&VirtualMachine{}, &Configuration{}, &DatabaseSchema{},
} } // + implicit join tables user_roles, user_claims (role_claims is an explicit model)
```

## Handoff notes (#421 / #424 — read before consuming this plan)

- **#421 user reads**: the only parity-safe read is `Preload("Claims").Preload("Roles.Claims")`;
  anything else reproduces `users.go:44-56` incorrectly (task 05 encodes this as a failing-negative test).
- **#421 host reads**: `constraint:-` on host has-manys is a *load* hint only — never
  `db.Save(&host)` with `VirtualMachines`/`ReverseProxyHosts` populated; write children per domain.
- **#424 join seeding order**: roles+claims → `role_claims` → users (associations) → rest;
  `vm_snapshots` requires the parent-key flattening described in mapping row 20 (`host_id`/`vm_id`
  are `json:"-"` carriers, they do not exist in legacy JSON — they come from wrapper keys).
- **#424 empty-ID rule**: `helpers.GenerateId` fills blank PKs (User/Role/Claim/VMSnapshot tolerate
  empties in legacy per `omitempty`); everything else fails fast on blank ID.
- Unique-index dedupe duty: `users.username`, `api_keys.key`, `…tokens.token`,
  `user_configs(user_id,slug)` — duplicates in the legacy store must fail migration loudly, not merge.

## Open questions (human in the loop; none block task 01)

1. **`orchestrator_hosts.host` unique?** No grep-visible legacy constraint → shipped as plain
   index; #424 may promote to unique if its loader proves host-keyed dedupe.
2. **`ParallelsVM` DTO family** assumed runtime-only (prlctl-shaped wire types, never a legacy
   JSON collection). If #424's loader enumerates it as persisted, mapping row 24 flips to a table —
   surface before #424 planning.
3. **`configurations` / `database_schemas`** mapped as single-row meta tables for losslessness;
   #424 may replace with `migration_meta` (rows documented so the decision is reversible).
4. **`users.email` uniqueness** deliberately deferred to a plain index (empty-string legacy rows
   would violate UNIQUE); revisit with data profiling in #424.
