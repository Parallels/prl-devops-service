package httpclient

import "github.com/Parallels/pd-api-service/models"

type HttpClientVerb string

const (
	HttpCallerVerbGet    HttpClientVerb = "GET"
	HttpCallerVerbPost   HttpClientVerb = "POST"
	HttpCallerVerbPut    HttpClientVerb = "PUT"
	HttpCallerVerbDelete HttpClientVerb = "DELETE"
)

func (v HttpClientVerb) String() string {
	return string(v)
}

type HttpClientAuthorization struct {
	BearerToken string
	ApiKey      string
}

type HttpClientResponse struct {
	StatusCode int
	Data       interface{}
	ApiError   *models.ApiErrorResponse
}

type AuthorizationModel struct {
	Username string
	Password string
	ApiKey   string
}
