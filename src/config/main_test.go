package config

import (
	"os"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	str "github.com/stretchr/testify/assert"
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	originalValue, existed := os.LookupEnv(key)
	str.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, originalValue)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func TestGetKeyUsesConfigFileValueWhenEnvIsUnset(t *testing.T) {
	unsetEnv(t, constants.API_PORT_ENV_VAR)
	unsetEnv(t, constants.DATABASE_FOLDER_ENV_VAR)

	cfg := New(basecontext.NewBaseContext())
	cfg.config.Environment = map[string]string{
		constants.API_PORT_ENV_VAR:        "8080",
		constants.DATABASE_FOLDER_ENV_VAR: "/config/data",
	}

	str.Equal(t, "8080", cfg.GetKey(constants.API_PORT_ENV_VAR))
	str.Equal(t, "/config/data", cfg.GetKey(constants.DATABASE_FOLDER_ENV_VAR))
}

func TestGetKeyEnvironmentOverridesConfigFileValue(t *testing.T) {
	t.Setenv(constants.API_PORT_ENV_VAR, "9090")
	t.Setenv(constants.DATABASE_FOLDER_ENV_VAR, "/env/data")

	cfg := New(basecontext.NewBaseContext())
	cfg.config.Environment = map[string]string{
		constants.API_PORT_ENV_VAR:        "8080",
		constants.DATABASE_FOLDER_ENV_VAR: "/config/data",
	}

	str.Equal(t, "9090", cfg.GetKey(constants.API_PORT_ENV_VAR))
	str.Equal(t, "/env/data", cfg.GetKey(constants.DATABASE_FOLDER_ENV_VAR))
}

func TestGetKeyMatchesConfigFileKeysCaseInsensitively(t *testing.T) {
	unsetEnv(t, constants.API_PORT_ENV_VAR)

	cfg := New(basecontext.NewBaseContext())
	cfg.config.Environment = map[string]string{
		"api_port": "8080",
	}

	str.Equal(t, "8080", cfg.GetKey(constants.API_PORT_ENV_VAR))
}

func TestGetKeyEmptyEnvironmentValueOverridesConfigFileValue(t *testing.T) {
	t.Setenv(constants.API_PORT_ENV_VAR, "")

	cfg := New(basecontext.NewBaseContext())
	cfg.config.Environment = map[string]string{
		constants.API_PORT_ENV_VAR: "8080",
	}

	str.Equal(t, "", cfg.GetKey(constants.API_PORT_ENV_VAR))
}
