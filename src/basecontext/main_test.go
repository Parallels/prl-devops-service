package basecontext

import (
	"context"
	"net/http"
	"testing"

	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	log "github.com/cjlapao/common-go-logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseContext(t *testing.T) {
	baseCtx := NewBaseContext()
	assert.NotNil(t, baseCtx)
	assert.True(t, baseCtx.shouldLog)
	assert.Equal(t, context.Background(), baseCtx.ctx)
	assert.Nil(t, baseCtx.authContext)
}

func TestNewRootBaseContext(t *testing.T) {
	baseCtx := NewRootBaseContext()
	assert.NotNil(t, baseCtx)
	assert.True(t, baseCtx.shouldLog)
	assert.Equal(t, context.Background(), baseCtx.ctx)
	assert.NotNil(t, baseCtx.authContext)
	assert.True(t, baseCtx.authContext.IsAuthorized)
	assert.Equal(t, "RootAuthorization", baseCtx.authContext.AuthorizedBy)
}

func TestNewBaseContextFromRequest(t *testing.T) {
	t.Run("Request without authorization context", func(t *testing.T) {
		// Create a new HTTP request
		req, _ := http.NewRequest("GET", "/", nil)

		// Test case 1: Request without authorization context
		baseCtx := NewBaseContextFromRequest(req)
		assert.NotNil(t, baseCtx)
		assert.True(t, baseCtx.shouldLog)
		assert.Equal(t, context.Background(), baseCtx.ctx)
		assert.Nil(t, baseCtx.authContext)
	})

	t.Run("Request with authorization context", func(t *testing.T) {
		// Create a new HTTP request
		req, _ := http.NewRequest("GET", "/", nil)

		authCtx := &AuthorizationContext{}
		req = req.WithContext(context.WithValue(req.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authCtx))

		baseCtx := NewBaseContextFromRequest(req)
		assert.NotNil(t, baseCtx)
		assert.True(t, baseCtx.shouldLog)
		assert.Equal(t, req.Context(), baseCtx.ctx)
		assert.Equal(t, authCtx, baseCtx.authContext)
	})

	t.Run("Request without authorization context", func(t *testing.T) {
		// Create a new HTTP request
		req, _ := http.NewRequest("GET", "/", nil)

		baseCtx := NewBaseContextFromRequest(req)
		assert.NotNil(t, baseCtx)
		assert.True(t, baseCtx.shouldLog)
		assert.Equal(t, req.Context(), baseCtx.ctx)
		assert.Nil(t, baseCtx.authContext)
	})

	t.Run("Request without wrong authorization context", func(t *testing.T) {
		// Create a new HTTP request
		req, _ := http.NewRequest("GET", "/", nil)
		// Test case 2: Request with authorization context

		authCtx := &BaseContext{}
		req = req.WithContext(context.WithValue(req.Context(), constants.AUTHORIZATION_CONTEXT_KEY, authCtx))

		baseCtx := NewBaseContextFromRequest(req)
		assert.NotNil(t, baseCtx)
		assert.True(t, baseCtx.shouldLog)
		assert.Equal(t, req.Context(), baseCtx.ctx)
		assert.Nil(t, baseCtx.authContext)
	})
}

func TestNewBaseContextFromContext(t *testing.T) {
	t.Run("Context without authorization context", func(t *testing.T) {
		// Create a new context
		ctx := context.Background()

		// Test case 1: Context without authorization context
		baseCtx := NewBaseContextFromContext(ctx)
		assert.NotNil(t, baseCtx)
		assert.True(t, baseCtx.shouldLog)
		assert.Equal(t, ctx, baseCtx.ctx)
		assert.Nil(t, baseCtx.authContext)
	})

	t.Run("Context with authorization context", func(t *testing.T) {
		// Create a new context
		ctx := context.Background()

		authCtx := &AuthorizationContext{}
		ctx = context.WithValue(ctx, constants.AUTHORIZATION_CONTEXT_KEY, authCtx)

		baseCtx := NewBaseContextFromContext(ctx)
		assert.NotNil(t, baseCtx)
		assert.True(t, baseCtx.shouldLog)
		assert.Equal(t, ctx, baseCtx.ctx)
		assert.Equal(t, authCtx, baseCtx.authContext)
	})

	t.Run("Context with wrong authorization context", func(t *testing.T) {
		// Create a new context
		ctx := context.Background()

		authCtx := &BaseContext{}
		ctx = context.WithValue(ctx, constants.AUTHORIZATION_CONTEXT_KEY, authCtx)

		baseCtx := NewBaseContextFromContext(ctx)
		assert.NotNil(t, baseCtx)
		assert.True(t, baseCtx.shouldLog)
		assert.Equal(t, ctx, baseCtx.ctx)
		assert.Nil(t, baseCtx.authContext)
	})
}

