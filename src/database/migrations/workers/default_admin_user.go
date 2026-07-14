package workers

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/stores"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	logging "github.com/cjlapao/common-go-logger"
	"gorm.io/gorm"
)

// DefaultAdminUserWorker creates the default admin user
type DefaultAdminUserWorker struct {
	db *gorm.DB
}

func NewDefaultAdminUserWorker(db *gorm.DB) *DefaultAdminUserWorker {
	return &DefaultAdminUserWorker{db: db}
}

func (w *DefaultAdminUserWorker) GetName() string {
	return "default-admin-user"
}

func (w *DefaultAdminUserWorker) GetDescription() string {
	return "Creates default admin user with admin/admin credentials"
}

func (w *DefaultAdminUserWorker) GetOrder() int {
	return 40 // Run after roles and claims
}

func (w *DefaultAdminUserWorker) Run(ctx basecontext.BaseContext) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("default_admin_user_migration")
	logger := logging.Get()

	// Initialize stores with BaseDataStore pattern
	userStore := &stores.UserDataStore{
		BaseDataStore: *common.NewBaseDataStore(w.db),
	}

	roleStore := &stores.RoleDataStore{
		BaseDataStore: *common.NewBaseDataStore(w.db),
	}

	logger.Info("Creating default admin user...")

	// Check if admin user already exists
	existingAdmin, getUserDiag := userStore.GetUserByUsername(ctx, "admin")
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}

	if existingAdmin != nil {
		logger.Info("Admin user already exists, skipping")
		return diag
	}

	// Get password from config or use default
	cfg := config.Get()
	adminPassword := "admin" // Default password
	if envPassword := cfg.RootPassword(); envPassword != "" {
		adminPassword = envPassword
		logger.Info("Using admin password from ROOT_PASSWORD environment variable")
	}

	// Get SUPER_USER role
	superUserRole, getRoleDiag := roleStore.GetRoleBySlugOrID(ctx, constants.SUPER_USER_ROLE)
	if getRoleDiag.HasErrors() {
		logger.Error("Failed to get SUPER_USER role: %v", getRoleDiag.GetSummary())
		diag.Append(getRoleDiag)
		return diag
	}

	if superUserRole == nil {
		diag.AddError("super_user_role_not_found", "SUPER_USER role not found", "default_admin_user_worker", nil)
		return diag
	}

	// Create admin user
	adminUser := &models.User{
		Username: "admin",
		Name:     "System Administrator",
		Email:    "admin@localhost",
		Password: adminPassword, // Will be hashed by user store
		Roles:    []models.Role{*superUserRole},
		Blocked:  false,
	}

	_, createDiag := userStore.CreateUser(ctx, adminUser)
	if createDiag.HasErrors() {
		logger.Error("Failed to create admin user: %v", createDiag.GetSummary())
		diag.Append(createDiag)
		return diag
	}

	logger.Info("✓ Default admin user created successfully")
	logger.Warn("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	logger.Warn("⚠️  SECURITY WARNING: Default admin credentials")
	logger.Warn("   Username: admin")
	if cfg.RootPassword() == "" {
		logger.Warn("   Password: admin")
		logger.Warn("   CHANGE THIS PASSWORD IMMEDIATELY!")
	} else {
		logger.Warn("   Password: (set via ROOT_PASSWORD env var)")
	}
	logger.Warn("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return diag
}

func (w *DefaultAdminUserWorker) Rollback(ctx basecontext.BaseContext) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("rollback_default_admin_user")
	logger := logging.Get()

	userStore := stores.NewUserStore()
	if err := userStore.Init(ctx.Context(), w.db); err != nil {
		return diag
	}

	// Find and delete admin user
	admin, getUserDiag := userStore.GetUserByUsername(ctx, "admin")
	if getUserDiag.HasErrors() {
		return diag
	}

	if admin == nil {
		logger.Info("Admin user not found, nothing to rollback")
		return diag
	}

	deleteDiag := userStore.DeleteUser(ctx, admin.ID)
	if deleteDiag != nil && deleteDiag.HasErrors() {
		logger.Error("Failed to rollback admin user: %v", deleteDiag.GetSummary())
		diag.Append(deleteDiag)
		return diag
	}

	logger.Info("Admin user rolled back successfully")
	return diag
}
