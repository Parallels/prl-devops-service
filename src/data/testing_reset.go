//go:build !release

package data

// ResetForTesting resets the in-memory database singleton so each integration
// test starts from a clean slate. Must only be called from tests.
//
// #421 task 03: tag widened from `//go:build integration` to `//go:build
// !release` — the parity harness (src/database/parity) needs engine
// isolation in plain unit-test builds; the function remains test-only by
// convention (the parity package's doc.go records the exception). Release
// builds keep it compiled out.
//
// Deprecated: legacy JSON data layer (#426); use the GORM-backed database.GormDatabase via serviceprovider.GetDatabaseService() — scheduled for removal in v1.3.0.
func ResetForTesting() {
	memoryDatabase = nil
}
