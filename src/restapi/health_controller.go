package restapi

import (
	"encoding/json"
	"net/http"
)

//	@Summary		Gets the API Health Probe
//	@Description	This endpoint returns the API Health Probe
//	@Tags			Config
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Failure		402	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/health/probe [get]
func (c *HttpListener) Probe() ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		response := make(map[string]string)
		response["status"] = "OK"

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
