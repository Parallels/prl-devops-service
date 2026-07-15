package orchestrator

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHostHealthProbeCheck_SkipsAuthenticationHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/health/probe", r.URL.Path)
		assert.Empty(t, r.Header.Get("Authorization"))
		assert.Empty(t, r.Header.Get("X-Api-Key"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"OK"}`))
	}))
	defer server.Close()

	svc := &OrchestratorService{
		healthCheckTimeout: 200 * time.Millisecond,
	}

	host := &data_models.OrchestratorHost{
		Host: server.URL,
		Authentication: &data_models.OrchestratorHostAuthentication{
			Username: "user",
			Password: "pass",
			ApiKey:   "api-key",
		},
	}

	response, err := svc.GetHostHealthProbeCheck(host)
	require.NoError(t, err)
	assert.Equal(t, &restapi.HealthProbeResponse{Status: "OK"}, response)
}

func TestGetHostHealthProbeCheck_FailsFastOnTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"OK"}`))
	}))
	defer server.Close()

	svc := &OrchestratorService{
		healthCheckTimeout: 50 * time.Millisecond,
	}

	host := &data_models.OrchestratorHost{
		Host: server.URL,
	}

	start := time.Now()
	response, err := svc.GetHostHealthProbeCheck(host)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Nil(t, response)
	assert.Less(t, elapsed, 150*time.Millisecond)
}
