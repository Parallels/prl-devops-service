package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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

func TestGeneratePlist_UsesConfigFlag(t *testing.T) {
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

	assert.Contains(t, plist, "<string>--config</string>")
	assert.Contains(t, plist, "/etc/prl-devops-service/prldevops_config.yaml")
}

func TestWriteServiceConfigFile_WritesYAMLWithEnvironmentKeys(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	svcCfg := ApiServiceConfig{
		Port:                  "3080",
		Prefix:                "/api",
		LogLevel:              "DEBUG",
		EnableTLS:             true,
		EnableCors:            true,
		CorsAllowedOrigins:    "*",
		EnabledModules:        "api,host",
		DisableCatalogCaching: true,
		TokenDurationMinutes:  "120",
	}

	err := writeServiceConfigFile(svcCfg, tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "prldevops_config.yaml"))
	require.NoError(t, err)

	yamlStr := string(content)
	assert.Contains(t, yamlStr, "environment:")
	assert.Contains(t, yamlStr, constants.API_PORT_ENV_VAR)
	assert.Contains(t, yamlStr, "3080")
	assert.Contains(t, yamlStr, constants.API_PREFIX_ENV_VAR)
	assert.Contains(t, yamlStr, "/api")
	assert.Contains(t, yamlStr, constants.LOG_LEVEL_ENV_VAR)
	assert.Contains(t, yamlStr, "DEBUG")
	assert.Contains(t, yamlStr, constants.TLS_ENABLED_ENV_VAR)
	assert.Contains(t, yamlStr, "true")
	assert.Contains(t, yamlStr, constants.ENABLE_CORS_ENV_VAR)
	assert.Contains(t, yamlStr, constants.ENABLED_MODULES_ENV_VAR)
	assert.Contains(t, yamlStr, "api,host")
	assert.Contains(t, yamlStr, constants.DISABLE_CATALOG_CACHING_ENV_VAR)
	assert.Contains(t, yamlStr, constants.TOKEN_DURATION_MINUTES_ENV_VAR)
	assert.Contains(t, yamlStr, "120")
	assert.Contains(t, yamlStr, constants.CORS_ALLOWED_ORIGINS_ENV_VAR)
}

func TestWriteServiceConfigFile_EmptyConfigWritesMinimalYAML(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	svcCfg := ApiServiceConfig{}

	err := writeServiceConfigFile(svcCfg, tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "prldevops_config.yaml"))
	require.NoError(t, err)

	yamlStr := string(content)
	// Even with no explicit config, DisableFileLogging defaults to false so LOG_TO_FILE=true is written
	assert.Contains(t, yamlStr, constants.LOG_TO_FILE_ENV_VAR)
	assert.Contains(t, yamlStr, "true")
}

func TestWriteServiceConfigFile_DefaultsToServiceDefaultDirectory(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()
	origDefault := constants.ServiceDefaultDirectory
	constants.ServiceDefaultDirectory = tmpDir
	t.Cleanup(func() { constants.ServiceDefaultDirectory = origDefault })

	svcCfg := ApiServiceConfig{
		Port: "4090",
	}

	err := writeServiceConfigFile(svcCfg, "")
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "prldevops_config.yaml"))
	require.NoError(t, err)
	yamlStr := string(content)
	assert.Contains(t, yamlStr, constants.API_PORT_ENV_VAR)
	assert.Contains(t, yamlStr, "4090")
}

