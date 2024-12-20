package basecontext

import (
	"context"
	"net/http"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	log "github.com/cjlapao/common-go-logger"
)

var logger = log.Get().WithTimestamp()

type BaseContext struct {
	shouldLog   bool
	ctx         context.Context
	authContext *AuthorizationContext
	User        models.ApiUser
}

func NewBaseContext() *BaseContext {
	baseContext := &BaseContext{
		shouldLog: true,
		ctx:       context.Background(),
	}

	return baseContext
}

func NewRootBaseContext() *BaseContext {
	baseContext := &BaseContext{
		ctx:       context.Background(),
		shouldLog: true,
		authContext: &AuthorizationContext{
			IsAuthorized: true,
			AuthorizedBy: "RootAuthorization",
		},
	}

	return baseContext
}

func NewBaseContextFromRequest(r *http.Request) *BaseContext {
	baseContext := &BaseContext{
		shouldLog: true,
		ctx:       r.Context(),
	}

	authContext := baseContext.ctx.Value(constants.AUTHORIZATION_CONTEXT_KEY)
	if authContext != nil {
		if ctx, ok := authContext.(*AuthorizationContext); ok {
			baseContext.authContext = ctx
		}
	}

	return baseContext
}

func NewBaseContextFromContext(c context.Context) *BaseContext {
	baseContext := &BaseContext{
		shouldLog: true,
		ctx:       c,
	}

	authContext := baseContext.ctx.Value(constants.AUTHORIZATION_CONTEXT_KEY)
	if authContext != nil {
		if ctx, ok := authContext.(*AuthorizationContext); ok {
			baseContext.authContext = ctx
		}
	}

	return baseContext
}

func (c *BaseContext) GetAuthorizationContext() *AuthorizationContext {
	return c.authContext
}

func (c *BaseContext) Context() context.Context {
	return c.ctx
}

func (c *BaseContext) GetRequestId() string {
	if c.ctx == nil {
		return ""
	}

	id := c.ctx.Value(constants.REQUEST_ID_KEY)
	if id == nil {
		return ""
	}

	return id.(string)
}

func (c *BaseContext) GetUser() *models.ApiUser {
	if c.authContext != nil {
		return c.authContext.User
	}

	return nil
}

func (c *BaseContext) Verbose() bool {
	return c.shouldLog
}

func (c *BaseContext) EnableLog() {
	c.shouldLog = true
}

func (c *BaseContext) DisableLog() {
	c.shouldLog = false
}

func (c *BaseContext) ToggleLogTimestamps(value bool) {
	logger.EnableTimestamp(value)
}

func (c *BaseContext) EnableLogFile(filename string) {
	logger.AddFileLogger(filename)
}

func (c *BaseContext) LogInfof(format string, a ...interface{}) {
	// log is disabled, returning
	if !c.shouldLog {
		return
	}

	msg := ""
	if c.GetRequestId() != "" {
		msg = "[" + c.GetRequestId() + "] "
	}

	msg += format
	logger.Info(msg, a...)
}

func (c *BaseContext) LogErrorf(format string, a ...interface{}) {
	// log is disabled, returning
	if !c.shouldLog {
		return
	}

	msg := ""
	if c.GetRequestId() != "" {
		msg = "[" + c.GetRequestId() + "] "
	}
	msg += format
	logger.Error(msg, a...)
}

func (c *BaseContext) LogDebugf(format string, a ...interface{}) {
	if !c.shouldLog {
		return
	}

	msg := ""
	if c.GetRequestId() != "" {
		msg = "[" + c.GetRequestId() + "] "
	}
	msg += format
	logger.Debug(msg, a...)
}

func (c *BaseContext) LogWarnf(format string, a ...interface{}) {
	// log is disabled, returning
	if !c.shouldLog {
		return
	}

	msg := ""
	if c.GetRequestId() != "" {
		msg = "[" + c.GetRequestId() + "] "
	}
	msg += format
	logger.Warn(msg, a...)
}

func (c *BaseContext) LogTracef(format string, a ...interface{}) {
	// log is disabled, returning
	if !c.shouldLog {
		return
	}

	msg := ""
	if c.GetRequestId() != "" {
		msg = "[" + c.GetRequestId() + "] "
	}
	msg += format
	logger.Trace(msg, a...)
}
