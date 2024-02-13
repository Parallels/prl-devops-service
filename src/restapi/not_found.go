package restapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/models"
)

func NotFoundController() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		baseCtx := basecontext.NewBaseContextFromRequest(r)

		// listener := globalHttpListener
		// path := r.URL.Path
		// if len(listener.Versions) > 0 {
		// 	if strings.HasPrefix(path, listener.Options.ApiPrefix) && !strings.HasPrefix(path, listener.GetFullPathPrefix()) {
		// 		protocol := "http"
		// 		if r.TLS != nil {
		// 			protocol = "https"
		// 		}
		// 		baseUrl := fmt.Sprintf("%s://%s", protocol, r.Host)
		// 		path = strings.ReplaceAll(path, listener.Options.ApiPrefix, "")
		// 		newUrl := fmt.Sprintf("%s%s", baseUrl, http_helper.JoinUrl(listener.GetFullPathPrefix(), path))
		// 		http.Redirect(w, r, newUrl, http.StatusTemporaryRedirect)
		// 		return
		// 	}
		// }

		response := models.ApiErrorResponse{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("resource %s not found", r.URL.Path),
		}

		SetContentType("application/json", w)
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(response)
		baseCtx.LogInfof("Resource %s not found", r.URL.Path)
	})
}
