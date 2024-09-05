package telemetry

type TelemetryEvent string

const (
	// Event types
	EventStartApi          TelemetryEvent = "PD-DEVOPS::START_API"
	EventStartOrchestrator TelemetryEvent = "PD-DEVOPS::START_ORCHESTRATOR"
	EventApiLog            TelemetryEvent = "PD-DEVOPS::CALL_API_ENDPOINT"
	EventCallHome          TelemetryEvent = "PD-DEVOPS::HEARTBEAT"
	EventCatalog           TelemetryEvent = "PD-DEVOPS::CATALOG"
	HeartbeatEvent         TelemetryEvent = "PD-DEVOPS::HEARTBEAT"
	StartEvent             TelemetryEvent = "PD-DEVOPS::START"
)
