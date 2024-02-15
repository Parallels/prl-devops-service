package telemetry

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/stretchr/testify/assert"
)

func TestNewTelemetryItem(t *testing.T) {
	ctx := &basecontext.BaseContext{} // Replace with your test context
	eventType := EventStartApi
	properties := map[string]interface{}{
		"property1": "value1",
		"property2": "value2",
	}
	options := map[string]interface{}{
		"option1": "value1",
		"option2": "value2",
	}

	item := NewTelemetryItem(ctx, eventType, properties, options)

	// Add your assertions here
	if item.Type != string(eventType) {
		t.Errorf("Expected Type to be %s, but got %s", eventType, item.Type)
	}

	// Add more assertions for other fields if needed

	// Example assertion for HardwareID
	if item.HardwareID == "" {
		t.Error("Expected HardwareID to be non-empty")
	}
}

func TestNewTelemetryItemEmptyProperties(t *testing.T) {
	ctx := &basecontext.BaseContext{} // Replace with your test context
	eventType := EventStartApi

	item := NewTelemetryItem(ctx, eventType, nil, nil)

	assert.Equal(t, item.Type, string(eventType))
	assert.NotNil(t, item.Properties)
	assert.NotNil(t, item.Options)

	// Example assertion for HardwareID
	if item.HardwareID == "" {
		t.Error("Expected HardwareID to be non-empty")
	}
}
