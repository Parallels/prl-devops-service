package orchestrator

import (
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
)

func TestShellFlag_QuotesValuesWithSpaces(t *testing.T) {
	got := shellFlag(constants.HOST_NAME_FLAG, "MacBook Pro M2")

	assert.Equal(t, "--host-name='MacBook Pro M2'", got)
}

func TestRegisterCommand_KeepsHostNameAndTagsAsSingleShellArgs(t *testing.T) {
	agentBaseURL := "http://10.0.5.236:3080"
	orchURL := "http://10.0.5.1:3080"
	token := "abc123"
	hostName := "MacBook Pro M2"
	tags := "macos,lab machine"
	agentPort := "3080"

	registerCmd := "BASE_URL=" + shellSingleQuote(agentBaseURL) +
		" /usr/local/bin/prldevops register-with-orchestrator " +
		shellFlag(constants.ORCHESTRATOR_URL_FLAG, orchURL) + " " +
		shellFlag(constants.ORCHESTRATOR_TOKEN_FLAG, token) + " " +
		shellFlag(constants.HOST_NAME_FLAG, hostName) + " " +
		shellFlag(constants.API_PORT_FLAG, agentPort) + " " +
		shellFlag(constants.TAGS_FLAG, tags)

	assert.Contains(t, registerCmd, "--host-name='MacBook Pro M2'")
	assert.Contains(t, registerCmd, "--tags='macos,lab machine'")
	assert.Contains(t, registerCmd, "BASE_URL='http://10.0.5.236:3080'")
}
