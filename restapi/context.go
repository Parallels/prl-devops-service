package restapi

import (
	"Parallels/pd-api-service/models"
	"net/http"
	"strings"

	"github.com/cjlapao/common-go/service_provider"
)

type AuthorizationContext struct {
	RequestId          string
	Issuer             string
	Scope              string
	Audiences          []string
	BaseUrl            string
	IsAuthorized       bool
	IsMicroService     bool
	IsSuperUser        bool
	AuthorizedBy       string
	User               *models.User
	AuthorizationError *models.OAuthErrorResponse
}

var (
	baseAuthorizationCtx *AuthorizationContext
)

func InitAuthorizationContext() *AuthorizationContext {
	if baseAuthorizationCtx == nil {
		context := AuthorizationContext{}

		baseAuthorizationCtx = &context
	}

	return baseAuthorizationCtx
}

func GetBaseContext() *AuthorizationContext {
	if baseAuthorizationCtx == nil {
		return InitAuthorizationContext()
	}

	return baseAuthorizationCtx
}

func CloneAuthorizationContext() *AuthorizationContext {
	// Creating the new context using the default values if it does not exist
	if baseAuthorizationCtx == nil {
		context := AuthorizationContext{}
		baseAuthorizationCtx = &context
	}

	newContext := AuthorizationContext{
		Issuer:       baseAuthorizationCtx.Issuer,
		Scope:        baseAuthorizationCtx.Scope,
		Audiences:    make([]string, 0),
		BaseUrl:      baseAuthorizationCtx.BaseUrl,
		IsAuthorized: false,
		RequestId:    "",
		AuthorizedBy: "",
		User:         nil,
	}

	return &newContext
}

func (a *AuthorizationContext) SetRequestIssuer(r *http.Request, tenantId string) string {
	if a.BaseUrl == "" {
		a.BaseUrl = service_provider.Get().GetBaseUrl(r)
	}

	if a.Issuer == "" {
		a.Issuer = a.GetBaseUrl(r) + "/auth/" + tenantId
		a.Issuer = strings.Trim(a.Issuer, "/")
	}

	return a.Issuer
}

func (a *AuthorizationContext) GetBaseUrl(r *http.Request) string {
	config := service_provider.Get().Configuration
	if a.BaseUrl == "" {
		return service_provider.Get().GetBaseUrl(r)
	}

	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}

	issuer := strings.ReplaceAll(a.BaseUrl, "https", "")
	issuer = strings.ReplaceAll(issuer, "http", "")
	issuer = strings.ReplaceAll(issuer, "://", "")
	if strings.HasSuffix(issuer, "/") {
		issuer = strings.Trim(issuer, "/")
	}

	baseUrl := protocol + "://" + issuer
	apiPrefix := config.GetString("API_PREFIX")
	if apiPrefix != "" {
		if strings.HasPrefix(apiPrefix, "/") {
			baseUrl += apiPrefix
		} else {
			baseUrl += "/" + apiPrefix
		}
	}

	return baseUrl
}
