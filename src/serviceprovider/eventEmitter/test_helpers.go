package eventemitter

import (
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/require"
)

// setupTestEmitterWithMode creates a test emitter in the specified mode
func setupTestEmitterWithMode(t *testing.T, mode string) (*EventEmitter, func()) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	// Set mode using t.Setenv for automatic cleanup (Go 1.17+)
	// This is safer than manual os.Setenv as it:
	// - Automatically restores the value after test
	// - Is safe for parallel tests
	// - Prevents environment pollution
	t.Setenv(constants.MODE_ENV_VAR, mode)

	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)
	diag := emitter.Initialize()
	require.False(t, diag.HasErrors(), "Initialization should not have errors")

	if mode == constants.API_MODE || mode == constants.ORCHESTRATOR_MODE {
		require.True(t, emitter.IsRunning(), "Emitter should be running in API/Orchestrator mode")
	}

	// Return cleanup function for emitter only
	// t.Setenv handles environment variable cleanup automatically
	return emitter, func() {
		if emitter.IsRunning() {
			emitter.Shutdown()
		}
	}
}
