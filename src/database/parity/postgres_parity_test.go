//go:build integration

// Postgres legs of the #425 parity surface (plan D-6 leg (b), AC-3
// "behaviour may differ — prove it does not"). Canonical run — snippet
// shared verbatim with #418's postgres_integration_test.go, keep the two
// copies in sync:
//
//	docker run --rm -d --name prl-pg418 -e POSTGRES_PASSWORD=*** \
//	  -e POSTGRES_DB=prl_test -p 5432:5432 postgres:16
//	POSTGRES_TEST_DSN='postgres://postgres:prl@127.0.0.1:5432/prl_test?sslmode=disable' \
//	  go test -tags integration ./database/... -v
//
// Skips cleanly when POSTGRES_TEST_DSN is unset (every engine leg via
// testutil.PostgresEngine; schema-isolated per run — concurrent CI jobs
// safe). SELECTED scenarios ride here (auth-core, one catalog, the
// error-matrix constraint subset); full-matrix-on-postgres is an open
// #426 follow-up noted in the PR.
package parity

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationParityPostgres runs the SELECTED wave scenarios against
// the real server — same bodies as the SQLite parity waves by
// construction (Scenario structs transposed from the merged files), so a
// green here = the scenario's claims are backend-independent.
func TestIntegrationParityPostgres(t *testing.T) {
	if os.Getenv("POSTGRES_TEST_DSN") == "" {
		t.Skip("POSTGRES_TEST_DSN unset — integration suite skipped")
	}
	testutil.RunOnEngines(t, pgScenarios(), testutil.PostgresEngine())
}

func pgScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "pg/auth_users_roundtrip", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db)
			u, err := db.CreateUser(ctx, models.User{
				Username: "pguser", Email: "pguser@example.com", Name: "PG", Password: "pw12345678",
			})
			require.NoError(t, err)
			got, err := db.GetUser(ctx, u.ID)
			require.NoError(t, err)
			assert.Equal(t, "pguser", got.Username)
			got.Name = "PG2"
			require.NoError(t, db.UpdateUser(ctx, *got))
			again, err := db.GetUser(ctx, u.ID)
			require.NoError(t, err)
			assert.Equal(t, "PG2", again.Name)
		}},
		{Name: "pg/catalog_roundtrip", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			m, err := db.CreateCatalogManifest(ctx, models.CatalogManifest{
				Name: "pg-manifest", Version: "1.0.0", RequiredRoles: []string{"r1", "r2"},
			})
			require.NoError(t, err)
			back, err := db.GetCatalogManifestByName(ctx, m.Name)
			require.NoError(t, err)
			assert.Equal(t, []string{"r1", "r2"}, back.RequiredRoles,
				"serialised array fidelity on the real backend")
		}},
		{Name: "pg/errors_not_found_subset", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			for _, tc := range []struct {
				name string
				call func() error
				want error
			}{
				{"users", func() error { _, err := db.GetUser(ctx, "nope"); return err }, datainterfaces.ErrUserNotFound},
				{"jobs", func() error { _, err := db.GetJob(ctx, "nope"); return err }, datainterfaces.ErrJobNotFound},
				{"orch_hosts", func() error { _, err := db.GetOrchestratorHost(ctx, "nope"); return err }, datainterfaces.ErrOrchestratorHostNotFound},
				{"packer", func() error { _, err := db.GetPackerTemplate(ctx, "nope"); return err }, datainterfaces.ErrPackerTemplateNotFound},
			} {
				t.Run(tc.name, func(t *testing.T) {
					assert.Equal(t, tc.want, tc.call(), "not-found sentinel identical to SQLite wave")
				})
			}
		}},
		{Name: "pg/errors_constraint_23505", Run: func(t *testing.T, _ ctxT, db dbT) {
			// THE AC-3 row: SQLite proved "constraint failed" maps to the
			// sentinels; Postgres must prove SQLSTATE 23505 does too.
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db)
			_, err := db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "pg-tpl", Name: "T"})
			require.NoError(t, err)
			_, err = db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "pg-tpl", Name: "T"})
			assert.Equal(t, datainterfaces.ErrPackerTemplateAlreadyExists, err, "23505 ⇒ same sentinel")
			_, err = db.CreateUser(ctx, models.User{Username: "pgdup", Email: "a@b.c", Name: "A", Password: "pw12345678"})
			require.NoError(t, err)
			_, err = db.CreateUser(ctx, models.User{Username: "pgdup", Email: "d@e.f", Name: "B", Password: "pw12345678"})
			assert.Equal(t, datainterfaces.ErrUserAlreadyExists, err)
			_, err = db.CreateApiKey(ctx, models.ApiKey{ID: "pg-k1", Name: "k", Key: "SAMEKEY", Secret: "s"})
			require.NoError(t, err)
			_, err = db.CreateApiKey(ctx, models.ApiKey{ID: "pg-k2", Name: "k2", Key: "SAMEKEY", Secret: "s"})
			assert.Equal(t, datainterfaces.ErrApiKeyAlreadyExists, err,
				"unique-violation path (gorm maps 23505, not a precheck) — 500 quirk preserved cross-backend")
		}},
	}
}

// Pool CRUD smoke through the SERVICE layer (the #418 integration file
// covers SELECT-1 concurrency + Stats caps directly; this complements
// with real DatabaseService CRUD) + the Stats() invariant on a handle
// built the same way the engine factory builds it.
func TestIntegrationPostgres_PoolCRUDSmoke(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN unset — integration suite skipped")
	}
	ctx := tplCtxRoot()
	gdb, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{
		Type: database.DBTypePostgres, DSN: dsn,
	})
	require.NoError(t, err)
	defer database.Close(ctx, gdb) //nolint:errcheck
	require.NoError(t, gdb.Exec("CREATE SCHEMA IF NOT EXISTS prl_t425_smoke").Error)
	defer gdb.Exec("DROP SCHEMA IF EXISTS prl_t425_smoke CASCADE") //nolint:errcheck
	require.NoError(t, gdb.Raw("SET search_path TO prl_t425_smoke").Error)
	require.NoError(t, database.RunAutoMigrate(ctx, gdb))
	svc := database.NewGormDatabase(ctx, gdb)
	require.NoError(t, svc.Connect(ctx))

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			j, err := svc.CreateJob(ctx, models.Job{ID: fmt.Sprintf("pg-smoke-%d", i), Owner: "smoke", State: constants.JobStatePending})
			if err != nil {
				if _, gerr := svc.GetJob(ctx, j.ID); gerr != nil {
					errs <- err
				}
				return
			}
			if _, err := svc.GetJob(ctx, j.ID); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		assert.NoError(t, err)
	}
	sqlDB, err := gdb.DB()
	require.NoError(t, err)
	stats := sqlDB.Stats()
	assert.Equal(t, database.DefaultPostgresMaxOpenConns, stats.MaxOpenConnections,
		"defaults propagate through the engine path (offline twin: postgres_offline_gaps_test)")
	assert.LessOrEqual(t, stats.OpenConnections, database.DefaultPostgresMaxOpenConns)
}
