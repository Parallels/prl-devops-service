package interfaces

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
)

// HostEventHandler interface for handling host events
type HostEventHandler interface {
	Handle(ctx basecontext.ApiContext, hostID string, eventType constants.EventType, payload []byte)
}

// HostRegistrar interface for registering handlers
type HostRegistrar interface {
	RegisterHandler(eventType []constants.EventType, handler HostEventHandler)
}