func TestBaseContextGetAuthorizationContext(t *testing.T) {
	baseCtx := &BaseContext{
		authContext: &AuthorizationContext{
			// Set the fields of the AuthorizationContext struct if needed
		},
	}

	authCtx := baseCtx.GetAuthorizationContext()
	assert.NotNil(t, authCtx)
	// Add assertions for the expected values of the AuthorizationContext fields
}

func TestBaseContext_Context(t *testing.T) {
	baseCtx := &BaseContext{
		ctx: context.TODO(),
	}

	ctx := baseCtx.Context()
	assert.Equal(t, baseCtx.ctx, ctx)
}

func TestBaseContext_GetRequestId(t *testing.T) {
	baseCtx := &BaseContext{
		ctx: context.WithValue(context.Background(), constants.REQUEST_ID_KEY, "12345"),
	}

	requestID := baseCtx.GetRequestId()
	assert.Equal(t, "12345", requestID)
}

func TestBaseContext_GetRequestId_NoContext(t *testing.T) {
	baseCtx := &BaseContext{}

	requestID := baseCtx.GetRequestId()
	assert.Equal(t, "", requestID)
}

func TestBaseContext_GetRequestId_NoValue(t *testing.T) {
	baseCtx := &BaseContext{
		ctx: context.Background(),
	}

	requestID := baseCtx.GetRequestId()
	assert.Equal(t, "", requestID)
}

func TestBaseContext_GetUser(t *testing.T) {
	t.Run("With AuthContext", func(t *testing.T) {
		// Create a new BaseContext with AuthContext
		authContext := &AuthorizationContext{
			User: &models.ApiUser{
				// Set the fields of the ApiUser struct if needed
			},
		}
		baseCtx := &BaseContext{
			authContext: authContext,
		}

		// Call the GetUser method
		user := baseCtx.GetUser()

		// Add assertions for the expected values of the user
		assert.NotNil(t, user)
		// Add assertions for the expected values of the user fields
	})

	t.Run("Without AuthContext", func(t *testing.T) {
		// Create a new BaseContext without AuthContext
		baseCtx := &BaseContext{}

		// Call the GetUser method
		user := baseCtx.GetUser()

		// Assert that the user is nil
		assert.Nil(t, user)
	})
}

func TestBaseContext_Verbose(t *testing.T) {
	baseCtx := &BaseContext{
		shouldLog: true,
	}

	verbose := baseCtx.Verbose()
	assert.True(t, verbose)
}

func TestBaseContext_Verbose_False(t *testing.T) {
	baseCtx := &BaseContext{
		shouldLog: false,
	}

	verbose := baseCtx.Verbose()
	assert.False(t, verbose)
}

func TestBaseContext_EnableLog(t *testing.T) {
	baseCtx := &BaseContext{
		shouldLog: false,
	}

	baseCtx.EnableLog()

	assert.True(t, baseCtx.shouldLog)
}

func TestBaseContext_DisableLog(t *testing.T) {
	baseCtx := &BaseContext{
		shouldLog: true,
	}

	baseCtx.DisableLog()

	assert.False(t, baseCtx.shouldLog)
}

func TestBaseContext_ToggleLogTimestamps(t *testing.T) {
	t.Run("Enable timestamps", func(t *testing.T) {
		baseCtx := &BaseContext{}
		baseCtx.ToggleLogTimestamps(true)
		assert.True(t, common.Logger.UseTimestamp)
	})

	t.Run("Disable timestamps", func(t *testing.T) {
		baseCtx := &BaseContext{}
		baseCtx.ToggleLogTimestamps(false)
		assert.False(t, common.Logger.UseTimestamp)
	})
}

func TestBaseContext_LogInfof(t *testing.T) {
	common.Logger = log.NewMockLogger()
	mockLogger, err := log.GetMockLogger()
	require.NoError(t, err)

	baseCtx := &BaseContext{
		shouldLog: true,
	}

	t.Run("Log is enabled", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx.LogInfof("Test log message: %s", "Hello, World!")
		assert.Equal(t, "Test log message: Hello, World!", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "info", mockLogger.LastPrintedMessage.Level)
	})

	t.Run("Log is enabled and request id is present", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx.ctx = context.WithValue(context.Background(), constants.REQUEST_ID_KEY, "12345")
		baseCtx.LogInfof("Test log message: %s", "Hello, World!")
		assert.Equal(t, "[12345] Test log message: Hello, World!", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "info", mockLogger.LastPrintedMessage.Level)
	})

	t.Run("Log is disabled", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx := &BaseContext{
			shouldLog: false,
		}

		baseCtx.LogInfof("Test log message: %s", "Hello, World!")
		assert.Equal(t, "", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "", mockLogger.LastPrintedMessage.Level)
	})
}

