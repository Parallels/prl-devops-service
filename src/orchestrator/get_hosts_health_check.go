package orchestrator

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	api_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
)

func (s *OrchestratorService) GetHostHealthProbeCheck(host *models.OrchestratorHost) (*restapi.HealthProbeResponse, error) {
	path := "/health/probe"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}

	timeout := s.healthCheckTimeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	cfg := config.Get()
	disableTLSValidation := false
	if cfg != nil {
		disableTLSValidation = cfg.DisableTlsValidation()
	}

	transport := &http.Transport{
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		ExpectContinueTimeout: timeout,
		IdleConnTimeout:       timeout,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: timeout,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: disableTLSValidation,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "PrlDevOpsService/HealthProbe")
	req.Header.Set("X-SOURCE", "ORCHESTRATOR_HEALTH_PROBE")
	req.Header.Set("X-LOGGING", "IGNORE")
	req.Header.Set("X-SOURCE-ID", "ORCHESTRATOR_HEALTH_PROBE")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewWithCodef(400, "Error getting health check for host %s: %v", host.Host, resp.StatusCode)
	}

	var response restapi.HealthProbeResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *OrchestratorService) GetHostSystemHealthCheck(host *models.OrchestratorHost) (*api_models.ApiHealthCheck, error) {
	httpClient := s.getApiClient(*host)
	httpClient.WithTimeout(s.healthCheckTimeout)

	path := "/health/system"
	url, err := helpers.JoinUrl([]string{host.GetHost(), path})
	if err != nil {
		return nil, err
	}
	var response api_models.ApiHealthCheck
	apiResponse, err := httpClient.Get(url.String(), &response)
	if err != nil {
		return nil, err
	}

	if apiResponse.StatusCode != 200 {
		return nil, errors.NewWithCodef(400, "Error getting health check for host %s: %v", host.Host, apiResponse.StatusCode)
	}

	return &response, nil
}

func (s *OrchestratorService) GetHostHealthCheckState(host *models.OrchestratorHost) string {
	healthCheck, err := s.GetHostHealthProbeCheck(host)
	if err != nil {
		return "unhealthy"
	}
	if healthCheck.Status == "OK" {
		return "healthy"
	} else {
		return "unhealthy"
	}
}
