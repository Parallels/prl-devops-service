package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/controllers"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSyncCreateHostServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/machines"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(models.CreateVirtualMachineResponse{
				ID:           "vm-123",
				Name:         "test-vm",
				Owner:        "admin",
				CurrentState: "stopped",
			})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/config/hardware"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(models.SystemUsageResponse{
				CpuType: "arm64",
				Total: &models.SystemUsageItem{
					LogicalCpuCount: 8,
					MemorySize:      16384,
				},
				TotalAvailable: &models.SystemUsageItem{
					LogicalCpuCount: 8,
					MemorySize:      16384,
				},
				TotalInUse:     &models.SystemUsageItem{},
				TotalReserved:  &models.SystemUsageItem{},
				SystemReserved: &models.SystemUsageItem{},
			})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/machines"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]models.ParallelsVM{})
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/cache"):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(models.VirtualMachineCatalogManifestList{
				Manifests: []models.CatalogManifest{},
			})
		default:
			http.NotFound(w, r)
		}
	}))

	t.Cleanup(func() {
		time.Sleep(200 * time.Millisecond)
		server.Close()
	})

	return server
}

func createHealthyHostRecord(t *testing.T, hostURL string) *data_models.OrchestratorHost {
	t.Helper()

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	db := serviceprovider.Get().JsonDatabase
	require.NotNil(t, db)

	host, err := db.CreateOrchestratorHost(ctx, data_models.OrchestratorHost{
		Host:         hostURL,
		Architecture: "arm64",
		State:        "healthy",
		Resources: &data_models.HostResources{
			TotalAvailable: data_models.HostResourceItem{
				LogicalCpuCount: 8,
				MemorySize:      16384,
			},
		},
	})
	require.NoError(t, err)

	return host
}

func TestCreateOrchestratorHostVirtualMachineHandler_Success(t *testing.T) {
	cleanup := setupDB(t)
	defer cleanup()

	hostServer := newSyncCreateHostServer(t)

	host := createHealthyHostRecord(t, hostServer.URL+"/api/v1")

	req := authRequest(t, http.MethodPost, "/v1/orchestrator/hosts/"+host.ID+"/machines", validBody(t))
	req = withHostID(req, host.ID)
	w := httptest.NewRecorder()

	controllers.CreateOrchestratorHostVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.CreateVirtualMachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "vm-123", resp.ID)
	assert.Equal(t, "test-vm", resp.Name)
	assert.Equal(t, hostServer.URL+"/api/v1", resp.Host)
}

func TestCreateOrchestratorVirtualMachineHandler_Success(t *testing.T) {
	cleanup := setupDB(t)
	defer cleanup()

	hostServer := newSyncCreateHostServer(t)

	_ = createHealthyHostRecord(t, hostServer.URL+"/api/v1")

	req := authRequest(t, http.MethodPost, "/v1/orchestrator/machines", validBody(t))
	w := httptest.NewRecorder()

	controllers.CreateOrchestratorVirtualMachineHandler()(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.CreateVirtualMachineResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "vm-123", resp.ID)
	assert.Equal(t, "test-vm", resp.Name)
	assert.Equal(t, hostServer.URL+"/api/v1", resp.Host)
}
