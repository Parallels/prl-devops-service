package basecontext

import (
	"context"

	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/errors"
	log "github.com/cjlapao/common-go-logger"
)

type ApiContext interface {
	Context() context.Context
	GetAuthorizationContext() *AuthorizationContext
	GetRequestId() string
	GetUser(diag *errors.Diagnostics) *models.ApiUser
	Verbose() bool
	EnableLog()
	DisableLog()
	ToggleLogTimestamps(value bool)
	Logger() *log.LoggerService
	LogInfof(format string, a ...interface{})
	LogErrorf(format string, a ...interface{})
	LogDebugf(format string, a ...interface{})
	LogWarnf(format string, a ...interface{})
	LogTracef(format string, a ...interface{})
}
