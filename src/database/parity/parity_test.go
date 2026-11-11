package parity

import (
	"errors"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParityUsersSmoke — #425 task-01 bootstrap: the scenario-per-engine
// testutil pattern proving both engines execute the same
// seed→get→filter→update→delete→not-found story identically.
func TestParityUsersSmoke(t *testing.T) {
	testutil.RunOnEngines(t, []testutil.Scenario{{
		Name: "users_crud_filter",
		Run: func(t *testing.T, ctx basecontext.ApiContext, db datainterfaces.DatabaseService) {
			// seed roles/claims first — idiom copied from src/data/users_test.go:14
			for _, roleName := range constants.DefaultRoles {
				_, _ = db.CreateRole(ctx, models.Role{Name: roleName, ID: roleName})
			}
			for _, claimName := range constants.DefaultClaims {
				_, _ = db.CreateClaim(ctx, models.Claim{Name: claimName, ID: claimName})
			}

			var ids []string
			seed := func(username string) *models.User {
				u, err := db.CreateUser(ctx, models.User{
					Username: username,
					Email:    username + "@parity.local",
					Name:     username,
					Password: "parity-password-1",
				})
				require.NoErrorf(t, err, "seed %s", username)
				require.NotEmpty(t, u.ID)
				ids = append(ids, u.ID)
				return u
			}
			seed("parity-alice")
			bob := seed("parity-bob")
			carol := seed("parity-carol")
			// self-cleaning contract (E-1): everything seeded dies below.
			t.Cleanup(func() {
				for _, id := range ids {
					_ = db.DeleteUser(ctx, id)
				}
			})

			all, err := db.GetUsers(ctx, "")
			require.NoError(t, err)
			assert.Len(t, all, 3)

			filtered, err := db.GetUsers(ctx, "username=parity-bob")
			require.NoError(t, err)
			require.Len(t, filtered, 1)
			assert.Equal(t, bob.ID, filtered[0].ID)

			bob.Blocked = true
			require.NoError(t, db.UpdateUserBlockStatus(ctx, *bob))
			reread, err := db.GetUser(ctx, bob.ID)
			require.NoError(t, err)
			assert.True(t, reread.Blocked, "block status must persist identically")

			require.NoError(t, db.DeleteUser(ctx, carol.ID))
			all, err = db.GetUsers(ctx, "")
			require.NoError(t, err)
			assert.Len(t, all, 2)

			_, err = db.GetUser(ctx, "parity-no-such-user")
			assert.Truef(t, errors.Is(err, data.ErrUserNotFound), "expected ErrUserNotFound, got %v", err)
		},
	}}, testutil.LegacyEngine(), testutil.GormEngine())
}
