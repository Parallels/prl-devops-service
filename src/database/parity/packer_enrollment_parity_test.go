package parity

// #425 task 03 — infra wave: packer templates + enrollment tokens.
// Packer fixtures from src/data/packer_template_test.go (1:1; the dup guard
// is by ID via GetPackerTemplate — packer_template.go:62 — surfacing the
// 400 sentinel; spec's "dup-name" wording lands on this implementation
// fact). Enrollment TTL edges need past/future expiries that no public
// create can produce (ttlMinutes<=0 ⇒ 15 min default, orchestrator_enrollment
// _token.go:24) ⇒ E-4 pre-seed: identical fixed tokens via JSON pre-Write
// and raw INSERT. Sweep outcome matrix mirrors the #421 SQL pins
// (expired/used pruned; empty & unparseable expiries stay) through the
// public API only.

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestParityPackerEnrollmentInfraWave(t *testing.T) {
	testutil.RunOnEngines(t, packerScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
	testutil.RunOnEngines(t, enrollmentScenarios(),
		authLegacyEngine(preWriteInfraTokens), authGormEngine(insertRawInfraTokens))
}

// ---------- packer ----------

func packerScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "packer/add_get", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.AddPackerTemplate(ctx, &models.PackerTemplate{
				ID: "infra-tpl-1", Name: "Ubuntu 22.04 Setup",
			})
			require.NoError(t, err)
			assert.NotEmpty(t, created.CreatedAt)
			assert.Equal(t, "infra-tpl-1", created.ID)
			loaded, err := db.GetPackerTemplate(ctx, "infra-tpl-1")
			require.NoError(t, err)
			assert.Equal(t, "Ubuntu 22.04 Setup", loaded.Name)
		}},
		{Name: "packer/update", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "infra-tpl-2", Name: "Initial Name"})
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond) // arrangement fidelity (packer_template_test.go:53)
			updatedT := *created
			updatedT.Name = "Updated Name"
			updated, err := db.UpdatePackerTemplate(ctx, &updatedT)
			require.NoError(t, err)
			assert.Equal(t, "Updated Name", updated.Name)
		}},
		{Name: "packer/delete", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			_, err := db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "infra-tpl-del", Name: "Delete Me"})
			require.NoError(t, err)
			require.NoError(t, db.DeletePackerTemplate(ctx, "infra-tpl-del"))
			_, err = db.GetPackerTemplate(ctx, "infra-tpl-del")
			require.Error(t, err)
			assert.Equal(t, data.ErrPackerTemplateNotFound, err)
		}},
		{Name: "packer/duplicate_rejected", Run: func(t *testing.T, _ ctxT, db dbT) {
			// dup guard fires on existing ID (GetPackerTemplate(id)!=nil —
			// packer_template.go:62), 400 sentinel per D-4.
			ctx := tplCtxRoot()
			_, err := db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "infra-tpl-dup", Name: "First"})
			require.NoError(t, err)
			_, err = db.AddPackerTemplate(ctx, &models.PackerTemplate{ID: "infra-tpl-dup", Name: "Second Name"})
			require.Error(t, err)
			assert.Equal(t, data.ErrPackerTemplateAlreadyExists, err)
		}},
		{Name: "packer/internal_delete_guard", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			_, err := db.AddPackerTemplate(ctx, &models.PackerTemplate{
				ID: "infra-tpl-internal", Name: "Internal Setup", Internal: true,
			})
			require.NoError(t, err)
			err = db.DeletePackerTemplate(ctx, "infra-tpl-internal")
			require.Error(t, err)
			assert.Equal(t, data.ErrRemovingInternalPackerTemplate, err)
		}},
	}
}

// ---------- enrollment ----------

type preseedToken struct {
	id      string
	expires string // "" ⇒ empty expiry; "not-a-stamp" ⇒ unparseable; past date ⇒ expired
	used    bool
}

var infraPreseedTokens = []preseedToken{
	{"infra-et-expired", "2020-01-01T00:00:00Z", false},
	{"infra-et-used-future", "2999-01-01T00:00:00Z", true}, // used ⇒ pruned regardless
	{"infra-et-used-empty", "", true},
	{"infra-et-garbage", "not-a-stamp", false}, // unparseable ⇒ never pruned
	{"infra-et-empty", "", false},              // empty ⇒ never pruned on time
}

