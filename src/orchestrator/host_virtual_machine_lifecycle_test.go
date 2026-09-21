package orchestrator

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostVirtualMachineLifecycleUsesPut(t *testing.T) {
	svc := &OrchestratorService{ctx: basecontext.NewRootBaseContext()}
	tests := []struct {
		name   string
		action string
		query  string
		call   func(*data_models.OrchestratorHost, string) (*models.VirtualMachineOperationResponse, error)
	}{
		{name: "start", action: "start", call: svc.CallStartHostVirtualMachine},
		{name: "stop", action: "stop", query: "force=false", call: func(host *data_models.OrchestratorHost, id string) (*models.VirtualMachineOperationResponse, error) {
			return svc.CallStopHostVirtualMachine(host, id, false)
		}},
		{name: "force stop", action: "stop", query: "force=true", call: func(host *data_models.OrchestratorHost, id string) (*models.VirtualMachineOperationResponse, error) {
			return svc.CallStopHostVirtualMachine(host, id, true)
		}},
		{name: "restart", action: "restart", call: svc.CallRestartHostVirtualMachine},
		{name: "pause", action: "pause", call: svc.CallPauseHostVirtualMachine},
		{name: "resume", action: "resume", call: svc.CallResumeHostVirtualMachine},
		{name: "reset", action: "reset", call: svc.CallResetHostVirtualMachine},
		{name: "suspend", action: "suspend", call: svc.CallSuspendHostVirtualMachine},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := models.VirtualMachineOperationResponse{ID: "vm-123", Operation: tt.action, Status: "success"}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !assert.Equal(t, http.MethodPut, r.Method) {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				assert.Equal(t, "/api/v1/machines/vm-123/"+tt.action, r.URL.Path)
				assert.Equal(t, tt.query, r.URL.RawQuery)
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.Empty(t, body)
				w.Header().Set("Content-Type", "application/json")
				assert.NoError(t, json.NewEncoder(w).Encode(expected))
			}))
			defer server.Close()

			response, err := tt.call(&data_models.OrchestratorHost{Host: server.URL + "/api/v1"}, expected.ID)
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, expected, *response)
		})
	}
}
