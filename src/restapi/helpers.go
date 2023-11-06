package restapi

import (
	"net/http"
	"strconv"

	"github.com/Parallels/pd-api-service/constants"
)

func SetContentType(contentType string, w http.ResponseWriter) {
	w.Header().Del("content-type")
	w.Header().Del("Content-Type")
	w.Header().Set("Content-Type", contentType)
}

func SetContentLength(contentLength int, w http.ResponseWriter) {
	w.Header().Del("content-length")
	w.Header().Del("Content-Length")
	w.Header().Set("Content-Length", strconv.Itoa(contentLength))
}

func GetRequestId(r *http.Request) string {
	id := r.Context().Value(constants.REQUEST_ID_KEY).(string)

	return id
}

func HasAuthorizationHeader(r *http.Request) bool {
	return r.Header.Get("Authorization") != ""
}

func HasApiKeyAuthorizationHeader(r *http.Request) bool {
	return r.Header.Get("X-Api-Key") != ""
}
