package orchestrator

import (
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
)

const (
	HealthyState = "healthy"
)

func (s *OrchestratorService) getApiClient(request models.OrchestratorHost) *apiclient.HttpClientService {
	apiClient := apiclient.NewHttpClient(s.ctx)
	apiClient.WithHeader("X-SOURCE", "ORCHESTRATOR_REQUEST")
	apiClient.WithHeader("X-LOGGING", "IGNORE")
	apiClient.WithHeader("X-SOURCE-ID", "ORCHESTRATOR_REQUEST")

	if request.Authentication != nil {
		if request.Authentication.ApiKey != "" {
			apiClient.AuthorizeWithApiKey(request.Authentication.ApiKey)
		} else {
			apiClient.AuthorizeWithUsernameAndPassword(request.Authentication.Username, request.Authentication.Password)
		}
	}

	return apiClient
}
