package data

import (
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	// Seed roles and claims first as CreateUser might require them if using defaults
	for _, roleName := range constants.DefaultRoles {
		_, _ = db.CreateRole(ctx, models.Role{Name: roleName, ID: roleName})
	}
	for _, claimName := range constants.DefaultClaims {
		_, _ = db.CreateClaim(ctx, models.Claim{Name: claimName, ID: claimName})
	}

	user := models.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Name:     "Test User",
		Password: "password123",
	}

	created, err := db.CreateUser(ctx, user)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.NotEmpty(t, created.CreatedAt)
	assert.Equal(t, "testuser", created.Username)

	// Verify encryption on password
	assert.NotEqual(t, "password123", created.Password)
	assert.NotEmpty(t, created.Password)

	loaded, err := db.GetUser(ctx, "testuser@example.com")
	require.NoError(t, err)
	assert.Equal(t, "testuser", loaded.Username)
}

func TestCreateUserFailsMissingFields(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	userEmptyEmail := models.User{
		Username: "testuser",
		Name:     "Test User",
		Password: "password123",
	}

	_, err := db.CreateUser(ctx, userEmptyEmail)
	require.Error(t, err)
	assert.Equal(t, ErrUserEmailCannotBeEmpty, err)

	userEmptyName := models.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "password123",
	}

	_, err = db.CreateUser(ctx, userEmptyName)
	require.Error(t, err)
	assert.Equal(t, ErrUserNameCannotBeEmpty, err)
}

func TestUpdateUser(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	for _, roleName := range constants.DefaultRoles {
		_, _ = db.CreateRole(ctx, models.Role{Name: roleName, ID: roleName})
	}
	for _, claimName := range constants.DefaultClaims {
		_, _ = db.CreateClaim(ctx, models.Claim{Name: claimName, ID: claimName})
	}

	user := models.User{
		Username: "updateuser",
		Email:    "update@example.com",
		Name:     "Update Test",
		Password: "password",
	}

	created, err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	updateReq := models.User{
		ID:    created.ID,
		Name:  "Updated Name",
		Email: "new@example.com",
	}

	err = db.UpdateUser(ctx, updateReq)
	require.NoError(t, err)

	loaded, err := db.GetUser(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", loaded.Name)
	assert.Equal(t, "new@example.com", loaded.Email)
}

func TestUpdateUserBlockStatus(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()

	for _, roleName := range constants.DefaultRoles {
		_, _ = db.CreateRole(ctx, models.Role{Name: roleName, ID: roleName})
	}
	for _, claimName := range constants.DefaultClaims {
		_, _ = db.CreateClaim(ctx, models.Claim{Name: claimName, ID: claimName})
	}

	user := models.User{
		Username: "blockuser",
		Email:    "block@example.com",
		Name:     "Block Test",
		Password: "password",
	}

	created, err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	blockReq := models.User{
		ID:                  created.ID,
		Blocked:             true,
		BlockedReason:       "Security Violation",
		FailedLoginAttempts: 5,
	}

	err = db.UpdateUserBlockStatus(ctx, blockReq)
	require.NoError(t, err)

	loaded, err := db.GetUser(ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, loaded.Blocked)
	assert.Equal(t, "Security Violation", loaded.BlockedReason)
	assert.Equal(t, 5, loaded.FailedLoginAttempts)
}

func TestAddAndRemoveRoleToUser(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	// Seed roles and user
	_, _ = db.CreateRole(ctx, models.Role{Name: "CUSTOM_ROLE", ID: "CUSTOM_ROLE"})
	for _, roleName := range constants.DefaultRoles {
		_, _ = db.CreateRole(ctx, models.Role{Name: roleName, ID: roleName})
	}
	for _, claimName := range constants.DefaultClaims {
		_, _ = db.CreateClaim(ctx, models.Claim{Name: claimName, ID: claimName})
	}

	user := models.User{
		Username: "roleuser",
		Email:    "role@example.com",
		Name:     "Role Test",
		Password: "password",
	}
	created, err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	// Add role
	err = db.AddRoleToUser(ctx, created.ID, "CUSTOM_ROLE")
	require.NoError(t, err)

	loaded, _ := db.GetUser(ctx, created.ID)
	found := false
	for _, r := range loaded.Roles {
		if r.ID == "CUSTOM_ROLE" {
			found = true
			break
		}
	}
	assert.True(t, found)

	// Remove role
	err = db.RemoveRoleFromUser(ctx, created.ID, "CUSTOM_ROLE")
	require.NoError(t, err)

	loaded, _ = db.GetUser(ctx, created.ID)
	found = false
	for _, r := range loaded.Roles {
		if r.ID == "CUSTOM_ROLE" {
			found = true
			break
		}
	}
	assert.False(t, found)
}

func TestDeleteUser(t *testing.T) {
	db, tmpDir := setupTestDB(t)
	defer cleanupTestDB(t, tmpDir, db)

	ctx := basecontext.NewBaseContext()
	ctx.DisableLog()

	for _, roleName := range constants.DefaultRoles {
		_, _ = db.CreateRole(ctx, models.Role{Name: roleName, ID: roleName})
	}
	for _, claimName := range constants.DefaultClaims {
		_, _ = db.CreateClaim(ctx, models.Claim{Name: claimName, ID: claimName})
	}

	user := models.User{
		Username: "deleteuser",
		Email:    "delete@example.com",
		Name:     "Delete Test",
		Password: "password",
	}
	created, err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	err = db.DeleteUser(ctx, created.ID)
	require.NoError(t, err)

	_, err = db.GetUser(ctx, created.ID)
	require.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}
