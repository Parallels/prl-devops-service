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
