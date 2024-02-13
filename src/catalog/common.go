package catalog

import (
	"github.com/Parallels/prl-devops-service/catalog/models"
	"github.com/Parallels/prl-devops-service/serviceprovider/apiclient"
)

func GetAuthenticator(provider *models.CatalogManifestProvider) apiclient.HttpClientServiceAuthorization {
	auth := apiclient.HttpClientServiceAuthorization{
		Username: provider.Username,
		Password: provider.Password,
		ApiKey:   provider.ApiKey,
	}

	return auth
}
