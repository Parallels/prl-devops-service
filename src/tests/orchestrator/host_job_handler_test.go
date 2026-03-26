package orchestrator_test

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/handlers"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureRegistrar records handler registrations and exposes the registered
// handler so tests can dispatch events directly without going through
// HostWebSocketManager.
type captureRegistrar struct {
	mu            sync.Mutex
	registrations map[constants.EventType]interfaces.HostEventHandler
}

func newCaptureRegistrar() *captureRegistrar {
	return &captureRegistrar{
		registrations: make(map[constants.EventType]interfaces.HostEventHandler),
	}
}

func (r *captureRegistrar) RegisterHandler(eventTypes []constants.EventType, handler interfaces.HostEventHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, et := range eventTypes {
		r.registrations[et] = handler
	}
}

func (r *captureRegistrar) handlerFor(et constants.EventType) interfaces.HostEventHandler {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.registrations[et]
}

// marshalEvent builds a JSON-encoded EventMessage as a host WebSocket would send.
func marshalEvent(t *testing.T, message string, body interface{}) []byte {
	t.Helper()
	msg := models.NewEventMessage(constants.EventTypeJobManager, message, body)
	b, err := json.Marshal(msg)
	require.NoError(t, err)
	return b
}

// newTestEmitter creates and initializes a real EventEmitter in API mode so
// that IsRunning() returns true. It shuts down the emitter when the test ends.
// Because EventEmitter uses a package-level singleton, all tests in this file
// share the same underlying instance — the emitter is started once on first
// call and reused thereafter.
func newTestEmitter(t *testing.T) *eventemitter.EventEmitter {
	t.Helper()
	t.Setenv(constants.MODE_ENV_VAR, constants.API_MODE)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	emitter := eventemitter.NewEventEmitter(ctx)
	if !emitter.IsRunning() {
		diag := emitter.Initialize()
		require.False(t, diag.HasErrors(), "emitter initialization must not fail")
	}
	return emitter
}

// sharedHandler is the HostJobEventHandler singleton. HostJobEventHandler uses
// sync.Once internally so we create it once here; all tests obtain the handler
// from the registrar after construction.
var (
	sharedRegistrar *captureRegistrar
	sharedHandler   interfaces.HostEventHandler
	handlerOnce     sync.Once
)

func getHandler(t *testing.T) interfaces.HostEventHandler {
	t.Helper()
	handlerOnce.Do(func() {
		sharedRegistrar = newCaptureRegistrar()
		handlers.NewHostJobEventHandler(sharedRegistrar)
		sharedHandler = sharedRegistrar.handlerFor(constants.EventTypeJobManager)
	})
	require.NotNil(t, sharedHandler, "HostJobEventHandler must be registered for EventTypeJobManager")
	return sharedHandler
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestHostJobEventHandler_RegistersForJobManager verifies that constructing the
// handler registers it with EventTypeJobManager.
func TestHostJobEventHandler_RegistersForJobManager(t *testing.T) {
	h := getHandler(t)
	assert.NotNil(t, h)
}

// TestHostJobEventHandler_IgnoresWrongEventType ensures the handler does nothing
// (no panic, no broadcast) when called with a non-JobManager event type.
func TestHostJobEventHandler_IgnoresWrongEventType(t *testing.T) {
	h := getHandler(t)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	payload, err := json.Marshal(models.NewEventMessage(constants.EventTypePDFM, "VM_STATE_CHANGED", nil))
	require.NoError(t, err)

	// Should not panic; emitter may or may not be running but the handler must
	// return before attempting to broadcast.
	assert.NotPanics(t, func() {
		h.Handle(ctx, "host-abc", constants.EventTypePDFM, payload)
	})
}

// TestHostJobEventHandler_HandlesInvalidJSON verifies that malformed JSON does
// not cause a panic. The handler must log the error and return.
func TestHostJobEventHandler_HandlesInvalidJSON(t *testing.T) {
	h := getHandler(t)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	assert.NotPanics(t, func() {
		h.Handle(ctx, "host-abc", constants.EventTypeJobManager, []byte(`{this is not valid json`))
	})
}

// TestHostJobEventHandler_ForwardsJobCreated verifies that a valid JOB_CREATED
// payload is processed without panicking. The handler builds a HostJobEvent and
// calls emitter.Broadcast; actual broadcast delivery to WebSocket clients is
// covered by the EventEmitter's own test suite.
func TestHostJobEventHandler_ForwardsJobCreated(t *testing.T) {
	// Ensure a running emitter is present so the handler can call Broadcast.
	newTestEmitter(t)

	h := getHandler(t)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	payload := marshalEvent(t, "JOB_CREATED", map[string]interface{}{"id": "job-123"})

	assert.NotPanics(t, func() {
		h.Handle(ctx, "host-abc", constants.EventTypeJobManager, payload)
		// Allow the goroutine inside Handle to run.
		time.Sleep(20 * time.Millisecond)
	})
}

// TestHostJobEventHandler_ForwardsJobUpdated verifies that a JOB_UPDATED payload
// is processed without panicking.
func TestHostJobEventHandler_ForwardsJobUpdated(t *testing.T) {
	newTestEmitter(t)

	h := getHandler(t)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	payload := marshalEvent(t, "JOB_UPDATED", map[string]interface{}{"id": "job-456", "state": "running"})

	assert.NotPanics(t, func() {
		h.Handle(ctx, "host-xyz", constants.EventTypeJobManager, payload)
		time.Sleep(20 * time.Millisecond)
	})
}

// TestHostJobEventHandler_ForwardsJobCompleted verifies that a JOB_COMPLETED
// payload is processed without panicking.
func TestHostJobEventHandler_ForwardsJobCompleted(t *testing.T) {
	newTestEmitter(t)

	h := getHandler(t)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	payload := marshalEvent(t, "JOB_COMPLETED", map[string]interface{}{
		"id":                 "job-789",
		"state":              "completed",
		"result_record_id":   "vm-uuid",
		"result_record_type": "virtual_machine",
	})

	assert.NotPanics(t, func() {
		h.Handle(ctx, "host-foo", constants.EventTypeJobManager, payload)
		time.Sleep(20 * time.Millisecond)
	})
}
