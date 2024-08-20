package models

type AmplitudeEvent struct {
	EventType       string                 `json:"event_type"`
	EventProperties map[string]interface{} `json:"event_properties"`
	UserProperties  map[string]interface{} `json:"user_properties"`
	DeviceId        string                 `json:"device_id"`
	UserId          string                 `json:"user_id"`
	Origin          string                 `json:"origin"`
}
