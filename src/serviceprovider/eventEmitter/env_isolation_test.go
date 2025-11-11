package eventemitter

import (
	"os"
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
)

// TestEnvironmentIsolation verifies that t.Setenv properly isolates tests
// and doesn't pollute the development environment
func TestEnvironmentIsolation(t *testing.T) {
	// Get original value before any test modification
	originalValue := os.Getenv(constants.MODE_ENV_VAR)

	t.Run("SetAPIMode", func(t *testing.T) {
		// This should only affect this test
		t.Setenv(constants.MODE_ENV_VAR, "api")
		assert.Equal(t, "api", os.Getenv(constants.MODE_ENV_VAR))
	})

	t.Run("SetOrchestratorMode", func(t *testing.T) {
		// This should only affect this test
		t.Setenv(constants.MODE_ENV_VAR, "orchestrator")
		assert.Equal(t, "orchestrator", os.Getenv(constants.MODE_ENV_VAR))
	})

	// After subtests complete, value should be restored to original
	currentValue := os.Getenv(constants.MODE_ENV_VAR)
	assert.Equal(t, originalValue, currentValue,
		"Environment variable should be restored after subtests")
}

// TestEnvironmentCleanup verifies cleanup happens even if test fails
func TestEnvironmentCleanup(t *testing.T) {
	originalValue := os.Getenv(constants.MODE_ENV_VAR)

	t.Run("TestThatChangesEnv", func(t *testing.T) {
		t.Setenv(constants.MODE_ENV_VAR, "test-value")
		// Simulate test doing work
		assert.Equal(t, "test-value", os.Getenv(constants.MODE_ENV_VAR))
	})

	// Verify cleanup happened
	assert.Equal(t, originalValue, os.Getenv(constants.MODE_ENV_VAR),
		"Environment should be restored after test completes")
}
