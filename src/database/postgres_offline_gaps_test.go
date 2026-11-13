package database

// Postgres offline gap-fill (PBI #425 task 05, plan D-6 leg (a)):
// server-free assertions that ride the DEFAULT `go test ./...` step.
// #418 shipped the per-unit matrices (redact_test, postgres_test,
// from_application_config_test); this file closes the spec's gaps —
// factory-level typed outcomes over the DSN FORM matrix (URL vs
// keyword/value, sslmode variants, missing host), the same paths
// re-asserted through the #420 FromApplicationConfig seam, pool-config
// propagation asserted offline, and the #419 D-1 string-timestamp
// fidelity that stands in for Postgres type handling (TEXT columns pass
// caller stamps through on BOTH backends — no driver coercion).
//
// Deviation note (spec step 1): the shipped #418 constructor DIALS in
// gorm.Open (postgres.go header) and has no no-ping option, so a live
// Postgres handle cannot exist offline. The Stats()==cfg propagation
// is therefore asserted against applyPool — the exact function the
// Postgres route calls — on an offline SQLite handle, and the same
// Stats invariant is proven on the live server in the integration leg.

import (
	"errors"
	"testing"

	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const d5Secret = "s3cr3t-pg-offline"

func openOfflineGORM(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.TempDir()+"/gaps.db?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

// Pool propagation, offline: applyPool is THE function the Postgres route
// invokes after Open (postgres.go:28) — asserting it on an offline handle
// proves cfg→sql.DB propagation without a server (see header for why the
// Postgres-opened Stats variant lives in the integration leg instead).
func TestPostgresOffline_PoolConfigPropagation(t *testing.T) {
	db := openOfflineGORM(t)
	require.NoError(t, applyPool(db, DatabaseConfig{Postgres: PostgresPoolConfig{
		MaxOpenConns: 4, MaxIdleConns: 2,
	}}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	assert.Equal(t, 4, sqlDB.Stats().MaxOpenConnections, "custom value propagates verbatim")

	db2 := openOfflineGORM(t)
	require.NoError(t, applyPool(db2, DatabaseConfig{})) // zero-means-default
	sqlDB2, err := db2.DB()
	require.NoError(t, err)
	assert.Equal(t, DefaultPostgresMaxOpenConns, sqlDB2.Stats().MaxOpenConnections)
	assert.Equal(t, DefaultPostgresMaxOpenConns, PostgresPoolConfig{}.MaxOpen())
	assert.Equal(t, DefaultPostgresMaxIdleConns, PostgresPoolConfig{}.MaxIdle())
}

// Factory typed outcomes over DSN FORMS (mirror of #418 task-04's table,
// extended to the keyword/value + sslmode space the unit file left at the
// seam level). All legs dial loopback:1 — connect-refused is instant and
// server-free; the typed outcome is the SAME sentinel wrap for every
// form, with the secret redacted and host:port preserved for triage.
func TestPostgresOffline_DSNFormsTypedOutcomes(t *testing.T) {
	cases := []struct {
		name, dsn  string
		wantTarget bool
	}{
		{"url_sslmode_require", "postgres://d5user:" + d5Secret + "@127.0.0.1:1/d5nope?sslmode=require", true},
		{"url_sslmode_verify_full", "postgres://d5user:" + d5Secret + "@127.0.0.1:1/d5nope?sslmode=verify-full", true},
		{"keyword_value", "host=127.0.0.1 port=1 dbname=d5nope user=d5user password=" + d5Secret + " sslmode=disable", true},
		// Empty host ⇒ pgx defaults to the LOCAL UNIX SOCKET (message
		// carries the socket path, not host:port) and RedactDSN fails
		// closed on the URL ([redacted-dsn]) — PINNED both engines' behaviour.
		{"missing_host_defaults_to_socket", "postgres://d5user:" + d5Secret + "@/d5nope?sslmode=disable", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := testCtx()
			_, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypePostgres, DSN: c.dsn})
			require.Error(t, err)
			assert.True(t, errors.Is(err, ErrDatabaseUnreachable),
				"every DSN FORM lands on the same typed wrap (#418 D-5)")
			assert.NotContains(t, err.Error(), d5Secret, "secret never echoes (D-8)")
			if c.wantTarget {
				assert.Contains(t, err.Error(), "127.0.0.1:1", "ops triage context kept")
			}
		})
	}
	t.Run("garbage_echo", func(t *testing.T) {
		ctx := testCtx()
		_, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypePostgres, DSN: "d5junk-not-a-dsn"})
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrDatabaseUnreachable))
		assert.NotContains(t, err.Error(), "d5junk-not-a-dsn",
			"T16 fix (hardening-02): the pgx verbatim echo is scrubbed at the wrap site")
	})
	t.Run("t16_parse_failure_secret_leak_FIXED", func(t *testing.T) {
		// T16 (backlog): the wrap used to embed pgx's raw-input echo, so an
		// unparseable DSN carrying a password leaked it into boot logs.
		// Was pinned as leaking EXISTING through 425-05; hardening-02 fixed
		// it at the wrap site (scrubDriverError) — pin flipped to absence.
		ctx := testCtx()
		leaky := d5Secret + "-not-a-dsn-at-all"
		_, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypePostgres, DSN: leaky})
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrDatabaseUnreachable))
		assert.NotContains(t, err.Error(), d5Secret, "T16: raw echo scrubbed")
		assert.Contains(t, err.Error(), "[redacted-dsn]", "fail-closed placeholder present")
	})
	t.Run("t16b_keyword_junk_with_secret", func(t *testing.T) {
		// Adversarial sibling of the pin: junk that CARRIES an "=" (so the
		// keyword-form RedactDSN pass would keep the prose) with the secret
		// OUTSIDE any password= token — pgx echoes it raw and masks nothing.
		// Regression pin for the fail-open hole scrubDriverError's exact-
		// input replacement closes.
		ctx := testCtx()
		_, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypePostgres,
			DSN: d5Secret + " host=broken="})
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrDatabaseUnreachable))
		assert.NotContains(t, err.Error(), d5Secret,
			"T16 fail-open hole: secret riding in keyword-ish junk echo must not survive")
	})
}

