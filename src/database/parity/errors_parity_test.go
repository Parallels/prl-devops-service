package parity

// D-4 error-sentinel mapping matrix as CODE — the AC-5 deliverable
// (PBI #425 task 04). Every row from plan D-4, implemented exactly once,
// quirks mirrored NOT fixed (fixing codes is an API behaviour change,
// out of #417 scope). Wave scenarios already assert domain errors by
// construction; THIS file is the single auditable table. Task 05 appends
// the Postgres constraint legs to the same shape without edits.
//
// Comparison style mirrors legacy: sentinel IDENTITY (both engines hand
// back the sentinel values themselves) plus Code() where the sentinel
// carries one (errors.SystemError). Constraint rows assert SENTINELS,
// never driver text (D-4).

import (
	"errors"
	"testing"

	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type d4Row struct {
	name string
	want error // sentinel identity
	code int   // expected SystemError code
	call func(t *testing.T, db dbT) error
}

func TestParityErrorsMatrixD4(t *testing.T) {
	testutil.RunOnEngines(t, []testutil.Scenario{
		{Name: "errors/not_found_404_matrix", Run: func(t *testing.T, _ ctxT, db dbT) {
			for _, r := range notFoundRows() {
				t.Run(r.name, func(t *testing.T) {
					err := r.call(t, db)
					assert.Equal(t, r.want, err, "D-4 not-found identity")
					assert.Equal(t, r.code, d4Code(err), "D-4 not-found code")
				})
			}
		}},
		{Name: "errors/no_code_quirk", Run: func(t *testing.T, _ ctxT, db dbT) {
			// ErrCatalogManagerNotFound = STDLIB errors.New(...) — no
			// Code() surface AT ALL (d4Code ⇒ -1). Controllers branch on
			// Code(); this error can never map to an HTTP class.
			// Quirk mirrored, NOT fixed.
			_, err := db.GetCatalogManager("no-such-manager")
			assert.Equal(t, datainterfaces.ErrCatalogManagerNotFound, err, "identity preserved")
			assert.Equal(t, -1, d4Code(err), "stdlib error: no Code() — quirk mirrored, not fixed")
			t.Log("quirk preserved, see plan D-4 — do not fix")
		}},
		{Name: "errors/duplicate_400_matrix", Run: func(t *testing.T, _ ctxT, db dbT) {
			for _, r := range dup400Rows() {
				t.Run(r.name, func(t *testing.T) {
					err := r.call(t, db)
					assert.Equal(t, r.want, err, "D-4 duplicate identity (NOT raw unique-violation text)")
					assert.Equal(t, r.code, d4Code(err))
				})
			}
		}},
		{Name: "errors/quirk_duplicates", Run: func(t *testing.T, _ ctxT, db dbT) {
			for _, r := range quirkRows() {
				t.Run(r.name, func(t *testing.T) {
					if r.call == nil {
						// dead sentinel (no create path reaches it):
						// mirror its CODE only.
						assert.Equal(t, r.code, d4Code(r.want), "quirk code preserved")
						t.Log("quirk preserved, see plan D-4 — do not fix")
						return
					}
					err := r.call(t, db)
					assert.Equal(t, r.want, err, "quirk identity preserved")
					assert.Equal(t, r.code, d4Code(err), "quirk code mirrored")
					t.Log("quirk preserved, see plan D-4 — do not fix")
				})
			}
		}},
	}, testutil.LegacyEngine(), testutil.GormEngine())
	t.Run("errors/unreachable_offline", testD4Unreachable)
}

// d4Code extracts SystemError.Code() without assuming the concrete type
// (the no-code quirk returns a *SystemError built by New ⇒ code 0).
func d4Code(err error) int {
	var coded interface{ Code() int }
	if errors.As(err, &coded) {
		return coded.Code()
	}
	return -1
}

func notFoundRows() []d4Row {
	call := func(fn func(t *testing.T, db dbT) error) func(t *testing.T, db dbT) error {
		return func(t *testing.T, db dbT) error { return fn(t, db) }
	}
	return []d4Row{
		{"users", datainterfaces.ErrUserNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetUser(tplCtxRoot(), "no-such-user")
			return err
		})},
		{"roles", datainterfaces.ErrRoleNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetRole(tplCtxRoot(), "no-such-role")
			return err
		})},
		{"claims", datainterfaces.ErrClaimNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetClaim(tplCtxRoot(), "no-such-claim")
			return err
		})},
		{"api_keys", datainterfaces.ErrApiKeyNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetApiKey(tplCtxRoot(), "no-such-key")
			return err
		})},
		{"catalog", datainterfaces.ErrCatalogManifestNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetCatalogManifestByName(tplCtxRoot(), "no-such-manifest")
			return err
		})},
		{"orch_hosts", datainterfaces.ErrOrchestratorHostNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetOrchestratorHost(tplCtxRoot(), "no-such-host")
			return err
		})},
		{"orch_rp_hosts", datainterfaces.ErrOrchestratorReverseProxyHostNotFound, 404, call(func(t *testing.T, db dbT) error {
			host, err := db.CreateOrchestratorHost(tplCtxRoot(), models.OrchestratorHost{Host: "d4-rp.example.com"})
			require.NoError(t, err)
			_, err = db.GetOrchestratorReverseProxyHost(tplCtxRoot(), host.ID, "nope:443")
			return err
		})},
		{"orch_vms", datainterfaces.ErrOrchestratorHostVirtualMachineNotFound, 404, call(func(t *testing.T, db dbT) error {
			host, err := db.CreateOrchestratorHost(tplCtxRoot(), models.OrchestratorHost{Host: "d4-vm.example.com"})
			require.NoError(t, err)
			_, err = db.GetOrchestratorHostVirtualMachine(tplCtxRoot(), host.ID, "no-such-vm")
			return err
		})},
		{"vm_snapshots", datainterfaces.ErrOrchestratorHostsVMSnapshotsNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetHostVMSnapshots(tplCtxRoot(), "no-such-host")
			return err
		})},
		{"packer", datainterfaces.ErrPackerTemplateNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetPackerTemplate(tplCtxRoot(), "no-such-template")
			return err
		})},
		{"jobs", datainterfaces.ErrJobNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetJob(tplCtxRoot(), "no-such-job")
			return err
		})},
		{"enrollment", datainterfaces.ErrEnrollmentTokenNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetEnrollmentToken(tplCtxRoot(), "no-such-token")
			return err
		})},
		{"user_configs", datainterfaces.ErrUserConfigNotFound, 404, call(func(t *testing.T, db dbT) error {
			_, err := db.GetUserConfig(tplCtxRoot(), "no-such-user", "no-such-slug")
			return err
		})},
		{"reverse_proxy_hosts", datainterfaces.ErrorReverseProxyHostNotFound, 404, call(func(t *testing.T, db dbT) error {
			host, err := db.CreateReverseProxyHost(tplCtxRoot(), models.ReverseProxyHost{Host: "d4.example.com", Port: "80"})
			require.NoError(t, err)
			route, err := db.CreateReverseProxyHostHttpRoute(tplCtxRoot(), host.ID,
				models.ReverseProxyHostHttpRoute{Path: "/x", TargetHost: "b.local", TargetPort: "80"})
			require.NoError(t, err)
			_, err = db.ReorderReverseProxyHostHttpRoute(tplCtxRoot(), "no-such-host", route.ID, 1)
			return err
		})},
		{"reverse_proxy_http_routes", datainterfaces.ErrorReverseProxyHttpRouteNotFound, 404, call(func(t *testing.T, db dbT) error {
			host, err := db.CreateReverseProxyHost(tplCtxRoot(), models.ReverseProxyHost{Host: "d4b.example.com", Port: "80"})
			require.NoError(t, err)
			_, err = db.CreateReverseProxyHostHttpRoute(tplCtxRoot(), host.ID,
				models.ReverseProxyHostHttpRoute{Path: "/x", TargetHost: "b.local", TargetPort: "80"})
			require.NoError(t, err)
			_, err = db.ReorderReverseProxyHostHttpRoute(tplCtxRoot(), host.ID, "no-such-route", 1)
			return err
		})},
		{"reverse_proxy_tcp_routes", datainterfaces.ErrorReverseProxyTcpRouteNotFound, 404, call(func(t *testing.T, db dbT) error {
			host, err := db.CreateReverseProxyHost(tplCtxRoot(), models.ReverseProxyHost{Host: "d4c.example.com", Port: "80"})
			require.NoError(t, err)
			_, err = db.UpdateReverseProxyHostTcpRoute(tplCtxRoot(), host.ID,
				models.ReverseProxyHostTcpRoute{ID: "no-such-tcp", TargetHost: "b.local", TargetPort: "22"})
			return err
		})},
	}
}

