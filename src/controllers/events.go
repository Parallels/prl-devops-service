package controllers

import (
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	eventemitter "github.com/Parallels/prl-devops-service/serviceprovider/eventEmitter"
)

func registerEventHandlers(ctx basecontext.ApiContext, version string) {

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/ws/subscribe").
		WithRequiredClaim(constants.READ_ONLY_CLAIM).
		WithHandler(WebSocketSubscribeHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/ws/unsubscribe").
		WithRequiredClaim(constants.READ_ONLY_CLAIM).
		WithHandler(WebSocketUnsubscribeHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/ws/clients").
		WithRequiredClaim(constants.READ_ONLY_CLAIM).
		WithHandler(WebSocketClientsHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/ws/stats").
		WithRequiredClaim(constants.READ_ONLY_CLAIM).
		WithHandler(WebSocketStatsHandler()).
		Register()
}

// @Summary		Subscribe to event notifications via WebSocket
// @Description	This endpoint upgrades the HTTP connection to WebSocket and subscribes to event notifications. Authentication is required via Authorization header (Bearer token) or query parameters (access_token or authorization).
// @Tags			Events
// @Produce		json
// @Param			event_types	query		string	false	"Comma-separated event types to subscribe to (e.g., global,pdfm,system). Valid types: global,pdfm and orchestrator. If omitted, subscribes to global events only."
// @Success		101			{string}	string	"Switching Protocols"
// @Failure		400			{object}	models.ApiErrorResponse
// @Failure		401			{object}	models.ApiErrorResponse
// @Failure		409			{object}	models.ApiErrorResponse
// @Failure		503			{object}	models.ApiErrorResponse
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

// @Summary		List connected WebSocket clients
// @Description	Returns all currently connected WebSocket clients with queue depth and ping/pong timestamps. Useful for diagnosing stale or dead clients whose queues are filling up.
// @Tags			Events
// @Produce		json
// @Success		200	{array}		models.EventClientInfo
// @Failure		503	{object}	models.ApiErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/ws/clients [get]
func WebSocketClientsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		emitter := eventemitter.Get()
		if emitter == nil || !emitter.IsRunning() {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "EventEmitter service is not available",
				Code:    http.StatusServiceUnavailable,
			})
			return
		}
		eventemitter.HandleGetClients(w, r, ctx)
	}
}

// @Summary		Get WebSocket event emitter statistics
// @Description	Returns aggregate statistics including total connected clients, subscription counts per event type, uptime, and per-client details with queue depths.
// @Tags			Events
// @Produce		json
// @Success		200	{object}	models.EventEmitterStats
// @Failure		503	{object}	models.ApiErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/ws/stats [get]
func WebSocketStatsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		emitter := eventemitter.Get()
		if emitter == nil || !emitter.IsRunning() {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "EventEmitter service is not available",
				Code:    http.StatusServiceUnavailable,
			})
			return
		}
		eventemitter.HandleGetStats(w, r, ctx)
	}
}

// @Summary		Unsubscribe from specific event types
// @Description	Unsubscribe an active WebSocket client from specific event types without disconnecting. The client must belong to the authenticated user.
// @Tags			Events
// @Accept			json
// @Produce		json
// @Param			body	body		models.UnsubscribeRequest	true	"Unsubscribe request with client ID and event types"
// @Success		200		{string}	string						"OK"
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
		eventemitter.HandleUnsubscribe(w, r, ctx)
	}
}
