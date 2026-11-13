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
// --- #425 wave map (per-function detail lives in LEDGER.md) --------------
//
// wave 02 (auth)     : users_parity / roles_claims_parity / api_key_parity
//                      + filter_differential (R-3 net, no legacy twin)
// wave 03 (infra)    : catalog_parity / reverse_proxy_parity /
//                      snapshots_parity / packer_enrollment_parity
// wave 04 (orch+jobs): orchestrator_parity / jobs_config_parity /
//                      errors_parity (the D-4 matrix as code)
//
// Each *_parity_test.go header carries its scenario list + Tier-1
// sources. LEDGER.md holds (a) the per-function conversion ledger for
// every wave, incl. [E-4] hook-engine seeds, and (b) the D-1 file
// ledger: one row per src/data/*_test.go file with landed-in targets or
// a one-line exclusion rationale (AC: every legacy test file accounted
// for). Engine-specific escapes cite LEDGER ids; free-form t.Skip is a
// review reject.
//
// Behaviours pinned as engine-agreed across waves (findings, not
// defects — LEDGER rows carry detail): create-path ID regeneration
// (catalog manifests, RP hosts/routes, RP routes), catalog "latest"-tag
// + upper-case CatalogId normalisation, SQL-only host-snapshot DTO
// enrichment, "host:port" reverse-proxy lookup key,
// ReplaceOrchestratorHostVMs full-sync field wipe, healthy+enabled
// host-level VM filter, ghost-job boot-posture divergence (legacy
// auto-recovers at Connect; SQL defers to the #421 watcher — terminal
// states converge).
//
// Deferred (#426/decision scope, tally in the feature ledger): T14
// RFC3339Nano ordering, T15 legacy concurrent reorder race (legacy leg
// of rp reorder stays sequential), T12 TempDir flakes → task-07 fix.
// Task 05 appends Postgres constraint legs to errors_parity_test.go's
// shape without edits.
