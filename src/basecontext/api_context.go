package basecontext

import (
	"context"

	"github.com/Parallels/pd-api-service/models"
)

type ApiContext interface {
	Context() context.Context
	GetAuthorizationContext() *AuthorizationContext
	GetRequestId() string
	GetUser() *models.ApiUser
	Verbose() bool
	EnableLog()
	DisableLog()
	ToggleLogTimestamps(value bool)
	LogInfof(format string, a ...interface{})
	LogErrorf(format string, a ...interface{})
	LogDebugf(format string, a ...interface{})
	LogWarnf(format string, a ...interface{})
}
