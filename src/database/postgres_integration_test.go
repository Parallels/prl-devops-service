//go:build integration

// Real-server suite for the postgres branch. Canonical run (snippet shared
// with #425's integration job — keep the two copies in sync there):
//
//	docker run --rm -d --name prl-pg418 -e POSTGRES_PASSWORD=*** your chosen local test password> \
//	  -e POSTGRES_DB=prl_test -p 5432:5432 postgres:16
//	POSTGRES_TEST_DSN='postgres://<user>:<password>@127.0.0.1:5432/prl_test?sslmode=disable' \
//	  go test -tags integration ./database/... -v
//
// Skips cleanly when POSTGRES_TEST_DSN is unset.
package database

import (
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pgScratch is declared HERE on purpose: no model code may leak from #419.
type pgScratch struct {
	ID   int `gorm:"primaryKey"`
	Name string
}

func integrationDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN unset — integration suite skipped")
	}
	return dsn
}

func TestIntegration_Postgres_Roundtrip(t *testing.T) {
	ctx := testCtx()
	db, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypePostgres, DSN: integrationDSN(t)})
	require.NoError(t, err)
	defer Close(ctx, db)

	require.NoError(t, db.Migrator().DropTable(&pgScratch{}))
	require.NoError(t, db.AutoMigrate(&pgScratch{}))
	require.NoError(t, db.Create(&pgScratch{ID: 1, Name: "roundtrip"}).Error)

	var got pgScratch
	require.NoError(t, db.First(&got, 1).Error)
	assert.Equal(t, "roundtrip", got.Name)
	require.NoError(t, db.Migrator().DropTable(&pgScratch{}))
}

func TestIntegration_Postgres_AuthFailureWrapsSentinel(t *testing.T) {
	ctx := testCtx()
	dsn := integrationDSN(t)

	// Deviation note: the spec sketch suggested a bad search_path variant, but
	// search_path never fails Ping (it is a plain session GUC) — wrong password
	// is the faithful unreachable-at-boot variant on a live server.
	bad := dsn
	if i := strings.Index(bad, "postgres://"); i == 0 {
		rest := bad[len("postgres://"):]
		if at := strings.LastIndex(rest, "@"); at > 0 {
			user := strings.SplitN(rest[:at], ":", 2)[0]
			bad = "postgres://" + user + ":wrong-password-" + user + "@" + rest[at+1:]
		}
	}

	_, err := NewDatabaseConnection(ctx, DatabaseConfig{Type: DBTypePostgres, DSN: bad})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrDatabaseUnreachable))
	assert.NotContains(t, err.Error(), "wrong-password-")
}

func TestIntegration_Postgres_PoolUnderConcurrency(t *testing.T) {
	ctx := testCtx()
	cfg := DatabaseConfig{Type: DBTypePostgres, DSN: integrationDSN(t)}
	cfg.Postgres.MaxOpenConns = 4
	cfg.Postgres.MaxIdleConns = 2

	db, err := NewDatabaseConnection(ctx, cfg)
	require.NoError(t, err)
	defer Close(ctx, db)

	var wg sync.WaitGroup
	errCh := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var one int
			if err := db.Raw("SELECT 1").Scan(&one).Error; err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		assert.NoError(t, err)
	}

	sqlDB, err := db.DB()
	require.NoError(t, err)
	stats := sqlDB.Stats()
	assert.Equal(t, 4, stats.MaxOpenConnections)
	assert.LessOrEqual(t, stats.OpenConnections, 4, "pool cap must hold under 10 concurrent opens")
}
