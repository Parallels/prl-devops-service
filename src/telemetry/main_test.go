package telemetry

import (
	"os"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	basecontext_test "github.com/Parallels/prl-devops-service/basecontext/test"
	"github.com/Parallels/prl-devops-service/constants"
	telemetry_test "github.com/Parallels/prl-devops-service/telemetry/test"
	"github.com/amplitude/analytics-go/amplitude/types"
	"github.com/stretchr/testify/assert"
)

func TestNewTelemetryService(t *testing.T) {
	mockContext := basecontext_test.NewMockBaseContext()
	logMessages := []string{}
	mockContext.On("LogErrorf", func(args ...interface{}) {
		logMessages = append(logMessages, args[0].(string))
	})
	mockContext.On("LogInfof", func(args ...interface{}) {
		logMessages = append(logMessages, args[0].(string))
	})
	mockAmplitudeClient := telemetry_test.NewMockAmplitudeClient()

	t.Run("Enable telemetry with environment key", func(t *testing.T) {
		_ = os.Setenv(constants.AmplitudeApiKeyEnvVar, "123")
		svc := New(mockContext)
		svc.client = mockAmplitudeClient
		svc.Flush()

		assert.True(t, svc.EnableTelemetry)
		assert.NotNil(t, svc.client)

		assert.Equal(t, 0, len(logMessages))
	})

	t.Run("Enable telemetry with embedded key", func(t *testing.T) {
		constants.AmplitudeApiKey = "123"
		svc := New(mockContext)
		svc.client = mockAmplitudeClient
		svc.Flush()

		assert.True(t, svc.EnableTelemetry)
		assert.NotNil(t, svc.client)

		assert.Equal(t, 0, len(logMessages))
	})

	t.Run("Disable telemetry", func(t *testing.T) {
		os.Clearenv()
		constants.AmplitudeApiKey = ""
		svc := New(mockContext)

		assert.False(t, svc.EnableTelemetry)
		assert.Nil(t, svc.client)

		assert.Equal(t, 0, len(logMessages))
	})
}

func TestGet(t *testing.T) {
	// mockCtx := &basecontext.BaseContext{}

	t.Run("Global telemetry service is nil", func(t *testing.T) {
		globalTelemetryService = nil

		telemetryService := Get()

		assert.NotNil(t, telemetryService)
		assert.NotNil(t, telemetryService.ctx)
	})

	t.Run("Global telemetry service is not nil", func(t *testing.T) {
		mockTelemetryService := &TelemetryService{}
		globalTelemetryService = mockTelemetryService

		telemetryService := Get()

		assert.Equal(t, mockTelemetryService, telemetryService)
	})
}

func TestTrackEvent(t *testing.T) {
	t.Run("Telemetry service available", func(t *testing.T) {
		var amplitudeTrackItem types.Event
		mockClient := telemetry_test.NewMockAmplitudeClient()
		mockClient.On("Track", func(args ...interface{}) {
			if item, ok := args[0].(types.Event); ok {
				amplitudeTrackItem = item
			}
			go func() {
				time.Sleep(1 * time.Second)
				globalTelemetryService.Callback(types.ExecuteResult{Code: 200, Message: "Success"})
			}()
		})

		globalTelemetryService = &TelemetryService{
			ctx:             &basecontext.BaseContext{},
			EnableTelemetry: true,
			client:          mockClient,
			CallBackChan:    make(chan types.ExecuteResult),
		}

		// Create a test telemetry item
		item := TelemetryItem{
			Type: "test_event",
			Properties: map[string]interface{}{
				"key": "value",
			},
		}

		TrackEvent(item)

		assert.Equal(t, amplitudeTrackItem.EventType, item.Type)
	})

	t.Run("Telemetry service available, error response", func(t *testing.T) {
		var amplitudeTrackItem types.Event
		mockClient := telemetry_test.NewMockAmplitudeClient()
		mockClient.On("Track", func(args ...interface{}) {
			if item, ok := args[0].(types.Event); ok {
				amplitudeTrackItem = item
			}
			go func() {
				time.Sleep(1 * time.Second)
				globalTelemetryService.Callback(types.ExecuteResult{Code: 400, Message: "Bad Request"})
			}()
		})

		globalTelemetryService = &TelemetryService{
			ctx:             &basecontext.BaseContext{},
			EnableTelemetry: true,
			client:          mockClient,
			CallBackChan:    make(chan types.ExecuteResult),
		}

		// Create a test telemetry item
		item := TelemetryItem{
			Type: "test_event",
			Properties: map[string]interface{}{
				"key": "value",
			},
		}

		TrackEvent(item)

		assert.Equal(t, amplitudeTrackItem.EventType, item.Type)
	})

	t.Run("Telemetry service not available", func(t *testing.T) {
		globalTelemetryService = nil

		// Create a test telemetry item
		item := TelemetryItem{
			Type: "test_event",
			Properties: map[string]interface{}{
				"key": "value",
			},
		}

		// Call the TrackEvent function
		TrackEvent(item)
		assert.NotNil(t, globalTelemetryService)
	})
}