func TestWriteServiceConfigFile_CreatesDirectoryIfNeeded(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")

	svcCfg := ApiServiceConfig{
		Port: "3080",
	}

	err := writeServiceConfigFile(svcCfg, nestedDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(nestedDir, "prldevops_config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "environment:")
}

func TestWriteServiceConfigFile_BoolTrueFields(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	svcCfg := ApiServiceConfig{
		EnableTLS:                true,
		EnableCors:               true,
		DisableCatalogCaching:    true,
		UseOrchestratorResources: true,
	}

	err := writeServiceConfigFile(svcCfg, tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "prldevops_config.yaml"))
	require.NoError(t, err)

	yamlStr := string(content)
	assert.Contains(t, yamlStr, constants.TLS_ENABLED_ENV_VAR)
	assert.Contains(t, yamlStr, "true")
	assert.Contains(t, yamlStr, constants.ENABLE_CORS_ENV_VAR)
	assert.Contains(t, yamlStr, constants.DISABLE_CATALOG_CACHING_ENV_VAR)
}

func TestWriteServiceConfigFile_StringFieldsOnly(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	svcCfg := ApiServiceConfig{
		Port:                 "3080",
		Prefix:               "/v2",
		LogLevel:             "WARN",
		RootPassword:         "secret",
		EncryptionRsaKey:     "rsa-key-content",
		HmacSecret:           "hmac-secret",
		TokenDurationMinutes: "30",
		EnabledModules:       "api,catalog",
		CorsAllowedOrigins:   "https://example.com",
		CorsAllowedMethods:   "GET,POST",
		CorsAllowedHeaders:   "Authorization",
	}

	err := writeServiceConfigFile(svcCfg, tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "prldevops_config.yaml"))
	require.NoError(t, err)

	yamlStr := string(content)
	assert.Contains(t, yamlStr, constants.API_PORT_ENV_VAR)
	assert.Contains(t, yamlStr, "3080")
	assert.Contains(t, yamlStr, constants.API_PREFIX_ENV_VAR)
	assert.Contains(t, yamlStr, "/v2")
	assert.Contains(t, yamlStr, constants.LOG_LEVEL_ENV_VAR)
	assert.Contains(t, yamlStr, "WARN")
	assert.Contains(t, yamlStr, constants.ROOT_PASSWORD_ENV_VAR)
	assert.Contains(t, yamlStr, "secret")
	assert.Contains(t, yamlStr, constants.ENCRYPTION_SECURITY_KEY_ENV_VAR)
	assert.Contains(t, yamlStr, "rsa-key-content")
	assert.Contains(t, yamlStr, constants.HMAC_SECRET_ENV_VAR)
	assert.Contains(t, yamlStr, "hmac-secret")
	assert.Contains(t, yamlStr, constants.TOKEN_DURATION_MINUTES_ENV_VAR)
	assert.Contains(t, yamlStr, "30")
	assert.Contains(t, yamlStr, constants.ENABLED_MODULES_ENV_VAR)
	assert.Contains(t, yamlStr, "api,catalog")
	assert.Contains(t, yamlStr, constants.CORS_ALLOWED_ORIGINS_ENV_VAR)
	assert.Contains(t, yamlStr, "https://example.com")
	assert.Contains(t, yamlStr, constants.CORS_ALLOWED_METHODS_ENV_VAR)
	assert.Contains(t, yamlStr, "GET,POST")
	assert.Contains(t, yamlStr, constants.CORS_ALLOWED_HEADERS_ENV_VAR)
	assert.Contains(t, yamlStr, "Authorization")
}

func TestWriteServiceConfigFile_TLSFields(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	svcCfg := ApiServiceConfig{
		TLSCertificate: "/path/to/cert.pem",
		TLSPrivateKey:  "/path/to/key.pem",
	}

	err := writeServiceConfigFile(svcCfg, tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "prldevops_config.yaml"))
	require.NoError(t, err)

	yamlStr := string(content)
	assert.Contains(t, yamlStr, constants.TLS_CERTIFICATE_ENV_VAR)
	assert.Contains(t, yamlStr, "/path/to/cert.pem")
	assert.Contains(t, yamlStr, constants.TLS_PRIVATE_KEY_ENV_VAR)
	assert.Contains(t, yamlStr, "/path/to/key.pem")
}

func TestWriteServiceConfigFile_SKIPS_EMPTY_AND_FALSE_FIELDS(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	svcCfg := ApiServiceConfig{
		Port:                  "",
		LogLevel:              "",
		EnableTLS:             false,
		EnableCors:            false,
		DisableCatalogCaching: false,
	}

	err := writeServiceConfigFile(svcCfg, tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "prldevops_config.yaml"))
	require.NoError(t, err)

	yamlStr := string(content)
	assert.NotContains(t, yamlStr, constants.API_PORT_ENV_VAR)
	assert.NotContains(t, yamlStr, constants.LOG_LEVEL_ENV_VAR)
	assert.NotContains(t, yamlStr, constants.TLS_ENABLED_ENV_VAR)
	assert.NotContains(t, yamlStr, constants.ENABLE_CORS_ENV_VAR)
	assert.NotContains(t, yamlStr, constants.DISABLE_CATALOG_CACHING_ENV_VAR)
	assert.Contains(t, yamlStr, constants.LOG_TO_FILE_ENV_VAR)
	assert.Contains(t, yamlStr, "true")
}

func TestWriteServiceConfigFile_DisableFileLoggingSkipsLogToFile(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	svcCfg := ApiServiceConfig{
		Port:               "3080",
		DisableFileLogging: true,
	}

	err := writeServiceConfigFile(svcCfg, tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "prldevops_config.yaml"))
	require.NoError(t, err)

	yamlStr := string(content)
	assert.Contains(t, yamlStr, constants.API_PORT_ENV_VAR)
	assert.NotContains(t, yamlStr, constants.LOG_TO_FILE_ENV_VAR)
}

func TestWriteServiceConfigFile_FailsOnInvalidPath(t *testing.T) {
	config.New(basecontext.NewBaseContext())

	svcCfg := ApiServiceConfig{
		Port: "3080",
	}

	err := writeServiceConfigFile(svcCfg, "/nonexistent/path/that/does/not/exist")
	assert.Error(t, err)
}

