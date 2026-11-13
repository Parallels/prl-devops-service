# Parity conversion ledger (PBI #425)

Single ledger for the `database/parity` scenario conversions (moved out of
`doc.go` at task 04 — the doc-size gate allows the split; doc.go keeps
methodology + pointers). Two sections: the per-function conversion ledger
(what landed where, incl. [E-4] hook-engine seeds) and the D-1 closure
ledger (one row per `src/data/*_test.go` file — AC "every legacy test file
accounted for"). Engine-specific escapes cite IDs here; free-form `t.Skip`
is a review reject.

## Per-function conversion ledger

### Task 02 — auth wave

Scenario bodies transcribed per the recipe with sentinels/arrangements
preserved; ctx posture NewRootBaseContext → tplCtxRoot(), NewBaseContext →
authCtxPlain() (both nil-equivalent postures verified in authorize_layer).

users_parity_test.go — data/users_test.go:
TestCreateUser→users/create_user · TestCreateUserFailsMissingFields→
users/create_user_fails_missing_fields · TestUpdateUser→users/
update_user (10 ms sleep kept) · TestUpdateUserBlockStatus→users/
update_user_block_status · TestAddAndRemoveRoleToUser→users/
add_and_remove_role_to_user (extended: embedded role Claims slice
round-trip, R-2 canary) · TestDeleteUser→users/delete_user

roles_claims_parity_test.go — data/roles_test.go + claims_test.go +
roles_claims_test.go: TestCreateRole/TestUpdateRole/TestDeleteRole/
TestDeleteInternalRoleFailsWithoutRoot→roles/* (all 1:1); TestCreateClaim
/TestUpdateClaim/TestDeleteClaim/TestDeleteInternalClaimFailsWithoutRoot→
claims/* (all 1:1); TestAddClaimToRole_{Success,Duplicate,RoleNotFound,
ClaimNotFound} + TestRemoveClaimFromRole_{Success,ClaimNotOnRole,
RoleNotFound}→roles_claims/* (all 1:1, ErrorIs kept). Lines 115-236
(Cascades, FailsWhenAssigned, GetRole(s)_IncludesClaims,
CreateRole_With{,Unknown}Claim) are NOT in the task-02 inventory — they
stay Tier 1 until their wave (covered by suite_roles_claims since #421).

api_key_parity_test.go — data/api_key_test.go: TestCreateApiKeyWith
Expiration→api_keys/create_with_expiration · …WithoutExpiration→api_keys/
create_without_expiration · TestApiKeyExpirationLogic→api_keys/
expiration_logic (no-DB pure leg, both engines) · TestApiKeyWithUser
Ownership→api_keys/with_user_ownership · TestGetApiKeysWithUserFiltering→
api_keys/get_api_keys_with_user_filtering · TestCreateApiKeyWithout
UserIdAutoAssignsToUser→api_keys/create_without_user_id_does_not_auto_
assign · TestCreateApiKeyDefaultTypeIsExternal→api_keys/create_default_
type_is_external · TestCreateApiKeyExplicitTypePreserved→api_keys/
create_explicit_type_preserved · TestGetApiKeyNormalizesType→api_keys/
get_api_key_normalizes_type [E-4] · TestGetApiKeysFiltersInternalKeys→
api_keys/get_api_keys_filters_internal_keys · TestGetApiKeysNormalizes
Types→api_keys/get_api_keys_normalizes_types [E-4] ·
TestInternalKeysNotInListButRetrievableByID→api_keys/internal_keys_not_in_
list_but_retrievable_by_id · TestBackwardCompatibilityWithOldApiKeys→
api_keys/backward_compatibility_with_old_api_keys [E-4].
Stay Tier 1 (pure domain fns, engine-agnostic): TestNormalizeApiKeyType,
TestFilterInternalKeys.

filter_differential_test.go — no legacy counterpart (R-3 net): 20
exact-sequence filter expressions (users+keys, regex/case/parse-error
corpus) + 3 catalog authz postures (nil-auth / no-roles / InjectedRoles
operator — RequiredRoles pairs with HasEffectiveRole, verified
authorize_layer.go), non-vacuity-guarded. Order runs reduced per E-3
rationale (file header).

### Task 03 — infra wave

catalog_parity_test.go — data/catalog_test.go: TestCreateCatalogManifest
+TestUpdateCatalogManifest→catalog/create_read_update_roundtrip (merged;
the sleep-10ms UpdatedAt-must-move arrangement kept) ·
TestDeleteCatalogManifest→catalog/delete_manifest ·
TestUpdateCatalogManifestTags→catalog/serialised_array_columns_fidelity
(extended: RequiredRoles/RequiredClaims exact round-trip + append-order
pin) · TestTaintAndRevokeCatalogManifestVersion→catalog/
taint_untaint_revoke_version; data/catalog_managers_test.go:
Test{Add,Update}CatalogManager→managers/add_update_get ·
TestDeleteCatalogManager→managers/delete · Test{Update,Delete}Internal
…Fails…→managers/internal_guards_without_root (both guards, one ctx
posture). PINNED engine-agreed behaviours: CreateCatalogManifest
REGENERATES the ID (content hash) and normalises CatalogId to UPPER
CASE, and appends the "latest" tag after supplied tags — scenarios key
survivors by CatalogId, never caller IDs. TestParityCatalogOrder
Differential — no legacy counterpart (D-5): identical out-of-order
UpdatedAt stamps pre-seeded on both engines (E-4 pre-Write vs raw
INSERT), GetCatalogManifests sequences asserted engine-equal AND pinned
UpdatedAt-desc.

reverse_proxy_parity_test.go — data/reverse_proxy_test.go: TestEnable
AndDisableProxyConfig→rp/config_enable_disable · TestUpdateReverseProxy→
rp/config_update_readback · TestCreateAndGet+TestUpdate+TestConfigure
TLSAndCORS→rp/host_create_update_tls_cors · (new coverage:
rp/tcp_route_on_clean_host pins the guard-order 400/404 sentinel edge) ·
TestReorderAndDeleteReverseProxyHostHttpRoute→rp/routes_reorder_and_
compacting_delete (gapless {1,2} after delete) ·
rp/reorder_concurrent_invariant = #421 TestReorderInvariant_Concurrent
rebuilt through DatabaseService methods only (legacy leg sequential —
T15 route-reorder race pinned at the row after T14; SQL leg concurrent) ·
rp/error_edges pins 3 D-4 sentinels. PINNED: Create{ReverseProxyHost,
HostHttpRoute} REGENERATE entity IDs on both engines — all follow-up
calls use returned IDs.

snapshots_parity_test.go — data/vm_snapshots_test.go: TestSetListVM
SnapshotsByVMId+TestGetListVMSnapshotsByVMId→snapshots/vm_list_replace_
per_vm (legacy poked a bare &JsonDatabase{connected:true}; converted to
the public ctx-free contract; adds replace-semantics + unknown-VM⇒empty
legs). PARITY FINDING (pinned per-engine, see file): snapshots/
host_upsert_idempotent — GetHostVMSnapshots on SQL ENRICHES each DTO
with HostID+VMID; legacy JSON returns caller values untouched (#421
suite normalised the fields away). snapshots/host_unknown_sentinel new.

packer_enrollment_parity_test.go — data/packer_template_test.go:
TestAddPackerTemplate→packer/add_get · TestUpdatePackerTemplate→packer/
update · TestDeletePackerTemplate→packer/delete · (spec "dup-name" lands
on the real guard: duplicate REJECTED BY ID, packer/duplicate_rejected) ·
TestDeleteInternalPackerTemplateFailsWithoutRoot→packer/
internal_delete_guard. Enrollment (no legacy Tier-1 tests; #421 SQL pins
via public API + E-4 pre-seed, spec §4 ±s no nowFn needed): create_
validate_roundtrip · used_gate_wins (Used beats expiry, orchestrator_
enrollment_token.go:14ff) · expired_edge · empty_and_garbage_stay_valid ·
sweep_prune_matrix (expired/used pruned; empty+unparseable expiries and
fresh survive — 5-token matrix, both engines).

### Task 04 — orchestrator + jobs/config + D-4 error matrix

orchestrator_parity_test.go (TestParityOrchestratorInfraWave) —
data/orchestrator_test.go (7 scenario legs; the remaining 34 originals
ride the #421 DomainSuite files suite_orchestrator_hosts/_vms —
split row in the file ledger):
orch_hosts/create_guard (empty name/id-or-host 500-code sentinels) ·
orch_hosts/host_roundtrip (create/get/update; no-diff Update does NOT
re-stamp; UpdateOrchestratorHostTimestamp verbatim + ⇒healthy clears
unhealthy fields) · orch_hosts/active_filter (Enabled-only, state-
agnostic golden) · orch_hosts/update_edges · orch_hosts/
resources_golden (HostResources aggregation golden {12,12,24,750} etc.,
D-3-concurrency legs excluded stay #421) · orch_hosts/heartbeat_result_
equality · orch_hosts/rp_views (GetOrchestratorReverseProxyHost key =
"host:port" — PINNED both engines; bare hostname 404s both) ·
orch_vms/upsert_replace_cascade (3-upsert list, per-VM state/uptime
updates, unknown-VM 404, ReplaceOrchestratorHostVMs FULL-SYNC field
wipe — vm-2 State/Uptime zeroed after Replace, PINNED; host-level list
observes the same rows through the healthy+enabled filter — legacy
fixture posture :366-369; DeleteOrchestratorHost cascades VM rows).
Scenario ctx posture: tplCtxRoot() — orchestrator reads ride the filter
layer and non-root ctx empties lists on BOTH engines (found this slice;
not a divergence). NOT converted: TestDeleteOrchestratorVirtualMachine
no-op-bug leg (legacy :352-359 documents the no-op — asserting it cross-
engine would pin a known bug; stays Tier 1) and
TestHostSelectionPreservesResourceFailuresInJobs (T12 TempDir flake,
task 07 fix).

jobs_config_parity_test.go (TestParityJobsConfigInfraWave +
TestParityGhostJobTimeout + TestParityAD4BothTablesExist) —
data/jobs_timeout_test.go + config_test.go:
jobs/state_walk (caller-assigned IDs; running walk; forwarded-failure
Error==Message persistence; owner scoping; ErrJobNotFound; DeleteJob("")
silently-nil quirk) · jobs/delete_by_state (DeleteJobsByState count
parity + multi-state form) · userconfig/owner_scoping (EqualFold owner
match, A⇏B leakage, cross-owner same-slug allowed, dup ⇒
ErrUserConfigAlreadyExists) · config/singleton_and_schema (SetId-before-
GetConfiguration ordering — GetConfiguration create-on-read seeds a
random ID that then locks SetId; "ID already exists" message pin;
UpdateSchemaVersion in-place upsert; SeedId≢overwrite).
TestParityGhostJobTimeout — the E-4 stale matrix pre-seed exposed a
BOOT-POSTURE DIVERGENCE (pinned per-engine, terminal-convergence is the
shared claim): legacy Connect AUTO-converts every running/pending row to
failed at boot (its Tier-1 ghost test mutates internals after Connect
because of exactly this), while the SQL engine defers to the explicit
RecoverOngoingJobs/DetectStaleJobs legs (#421 jobs_watcher.go) — so the
ghost timeout is the PRIMARY recovery path in SQL mode. Timeout env is
set in the test-parent scope BEFORE engine creation (config.Get() is a
process-cached singleton; in-hook t.Setenv reaches only whichever engine
rebuilds later — flaky by construction). GORM leg pins: 2 notifications
(stale pending+running), reason carries timeout + last-update stamp,
Error==Message, Progress/Steps/IsOrchestratorJob/ResultRecordId survive
the timeout write, recent/terminal rows untouched, no re-fire.
TestParityAD4BothTablesExist — SQL-leg-only raw-Migrator guard (AD-4):
schema_metadata + configurations coexist post-write; justified gorm-
only (legacy has no schema_metadata table).

errors_parity_test.go (TestParityErrorsMatrixD4) — plan D-4 as CODE,
every row once, quirks mirrored NOT fixed: 16 not-found 404 rows × both
engines (identity + Code()==404) · stdlib no-code quirk
ErrCatalogManagerNotFound (implements NO Code() surface — d4Code -1) ·
5 duplicate-400 rows (ErrUserAlreadyExists needs default roles seeded —
CreateUser assigns them; ErrPackerTemplateAlreadyExists; role/claim +
user/role + user/claim dup-adds) · quirk duplicates ErrApiKeyAlready-
Exists(500), ErrUserConfigAlreadyExists(500), ErrCatalogAlreadyExists
(404, DEAD sentinel — metadata pinned, call==nil documents it) ·
errors/unreachable_offline (engine-agnostic): factory wrap
errors.Is(ErrDatabaseUnreachable) via loopback :1 postgres DSN, password
redacted (RedactDSN), host:port kept for triage. Postgres constraint
legs append in task 05 WITHOUT edits to this shape.
PINNED: ReverseProxyHostTcpRoute carries no Host/Port (target-only);
GetCatalogManager is ctx-free (relocated #421 D-3).

## D-1 file ledger (one row per data/*_test.go — every file accounted)

| Tier-1 file | fn | Landed in | Split / exclusion rationale |
|---|---|---|---|
| api_key_test.go | 15 | api_key_parity, suite_api_keys | 13 converted (02); TestNormalizeApiKeyType/TestFilterInternalKeys = pure domain fns, engine-agnostic |
| catalog_managers_test.go | 5 | catalog_parity (managers/*) | all 5 converted (03) |
| catalog_test.go | 5 | catalog_parity, suite_catalog | all converted (03) + order-differential (new, D-5) |
| claims_test.go | 4 | roles_claims_parity, suite_roles_claims | all 4 converted (02) |
| concurrency_test.go | 4 | — | D-3: exercises legacy mutex/SaveNow locking semantics — no SQL analogue; concurrency legs ride rp/reorder_concurrent_invariant + #421 suites |
| config_test.go | 3 | jobs_config_parity (config/*), suite_config_schema | all 3 legs converted (04); AD-4 both-tables guard added (gorm-only) |
| interfaces_compilation_test.go | 1 | — | compile-only structural guarantee, no runtime behaviour to mirror |
| jobs_timeout_test.go | 2 | jobs_config_parity (ghost_timeout), suite_jobs | matrix adapted for boot-posture divergence (pinned above) |
| legacy_load_test.go | 9 | — | file-format loader tests (data.json ingest) — file-engine-only, no SQL counterpart |
| main_test.go | 13 | engine_factory_test, harness_test (lifecycle) | 11/13 = file-engine lifecycle: TestNewJsonDatabase/Connect/LoadSave/SaveAs*/Locking/ProcessSaveQueue/Load*/IsDataFileEmpty/Recovery*/SaveNow* = file-engine-only (TestSaveAs* = backup file ops); engine lifecycle covered by harness/engine_factory |
| orchestrator_test.go | 41 | orchestrator_parity (7 legs), suite_orchestrator_hosts/_vms | biggest split: DomainSuite files carry 32; wave adds 7; DeleteOrchestratorVirtualMachine no-op-bug leg stays Tier 1; T12-flaky host-selection leg → task 07 |
| packer_template_test.go | 4 | packer_enrollment_parity, suite_packer_templates | all converted (03); dup leg lands on real ID guard |
| reverse_proxy_test.go | 7 | reverse_proxy_parity, suite_reverse_proxy | all converted (03); concurrent leg split per T15 |
| roles_claims_test.go | 13 | roles_claims_parity, suite_roles_claims | 7 relation legs converted (02); cascade/includes block (lines 115-236) carried by suite_roles_claims since #421 |
| roles_test.go | 4 | roles_claims_parity | all 4 converted (02) |
| users_test.go | 6 | users_parity, suite_users | all 6 converted (02); root-protection (E-1) → D-4 matrix scope |
| vm_snapshots_test.go | 2 | snapshots_parity | both converted (03) + enrichment divergence pinned |