func TestBaseContext_LogErrorf(t *testing.T) {
	common.Logger = log.NewMockLogger()
	mockLogger, err := log.GetMockLogger()
	require.NoError(t, err)

	baseCtx := &BaseContext{
		shouldLog: true,
	}

	t.Run("Log is enabled", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx.LogErrorf("Test log message: %s", "Hello, World!")
		assert.Equal(t, "Test log message: Hello, World!", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "error", mockLogger.LastPrintedMessage.Level)
	})

	t.Run("Log is enabled and request id is present", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx.ctx = context.WithValue(context.Background(), constants.REQUEST_ID_KEY, "12345")
		baseCtx.LogErrorf("Test log message: %s", "Hello, World!")
		assert.Equal(t, "[12345] Test log message: Hello, World!", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "error", mockLogger.LastPrintedMessage.Level)
	})

	t.Run("Log is disabled", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx := &BaseContext{
			shouldLog: false,
		}

		baseCtx.LogErrorf("Test log message: %s", "Hello, World!")
		assert.Equal(t, "", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "", mockLogger.LastPrintedMessage.Level)
	})
}

func TestBaseContext_LogWarnf(t *testing.T) {
	common.Logger = log.NewMockLogger()
	mockLogger, err := log.GetMockLogger()
	require.NoError(t, err)

	baseCtx := &BaseContext{
		shouldLog: true,
	}

	t.Run("Log is enabled", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx.LogWarnf("Test log message: %s", "Hello, World!")
		assert.Equal(t, "Test log message: Hello, World!", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "warn", mockLogger.LastPrintedMessage.Level)
	})

	t.Run("Log is enabled and request id is present", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx.ctx = context.WithValue(context.Background(), constants.REQUEST_ID_KEY, "12345")
		baseCtx.LogWarnf("Test log message: %s", "Hello, World!")
		assert.Equal(t, "[12345] Test log message: Hello, World!", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "warn", mockLogger.LastPrintedMessage.Level)
	})

	t.Run("Log is disabled", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx := &BaseContext{
			shouldLog: false,
		}

		baseCtx.LogWarnf("Test log message: %s", "Hello, World!")
		assert.Equal(t, "", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "", mockLogger.LastPrintedMessage.Level)
	})
}

func TestBaseContext_LogDebugf(t *testing.T) {
	common.Logger = log.NewMockLogger()
	mockLogger, err := log.GetMockLogger()
	require.NoError(t, err)

	baseCtx := &BaseContext{
		shouldLog: true,
	}

	t.Run("Log is enabled and debug level is on", func(t *testing.T) {
		mockLogger.Clear()
		common.Logger.LogLevel = log.Debug
		baseCtx.LogDebugf("Test log message: %s", "Hello, World!")
		assert.Equal(t, "Test log message: Hello, World!", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "debug", mockLogger.LastPrintedMessage.Level)
	})

	t.Run("Log is enabled and request id is present and debug level is on", func(t *testing.T) {
		mockLogger.Clear()
		common.Logger.LogLevel = log.Debug
		baseCtx.ctx = context.WithValue(context.Background(), constants.REQUEST_ID_KEY, "12345")
		baseCtx.LogDebugf("Test log message: %s", "Hello, World!")
		assert.Equal(t, "[12345] Test log message: Hello, World!", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "debug", mockLogger.LastPrintedMessage.Level)
	})

	t.Run("Log is disabled and debug level is on", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx := &BaseContext{
			shouldLog: false,
		}
		common.Logger.LogLevel = log.Debug

		baseCtx.LogDebugf("Test log message: %s", "Hello, World!")
		assert.Equal(t, "", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "", mockLogger.LastPrintedMessage.Level)
	})

	t.Run("Log is disabled and debug level is off", func(t *testing.T) {
		mockLogger.Clear()
		baseCtx := &BaseContext{
			shouldLog: false,
		}

		baseCtx.LogDebugf("Test log message: %s", "Hello, World!")
		assert.Equal(t, "", mockLogger.LastPrintedMessage.Message)
		assert.Equal(t, "", mockLogger.LastPrintedMessage.Level)
	})
}
