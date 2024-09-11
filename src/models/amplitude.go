package models

type AmplitudeEvent struct {
	EventType       string                 `json:"event_type"`
	EventProperties map[string]interface{} `json:"event_properties"`
	UserProperties  map[string]interface{} `json:"user_properties"`
	DeviceId        string                 `json:"device_id"`
	AppId           string                 `json:"app_id"`
	Origin          string                 `json:"origin"`
}
