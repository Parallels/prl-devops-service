package orchestrator

import (
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
)

func (s *OrchestratorService) getApiClient(request models.OrchestratorHost) *apiclient.HttpClientService {
	apiClient := apiclient.NewHttpClient(s.ctx)
	if request.Authentication != nil {
		if request.Authentication.ApiKey != "" {
			apiClient.AuthorizeWithApiKey(request.Authentication.ApiKey)
		} else {
			apiClient.AuthorizeWithUsernameAndPassword(request.Authentication.Username, request.Authentication.Password)
		}
	}

	return apiClient
}
