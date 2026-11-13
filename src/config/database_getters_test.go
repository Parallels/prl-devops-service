package config

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	str "github.com/stretchr/testify/require"
)

// DB_* getters (PBI #420 task 02, D-3): dumb raw readers on GetKey. Seeding
// idiom copied from main_test.go (direct Environment map + unsetEnv helper).

func seedCfg(t *testing.T, env map[string]string) *Config {
	t.Helper()
	cfg := New(basecontext.NewBaseContext())
	cfg.config.Environment = env
	return cfg
}

func TestDatabaseGetters_ReadFlatConfigKeys(t *testing.T) {
	for _, k := range []string{constants.DB_TYPE_ENV_VAR, constants.DB_PATH_ENV_VAR, constants.DB_DSN_ENV_VAR, constants.DB_AUTO_MIGRATE_ENV_VAR} {
		unsetEnv(t, k)
	}
	cfg := seedCfg(t, map[string]string{
		"db_type":         "postgres",
		"db_path":         "/srv/prl.db",
		"db_dsn":          testDSN("user", "pw", "h", "d"),
		"db_auto_migrate": "false",
	})
	str.Equal(t, "postgres", cfg.DatabaseType())
	str.Equal(t, "/srv/prl.db", cfg.DatabasePath())
	str.Equal(t, testDSN("user", "pw", "h", "d"), cfg.DatabaseDSN())
	v, err := cfg.DatabaseAutoMigrate()
	str.NoError(t, err)
	str.False(t, v)
}

func TestDatabaseGetters_EnvBeatsFileKey(t *testing.T) {
	t.Setenv("DB_TYPE", "sqlite")
	cfg := seedCfg(t, map[string]string{"db_type": "postgres"})
	str.Equal(t, "sqlite", cfg.DatabaseType()) // AC: env-var precedence (GetKey mechanism)
}

func TestDatabaseAutoMigrate_TruthyFalsyVocabulary(t *testing.T) {
	cases := map[string]bool{
		"true": true, "1": true, "t": true, "yes": true, "y": true, "on": true, "enabled": true,
		"TRUE": true, " Yes ": true, // case + whitespace tolerant
		"false": false, "0": false, "f": false, "no": false, "n": false, "off": false, "disabled": false,
	}
	for raw, want := range cases {
		t.Setenv(constants.DB_AUTO_MIGRATE_ENV_VAR, raw)
		got, err := seedCfg(t, nil).DatabaseAutoMigrate()
		str.NoErrorf(t, err, "raw=%q", raw)
		str.Equalf(t, want, got, "raw=%q", raw)
	}
}

func TestDatabaseAutoMigrate_UnsetDefaultsTrue(t *testing.T) {
	unsetEnv(t, constants.DB_AUTO_MIGRATE_ENV_VAR)
	got, err := seedCfg(t, nil).DatabaseAutoMigrate()
	str.NoError(t, err)
	str.True(t, got) // AC: "Defaults to true"
}

func TestDatabaseAutoMigrate_GarbageErrors(t *testing.T) {
	t.Setenv(constants.DB_AUTO_MIGRATE_ENV_VAR, "maybe")
	_, err := seedCfg(t, nil).DatabaseAutoMigrate()
	str.Error(t, err)
	str.Contains(t, err.Error(), "db_auto_migrate") // actionable: names the setting + key
	str.Contains(t, err.Error(), "true")
	str.Contains(t, err.Error(), "false")
}

func TestDatabaseAutoMigrate_GarbageEchoCapped(t *testing.T) {
	long := ""
	for i := 0; i < 100; i++ {
		long += "abcdefgh"
	}
	t.Setenv(constants.DB_AUTO_MIGRATE_ENV_VAR, long)
	_, err := seedCfg(t, nil).DatabaseAutoMigrate()
	str.Error(t, err)
	str.Contains(t, err.Error(), "…")     // capped echo marker
	str.NotContains(t, err.Error(), long) // full garbage never echoed
}
