package workers

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/stores"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	logging "github.com/cjlapao/common-go-logger"
	"gorm.io/gorm"
)

// DefaultRolesWorker seeds default system roles
type DefaultRolesWorker struct {
	db *gorm.DB
}

func NewDefaultRolesWorker(db *gorm.DB) *DefaultRolesWorker {
	return &DefaultRolesWorker{db: db}
}

func (w *DefaultRolesWorker) GetName() string {
	return "default-roles"
}

func (w *DefaultRolesWorker) GetDescription() string {
	return "Seeds default system roles: USER, ADMIN, SUPER_USER"
}

func (w *DefaultRolesWorker) GetOrder() int {
	return 20 // Run after claims
}

func (w *DefaultRolesWorker) Run(ctx basecontext.BaseContext) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("default_roles_migration")
	logger := logging.Get()

	// Initialize role store with BaseDataStore pattern
	roleStore := &stores.RoleDataStore{
		BaseDataStore: *common.NewBaseDataStore(w.db),
	}

	logger.Info("Seeding default roles...")

	// Define roles with their internal flags
	// Only SUPER_USER is locked (internal=true); USER and ADMIN are editable
	roleConfigs := []struct {
		ID          string
		Name        string
		Description string
		Internal    bool
	}{
		{
			ID:          constants.USER_ROLE,
			Name:        constants.USER_ROLE,
			Description: constants.RoleDescriptionMap[constants.USER_ROLE],
			Internal:    false,
		},
		{
			ID:          constants.ADMIN_ROLE,
			Name:        constants.ADMIN_ROLE,
			Description: constants.RoleDescriptionMap[constants.ADMIN_ROLE],
			Internal:    false,
		},
		{
			ID:          constants.SUPER_USER_ROLE,
			Name:        constants.SUPER_USER_ROLE,
			Description: constants.RoleDescriptionMap[constants.SUPER_USER_ROLE],
			Internal:    true, // Locked, cannot be modified
		},
	}

	created := 0
	skipped := 0

	for _, config := range roleConfigs {
		// Check if role already exists
		existing, getDiag := roleStore.GetRoleBySlugOrID(ctx, config.ID)
		if getDiag.HasErrors() {
			diag.Append(getDiag)
			continue
		}

		if existing != nil {
			logger.Debug("Role '%s' already exists, skipping", config.Name)
			skipped++
			continue
		}

		// Create role
		role := &models.Role{
			Name:        config.Name,
			Description: config.Description,
			Internal:    config.Internal,
		}

		_, createDiag := roleStore.CreateRole(ctx, role)
		if createDiag.HasErrors() {
			logger.Error("Failed to create role '%s': %v", config.Name, createDiag.GetSummary())
			diag.Append(createDiag)
			continue
		}

		logger.Info("Created role: %s (%s)", config.Name, config.Description)
		created++
	}

	logger.Info("Roles seeding completed: %d created, %d skipped", created, skipped)

	return diag
}

func (w *DefaultRolesWorker) Rollback(ctx basecontext.BaseContext) *apperrors.Diagnostics {
	// Rollback not supported for roles (they might be assigned to users)
	return nil
}
