package telemetry

type TelemetryEvent string

const (
	// Event types
	EventStartApi          TelemetryEvent = "DEVOPS_START_API"
	EventStartOrchestrator TelemetryEvent = "DEVOPS_START_ORCHESTRATOR"
	EventApiLog            TelemetryEvent = "DEVOPS_API_ENDPOINT"
	EventCallHome          TelemetryEvent = "DEVOPS_CALL_HOME"
	EventCatalog           TelemetryEvent = "DEVOPS_CATALOG"
	CallHomeEvent          TelemetryEvent = "DEVOPS_CALL_HOME"
)
