package basecontext

import (
	"context"
	"net/http"

	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/constants"
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
	LogInfo(format string, a ...interface{})
	LogError(format string, a ...interface{})
	LogDebug(format string, a ...interface{})
	LogWarn(format string, a ...interface{})
}

type BaseContext struct {
	shouldLog   bool
	ctx         context.Context
	authContext *AuthorizationContext
	User        models.ApiUser
}

func NewBaseContext() *BaseContext {
	baseContext := &BaseContext{
		ctx: context.Background(),
	}

	return baseContext
}

func NewRootBaseContext() *BaseContext {
	baseContext := &BaseContext{
		ctx: context.Background(),
		authContext: &AuthorizationContext{
			IsAuthorized: true,
			AuthorizedBy: "RootAuthorization",
		},
	}

	return baseContext
}

func NewBaseContextFromRequest(r *http.Request) *BaseContext {
	baseContext := &BaseContext{
		ctx: r.Context(),
	}

	authContext := baseContext.ctx.Value(constants.AUTHORIZATION_CONTEXT_KEY)
	if authContext != nil {
		baseContext.authContext = authContext.(*AuthorizationContext)
	}

	return baseContext
}

func NewBaseContextFromContext(c context.Context) *BaseContext {
	baseContext := &BaseContext{
		ctx: c,
	}

	authContext := baseContext.ctx.Value(constants.AUTHORIZATION_CONTEXT_KEY)
	if authContext != nil {
		baseContext.authContext = authContext.(*AuthorizationContext)
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
	common.Logger.EnableTimestamp(value)
}

func (c *BaseContext) LogInfo(format string, a ...interface{}) {
	// log is disabled, returning
	if !c.shouldLog {
		return
	}

	msg := ""
	if c.GetRequestId() != "" {
		msg = "[" + c.GetRequestId() + "] "
	}
	msg += format
	common.Logger.Info(msg, a...)
}

func (c *BaseContext) LogError(format string, a ...interface{}) {
	// log is disabled, returning
	if !c.shouldLog {
		return
	}

	msg := ""
	if c.GetRequestId() != "" {
		msg = "[" + c.GetRequestId() + "] "
	}
	msg += format
	common.Logger.Error(msg, a...)
}

func (c *BaseContext) LogDebug(format string, a ...interface{}) {
	// log is disabled, returning
	if !c.shouldLog {
		return
	}

	msg := ""
	if c.GetRequestId() != "" {
		msg = "[" + c.GetRequestId() + "] "
	}
	msg += format
	common.Logger.Debug(msg, a...)
}

func (c *BaseContext) LogWarn(format string, a ...interface{}) {
	// log is disabled, returning
	if !c.shouldLog {
		return
	}

	msg := ""
	if c.GetRequestId() != "" {
		msg = "[" + c.GetRequestId() + "] "
	}
	msg += format
	common.Logger.Warn(msg, a...)
}
