package restapi

import (
	"net/http"

	"github.com/Parallels/prl-devops-service/constants"
)

func SetContentType(contentType string, w http.ResponseWriter) {
	w.Header().Del("content-type")
	w.Header().Del("Content-Type")
	w.Header().Set("Content-Type", contentType)
}

func GetRequestId(r *http.Request) string {
	id, _ := r.Context().Value(constants.REQUEST_ID_KEY).(string)

	return id
}

func HasAuthorizationHeader(r *http.Request) bool {
	return r.Header.Get("Authorization") != ""
}

func HasApiKeyAuthorizationHeader(r *http.Request) bool {
	return r.Header.Get("X-Api-Key") != ""
}
