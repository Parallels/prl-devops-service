//go:build integration

package data

// ResetForTesting resets the in-memory database singleton so each integration
// test starts from a clean slate. Must only be called from tests.
func ResetForTesting() {
	memoryDatabase = nil
}
