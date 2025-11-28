package orchestrator

import (
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
)

const (
	HealthyState      = "healthy"
	MaxNumberAppleVms = 2
)

// getApiClient creates and configures an HTTP client for orchestrator requests
func getApiClient(ctx basecontext.ApiContext, host models.OrchestratorHost) *apiclient.HttpClientService {
	apiClient := apiclient.NewHttpClient(ctx)

	apiClient.WithHeader("X-SOURCE", "ORCHESTRATOR_REQUEST")
	apiClient.WithHeader("X-LOGGING", "IGNORE")
	apiClient.WithHeader("X-SOURCE-ID", "ORCHESTRATOR_REQUEST")

	if host.Authentication != nil {
		if host.Authentication.ApiKey != "" {
			apiClient.AuthorizeWithApiKey(host.Authentication.ApiKey)
		} else {
			apiClient.AuthorizeWithUsernameAndPassword(host.Authentication.Username, host.Authentication.Password)
		}
	}

	return apiClient
}

// getAuthHeaderForWebSocket returns the authorization header for WebSocket connections
func getAuthHeaderForWebSocket(ctx basecontext.ApiContext, host models.OrchestratorHost) (http.Header, error) {
	header := http.Header{}

	if host.Authentication == nil {
		return header, nil
	}

	client := getApiClient(ctx, host)
	authorizer, err := client.Authorize(ctx, host.GetHost())
	if err != nil {
		return header, err
	}

	if authorizer.BearerToken != "" {
		header.Set("Authorization", "Bearer "+authorizer.BearerToken)
	} else if authorizer.ApiKey != "" {
		header.Set("X-Api-Key", authorizer.ApiKey)
	}

	return header, nil
}

// OrchestratorService method wrapper for backward compatibility
func (s *OrchestratorService) getApiClient(request models.OrchestratorHost) *apiclient.HttpClientService {
	return getApiClient(s.ctx, request)
}
