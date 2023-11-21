package apiclient

type HttpClientServiceAuthorizer struct {
	BearerToken string
	ApiKey      string
}

type HttpClientServiceAuthorization struct {
	Username string
	Password string
	ApiKey   string
}