func dup400Rows() []d4Row {
	return []d4Row{
		{"user_exists", datainterfaces.ErrUserAlreadyExists, 400, func(t *testing.T, db dbT) error {
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db) // CreateUser assigns default roles (must exist)
			_, err := db.CreateUser(ctx, models.User{Username: "d4dup", Email: "d4dup@example.com", Name: "D", Password: "pw12345678"})
			require.NoError(t, err)
			_, err = db.CreateUser(ctx, models.User{Username: "d4dup", Email: "d4dup2@example.com", Name: "D2", Password: "pw12345678"})
			return err
		}},
		{"packer_exists", datainterfaces.ErrPackerTemplateAlreadyExists, 400, func(t *testing.T, db dbT) error {
			ctx := tplCtxRoot()
			_, err := db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "d4-tpl", Name: "T1"})
			require.NoError(t, err)
			_, err = db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "d4-tpl", Name: "T1"})
			return err
		}},
		{"role_claim_dup", datainterfaces.ErrRoleAlreadyContainsClaim, 400, func(t *testing.T, db dbT) error {
			ctx := tplCtxRoot()
			_, err := db.CreateRole(ctx, models.Role{Name: "d4-role"})
			require.NoError(t, err)
			_, err = db.CreateClaim(ctx, models.Claim{Name: "d4-claim"})
			require.NoError(t, err)
			require.NoError(t, db.AddClaimToRole(ctx, "d4-role", "d4-claim"))
			return db.AddClaimToRole(ctx, "d4-role", "d4-claim")
		}},
		{"user_role_dup", datainterfaces.ErrUserAlreadyContainsRole, 400, func(t *testing.T, db dbT) error {
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db)
			u, err := db.CreateUser(ctx, models.User{Username: "d4ur", Email: "d4ur@example.com", Name: "U", Password: "pw12345678"})
			require.NoError(t, err)
			r, err := db.CreateRole(ctx, models.Role{Name: "d4-role2"})
			require.NoError(t, err)
			require.NoError(t, db.AddRoleToUser(ctx, u.ID, r.ID))
			return db.AddRoleToUser(ctx, u.ID, r.ID)
		}},
		{"user_claim_dup", datainterfaces.ErrUserAlreadyContainsClaim, 400, func(t *testing.T, db dbT) error {
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db)
			u, err := db.CreateUser(ctx, models.User{Username: "d4uc", Email: "d4uc@example.com", Name: "U", Password: "pw12345678"})
			require.NoError(t, err)
			c, err := db.CreateClaim(ctx, models.Claim{Name: "d4-claim2"})
			require.NoError(t, err)
			require.NoError(t, db.AddClaimToUser(ctx, u.ID, c.ID))
			return db.AddClaimToUser(ctx, u.ID, c.ID)
		}},
	}
}

