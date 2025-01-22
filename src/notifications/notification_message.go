package notifications

import (
	"encoding/base64"
	"fmt"
	"time"
)

type NotificationMessage struct {
	correlationId   string
	Message         string
	CurrentProgress float64
	totalSize       int64
	currentSize     int64
	IsProgress      bool
	prefix          string
	closed          bool
	startingTime    time.Time
	Level           NotificationMessageLevel
}

func NewNotificationMessage(message string, level NotificationMessageLevel) *NotificationMessage {
	return &NotificationMessage{
		Message: message,
		Level:   level,
	}
}

func NewProgressNotificationMessage(correlationId string, message string, progress float64) *NotificationMessage {
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

func (nm *NotificationMessage) SetStartingTime(startingTime time.Time) *NotificationMessage {
	nm.startingTime = startingTime
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

func calculateETA(started time.Time, downloaded int64, total int64) string {
	if downloaded == 0 {
		return "calculating..."
	}

	elapsed := time.Since(started)
	rate := float64(downloaded) / elapsed.Seconds()
	remaining := float64(total-downloaded) / rate
	eta := time.Duration(remaining) * time.Second

	if eta < time.Minute {
		return fmt.Sprintf("%d seconds", int(eta.Seconds()))
	} else if eta < time.Hour {
		minutes := int(eta.Minutes())
		seconds := int(eta.Seconds()) % 60
		return fmt.Sprintf("%d minutes %d seconds", minutes, seconds)
	} else if eta < 24*time.Hour {
		hours := int(eta.Hours())
		minutes := int(eta.Minutes()) % 60
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	} else {
		days := int(eta.Hours()) / 24
		hours := int(eta.Hours()) % 24
		return fmt.Sprintf("%d days %d hours", days, hours)
	}
}
