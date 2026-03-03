package handlers

import (
	"encoding/json"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	global_models "github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

type HostCatalogCacheEventHandler struct {
	registrar    interfaces.HostRegistrar
	onCacheEvent func(hostId string)
}

var (
	cacheInstance *HostCatalogCacheEventHandler
	cacheOnce     sync.Once
)

func NewHostCatalogCacheEventHandler(registrar interfaces.HostRegistrar, onCacheEvent func(hostId string)) *HostCatalogCacheEventHandler {
	cacheOnce.Do(func() {
		cacheInstance = &HostCatalogCacheEventHandler{
			registrar:    registrar,
			onCacheEvent: onCacheEvent,
		}
		registrar.RegisterHandler([]constants.EventType{constants.EventTypeCatalogCache}, cacheInstance)
	})
	return cacheInstance
}

func (h *HostCatalogCacheEventHandler) Handle(ctx basecontext.ApiContext, hostID string, eventType constants.EventType, payload []byte) {
	if eventType != constants.EventTypeCatalogCache {
		return
	}

	var event global_models.EventMessage
	if err := json.Unmarshal(payload, &event); err != nil {
		ctx.LogErrorf("[HostCatalogCacheEventHandler] Error unmarshalling event message: %v", err)
		return
	}

	ctx.LogDebugf("[HostCatalogCacheEventHandler] Received catalog cache event from host %s: %s", hostID, event.Message)

	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := global_models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_CATALOG_CACHE_EVENT", map[string]interface{}{
			"host_id": hostID,
			"event":   event,
		})
		go func() {
			if err := emitter.Broadcast(msg); err != nil {
				ctx.LogErrorf("[HostCatalogCacheEventHandler] Failed to broadcast event HOST_CATALOG_CACHE_EVENT: %v", err)
			}
		}()
	}

	if h.onCacheEvent != nil {
		h.onCacheEvent(hostID)
	}
}
