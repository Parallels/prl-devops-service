package notifications

type NotificationMessageLevel int

const (
	NotificationMessageLevelInfo NotificationMessageLevel = iota
	NotificationMessageLevelWarning
	NotificationMessageLevelError
	NotificationMessageLevelDebug
)
