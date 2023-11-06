package catalog

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/serviceprovider/httpclient"
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
