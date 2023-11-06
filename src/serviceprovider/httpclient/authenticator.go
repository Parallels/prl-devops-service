package httpclient

import (
	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/errors"
)

func GetAuthenticator(ctx basecontext.ApiContext, host string, authenticator *AuthorizationModel) (*HttpClientAuthorization, error) {
	client := NewHttpCaller()
	var auth HttpClientAuthorization
	if authenticator == nil {
		ctx.LogError("Authenticator cannot be null")
		return nil, errors.New("Authenticator cannot be null")
	}

	if authenticator.Username != "" {
		password := authenticator.Password
		token, err := client.GetJwtToken(ctx, host, authenticator.Username, password)
		if err != nil {
			return nil, err
		}
		auth = HttpClientAuthorization{
			BearerToken: token,
		}
	} else {
		auth = HttpClientAuthorization{
			ApiKey: authenticator.ApiKey,
		}
	}

	return &auth, nil
}
