package install

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"
)

func TestCompleteConfigurationPreservesSectionsAndProtectsSecrets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prldevops_config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("environment:\n  STALE: old\nreverse_proxy:\n  enabled: true\n"), 0644))
	cfg := ApiServiceConfig{TLSPort: "8443", UseOrchestratorResources: true, SystemReservedMemory: "0", SystemReservedCPU: "15", SystemReservedDisk: "100", RootPassword: "quote'\"$`\npassword", Environment: map[string]string{"TLS_ENABLED": "false", "LOG_TO_FILE": "false", "CUSTOM": "0"}}
	require.NoError(t, writeServiceConfigFile(cfg, dir))
	bytes, err := os.ReadFile(path)
	require.NoError(t, err)
	var doc map[string]interface{}
	require.NoError(t, yaml.Unmarshal(bytes, &doc))
	require.Contains(t, doc, "reverse_proxy")
	env := doc["environment"].(map[string]interface{})
	require.Equal(t, "8443", env["TLS_PORT"])
	require.Equal(t, "false", env["TLS_ENABLED"])
	require.Equal(t, "false", env["LOG_TO_FILE"])
	require.Equal(t, "true", env["USE_ORCHESTRATOR_RESOURCES"])
	require.Equal(t, "0", env["SYSTEM_RESERVED_MEMORY"])
	require.Equal(t, cfg.RootPassword, env["ROOT_PASSWORD"])
	require.NotContains(t, env, "STALE")
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestInstallerReadsCertificateAndCachingFromCorrectKeys(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	t.Setenv(constants.TLS_CERTIFICATE_ENV_VAR, "certificate")
	t.Setenv(constants.TLS_PRIVATE_KEY_ENV_VAR, "private-key")
	t.Setenv(constants.ROOT_PASSWORD_ENV_VAR, "not-a-boolean")
	t.Setenv(constants.DISABLE_CATALOG_CACHING_ENV_VAR, "true")
	cfg := getConfigFromEnv()
	require.Equal(t, "certificate", cfg.TLSCertificate)
	require.Equal(t, "private-key", cfg.TLSPrivateKey)
	require.True(t, cfg.DisableCatalogCaching)
}