// The same matrix re-asserted through the #420 seam: what the app
// config holds must reach the factory byte-identically (DSN passthrough
// was seam-tested in #420; the FULL seam→factory typed outcome is new).
func TestPostgresOffline_SeamToFactoryWiring(t *testing.T) {
	resetConfig(t)
	dsn := "host=127.0.0.1 port=1 dbname=d5nope user=d5user password=" + d5Secret
	t.Setenv("DB_TYPE", "postgres")
	t.Setenv("DB_DSN", dsn)
	cfg, err := FromApplicationConfig()
	require.NoError(t, err)
	assert.Equal(t, dsn, cfg.DSN, "seam hands the keyword-form DSN over verbatim")

	_, err = NewDatabaseConnection(testCtx(), cfg)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDatabaseUnreachable))
	assert.NotContains(t, err.Error(), d5Secret)
}

// String-date fidelity (#419 D-1): CreatedAt/UpdatedAt are STRING columns
// by design; the offline stand-in for Postgres type handling (header).
// Byte-equality across a create+reload must hold for the full-precision
// AND the trailing-zero-trimmed RFC3339Nano shapes (T14 is an ORDERING
// defect pinned for #426 — fidelity is orthogonal and asserted here).
func TestPostgresOffline_StringDateFidelity(t *testing.T) {
	db := openOfflineGORM(t)
	require.NoError(t, db.AutoMigrate(&models.Job{}))
	full := "2026-11-01T00:00:00.123456789Z"
	trimmed := "2026-11-01T00:00:00.5Z"
	require.NoError(t, db.Table("jobs").Create(&models.Job{
		ID: "d5-dates", CreatedAt: full, UpdatedAt: trimmed, State: "running",
	}).Error)
	var got models.Job
	require.NoError(t, db.Table("jobs").First(&got, "id = ?", "d5-dates").Error)
	assert.Equal(t, full, got.CreatedAt, "full-precision stamp byte-equal")
	assert.Equal(t, trimmed, got.UpdatedAt, "trimmed stamp byte-equal (driver must not normalise TEXT)")
}
