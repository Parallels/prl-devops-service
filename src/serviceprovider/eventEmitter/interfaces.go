package eventemitter

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
)

// MessageHandler defines the contract for any service that wants to handle messages.
// This allows the Hub to be polymorphic - it can handle any service.
type MessageHandler interface {
	// Handle processes the message.
	// It takes the context, clientID to know who sent it, the message type, the raw payload,
	// and the message ID for correlation (replies).
	Handle(ctx basecontext.ApiContext, clientID string, msgType constants.EventType, payload []byte, msgID string)
}

// Registrar defines the interface for registering message handlers
type Registrar interface {
	RegisterHandler(eventType []constants.EventType, handler MessageHandler)
}
