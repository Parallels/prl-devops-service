// Package testutil is the exported harness for the #425 data-layer parity
// suite (and any future database-layer test). It exists because the AC's
// "file::memory:?cache=shared" is ONE logical DB per process: unguarded,
// every parity test would silently share state with every other.
//
// # Isolation rules (D-3 — review checklist; sweeps in task 07)
//
//  1. One DB per test. Every test opens its own uniquely-named memory DB
//     (OpenMemoryDB / memoryDSN); never share handles across tests.
//  2. No cross-test globals in new code. src/database/** and parity/**
//     never touch the serviceprovider singleton Get(), never assign data.memoryDatabase
//     (can't — unexported, wrong package). The parity suite's legacy leg
//     constructs local JsonDatabases exactly like setupTestDB and never
//     relies on the singleton being nil.
//  3. Config: only via WithCleanConfig / t.Setenv; tests using them are
//     sequential (t.Setenv forbids t.Parallel). New packages read config
//     through the #420 FromApplicationConfig seam, never config.Get() at
//     package-init.
//  4. Order-independence gates: new suites must pass
//     `go test -count=5 ./database/...` and `go test -shuffle=on ./database/...`.
//     The LEGACY suite gets the same runs informationally only; any legacy
//     order-dependency is recorded, not fixed (Tier 1).
//  5. No disk for the GORM legs (memory DSN) — the one honest exception is
//     migration fixtures: data.json files in t.TempDir() are INPUTS, not
//     test-written DB state (see task 06).
//
// # DSN scheme (D-2)
//
// memoryDSN(t) = "file:t425-<sanitized-test-name>-<seq>?mode=memory&cache=shared"
// — a NAMED shared-cache DB: every connection to the same name shares, every
// other name is isolated, nothing touches disk. t.Cleanup closes the handle
// (a named memory DB otherwise lives until EVERY connection closes). The
// #418 factory passes file:-URIs through verbatim (#425 task 01 guard).
//
// # Usage
//
//	db := testutil.OpenMemoryDB(t)   // isolated, closed via t.Cleanup
//	testutil.MigrateSchema(t, db)     // #419 AutoMigrate entrypoint
//	ctx := testutil.WithCleanConfig(t)
//	testutil.RunOnEngines(t, scenarios, testutil.LegacyEngine(), testutil.GormEngine())
package testutil
