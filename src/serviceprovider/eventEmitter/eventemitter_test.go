package eventemitter

import (
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventEmitter_Singleton(t *testing.T) {
	// Reset singleton for testing
	globalEventEmitter = nil
	once = sync.Once{}

	ctx := basecontext.NewBaseContext()
	emitter1 := NewEventEmitter(ctx)
	emitter2 := NewEventEmitter(ctx)

	assert.Same(t, emitter1, emitter2, "Should return the same instance")
	assert.NotNil(t, emitter1.ctx, "Context should be initialized")
}

func TestEventEmitter_Initialize_APIMode(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	assert.True(t, emitter.IsRunning(), "Should be running in API mode")
	assert.NotNil(t, emitter.hub, "Hub should be created")
	assert.NotNil(t, emitter.ctx, "Context should be set")
}

func TestEventEmitter_Initialize_OrchestratorMode(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "orchestrator")
	defer cleanup()

	assert.True(t, emitter.IsRunning(), "Should be running in Orchestrator mode")
	assert.NotNil(t, emitter.hub, "Hub should be created")
}

func TestEventEmitter_Initialize_CLIMode(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	// Use t.Setenv for automatic cleanup
	t.Setenv(constants.MODE_ENV_VAR, "cli")

	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)
	diag := emitter.Initialize()

	assert.NotNil(t, diag, "Diagnostics should be returned")
	assert.False(t, diag.HasErrors(), "Should not have errors")
	assert.False(t, emitter.IsRunning(), "Should NOT be running in CLI mode")
	assert.Nil(t, emitter.hub, "Hub should not be created")
}

func TestEventEmitter_Initialize_AlreadyRunning(t *testing.T) {
	emitter, cleanup := setupTestEmitterWithMode(t, "api")
	defer cleanup()

	require.True(t, emitter.IsRunning())

	// Try to initialize again
	diag2 := emitter.Initialize()
	assert.True(t, diag2.HasWarnings(), "Should have warning about already running")
	assert.Equal(t, 1, diag2.GetWarningCount(), "Should have exactly one warning")
}

func TestEventEmitter_Shutdown(t *testing.T) {
	emitter, _ := setupTestEmitterWithMode(t, "api")

	require.True(t, emitter.IsRunning())

	emitter.Shutdown()
	assert.False(t, emitter.IsRunning(), "Should not be running after shutdown")

	// Note: We don't call cleanup() here since we manually shut down
	// Reset singleton for other tests
	globalEventEmitter = nil
	once = sync.Once{}
}

func TestEventEmitter_IsRunning(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	ctx := basecontext.NewBaseContext()
	emitter := NewEventEmitter(ctx)

	assert.False(t, emitter.IsRunning(), "Should not be running initially")
}

func TestEventEmitter_Get(t *testing.T) {
	// Reset singleton
	globalEventEmitter = nil
	once = sync.Once{}

	ctx := basecontext.NewBaseContext()
	NewEventEmitter(ctx)

	retrieved := Get()
	assert.NotNil(t, retrieved, "Get() should return the singleton instance")
	assert.Same(t, globalEventEmitter, retrieved, "Should return the same instance")
}
