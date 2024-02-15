package telemetry

type TelemetryEvent string

const (
	// Event types
	EventStartApi          TelemetryEvent = "START_API"
	EventStartOrchestrator TelemetryEvent = "START_ORCHESTRATOR"
)
