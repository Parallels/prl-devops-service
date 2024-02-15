package telemetry_test

import (
	"strings"

	"github.com/amplitude/analytics-go/amplitude/types"
)

type MockAmplitudeClient struct {
	callbackFunctions map[string]func(args ...interface{})
}

func NewMockAmplitudeClient() *MockAmplitudeClient {
	return &MockAmplitudeClient{
		callbackFunctions: make(map[string]func(args ...interface{})),
	}
}

func (c *MockAmplitudeClient) On(functionName string, callback func(args ...interface{})) {
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

func (c *MockAmplitudeClient) Track(event types.Event) {
	if c.callbackFunctions["Track"] != nil {
		c.callbackFunctions["Track"](event)
	}
}

func (c *MockAmplitudeClient) Identify(identify types.Identify, eventOptions types.EventOptions) {
	if c.callbackFunctions["Identify"] != nil {
		c.callbackFunctions["Identify"](identify, eventOptions)
	}
}

func (c *MockAmplitudeClient) GroupIdentify(groupType string, groupName string, identify types.Identify, eventOptions types.EventOptions) {
	if c.callbackFunctions["GroupIdentify"] != nil {
		c.callbackFunctions["GroupIdentify"](groupType, groupName, identify, eventOptions)
	}
}

func (c *MockAmplitudeClient) SetGroup(groupType string, groupName []string, eventOptions types.EventOptions) {
	if c.callbackFunctions["SetGroup"] != nil {
		c.callbackFunctions["SetGroup"](groupType, groupName, eventOptions)
	}
}

func (c *MockAmplitudeClient) Revenue(revenue types.Revenue, eventOptions types.EventOptions) {
	if c.callbackFunctions["Revenue"] != nil {
		c.callbackFunctions["Revenue"](revenue, eventOptions)
	}
}

func (c *MockAmplitudeClient) Flush() {
	if c.callbackFunctions["Flush"] != nil {
		c.callbackFunctions["Flush"]()
	}
}

func (c *MockAmplitudeClient) Shutdown() {
	if c.callbackFunctions["Shutdown"] != nil {
		c.callbackFunctions["Shutdown"]()
	}
}

func (c *MockAmplitudeClient) Add(plugin types.Plugin) {
	if c.callbackFunctions["Add"] != nil {
		c.callbackFunctions["Add"](plugin)
	}
}

func (c *MockAmplitudeClient) Remove(pluginName string) {
	if c.callbackFunctions["Remove"] != nil {
		c.callbackFunctions["Remove"](pluginName)
	}
}

func (c *MockAmplitudeClient) Config() types.Config {
	if c.callbackFunctions["Config"] != nil {
		c.callbackFunctions["Config"]()
	}

	return types.Config{}
}
