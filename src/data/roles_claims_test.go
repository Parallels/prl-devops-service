package data

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRoleClaimTestDB(t *testing.T) (*JsonDatabase, string, basecontext.ApiContext) {
	t.Helper()
	db, tmpDir := setupTestDB(t)
	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()
	return db, tmpDir, ctx
}

// seedRoleAndClaim is a helper that creates a role and a claim, returning both.
func seedRoleAndClaim(t *testing.T, db *JsonDatabase, ctx basecontext.ApiContext, roleName, claimName string) (*models.Role, *models.Claim) {
	t.Helper()

	claim, err := db.CreateClaim(ctx, models.Claim{Name: claimName})
	require.NoError(t, err)

	role, err := db.CreateRole(ctx, models.Role{Name: roleName})
	require.NoError(t, err)

	return role, claim
}

func TestAddClaimToRole_Success(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	role, claim := seedRoleAndClaim(t, db, ctx, "TESTROLE", "TESTCLAIM")

	err := db.AddClaimToRole(ctx, role.ID, claim.ID)
	require.NoError(t, err)

	fetched, err := db.GetRole(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, fetched.Claims, 1)
	assert.Equal(t, claim.ID, fetched.Claims[0].ID)
}

func TestAddClaimToRole_Duplicate(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	role, claim := seedRoleAndClaim(t, db, ctx, "TESTROLE", "TESTCLAIM")

	require.NoError(t, db.AddClaimToRole(ctx, role.ID, claim.ID))

	err := db.AddClaimToRole(ctx, role.ID, claim.ID)
	assert.ErrorIs(t, err, ErrRoleAlreadyContainsClaim)
}

func TestAddClaimToRole_RoleNotFound(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	_, err := db.CreateClaim(ctx, models.Claim{Name: "TESTCLAIM"})
	require.NoError(t, err)

	err = db.AddClaimToRole(ctx, "NONEXISTENT_ROLE", "TESTCLAIM")
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

func TestAddClaimToRole_ClaimNotFound(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	role, err := db.CreateRole(ctx, models.Role{Name: "TESTROLE"})
	require.NoError(t, err)

	err = db.AddClaimToRole(ctx, role.ID, "NONEXISTENT_CLAIM")
	assert.ErrorIs(t, err, ErrClaimNotFound)
}

func TestRemoveClaimFromRole_Success(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	role, claim := seedRoleAndClaim(t, db, ctx, "TESTROLE", "TESTCLAIM")
	require.NoError(t, db.AddClaimToRole(ctx, role.ID, claim.ID))

	err := db.RemoveClaimFromRole(ctx, role.ID, claim.ID)
	require.NoError(t, err)

	fetched, err := db.GetRole(ctx, role.ID)
	require.NoError(t, err)
	assert.Empty(t, fetched.Claims)
}

func TestRemoveClaimFromRole_ClaimNotOnRole(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	role, claim := seedRoleAndClaim(t, db, ctx, "TESTROLE", "TESTCLAIM")

	err := db.RemoveClaimFromRole(ctx, role.ID, claim.ID)
	assert.ErrorIs(t, err, ErrRoleDoesNotContainClaim)
}

func TestRemoveClaimFromRole_RoleNotFound(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	err := db.RemoveClaimFromRole(ctx, "NONEXISTENT_ROLE", "SOMECLAIM")
	assert.ErrorIs(t, err, ErrRoleNotFound)
}

func TestDeleteClaim_CascadesToRoles(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	role, claim := seedRoleAndClaim(t, db, ctx, "TESTROLE", "TESTCLAIM")
	require.NoError(t, db.AddClaimToRole(ctx, role.ID, claim.ID))

	err := db.DeleteClaim(ctx, claim.ID)
	require.NoError(t, err)

	fetched, err := db.GetRole(ctx, role.ID)
	require.NoError(t, err)
	assert.Empty(t, fetched.Claims, "claim should be removed from role on deletion")
}

func TestDeleteClaim_CascadesToUsers(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	claim, err := db.CreateClaim(ctx, models.Claim{Name: "TESTCLAIM"})
	require.NoError(t, err)

	// Directly inject a user that already has the claim, bypassing the
	// CreateUser validation which requires all DefaultClaims to be seeded.
	db.dataMutex.Lock()
	db.data.Users = append(db.data.Users, models.User{
		ID:       "test-user-id",
		Username: "testuser",
		Name:     "Test User",
		Email:    "test@example.com",
		Claims:   []models.Claim{*claim},
	})
	db.dataMutex.Unlock()

	err = db.DeleteClaim(ctx, claim.ID)
	require.NoError(t, err)

	fetched, err := db.GetUser(ctx, "test-user-id")
	require.NoError(t, err)
	for _, c := range fetched.Claims {
		assert.NotEqual(t, claim.ID, c.ID, "deleted claim should be removed from user")
	}
}

func TestGetRole_IncludesClaims(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	role, claim := seedRoleAndClaim(t, db, ctx, "TESTROLE", "TESTCLAIM")
	require.NoError(t, db.AddClaimToRole(ctx, role.ID, claim.ID))

	fetched, err := db.GetRole(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, fetched.Claims, 1)
	assert.Equal(t, claim.ID, fetched.Claims[0].ID)
	assert.Equal(t, claim.Name, fetched.Claims[0].Name)
}

func TestGetRoles_IncludesClaims(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	claim1, err := db.CreateClaim(ctx, models.Claim{Name: "CLAIM_A"})
	require.NoError(t, err)
	claim2, err := db.CreateClaim(ctx, models.Claim{Name: "CLAIM_B"})
	require.NoError(t, err)

	roleA, err := db.CreateRole(ctx, models.Role{Name: "ROLE_A"})
	require.NoError(t, err)
	roleB, err := db.CreateRole(ctx, models.Role{Name: "ROLE_B"})
	require.NoError(t, err)

	require.NoError(t, db.AddClaimToRole(ctx, roleA.ID, claim1.ID))
	require.NoError(t, db.AddClaimToRole(ctx, roleB.ID, claim2.ID))

	roles, err := db.GetRoles(ctx, "")
	require.NoError(t, err)

	roleMap := make(map[string][]models.Claim)
	for _, r := range roles {
		roleMap[r.ID] = r.Claims
	}

	require.Len(t, roleMap["ROLE_A"], 1)
	assert.Equal(t, "CLAIM_A", roleMap["ROLE_A"][0].ID)

	require.Len(t, roleMap["ROLE_B"], 1)
	assert.Equal(t, "CLAIM_B", roleMap["ROLE_B"][0].ID)
}

func TestCreateRole_WithClaims(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	claim, err := db.CreateClaim(ctx, models.Claim{Name: "MYCLLAIM"})
	require.NoError(t, err)

	role, err := db.CreateRole(ctx, models.Role{
		Name:   "MYROLE",
		Claims: []models.Claim{{ID: claim.ID, Name: claim.Name}},
	})
	require.NoError(t, err)
	require.Len(t, role.Claims, 1)
	assert.Equal(t, claim.ID, role.Claims[0].ID)
}

func TestCreateRole_WithUnknownClaim_Fails(t *testing.T) {
	db, tmpDir, ctx := setupRoleClaimTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	_, err := db.CreateRole(ctx, models.Role{
		Name:   "MYROLE",
		Claims: []models.Claim{{ID: "NONEXISTENT", Name: "NONEXISTENT"}},
	})
	assert.Error(t, err, "creating a role with an unknown claim should fail")
}
