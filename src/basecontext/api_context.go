package basecontext

import (
	"context"

	"github.com/Parallels/prl-devops-service/models"
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
	LogInfo(format string, a ...interface{})
	LogError(format string, a ...interface{})
	LogDebug(format string, a ...interface{})
	LogWarn(format string, a ...interface{})
}
