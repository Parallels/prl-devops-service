// Package parity is the two-engine differential harness of PBI #421 (plan
// D-5): one seed function runs against the legacy JSON engine and the SQL
// engine, and Checks compare observable outputs — R-3's structural mitigation
// that every domain task (05-16) plugs a DomainSuite into.
//
// TEST-ONLY DATA IMPORT EXCEPTION (plan AD-1): this package imports
// …/data (JsonDatabase) BY DESIGN — the legacy engine IS the oracle. That
// import must never appear outside _test scopes in src/database/**; the
// #426 bypass audit enforces the rule for production code.
//
// The legacy engine is a process-wide singleton factory (data.NewJsonDatabase,
// data/main.go), so legacy engines cannot run in parallel within one process;
// the harness serialises them with a mutex and DomainSuite tests must not
// call t.Parallel().
package parity

// --- #425 scenario harness (task 01) -------------------------------------
//
// On top of the DomainSuite machinery, #425 adds scenario-per-engine
// tests via the exported testutil package:
// testutil.RunOnEngines(t, []testutil.Scenario{...},
// testutil.LegacyEngine(), testutil.GormEngine()) →
// subtests <scenario>/legacy and <scenario>/gorm, each on an ISOLATED
// store (legacy: ResetForTesting + fresh t.TempDir, mutex-serialised,
// exactly like NewLegacyEngine; gorm: unique memory DSN per
// testutil/doc.go D-2 scheme — never the shared-cache AC literal).
//
// Adding a scenario: (1) use only the DatabaseService surface + data/*
// sentinels; (2) engine-specific escapes cite an EXACT ledger entry id —
// free-form t.Skip is a review reject; (3) no t.Parallel (t.Setenv in
// testutil.WithCleanConfig); (4) self-cleaning is belt-and-braces even
// though legacy legs are isolated — later waves reuse seed ids.
//
// Exclusion ledger (DRAFT — task 04 finalises):
//   - E-1 root protection: DeleteUser refuses Name/Username=="root" on
//     BOTH engines; scenarios must not seed root (its parity behaviour is
//     a task-04 error-matrix item).
//   - E-2 memory URIs skip the WAL verify + journal pragma (memory DBs
//     report journal_mode=memory); foreign_keys(ON) IS applied to match
//     production SQL boots.
//   - E-3 catalog order-parity is set-based: list comparison sorts by
//     business key and head-addressing probes assert per-engine self-
//     consistency, because per-engine ORDER over wall-clock UpdatedAt is a
//     clock race, not a parity signal — helpers.GetUtcCurrentDateTime
//     returns RFC3339Nano, whose trailing-zero trimming makes the stored
//     string lexicographically ≠ chronological (backlog T14, production
//     scope). Order SEMANTICS parity is covered deterministically by the
//     order-layer unit tests on both sides (shared tiebreak contract).
//   - E-4 pre-Type rows (empty ApiKey.Type): unreachable through
//     CreateApiKey on either engine (both default "external" at create),
//     so the old-data scenarios seed them engine-faithfully on the
//     seed_auth_test.go hook engines — legacy = pre-Write data.json
//     (loader ingests verbatim), gorm = raw INSERT post-migrate (same
//     shape legacy_migrator lands). The legacy originals poked private
//     db.data.ApiKeys; assertions here are those originals verbatim.
//
// --- #425 task 02 conversion ledger (auth wave) --------------------------
// Legacy fn (file) → scenario. "1:1" = body transcribed per the recipe
// with sentinels/arrangements preserved; ctx posture NewRootBaseContext →
// tplCtxRoot(), NewBaseContext → authCtxPlain() (both nil-equivalent
// postures verified in authorize_layer).
//
// users_parity_test.go — data/users_test.go:
//   TestCreateUser→users/create_user · TestCreateUserFailsMissingFields→
//   users/create_user_fails_missing_fields · TestUpdateUser→users/
//   update_user (10 ms sleep kept) · TestUpdateUserBlockStatus→users/
//   update_user_block_status · TestAddAndRemoveRoleToUser→users/
//   add_and_remove_role_to_user (extended: embedded role Claims slice
//   round-trip, R-2 canary) · TestDeleteUser→users/delete_user
//
// roles_claims_parity_test.go — data/roles_test.go + claims_test.go +
// roles_claims_test.go: TestCreateRole/TestUpdateRole/TestDeleteRole/
// TestDeleteInternalRoleFailsWithoutRoot→roles/* (all 1:1); TestCreateClaim
// /TestUpdateClaim/TestDeleteClaim/TestDeleteInternalClaimFailsWithoutRoot→
// claims/* (all 1:1); TestAddClaimToRole_{Success,Duplicate,RoleNotFound,
// ClaimNotFound} + TestRemoveClaimFromRole_{Success,ClaimNotOnRole,
// RoleNotFound}→roles_claims/* (all 1:1, ErrorIs kept). Lines 115-236
// (Cascades, FailsWhenAssigned, GetRole(s)_IncludesClaims,
// CreateRole_With{,Unknown}Claim) are NOT in the task-02 inventory — they
// stay Tier 1 until their wave.
//
// api_key_parity_test.go — data/api_key_test.go: TestCreateApiKeyWith
// Expiration→api_keys/create_with_expiration · …WithoutExpiration→api_keys/
// create_without_expiration · TestApiKeyExpirationLogic→api_keys/
// expiration_logic (no-DB pure leg, both engines) · TestApiKeyWithUser
// Ownership→api_keys/with_user_ownership · TestGetApiKeysWithUserFiltering→
// api_keys/get_api_keys_with_user_filtering · TestCreateApiKeyWithout
// UserIdAutoAssignsToUser→api_keys/create_without_user_id_does_not_auto_
// assign · TestCreateApiKeyDefaultTypeIsExternal→api_keys/create_default_
// type_is_external · TestCreateApiKeyExplicitTypePreserved→api_keys/
// create_explicit_type_preserved · TestGetApiKeyNormalizesType→api_keys/
// get_api_key_normalizes_type [E-4] · TestGetApiKeysFiltersInternalKeys→
// api_keys/get_api_keys_filters_internal_keys · TestGetApiKeysNormalizes
// Types→api_keys/get_api_keys_normalizes_types [E-4] ·
// TestInternalKeysNotInListButRetrievableByID→api_keys/internal_keys_not_in_
// list_but_retrievable_by_id · TestBackwardCompatibilityWithOldApiKeys→
// api_keys/backward_compatibility_with_old_api_keys [E-4].
// Stay Tier 1 (pure domain fns, engine-agnostic): TestNormalizeApiKeyType,
// TestFilterInternalKeys.
//
// filter_differential_test.go — no legacy counterpart (R-3 net): 20
// exact-sequence filter expressions (users+keys, regex/case/parse-error
// corpus) + 3 catalog authz postures (nil-auth / no-roles / InjectedRoles
// operator — RequiredRoles pairs with HasEffectiveRole, verified
// authorize_layer.go), non-vacuity-guarded. Order runs reduced per E-3
// rationale (file header).
