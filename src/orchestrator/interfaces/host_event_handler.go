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

// HardwareEnqueuer is implemented by HardwareUpdateQueue.
// PDfMEventHandler uses this to request a hardware refresh without knowing
// about the queue implementation.
type HardwareEnqueuer interface {
	Enqueue(hostID string)
}
