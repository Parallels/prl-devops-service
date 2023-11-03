package catalog

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/catalog/models"
	"Parallels/pd-api-service/serviceprovider/httpclient"
)

func GetAuthenticator(ctx basecontext.ApiContext, provider *models.CatalogManifestProvider) (*httpclient.HttpClientAuthorization, error) {
	auth, err := httpclient.GetAuthenticator(ctx, provider.GetUrl(), &httpclient.AuthorizationModel{
		Username: provider.Username,
		Password: provider.Password,
		ApiKey:   provider.ApiKey,
	})

	if err != nil {
		return nil, err
	}

	return auth, nil
}
