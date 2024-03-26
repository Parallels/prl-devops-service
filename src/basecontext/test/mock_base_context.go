package basecontext_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/models"
)

type MockBaseContext struct {
	shouldLog                bool
	MockContext              context.Context
	MockRequestID            string
	MockAuthorizationContext *basecontext.AuthorizationContext
	MockApiUser              *models.ApiUser
	callbackFunctions        map[string]func(args ...interface{})
}

func NewMockBaseContext() *MockBaseContext {
	return &MockBaseContext{
		callbackFunctions: make(map[string]func(args ...interface{})),
	}
}

func (c *MockBaseContext) On(functionName string, callback func(args ...interface{})) {
	found := false
	for k := range c.callbackFunctions {
		if strings.EqualFold(k, functionName) {
			c.callbackFunctions[k] = callback
			return
		}
	}

	if !found {
		c.callbackFunctions[functionName] = callback
	}
}

func (m *MockBaseContext) Context() context.Context {
	if m.callbackFunctions["Context"] != nil {
		m.callbackFunctions["Context"]()
	}

	return m.MockContext
}

func (m *MockBaseContext) GetAuthorizationContext() *basecontext.AuthorizationContext {
	if m.callbackFunctions["GetAuthorizationContext"] != nil {
		m.callbackFunctions["GetAuthorizationContext"]()
	}

	return m.MockAuthorizationContext
}

func (m *MockBaseContext) GetRequestId() string {
	if m.callbackFunctions["GetRequestId"] != nil {
		m.callbackFunctions["GetRequestId"]()
	}

	return m.MockRequestID
}

func (m *MockBaseContext) GetUser() *models.ApiUser {
	if m.callbackFunctions["GetUser"] != nil {
		m.callbackFunctions["GetUser"]()
	}

	return m.MockApiUser
}

func (m *MockBaseContext) Verbose() bool {
	if m.callbackFunctions["Verbose"] != nil {
		m.callbackFunctions["Verbose"]()
	}

	return m.shouldLog
}

func (m *MockBaseContext) EnableLog() {
	if m.callbackFunctions["EnableLog"] != nil {
		m.callbackFunctions["EnableLog"]()
	}
}

func (m *MockBaseContext) DisableLog() {
	if m.callbackFunctions["DisableLog"] != nil {
		m.callbackFunctions["DisableLog"]()
	}
}

func (m *MockBaseContext) ToggleLogTimestamps(value bool) {
	if m.callbackFunctions["ToggleLogTimestamps"] != nil {
		m.callbackFunctions["ToggleLogTimestamps"](value)
	}
}

func (m *MockBaseContext) LogInfof(format string, a ...interface{}) {
	if m.callbackFunctions["LogInfof"] != nil {
		value := fmt.Sprintf(format, a...)
		m.callbackFunctions["LogInfof"](value)
	}
}

func (m *MockBaseContext) LogErrorf(format string, a ...interface{}) {
	if m.callbackFunctions["LogErrorf"] != nil {
		value := fmt.Sprintf(format, a...)
		m.callbackFunctions["LogErrorf"](value)
	}
}

func (m *MockBaseContext) LogDebugf(format string, a ...interface{}) {
	if m.callbackFunctions["LogDebugf"] != nil {
		value := fmt.Sprintf(format, a...)
		m.callbackFunctions["LogDebugf"](value)
	}
}

func (m *MockBaseContext) LogWarnf(format string, a ...interface{}) {
	if m.callbackFunctions["LogWarnf"] != nil {
		value := fmt.Sprintf(format, a...)
		m.callbackFunctions["LogWarnf"](value)
	}
}

func (m *MockBaseContext) LogTracef(format string, a ...interface{}) {
	if m.callbackFunctions["LogTracef"] != nil {
		value := fmt.Sprintf(format, a...)
		m.callbackFunctions["LogTracef"](value)
	}
}
