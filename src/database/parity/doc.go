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
