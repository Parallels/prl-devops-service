package parity

// #425 task 02 — auth wave, roles (4, data/roles_test.go:12/34/57/80) +
// claims (4, data/claims_test.go:12/34/57/80) + roles_claims matrix (7,
// data/roles_claims_test.go:33/48/60/71/82/97/107). 1:1 conversions;
// DeleteInternal*FailsWithoutRoot pins the authorization semantics of the
// non-root (nil-auth, authCtxPlain) posture; the 7 rows pre-map task 04's
// error matrix.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParityRolesClaimsAuthWave(t *testing.T) {
	testutil.RunOnEngines(t, rolesScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
	testutil.RunOnEngines(t, claimsScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
	testutil.RunOnEngines(t, rolesClaimsScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
}

func rolesScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "roles/create_role", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.CreateRole(ctx, models.Role{Name: "API_CONSUMER"})
			require.NoError(t, err)
			assert.Equal(t, "API_CONSUMER", created.ID)
			assert.Equal(t, "API_CONSUMER", created.Name)
			loaded, err := db.GetRole(ctx, "API_CONSUMER")
			require.NoError(t, err)
			assert.Equal(t, "API_CONSUMER", loaded.Name)
		}},
		{Name: "roles/update_role", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.CreateRole(ctx, models.Role{Name: "ROLE_UPDATER"})
			require.NoError(t, err)
			updatedRole := *created
			updatedRole.Name = "Updated Name"
			updated, err := db.UpdateRole(ctx, &updatedRole)
			require.NoError(t, err)
			assert.Equal(t, "Updated Name", updated.Name)
		}},
		{Name: "roles/delete_role", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			_, err := db.CreateRole(ctx, models.Role{Name: "ROLE_DELETER"})
			require.NoError(t, err)
			require.NoError(t, db.DeleteRole(ctx, "ROLE_DELETER"))
			_, err = db.GetRole(ctx, "ROLE_DELETER")
			require.Error(t, err)
			assert.Equal(t, data.ErrRoleNotFound, err)
		}},
		{Name: "roles/delete_internal_role_fails_without_root", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain() // NewBaseContext posture
			_, err := db.CreateRole(ctx, models.Role{Name: "SYSTEM_ADMIN", Internal: true})
			require.NoError(t, err)
			err = db.DeleteRole(ctx, "SYSTEM_ADMIN")
			require.Error(t, err)
			assert.Equal(t, data.ErrRemoveInternalRole, err)
		}},
	}
}

func claimsScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "claims/create_claim", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.CreateClaim(ctx, models.Claim{Name: "READ_ONLY_ACCESS"})
			require.NoError(t, err)
			assert.Equal(t, "READ_ONLY_ACCESS", created.ID)
			assert.Equal(t, "READ_ONLY_ACCESS", created.Name)
			loaded, err := db.GetClaim(ctx, "READ_ONLY_ACCESS")
			require.NoError(t, err)
			assert.Equal(t, "READ_ONLY_ACCESS", loaded.Name)
		}},
		{Name: "claims/update_claim", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			created, err := db.CreateClaim(ctx, models.Claim{Name: "CLAIM_UPDATER"})
			require.NoError(t, err)
			updatedClaim := *created
			updatedClaim.Name = "Updated Name"
			updated, err := db.UpdateClaim(ctx, &updatedClaim)
			require.NoError(t, err)
			assert.Equal(t, "Updated Name", updated.Name)
		}},
		{Name: "claims/delete_claim", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			_, err := db.CreateClaim(ctx, models.Claim{Name: "CLAIM_DELETER"})
			require.NoError(t, err)
			require.NoError(t, db.DeleteClaim(ctx, "CLAIM_DELETER"))
			_, err = db.GetClaim(ctx, "CLAIM_DELETER")
			require.Error(t, err)
			assert.Equal(t, data.ErrClaimNotFound, err)
		}},
		{Name: "claims/delete_internal_claim_fails_without_root", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			_, err := db.CreateClaim(ctx, models.Claim{Name: "SYSTEM_CLAIM", Internal: true})
			require.NoError(t, err)
			err = db.DeleteClaim(ctx, "SYSTEM_CLAIM")
			require.Error(t, err)
			assert.Equal(t, data.ErrRemoveInternalClaim, err)
		}},
	}
}

// seedRoleAndClaim mirrors data/roles_claims_test.go:22 (same helpers,
// DatabaseService surface).
func seedRoleAndClaim(t *testing.T, ctx ctxT, db dbT, roleName, claimName string) (*models.Role, *models.Claim) {
	t.Helper()
	claim, err := db.CreateClaim(ctx, models.Claim{Name: claimName})
	require.NoError(t, err)
	role, err := db.CreateRole(ctx, models.Role{Name: roleName})
	require.NoError(t, err)
	return role, claim
}

func rolesClaimsScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "roles_claims/add_claim_to_role_success", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			role, claim := seedRoleAndClaim(t, ctx, db, "TESTROLE", "TESTCLAIM")
			require.NoError(t, db.AddClaimToRole(ctx, role.ID, claim.ID))
			fetched, err := db.GetRole(ctx, role.ID)
			require.NoError(t, err)
			require.Len(t, fetched.Claims, 1)
			assert.Equal(t, claim.ID, fetched.Claims[0].ID)
		}},
		{Name: "roles_claims/add_claim_to_role_duplicate", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			role, claim := seedRoleAndClaim(t, ctx, db, "TESTROLE", "TESTCLAIM")
			require.NoError(t, db.AddClaimToRole(ctx, role.ID, claim.ID))
			err := db.AddClaimToRole(ctx, role.ID, claim.ID)
			assert.ErrorIs(t, err, data.ErrRoleAlreadyContainsClaim)
		}},
		{Name: "roles_claims/add_claim_to_role_role_not_found", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			_, err := db.CreateClaim(ctx, models.Claim{Name: "TESTCLAIM"})
			require.NoError(t, err)
			err = db.AddClaimToRole(ctx, "NONEXISTENT_ROLE", "TESTCLAIM")
			assert.ErrorIs(t, err, data.ErrRoleNotFound)
		}},
		{Name: "roles_claims/add_claim_to_role_claim_not_found", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			role, err := db.CreateRole(ctx, models.Role{Name: "TESTROLE"})
			require.NoError(t, err)
			err = db.AddClaimToRole(ctx, role.ID, "NONEXISTENT_CLAIM")
			assert.ErrorIs(t, err, data.ErrClaimNotFound)
		}},
		{Name: "roles_claims/remove_claim_from_role_success", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			role, claim := seedRoleAndClaim(t, ctx, db, "TESTROLE", "TESTCLAIM")
			require.NoError(t, db.AddClaimToRole(ctx, role.ID, claim.ID))
			require.NoError(t, db.RemoveClaimFromRole(ctx, role.ID, claim.ID))
			fetched, err := db.GetRole(ctx, role.ID)
			require.NoError(t, err)
			assert.Empty(t, fetched.Claims)
		}},
		{Name: "roles_claims/remove_claim_from_role_claim_not_on_role", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			role, claim := seedRoleAndClaim(t, ctx, db, "TESTROLE", "TESTCLAIM")
			err := db.RemoveClaimFromRole(ctx, role.ID, claim.ID)
			assert.ErrorIs(t, err, data.ErrRoleDoesNotContainClaim)
		}},
		{Name: "roles_claims/remove_claim_from_role_role_not_found", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			err := db.RemoveClaimFromRole(ctx, "NONEXISTENT_ROLE", "SOMECLAIM")
			assert.ErrorIs(t, err, data.ErrRoleNotFound)
		}},
	}
}
