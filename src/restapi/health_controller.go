package restapi

import (
	"encoding/json"
	"net/http"
)

type HealthProbeResponse struct {
	Status string `json:"status"`
}

// @Summary		Gets the API Health Probe
// @Description	This endpoint returns the API Health Probe
// @Tags			Config
// @Produce		json
// @Success		200	{object}	map[string]string
// @Failure		402	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/health/probe [get]
func (c *HttpListener) Probe() ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		response := HealthProbeResponse{
			Status: "OK",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