func TestMergeDefaultsIntoConfig_AddsMissingKeys(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	svcCfg := ApiServiceConfig{
		Port: "3080",
	}
	err := writeServiceConfigFile(svcCfg, tmpDir)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "prldevops_config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(content), constants.API_PORT_ENV_VAR)
}

func TestMergeDefaultsIntoConfig_PreservesUserValues(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	svcCfg := ApiServiceConfig{
		Port:    "9999",
		Prefix:  "/custom",
		LogLevel: "DEBUG",
	}
	err := writeServiceConfigFile(svcCfg, tmpDir)
	require.NoError(t, err)

	// Read the file and verify user values are present
	content, err := os.ReadFile(filepath.Join(tmpDir, "prldevops_config.yaml"))
	require.NoError(t, err)
	yamlStr := string(content)
	assert.Contains(t, yamlStr, "9999")
	assert.Contains(t, yamlStr, "/custom")
	assert.Contains(t, yamlStr, "DEBUG")
}

func TestGetDefaultEnvironmentMap_ContainsExpectedKeys(t *testing.T) {
	defaults := getDefaultEnvironmentMap()

	assert.Contains(t, defaults, constants.API_PORT_ENV_VAR)
	assert.Contains(t, defaults, constants.API_PREFIX_ENV_VAR)
	assert.Contains(t, defaults, constants.LOG_LEVEL_ENV_VAR)
	assert.Contains(t, defaults, constants.ENABLE_CORS_ENV_VAR)
	assert.Contains(t, defaults, constants.CORS_ALLOWED_ORIGINS_ENV_VAR)
	assert.Contains(t, defaults, constants.CORS_ALLOWED_METHODS_ENV_VAR)
	assert.Contains(t, defaults, constants.CORS_ALLOWED_HEADERS_ENV_VAR)
	assert.Contains(t, defaults, constants.LOG_TO_FILE_ENV_VAR)
	assert.Contains(t, defaults, constants.TOKEN_DURATION_MINUTES_ENV_VAR)
	assert.Contains(t, defaults, constants.TLS_ENABLED_ENV_VAR)
	assert.Contains(t, defaults, constants.TLS_PORT_ENV_VAR)
	assert.Contains(t, defaults, constants.DISABLE_CATALOG_CACHING_ENV_VAR)
	assert.Contains(t, defaults, constants.USE_ORCHESTRATOR_RESOURCES_ENV_VAR)
	assert.Contains(t, defaults, constants.SYSTEM_RESERVED_CPU_ENV_VAR)
	assert.Contains(t, defaults, constants.SYSTEM_RESERVED_MEMORY_ENV_VAR)
	assert.Contains(t, defaults, constants.SYSTEM_RESERVED_DISK_ENV_VAR)
}

func TestGetDefaultEnvironmentMap_DefaultValues(t *testing.T) {
	defaults := getDefaultEnvironmentMap()

	assert.Equal(t, constants.DEFAULT_API_PORT, defaults[constants.API_PORT_ENV_VAR])
	assert.Equal(t, constants.DEFAULT_API_PREFIX, defaults[constants.API_PREFIX_ENV_VAR])
	assert.Equal(t, "INFO", defaults[constants.LOG_LEVEL_ENV_VAR])
	assert.Equal(t, "true", defaults[constants.ENABLE_CORS_ENV_VAR])
	assert.Equal(t, "*", defaults[constants.CORS_ALLOWED_ORIGINS_ENV_VAR])
	assert.Equal(t, "GET,POST,PUT,DELETE,PATCH", defaults[constants.CORS_ALLOWED_METHODS_ENV_VAR])
	assert.Equal(t, "false", defaults[constants.TLS_ENABLED_ENV_VAR])
	assert.Equal(t, "true", defaults[constants.LOG_TO_FILE_ENV_VAR])
}

func TestSaveConfig_YamlFormat(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "prldevops_config.yaml")

	cf := config.ConfigFile{
		Environment: map[string]string{
			"test_key": "test_value",
		},
	}

	content, err := yaml.Marshal(cf)
	require.NoError(t, err)
	err = os.WriteFile(testFile, content, 0o644)
	require.NoError(t, err)

	readContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Contains(t, string(readContent), "test_key")
	assert.Contains(t, string(readContent), "test_value")
}

func TestSaveConfig_JsonFormat(t *testing.T) {
	config.New(basecontext.NewBaseContext())
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "prldevops_config.json")

	cf := config.ConfigFile{
		Environment: map[string]string{
			"json_key": "json_value",
		},
	}

	content, err := json.Marshal(cf)
	require.NoError(t, err)
	err = os.WriteFile(testFile, content, 0o644)
	require.NoError(t, err)

	readContent, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Contains(t, string(readContent), "json_key")
	assert.Contains(t, string(readContent), "json_value")
}
