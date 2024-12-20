package notifications

import (
	"encoding/base64"
)

type NotificationMessage struct {
	correlationId   string
	Message         string
	CurrentProgress int
	totalSize       int64
	currentSize     int64
	IsProgress      bool
	prefix          string
	closed          bool
	Level           NotificationMessageLevel
}

func NewNotificationMessage(message string, level NotificationMessageLevel) *NotificationMessage {
	return &NotificationMessage{
		Message: message,
		Level:   level,
	}
}

func NewProgressNotificationMessage(correlationId string, message string, progress int) *NotificationMessage {
	cid := base64.StdEncoding.EncodeToString([]byte(correlationId))
	return &NotificationMessage{
		correlationId:   cid,
		Message:         message,
		CurrentProgress: progress,
		IsProgress:      true,
	}
}

func (nm *NotificationMessage) String() string {
	return nm.Message
}

func (nm *NotificationMessage) SetCorrelationId(id string) *NotificationMessage {
	nm.correlationId = id
	return nm
}

func (nm *NotificationMessage) CorrelationId() string {
	return nm.correlationId
}

func (nm *NotificationMessage) SetTotalSize(size int64) *NotificationMessage {
	nm.totalSize = size
	return nm
}

func (nm *NotificationMessage) TotalSize() int64 {
	return nm.totalSize
}

func (nm *NotificationMessage) SetCurrentSize(size int64) *NotificationMessage {
	nm.currentSize = size
	return nm
}

func (nm *NotificationMessage) CurrentSize() int64 {
	return nm.currentSize
}

func (nm *NotificationMessage) SetPrefix(prefix string) *NotificationMessage {
	nm.prefix = prefix
	return nm
}

func (nm *NotificationMessage) Prefix() string {
	return nm.prefix
}

func (nm *NotificationMessage) Closed() bool {
	return nm.closed
}

func (nm *NotificationMessage) Close() *NotificationMessage {
	nm.closed = true
	return nm
}
