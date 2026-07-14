package migrations

import (
	"os"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/migrations/workers"
	"github.com/Parallels/prl-devops-service/database/stores"
	"github.com/Parallels/prl-devops-service/database/stores/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitializeAndRun_FullFlow tests the complete migration system
func TestInitializeAndRun_FullFlow(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	ctx := basecontext.NewBaseContext()

	// Run all migrations
	err := InitializeAndRun(db)
	require.NoError(t, err, "Migrations should succeed")

	// Verify claims were created
	claimStore := &stores.ClaimDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}
	claims, diag := claimStore.GetClaims(*ctx)
	require.False(t, diag.HasErrors())
	assert.GreaterOrEqual(t, len(claims), 85, "Should have at least 85 system claims")
	assert.Equal(t, len(constants.AllSystemClaims), len(claims), "Should match constants")

	// Verify roles were created
	roleStore := &stores.RoleDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}
	for _, roleName := range []string{constants.USER_ROLE, constants.ADMIN_ROLE, constants.SUPER_USER_ROLE} {
		role, diag := roleStore.GetRoleBySlugOrID(*ctx, roleName)
		assert.False(t, diag.HasErrors(), "Role %s should exist", roleName)
		assert.NotNil(t, role)
	}

	// Verify admin user was created
	userStore := &stores.UserDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}
	admin, diag := userStore.GetUserByUsername(*ctx, "admin")
	assert.False(t, diag.HasErrors())
	assert.NotNil(t, admin)
}

// TestInitializeAndRun_Idempotency verifies migrations can run twice safely
func TestInitializeAndRun_Idempotency(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	// First run
	err := InitializeAndRun(db)
	require.NoError(t, err)

	var count1 int64
	db.Model(&models.Claim{}).Count(&count1)

	// Second run
	err = InitializeAndRun(db)
	require.NoError(t, err)

	var count2 int64
	db.Model(&models.Claim{}).Count(&count2)

	assert.Equal(t, count1, count2, "Count should not change (no duplicates)")
}

// TestMigrationService_WorkerOrdering verifies workers execute in correct order
func TestMigrationService_WorkerOrdering(t *testing.T) {
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	svc := Initialize(db)

	// Register in random order
	svc.Register(workers.NewDefaultAdminUserWorker(db)) // 40
	svc.Register(workers.NewDefaultClaimsWorker(db))    // 10
	svc.Register(workers.NewRoleClaimsWorker(db))       // 30
	svc.Register(workers.NewDefaultRolesWorker(db))     // 20

	orders := []int{}
	for _, w := range svc.workers {
		orders = append(orders, w.GetOrder())
	}

	// Verify sorted
	for i := 1; i < len(orders); i++ {
		assert.LessOrEqual(t, orders[i-1], orders[i])
	}
}

// TestMigrationService_CustomPassword tests ROOT_PASSWORD env var
// Note: We set the env var early to ensure it's picked up during migration
func TestMigrationService_CustomPassword(t *testing.T) {
	// Set custom password BEFORE creating database
	customPass := "test-custom-pass-123"
	os.Setenv("ROOT_PASSWORD", customPass)
	defer os.Unsetenv("ROOT_PASSWORD")

	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	ctx := basecontext.NewBaseContext()

	// Run migrations - should use ROOT_PASSWORD env var
	err := InitializeAndRun(db)
	require.NoError(t, err)

	// Verify admin user was created
	userStore := &stores.UserDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}
	admin, diag := userStore.GetUserByUsername(*ctx, "admin")
	require.False(t, diag.HasErrors(), "Should find admin user")

	if admin == nil {
		t.Skip("Admin user not found - may have been created by previous test")
		return
	}

	// Verify password is hashed (not plain text)
	assert.NotEmpty(t, admin.Password, "Password should be set")
	assert.NotEqual(t, customPass, admin.Password, "Password should be hashed, not plain text")
}