func quirkRows() []d4Row {
	return []d4Row{
		{"apikey_exists_500", datainterfaces.ErrApiKeyAlreadyExists, 500, func(t *testing.T, db dbT) error {
			ctx := tplCtxRoot()
			_, err := db.CreateApiKey(ctx, models.ApiKey{ID: "d4-k1", Name: "d4 key", Key: "D4K", Secret: "s"})
			require.NoError(t, err)
			_, err = db.CreateApiKey(ctx, models.ApiKey{ID: "d4-k2", Name: "d4 key", Key: "D4K2", Secret: "s"})
			return err
		}},
		{"userconfig_exists_500", datainterfaces.ErrUserConfigAlreadyExists, 500, func(t *testing.T, db dbT) error {
			ctx := tplCtxRoot()
			_, err := db.CreateUserConfig(ctx, models.UserConfig{UserID: "d4u", Slug: "s1", Type: "string", Value: "1"})
			require.NoError(t, err)
			_, err = db.CreateUserConfig(ctx, models.UserConfig{UserID: "d4u", Slug: "s1", Type: "string", Value: "2"})
			return err
		}},
		// ErrCatalogAlreadyExists is DEFINED-WITH-404 but no create path
		// reaches it (dead sentinel — inventory grep 2026-11-03). Its
		// absurd code is still what controllers would surface if wired —
		// assert the metadata only.
		{"catalog_exists_404_dead", datainterfaces.ErrCatalogAlreadyExists, 404, nil},
	}
}

// testD4Unreachable — offline, loopback only: the factory/ping error
// wrap (D-5 of #418) must surface ErrDatabaseUnreachable AND the
// redacted DSN (password never leaks). Postgres legs of the other
// classes land here in task 05 without needing edits.
func testD4Unreachable(t *testing.T) {
	ctx := tplCtxRoot()
	_, err := database.NewDatabaseConnection(ctx, database.DatabaseConfig{
		Type: database.DBTypePostgres,
		DSN:  "postgres://d4user:sup3rs3cr3t@127.0.0.1:1/d4nope",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, database.ErrDatabaseUnreachable), "unreachable wrap identity")
	assert.NotContains(t, err.Error(), "sup3rs3cr3t", "password redacted (D-8)")
	assert.Contains(t, err.Error(), "127.0.0.1:1", "host:port kept for ops triage")
}
