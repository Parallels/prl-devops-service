package parity

// #425 task 02 — auth wave, users (6 scenarios converted 1:1 from
// src/data/users_test.go:14/51/79/120/161/215 per the conversion recipe:
// arrangements, expected sentinels and list expectations preserved, never
// weakened; ctx posture mirrored — NewRootBaseContext → tplCtxRoot(),
// NewBaseContext → authCtxPlain() (nil AuthorizationContext)).

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParityUsersAuthWave(t *testing.T) {
	testutil.RunOnEngines(t, usersScenarios(), testutil.LegacyEngine(), testutil.GormEngine())
}

func usersScenarios() []testutil.Scenario {
	return []testutil.Scenario{
		{Name: "users/create_user", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db)
			created, err := db.CreateUser(ctx, models.User{
				Username: "testuser", Email: "testuser@example.com",
				Name: "Test User", Password: fixtureSecret("pw123"),
			})
			require.NoError(t, err)
			assert.NotEmpty(t, created.ID)
			assert.NotEmpty(t, created.CreatedAt)
			assert.Equal(t, "testuser", created.Username)
			assert.NotEqual(t, fixtureSecret("pw123"), created.Password) // encryption at rest
			assert.NotEmpty(t, created.Password)
			loaded, err := db.GetUser(ctx, "testuser@example.com")
			require.NoError(t, err)
			assert.Equal(t, "testuser", loaded.Username)
		}},
		{Name: "users/create_user_fails_missing_fields", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			_, err := db.CreateUser(ctx, models.User{Username: "testuser", Name: "Test User", Password: fixtureSecret("pw123")})
			require.Error(t, err)
			assert.Equal(t, data.ErrUserEmailCannotBeEmpty, err)
			_, err = db.CreateUser(ctx, models.User{Username: "testuser", Email: "testuser@example.com", Password: fixtureSecret("pw123")})
			require.Error(t, err)
			assert.Equal(t, data.ErrUserNameCannotBeEmpty, err)
		}},
		{Name: "users/update_user", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db)
			created, err := db.CreateUser(ctx, models.User{
				Username: "updateuser", Email: "update@example.com",
				Name: "Update Test", Password: "password",
			})
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond) // arrangement fidelity (users_test.go:103)
			require.NoError(t, db.UpdateUser(ctx, models.User{
				ID: created.ID, Name: "Updated Name", Email: "new@example.com",
			}))
			loaded, err := db.GetUser(ctx, created.ID)
			require.NoError(t, err)
			assert.Equal(t, "Updated Name", loaded.Name)
			assert.Equal(t, "new@example.com", loaded.Email)
		}},
		{Name: "users/update_user_block_status", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := tplCtxRoot()
			seedDefaultRolesClaims(ctx, db)
			created, err := db.CreateUser(ctx, models.User{
				Username: "blockuser", Email: "block@example.com",
				Name: "Block Test", Password: "password",
			})
			require.NoError(t, err)
			require.NoError(t, db.UpdateUserBlockStatus(ctx, models.User{
				ID: created.ID, Blocked: true,
				BlockedReason: "Security Violation", FailedLoginAttempts: 5,
			}))
			loaded, err := db.GetUser(ctx, created.ID)
			require.NoError(t, err)
			assert.True(t, loaded.Blocked)
			assert.Equal(t, "Security Violation", loaded.BlockedReason)
			assert.Equal(t, 5, loaded.FailedLoginAttempts)
		}},
		{Name: "users/add_and_remove_role_to_user", Run: func(t *testing.T, _ ctxT, db dbT) {
			// R-2 canary: the embedded role(+claim) slice must round-trip.
			ctx := authCtxPlain()
			_, _ = db.CreateRole(ctx, models.Role{Name: "CUSTOM_ROLE", ID: "CUSTOM_ROLE"})
			seedDefaultRolesClaims(ctx, db)
			created, err := db.CreateUser(ctx, models.User{
				Username: "roleuser", Email: "role@example.com",
				Name: "Role Test", Password: "password",
			})
			require.NoError(t, err)
			require.NoError(t, db.AddRoleToUser(ctx, created.ID, "CUSTOM_ROLE"))
			loaded, _ := db.GetUser(ctx, created.ID)
			assert.True(t, hasRole(loaded, "CUSTOM_ROLE"), "role survived re-read")
			require.NoError(t, db.RemoveRoleFromUser(ctx, created.ID, "CUSTOM_ROLE"))
			loaded, _ = db.GetUser(ctx, created.ID)
			assert.False(t, hasRole(loaded, "CUSTOM_ROLE"), "role removal persisted")
			// claims sub-slice round-trip on an assigned role:
			_, _ = db.CreateClaim(ctx, models.Claim{Name: "EMBEDDED_CLAIM"})
			require.NoError(t, db.AddClaimToRole(ctx, "CUSTOM_ROLE", "EMBEDDED_CLAIM"))
			require.NoError(t, db.AddRoleToUser(ctx, created.ID, "CUSTOM_ROLE"))
			loaded, _ = db.GetUser(ctx, created.ID)
			require.True(t, hasRole(loaded, "CUSTOM_ROLE"))
			for _, r := range loaded.Roles {
				if r.ID == "CUSTOM_ROLE" {
					assert.Len(t, r.Claims, 1, "claim slice survives embedded re-read")
					assert.Equal(t, "EMBEDDED_CLAIM", r.Claims[0].ID)
				}
			}
		}},
		{Name: "users/delete_user", Run: func(t *testing.T, _ ctxT, db dbT) {
			ctx := authCtxPlain()
			seedDefaultRolesClaims(ctx, db)
			created, err := db.CreateUser(ctx, models.User{
				Username: "deleteuser", Email: "delete@example.com",
				Name: "Delete Test", Password: "password",
			})
			require.NoError(t, err)
			require.NoError(t, db.DeleteUser(ctx, created.ID))
			_, err = db.GetUser(ctx, created.ID)
			require.Error(t, err)
			assert.Equal(t, data.ErrUserNotFound, err)
		}},
	}
}

func hasRole(u *models.User, id string) bool {
	if u == nil {
		return false
	}
	for _, r := range u.Roles {
		if r.ID == id {
			return true
		}
	}
	return false
}
