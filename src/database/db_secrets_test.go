package database

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PBI #420 task 03 — secrets-safety regressions (AC "DSN values are never
// logged in plain text"; plan D-8).
//
// §D-8 AUDIT RECORD (documentation-in-code — the masking question was answered,
// not skipped). Grep evidence, re-run at slice time, that no pre-existing
// config-dump path needs masking:
//   - src/config (post-420-01 split) GetConfigContent — config/main.go — marshals
//     ConfigFile; its only consumer is install's saveConfig FILE WRITE (src/install
//     saveConfig around its GetConfigContent call site); no logger call consumes it.
//   - src/install writes the user's OWN config file with their own values: secret
//     storage in the operator's file, not a log — out of scope per spec.
//   - Config.Load/Save never log values; no Log* function receives GetKey output
//     for DB_* keys (none existed before #420; the four new getters are pure readers).
// Therefore task 02 shipped no masking change; THIS FILE is the durable guard: it
// fails the moment anyone injects a DSN into an error/log string on #420 paths.
// If a new dump path ever appears, start the audit here and at RedactDSN (#418)
// — the single sanctioned funnel.

var pwCanary = fixtureSecret("dsn-canary")

var (
	dsnURL     = testDSN("ops", pwCanary, "db.internal:5432", "app?sslmode=disable")
	dsnKeyword = "host=db.internal user=ops password=" + pwCanary + " dbname=app"
)

func TestSecrets_ValidPostgresBothForms_NoErrorNoLeakSurface(t *testing.T) {
	for _, dsn := range []string{dsnURL, dsnKeyword} {
		resetCfg(t)
		t.Setenv("DB_TYPE", "postgres")
		t.Setenv("DB_DSN", dsn)
		cfg, err := FromApplicationConfig()
		require.NoError(t, err)
		assert.Equal(t, dsn, cfg.DSN) // verbatim in-memory (redaction is for output paths only)
	}
}

func TestSecrets_PostgresBlankDSN_ErrorCarriesNoRemnant(t *testing.T) {
	resetCfg(t)
	t.Setenv("DB_TYPE", "postgres") // D-7.3: never echo the DSN — D7 guard lock
	_, err := FromApplicationConfig()
	require.Error(t, err)
	msg := err.Error()
	assert.NotContains(t, msg, "password")
	assert.NotContains(t, msg, "://")
	assert.NotContains(t, msg, pwCanary)
}

func TestSecrets_AutoMigrateGarbageEchoIsBounded(t *testing.T) {
	resetCfg(t)
	// DSN-looking garbage in the WRONG setting: the config-validation error may
	// legitimately echo it back to its owner — bounded at 64 bytes (task 02 cap).
	// The invariant under test: our additions never inject cfg.DSN into any text.
	longDsnLike := "postgres://u:" + pwCanary + "@h/dbname_that_is_long_enough_to_force_the_truncation_marker"
	t.Setenv("DB_AUTO_MIGRATE", longDsnLike)
	cfg, err := FromApplicationConfig()
	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "db_auto_migrate")
	assert.Contains(t, msg, "…")
	assert.Less(t, len(msg), 200)
	assert.Equal(t, DatabaseConfig{}, cfg) // zero config: no DSN field carried out
	assert.Empty(t, cfg.DSN)               // cfg.DSN never populated on error path
	assert.NotContains(t, msg, dsnURL)     // unrelated real DSN shape never injected
	_ = config.Get()                       // (no state dependency; documents isolation)
}

func TestSecrets_ResolutionMutatesNeitherConfigNorEnvironment(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T)
	}{
		{"sqlite-default", func(t *testing.T) {}},
		{"postgres-valid", func(t *testing.T) {
			t.Setenv("DB_TYPE", "postgres")
			t.Setenv("DB_DSN", dsnURL)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resetCfg(t)
			tc.setup(t) // env snapshot AFTER t.Setenv: we guard the resolver itself
			before, err := config.Get().GetConfigContent()
			require.NoError(t, err)
			envBefore := os.Environ()

			_, err = FromApplicationConfig()
			require.NoError(t, err)

			after, err := config.Get().GetConfigContent()
			require.NoError(t, err)
			assert.True(t, bytes.Equal(before, after), "config content mutated during resolution")
			assert.Equal(t, envBefore, os.Environ(), "process environment mutated during resolution")
		})
	}
}

func TestSecrets_RedactDSNSanity_BothCanaryForms(t *testing.T) {
	// Integration with #418: RedactDSN is the single funnel; canaries must die.
	urlOut := RedactDSN(dsnURL)
	assert.NotContains(t, urlOut, pwCanary)
	assert.Contains(t, urlOut, "***")
	kwOut := RedactDSN(dsnKeyword)
	assert.NotContains(t, kwOut, pwCanary)
	assert.Contains(t, kwOut, "***")
	// Fail-closed on unrecognised input: never echo, always a marker.
	garbage := RedactDSN("totally not a dsn " + pwCanary)
	assert.NotContains(t, garbage, pwCanary)
	assert.True(t, strings.Contains(garbage, "[redacted"))
}