func preWriteInfraTokens(t *testing.T, dbPath string) {
	store := data.Data{}
	for _, p := range infraPreseedTokens {
		store.EnrollmentTokens = append(store.EnrollmentTokens, models.OrchestratorEnrollmentToken{
			ID: p.id, Token: "tok-" + p.id, HostName: "pre-host", Used: p.used,
			ExpiresAt: p.expires, CreatedAt: "2020-01-01T00:00:00Z",
		})
	}
	b, err := json.Marshal(store)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(dbPath, b, 0o640))
}

func insertRawInfraTokens(t *testing.T, gdb *gorm.DB) {
	for _, p := range infraPreseedTokens {
		require.NoError(t, gdb.Table("orchestrator_enrollment_tokens").Create(
			&models.OrchestratorEnrollmentToken{
				ID: p.id, Token: "tok-" + p.id, HostName: "pre-host", Used: p.used,
				ExpiresAt: p.expires, CreatedAt: "2020-01-01T00:00:00Z",
			}).Error)
	}
}

func enrollmentScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "enrollment/create_validate_roundtrip", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			tok, err := db.CreateEnrollmentToken(ctx, "fresh-host", 15)
			require.NoError(t, err)
			validated, err := db.ValidateEnrollmentToken(ctx, tok.Token)
			require.NoError(t, err)
			assert.Equal(t, "fresh-host", validated.HostName)
			assert.False(t, validated.Used)
			_, err = db.GetEnrollmentToken(ctx, "no-such-token")
			require.Error(t, err)
			assert.Equal(t, datainterfaces.ErrEnrollmentTokenNotFound, err)
		}},
		{Name: "enrollment/used_gate_wins", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			tok, err := db.CreateEnrollmentToken(ctx, "used-host", 15)
			require.NoError(t, err)
			require.NoError(t, db.MarkEnrollmentTokenUsed(ctx, tok.ID))
			_, err = db.ValidateEnrollmentToken(ctx, tok.Token)
			require.Error(t, err)
			assert.Equal(t, datainterfaces.ErrEnrollmentTokenUsed, err)
		}},
		{Name: "enrollment/expired_edge", Run: func(t *testing.T, _ ctxT, db dbT) {
			// past expiry (±2 s margin vs years-past stamp): Get finds it,
			// Validate refuses — observable tri-state, no clock injection.
			ctx := authCtxPlain()
			found, err := db.GetEnrollmentToken(ctx, "tok-infra-et-expired")
			require.NoError(t, err)
			assert.Equal(t, "pre-host", found.HostName)
			_, err = db.ValidateEnrollmentToken(ctx, "tok-infra-et-expired")
			require.Error(t, err)
			assert.Equal(t, datainterfaces.ErrEnrollmentTokenExpired, err)
		}},
		{Name: "enrollment/empty_and_garbage_stay_valid", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			for _, token := range []string{"tok-infra-et-empty", "tok-infra-et-garbage"} {
				validated, err := db.ValidateEnrollmentToken(ctx, token)
				require.NoError(t, err, "%s: empty/unparseable expiry ⇒ never expired", token)
				assert.False(t, validated.Used)
			}
		}},
		{Name: "enrollment/sweep_prune_matrix", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			fresh, err := db.CreateEnrollmentToken(ctx, "sweep-host", 60)
			require.NoError(t, err)
			require.NoError(t, db.DeleteExpiredEnrollmentTokens(ctx))
			probes := map[string]error{
				"tok-infra-et-expired":     datainterfaces.ErrEnrollmentTokenNotFound, // pruned (past)
				"tok-infra-et-used-future": datainterfaces.ErrEnrollmentTokenNotFound, // pruned (used wins)
				"tok-infra-et-used-empty":  datainterfaces.ErrEnrollmentTokenNotFound, // pruned (used wins)
				"tok-infra-et-garbage":     nil,                                       // unparseable stays
				"tok-infra-et-empty":       nil,                                       // empty stays
			}
			for token, want := range probes {
				_, err := db.GetEnrollmentToken(ctx, token)
				if want == nil {
					require.NoError(t, err, "%s must survive sweep", token)
				} else {
					require.Error(t, err, "%s must be pruned", token)
					assert.Equal(t, want, err)
				}
			}
			_, err = db.GetEnrollmentToken(ctx, fresh.Token)
			require.NoError(t, err, "fresh runtime token survives")
		}},
	}
}
