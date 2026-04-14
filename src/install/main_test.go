package install

import (
	"os"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigFromEnv_UsesSpaceSeparatedServiceFlags(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	t.Setenv(constants.API_PORT_ENV_VAR, "")
	t.Setenv(constants.ENABLED_MODULES_ENV_VAR, "")
	t.Setenv(constants.MODE_ENV_VAR, "")

	originalArgs := os.Args
	os.Args = []string{"prldevops", "install", "service", "--modules", "orchestrator", "--api-port", "4090"}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	cfg := getConfigFromEnv()

	assert.Equal(t, "4090", cfg.Port)
	assert.Equal(t, "api,orchestrator", cfg.EnabledModules)
}

func TestGeneratePlist_IncludesModulesFromEqualsStyleServiceFlag(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	t.Setenv(constants.ENABLED_MODULES_ENV_VAR, "")
	t.Setenv(constants.MODE_ENV_VAR, "")

	originalArgs := os.Args
	os.Args = []string{"prldevops", "install", "service", "--modules=host,orchestrator"}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	cfg := getConfigFromEnv()
	plist, err := generatePlist("/usr/local/bin", cfg)
	require.NoError(t, err)

	assert.Contains(t, plist, "<key>"+constants.ENABLED_MODULES_ENV_VAR+"</key>")
	assert.Contains(t, plist, "<string>api,orchestrator,host</string>")
}
