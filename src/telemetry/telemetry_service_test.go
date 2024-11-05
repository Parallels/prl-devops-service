package telemetry

import (
	"os"
	"testing"
	"time"

	basecontext_test "github.com/Parallels/prl-devops-service/basecontext/test"
	"github.com/Parallels/prl-devops-service/constants"
	telemetry_test "github.com/Parallels/prl-devops-service/telemetry/test"
	"github.com/amplitude/analytics-go/amplitude/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelemetryService_Flush(t *testing.T) {
	mockClient := telemetry_test.NewMockAmplitudeClient()
	telemetryService := &TelemetryService{client: mockClient}
	flushWasCalled := false
	mockClient.On("Flush", func(args ...interface{}) {
		flushWasCalled = true
	})

	telemetryService.Flush()
	assert.True(t, flushWasCalled)
}

func TestTelemetryService_Shutdown(t *testing.T) {
	mockClient := telemetry_test.NewMockAmplitudeClient()
	telemetryService := &TelemetryService{client: mockClient}
	shutdownWasCalled := false
	mockClient.On("Shutdown", func(args ...interface{}) {
		shutdownWasCalled = true
	})

	telemetryService.Close()
	assert.True(t, shutdownWasCalled)
}

func TestTelemetryService_Callback(t *testing.T) {
	t.Run("result.Code is greater than 200 and less than 300", func(t *testing.T) {
		os.Setenv(constants.LOG_LEVEL_ENV_VAR, "debug")
		mockContext := basecontext_test.NewMockBaseContext()
		mockClient := telemetry_test.NewMockAmplitudeClient()
		telemetryService := &TelemetryService{
			ctx:             mockContext,
			client:          mockClient,
			EnableTelemetry: true,
			CallBackChan:    make(chan types.ExecuteResult),
		}
		logMessages := []string{}
		mockContext.On("LogErrorf", func(args ...interface{}) {
			logMessages = append(logMessages, args[0].(string))
		})
		mockContext.On("LogInfof", func(args ...interface{}) {
			logMessages = append(logMessages, args[0].(string))
		})
		mockClient.On("Track", func(args ...interface{}) {
			go func() {
				time.Sleep(100 * time.Millisecond)
				telemetryService.Callback(types.ExecuteResult{Code: 201, Message: "Success"})
			}()
		})

		item := TelemetryItem{
			Type: "test_event",
			Properties: map[string]interface{}{
				"key": "value",
			},
		}

		telemetryService.TrackEvent(item)

		result := <-telemetryService.CallBackChan // wait for the callback
		assert.Equal(t, 201, result.Code)

		assert.True(t, telemetryService.EnableTelemetry)
		logMessages = []string{}
	})

	t.Run("result.Code is equal to 400", func(t *testing.T) {
		mockContext := basecontext_test.NewMockBaseContext()
		mockClient := telemetry_test.NewMockAmplitudeClient()
		telemetryService := &TelemetryService{
			ctx:             mockContext,
			client:          mockClient,
			EnableTelemetry: true,
			CallBackChan:    make(chan types.ExecuteResult),
		}
		logMessages := []string{}
		mockContext.On("LogErrorf", func(args ...interface{}) {
			logMessages = append(logMessages, args[0].(string))
		})
		mockContext.On("LogInfof", func(args ...interface{}) {
			logMessages = append(logMessages, args[0].(string))
		})

		mockClient.On("Track", func(args ...interface{}) {
			go func() {
				time.Sleep(100 * time.Millisecond)
				telemetryService.Callback(types.ExecuteResult{Code: 400, Message: "Bad Request"})
			}()
		})

		item := TelemetryItem{
			Type: "test_event",
			Properties: map[string]interface{}{
				"key": "value",
			},
		}

		telemetryService.TrackEvent(item)

		result := <-telemetryService.CallBackChan // wait for the callback
		assert.Equal(t, 400, result.Code)

		require.Equal(t, 1, len(logMessages))
		assert.True(t, telemetryService.EnableTelemetry)
		logMessages = []string{}
	})

	t.Run("result.Code is equal to 400 with Invalid Api Key", func(t *testing.T) {
		mockContext := basecontext_test.NewMockBaseContext()
		mockClient := telemetry_test.NewMockAmplitudeClient()
		telemetryService := &TelemetryService{
			ctx:             mockContext,
			client:          mockClient,
			EnableTelemetry: true,
			CallBackChan:    make(chan types.ExecuteResult),
		}
		logMessages := []string{}
		mockContext.On("LogErrorf", func(args ...interface{}) {
			logMessages = append(logMessages, args[0].(string))
		})
		mockContext.On("LogInfof", func(args ...interface{}) {
			logMessages = append(logMessages, args[0].(string))
		})

		mockClient.On("Track", func(args ...interface{}) {
			go func() {
				time.Sleep(100 * time.Millisecond)
				telemetryService.Callback(types.ExecuteResult{Code: 400, Message: "Invalid API key"})
			}()
		})

		item := TelemetryItem{
			Type: "test_event",
			Properties: map[string]interface{}{
				"key": "value",
			},
		}

		telemetryService.TrackEvent(item)

		result := <-telemetryService.CallBackChan // wait for the callback
		assert.Equal(t, 400, result.Code)

		require.Equal(t, 2, len(logMessages))
		// assert.Equal(t, "[Telemetry] Failed to send event to Amplitude: Invalid API key", logMessages[1])
		assert.Equal(t, "[Telemetry] Disabling telemetry as received invalid key", logMessages[1])
		assert.False(t, telemetryService.EnableTelemetry)
		logMessages = []string{}
	})

	t.Run("result.Code is equal to 401", func(t *testing.T) {
		mockContext := basecontext_test.NewMockBaseContext()
		mockClient := telemetry_test.NewMockAmplitudeClient()
		telemetryService := &TelemetryService{
			ctx:             mockContext,
			client:          mockClient,
			EnableTelemetry: true,
			CallBackChan:    make(chan types.ExecuteResult),
		}
		logMessages := []string{}
		mockContext.On("LogErrorf", func(args ...interface{}) {
			logMessages = append(logMessages, args[0].(string))
		})
		mockContext.On("LogInfof", func(args ...interface{}) {
			logMessages = append(logMessages, args[0].(string))
		})

		mockClient.On("Track", func(args ...interface{}) {
			go func() {
				time.Sleep(100 * time.Millisecond)
				telemetryService.Callback(types.ExecuteResult{Code: 401, Message: "Invalid API key"})
			}()
		})

		item := TelemetryItem{
			Type: "test_event",
			Properties: map[string]interface{}{
				"key": "value",
			},
		}

		telemetryService.TrackEvent(item)
		result := <-telemetryService.CallBackChan // wait for the callback
		assert.Equal(t, 401, result.Code)

		require.Equal(t, 2, len(logMessages))
		// assert.Equal(t, "[Telemetry] Failed to send event to Amplitude: Invalid API key", logMessages[1])
		assert.Equal(t, "[Telemetry] Disabling telemetry as received invalid key", logMessages[1])
		assert.False(t, telemetryService.EnableTelemetry)
		logMessages = []string{}
	})

	t.Run("result.Code is equal to 403", func(t *testing.T) {
		mockContext := basecontext_test.NewMockBaseContext()
		mockClient := telemetry_test.NewMockAmplitudeClient()
		telemetryService := &TelemetryService{
			ctx:             mockContext,
			client:          mockClient,
			EnableTelemetry: true,
			CallBackChan:    make(chan types.ExecuteResult),
		}
		logMessages := []string{}
		mockContext.On("LogErrorf", func(args ...interface{}) {
			logMessages = append(logMessages, args[0].(string))
		})
		mockContext.On("LogInfof", func(args ...interface{}) {
			logMessages = append(logMessages, args[0].(string))
		})

		mockClient.On("Track", func(args ...interface{}) {
			go func() {
				time.Sleep(100 * time.Millisecond)
				telemetryService.Callback(types.ExecuteResult{Code: 403, Message: "Invalid API key"})
			}()
		})

		item := TelemetryItem{
			Type: "test_event",
			Properties: map[string]interface{}{
				"key": "value",
			},
		}

		telemetryService.TrackEvent(item)

		result := <-telemetryService.CallBackChan // wait for the callback
		assert.Equal(t, 403, result.Code)

		require.Equal(t, 2, len(logMessages))
		// assert.Equal(t, "[Telemetry] Failed to send event to Amplitude: Invalid API key", logMessages[1])
		assert.Equal(t, "[Telemetry] Disabling telemetry as received invalid key", logMessages[1])
		assert.False(t, telemetryService.EnableTelemetry)
		logMessages = []string{}
	})
}
