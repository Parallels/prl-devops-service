package mappers

import (
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
)

func MapOrchestratorHostAuthenticationToHttpClientService(m models.OrchestratorHostAuthentication) apiclient.HttpClientServiceAuthorization {
	return apiclient.HttpClientServiceAuthorization{
		Username: m.Username,
		Password: m.Password,
		ApiKey:   m.ApiKey,
	}
}
