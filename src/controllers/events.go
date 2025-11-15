package controllers

import (
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
)

func registerEventHandlers(ctx basecontext.ApiContext, version string) {

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/ws/subscribe").
		WithHandler(WebSocketSubscribeHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/ws/unsubscribe").
		WithHandler(WebSocketUnsubscribeHandler()).
		Register()
}

// @Summary		Subscribe to event notifications via WebSocket
// @Description	This endpoint upgrades the HTTP connection to WebSocket and subscribes to event notifications. Supports both JWT Bearer tokens and API Keys for authentication.
// @Tags			Events
// @Produce		json
// @Param			event_types	query		string	false	"Comma-separated event types to subscribe to (e.g., vm,host,system). Valid types: global, system, vm, host, pdfm. If omitted, subscribes to global events only."
// @Param			token	query		string	false	"JWT token for authentication (alternative to Authorization header)"
// @Param			api_key	query		string	false	"API key for authentication (alternative to X-API-KEY header)"
// @Success		101		{string}	string	"Switching Protocols"
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.ApiErrorResponse
// @Failure		409		{object}	models.ApiErrorResponse
// @Failure		503		{object}	models.ApiErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/ws/subscribe [get]
func WebSocketSubscribeHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)

		emitter := eventemitter.Get()
		if emitter == nil || !emitter.IsRunning() {
			ctx.LogErrorf("EventEmitter service is not available")
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "EventEmitter service is not available",
				Code:    http.StatusServiceUnavailable,
			})
			return
		}
		if err := eventemitter.HandleWebSocketConnection(w, r, ctx); err != nil {
			ReturnApiError(ctx, w, *err)
			return
		}
	}
}

// @Summary		Unsubscribe from specific event types
// @Description	Unsubscribe an active WebSocket client from specific event types without disconnecting. The client must belong to the authenticated user.
// @Tags			Events
// @Accept			json
// @Produce		json
// @Param			body	body		models.UnsubscribeRequest	true	"Unsubscribe request with client ID and event types"
// @Success		200		{string}	string	"OK"
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.ApiErrorResponse
// @Failure		404		{object}	models.ApiErrorResponse
// @Failure		503		{object}	models.ApiErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/ws/unsubscribe [post]
func WebSocketUnsubscribeHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		emitter := eventemitter.Get()
		if emitter == nil || !emitter.IsRunning() {
			ctx.LogErrorf("EventEmitter service is not available")
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "EventEmitter service is not available",
				Code:    http.StatusServiceUnavailable,
			})
			return
		}

		if err := eventemitter.HandleUnsubscribe(r, ctx); err != nil {
			ReturnApiError(ctx, w, *err)
			return
		}
		ReturnApiCommonResponse(w)
	}
}
