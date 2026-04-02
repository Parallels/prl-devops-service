package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/controllers"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupDB creates a temp-dir backed JSON database, registers it with a fresh
// mock provider and returns a cleanup function.
func setupDB(t *testing.T) func() {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test_db.json")
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	db := data.NewJsonDatabase(ctx, dbPath)
	require.NoError(t, db.Connect(ctx))

	sp := serviceprovider.NewMockProvider()
	sp.JsonDatabase = db
	jobs.New(ctx)

	return func() {
		// Nothing to clean up — the temp dir is removed by t.TempDir automatically.
	}
}

func waitForJobToSettle(t *testing.T, jobID string) {
	t.Helper()

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	db := serviceprovider.Get().JsonDatabase
	require.NotNil(t, db)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		job, err := db.GetJob(ctx, jobID)
		if err == nil && job.State != constants.JobStatePending && job.State != constants.JobStateRunning && job.State != constants.JobStateInit {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("job %s did not settle before timeout", jobID)
}

// authRequest creates a request carrying a user in the authorization context.
func authRequest(t *testing.T, method, target string, body []byte) *http.Request {
	t.Helper()
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, target, http.NoBody)
	}
	req.Header.Set("Content-Type", "application/json")

	authCtx := &basecontext.AuthorizationContext{
		IsAuthorized: true,
		AuthorizedBy: "test",
		User: &models.ApiUser{
			ID:    "test-user-id",
			Name:  "Test User",
			Email: "test@example.com",
		},
	}
	ctx := context.WithValue(req.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authCtx)
	return req.WithContext(ctx)
}

// unauthRequest creates a request without any authorization context.
func unauthRequest(t *testing.T, method, target string, body []byte) *http.Request {
	t.Helper()
	if body != nil {
		return httptest.NewRequest(method, target, bytes.NewReader(body))
	}
	return httptest.NewRequest(method, target, http.NoBody)
}

// validBody returns a minimal valid CreateVirtualMachineRequest with name and architecture.
// The orchestrator async handlers accept any valid combination (not restricted to catalog only).
func validBody(t *testing.T) []byte {
	t.Helper()
	req := models.CreateVirtualMachineRequest{
		Name:         "test-vm",
		Architecture: "arm64",
	}
	b, err := json.Marshal(req)
	require.NoError(t, err)
	return b
}

// invalidBodyMissingName returns a request body missing the required Name field.
func invalidBodyMissingName(t *testing.T) []byte {
	t.Helper()
	req := models.CreateVirtualMachineRequest{
		Architecture: "arm64",
	}
	b, err := json.Marshal(req)
	require.NoError(t, err)
	return b
}

// --- AsyncCreateOrchestratorVirtualMachineHandler ---

func TestAsyncCreateOrchestratorVirtualMachineHandler_NoUser(t *testing.T) {
	cleanup := setupDB(t)
	defer cleanup()

	req := unauthRequest(t, http.MethodPost, "/v1/orchestrator/machines/async", validBody(t))
	w := httptest.NewRecorder()

	controllers.AsyncCreateOrchestratorVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAsyncCreateOrchestratorVirtualMachineHandler_InvalidBody(t *testing.T) {
	cleanup := setupDB(t)
	defer cleanup()

	req := authRequest(t, http.MethodPost, "/v1/orchestrator/machines/async", []byte(`{not valid json`))
	w := httptest.NewRecorder()

	controllers.AsyncCreateOrchestratorVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAsyncCreateOrchestratorVirtualMachineHandler_InvalidRequest_MissingName(t *testing.T) {
	cleanup := setupDB(t)
	defer cleanup()

	req := authRequest(t, http.MethodPost, "/v1/orchestrator/machines/async", invalidBodyMissingName(t))
	w := httptest.NewRecorder()

	controllers.AsyncCreateOrchestratorVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAsyncCreateOrchestratorVirtualMachineHandler_NoJobManager(t *testing.T) {
	// Provider with no database — jobs.Get returns nil → 500.
	serviceprovider.NewMockProvider()

	req := authRequest(t, http.MethodPost, "/v1/orchestrator/machines/async", validBody(t))
	w := httptest.NewRecorder()

	controllers.AsyncCreateOrchestratorVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAsyncCreateOrchestratorVirtualMachineHandler_Success(t *testing.T) {
	cleanup := setupDB(t)
	defer cleanup()

	req := authRequest(t, http.MethodPost, "/v1/orchestrator/machines/async", validBody(t))
	w := httptest.NewRecorder()

	controllers.AsyncCreateOrchestratorVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp models.JobResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotEmpty(t, resp.ID, "response must contain a job ID")
	assert.Equal(t, "orchestrator", resp.JobType)
	assert.Equal(t, "create", resp.JobOperation)
	waitForJobToSettle(t, resp.ID)
}

// --- AsyncCreateOrchestratorHostVirtualMachineHandler ---

// withHostID injects a gorilla/mux "id" path variable into the request context.
func withHostID(r *http.Request, hostID string) *http.Request {
	return mux.SetURLVars(r, map[string]string{"id": hostID})
}

func TestAsyncCreateOrchestratorHostVirtualMachineHandler_NoUser(t *testing.T) {
	cleanup := setupDB(t)
	defer cleanup()

	req := unauthRequest(t, http.MethodPost, "/v1/orchestrator/hosts/host-1/machines/async", validBody(t))
	req = withHostID(req, "host-1")
	w := httptest.NewRecorder()

	controllers.AsyncCreateOrchestratorHostVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAsyncCreateOrchestratorHostVirtualMachineHandler_InvalidBody(t *testing.T) {
	cleanup := setupDB(t)
	defer cleanup()

	req := authRequest(t, http.MethodPost, "/v1/orchestrator/hosts/host-1/machines/async", []byte(`{bad json`))
	req = withHostID(req, "host-1")
	w := httptest.NewRecorder()

	controllers.AsyncCreateOrchestratorHostVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAsyncCreateOrchestratorHostVirtualMachineHandler_InvalidRequest_MissingName(t *testing.T) {
	cleanup := setupDB(t)
	defer cleanup()

	req := authRequest(t, http.MethodPost, "/v1/orchestrator/hosts/host-1/machines/async", invalidBodyMissingName(t))
	req = withHostID(req, "host-1")
	w := httptest.NewRecorder()

	controllers.AsyncCreateOrchestratorHostVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAsyncCreateOrchestratorHostVirtualMachineHandler_Success(t *testing.T) {
	cleanup := setupDB(t)
	defer cleanup()

	req := authRequest(t, http.MethodPost, "/v1/orchestrator/hosts/host-1/machines/async", validBody(t))
	req = withHostID(req, "host-1")
	w := httptest.NewRecorder()

	controllers.AsyncCreateOrchestratorHostVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var resp models.JobResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotEmpty(t, resp.ID, "response must contain a job ID")
	assert.Equal(t, "orchestrator", resp.JobType)
	assert.Equal(t, "create", resp.JobOperation)
	waitForJobToSettle(t, resp.ID)
}
