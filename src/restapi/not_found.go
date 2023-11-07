package restapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/models"
)

func NotFoundController() Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		baseCtx := basecontext.NewBaseContextFromRequest(r)
		response := models.ApiErrorResponse{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("resource %s not found", r.URL.Path),
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		baseCtx.LogInfo("Resource not found")
	}
}
