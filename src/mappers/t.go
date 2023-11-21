package mappers

import (
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/serviceprovider/apiclient"
)

func MapOrchestratorHostAuthenticationToHttpClientService(m models.OrchestratorHostAuthentication) apiclient.HttpClientServiceAuthorization {
	return apiclient.HttpClientServiceAuthorization{
		Username: m.Username,
		Password: m.Password,
		ApiKey:   m.ApiKey,
	}
}
