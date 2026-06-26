package stores

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/database/interfaces"

	"github.com/cjlapao/common-go-logger"

	pkg_utils "github.com/Parallels/prl-devops-service/helpers"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	roleDataStoreInstance *RoleDataStore
	roleDataStoreOnce     sync.Once
)

type RoleDataStoreInterface interface {
	interfaces.Store
	GetRoles(ctx basecontext.BaseContext, tenantID string) ([]entities.Role, *apperrors.Diagnostics)

	GetRolesByQuery(ctx basecontext.BaseContext, tenantID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *apperrors.Diagnostics)
	GetRoleBySlugOrID(ctx basecontext.BaseContext, tenantID string, slugOrID string) (*entities.Role, *apperrors.Diagnostics)
	GetRoleUsers(ctx basecontext.BaseContext, tenantID string, roleID string) ([]entities.User, *apperrors.Diagnostics)
	GetRoleUsersByQuery(ctx basecontext.BaseContext, tenantID string, roleID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *apperrors.Diagnostics)
	CreateRole(ctx basecontext.BaseContext, tenantID string, role *entities.Role) (*entities.Role, *apperrors.Diagnostics)
	UpdateRole(ctx basecontext.BaseContext, tenantID string, role *entities.Role) *apperrors.Diagnostics
	DeleteRole(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics
	GetRoleClaims(ctx basecontext.BaseContext, tenantID string, roleID string) ([]entities.Claim, *apperrors.Diagnostics)
	GetRoleClaimsByQuery(ctx basecontext.BaseContext, tenantID string, roleID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *apperrors.Diagnostics)
	GetUserRoles(ctx basecontext.BaseContext, tenantID string, userID string) ([]entities.Role, *apperrors.Diagnostics)
	AddUserToRole(ctx basecontext.BaseContext, tenantID string, userID string, roleIdOrSlug string) *apperrors.Diagnostics
	RemoveUserFromRole(ctx basecontext.BaseContext, tenantID string, userID string, roleIdOrSlug string) *apperrors.Diagnostics
	AddClaimToRole(ctx basecontext.BaseContext, tenantID string, roleID string, claimID string) *apperrors.Diagnostics
	RemoveClaimFromRole(ctx basecontext.BaseContext, tenantID string, roleID string, claimID string) *apperrors.Diagnostics
}

type RoleDataStore struct {
	common.BaseDataStore
}

func GetRoleDataStoreInstance() RoleDataStoreInterface {
	if roleDataStoreInstance == nil {
		return NewRoleStore()
	}
	return roleDataStoreInstance
}

func NewRoleStore() *RoleDataStore {
	return &RoleDataStore{}
}

func (s *RoleDataStore) Name() string {
	return "role_store"
}

func (s *RoleDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	roleDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *RoleDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *RoleDataStore) IsEnabled() bool {
	return true
}

func (s *RoleDataStore) Dependencies() []string {
	return []string{}
}

func (s *RoleDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.GetInstance().Get()
	logging.Info("Initializing role store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get(config.DatabaseMigrateKey).GetBool() {
		logging.Info("Running role migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate role store: %v", err)
		}
		logging.Info("Role migrations completed")
	}

	roleDataStoreInstance = s
	return nil
}

// Kept for backward compatibility
func InitializeRoleDataStore(db *gorm.DB) (RoleDataStoreInterface, *apperrors.Diagnostics) {
	if roleDataStoreInstance != nil {
		return roleDataStoreInstance, nil
	}
	s := NewRoleStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := errors.NewDiagnostics("initialize_role_data_store")
		diag.AddError("failed_to_initialize_role_store", err.Error(), "role_data_store", err)
		return nil, diag
	}
	return roleDataStoreInstance, nil
}

func (s *RoleDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.RoleClaims{}); err != nil {
		return fmt.Errorf("failed to migrate role claim table: %v", err)
	}

	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_role_claims_unique ON role_claims(role_id, claim_id);").Error; err != nil {
		return fmt.Errorf("failed to create unique index on role claims: %v", err)
	}

	if err := s.GetDB().AutoMigrate(&entities.Role{}); err != nil {
		return fmt.Errorf("failed to migrate role table: %v", err)
	}

	return nil
}

func (s *RoleDataStore) GetRoles(ctx basecontext.BaseContext, tenantID string) ([]entities.Role, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_roles")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "role_data_store", nil)
		return nil, diag
	}

	var roles []entities.Role
	result := s.GetDB().WithContext(ctx).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Where("tenant_id = ?", tenantID).
		Find(&roles)
	if result.Error != nil {
		diag.AddError("failed_to_get_roles", fmt.Sprintf("failed to get roles: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"tenant_id": tenantID,
		})
		return nil, diag
	}
	return roles, diag
}

func (s *RoleDataStore) GetRolesByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_roles_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "role_data_store", nil)
		return nil, diag
	}

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	db := s.GetDB().WithContext(ctx)
	db = db.Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	})

	result, err := filters.QueryDatabase[entities.Role](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_roles", fmt.Sprintf("failed to get roles: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"tenant_id": tenantID,
		})
		return nil, diag
	}
	return result, diag
}

func (s *RoleDataStore) GetRoleBySlugOrID(ctx basecontext.BaseContext, tenantID string, slugOrID string) (*entities.Role, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_role_by_slug_or_id")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "role_data_store", nil)
		return nil, diag
	}

	var role entities.Role
	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Where("tenant_id = ?", tenantID).
		First(&role, "(slug = ? OR id = ?)", slugOrID, slugOrID)
	if db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %s", common.MapError(db.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id": slugOrID,
		})
		return nil, diag
	}

	return &role, diag
}

func (s *RoleDataStore) GetRoleUsersByQuery(ctx basecontext.BaseContext, tenantID string, roleID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_role_users_by_query")
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, roleID)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return nil, diag
	}
	if role == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}

	db := s.GetDB().
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", role.ID)
	result, err := filters.QueryDatabase[entities.User](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_role_users", fmt.Sprintf("failed to get role users: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}

	return result, diag
}

func (s *RoleDataStore) GetRoleUsers(ctx basecontext.BaseContext, tenantID string, roleID string) ([]entities.User, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_role_users")
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, roleID)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return nil, diag
	}
	if role == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}

	db := s.GetDB().
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", role.ID)

	if tenantID != "" {
		db = db.Where("users.tenant_id = ?", tenantID)
	}

	var users []entities.User
	result := db.Find(&users)
	if result.Error != nil {
		diag.AddError("failed_to_get_role_users", fmt.Sprintf("failed to get role users: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}

	return users, diag
}

func (s *RoleDataStore) GetRoleClaims(ctx basecontext.BaseContext, tenantID string, roleID string) ([]entities.Claim, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_role_claims")
	var role entities.Role
	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Where("tenant_id = ?", tenantID).
		Where("id = ?", roleID).
		Find(&role)
	if db.Error != nil {
		diag.AddError("failed_to_get_role_claims", fmt.Sprintf("failed to get role claims: %s", common.MapError(db.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}
	return role.Claims, diag
}

func (s *RoleDataStore) GetRoleClaimsByQuery(ctx basecontext.BaseContext, tenantID string, roleID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_role_claims_by_query")
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, roleID)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return nil, diag
	}
	if role == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}
	db := s.GetDB().WithContext(ctx).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN role_claims ON claims.id = role_claims.claim_id").
		Where("role_claims.role_id = ?", role.ID)
	result, err := filters.QueryDatabase[entities.Claim](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_role_claims", fmt.Sprintf("failed to get role claims: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}
	return result, diag
}

func (s *RoleDataStore) CreateRole(ctx basecontext.BaseContext, tenantID string, role *entities.Role) (*entities.Role, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_create_role")
	if role.ID == "" {
		role.ID = uuid.New().String()
	}
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	role.TenantID = tenantID
	if role.Slug != "" {
		role.Slug = pkg_utils.Slugify(role.Slug)
	}

	// Store the claims to associate after creating the role
	claimsToAssociate := role.Claims
	role.Claims = nil // Clear claims to avoid GORM trying to create them

	result := s.GetDB().WithContext(ctx).Create(role)
	if result.Error != nil {
		diag.AddError("failed_to_create_role", fmt.Sprintf("failed to create role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role": role,
		})
		return nil, diag
	}

	// Associate claims if any were provided
	if len(claimsToAssociate) > 0 {
		// Get the full claim entities from the database
		var dbClaims []entities.Claim
		for _, claim := range claimsToAssociate {
			var dbClaim entities.Claim
			if result := s.GetDB().WithContext(ctx).Where("id = ?", claim.ID).First(&dbClaim); result.Error != nil {
				diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim with id %s: %s", claim.ID, common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
					"claim_id": claim.ID,
				})
				return nil, diag
			}
			dbClaims = append(dbClaims, dbClaim)
		}

		// Associate the claims with the role using Replace to avoid duplicates
		// First, clear any existing associations
		if err := s.GetDB().WithContext(ctx).Model(role).Association("Claims").Clear(); err != nil {
			diag.AddError("failed_to_clear_existing_claims_associations", fmt.Sprintf("failed to clear existing claims associations: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
				"role_id": role.ID,
			})
			return nil, diag
		}

		// Then add the new associations
		if err := s.GetDB().WithContext(ctx).Model(role).Association("Claims").Append(dbClaims); err != nil {
			diag.AddError("failed_to_associate_claims_with_role", fmt.Sprintf("failed to associate claims with role: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
				"role_id": role.ID,
			})
			return nil, diag
		}
	}

	return role, diag
}

func (s *RoleDataStore) UpdateRole(ctx basecontext.BaseContext, tenantID string, role *entities.Role) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("store_update_role")
	role.UpdatedAt = time.Now()
	if role.Slug != "" {
		role.Slug = pkg_utils.Slugify(role.Slug)
	}
	// check if the role exists in the database
	existingRole, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, role.Slug)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return diag
	}
	if existingRole == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": role.ID,
		})
		return diag
	}

	// using the partial update map to update the role
	updates := common.PartialUpdateMap(existingRole, role, "updated_at", "slug")
	if err := s.GetDB().WithContext(ctx).Model(&entities.Role{}).Where("id = ?", role.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_role", fmt.Sprintf("failed to update role: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"role_id": role.ID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) DeleteRole(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("store_delete_role")
	result := s.GetDB().WithContext(ctx).Where("tenant_id = ?", tenantID).Delete(&entities.Role{}, "id = ?", id)
	if result.Error != nil {
		diag.AddError("failed_to_delete_role", fmt.Sprintf("failed to delete role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id": id,
		})
		return diag
	}
	return diag
}

func (s *RoleDataStore) GetUserRoles(ctx basecontext.BaseContext, tenantID string, userID string) ([]entities.Role, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_user_roles")
	var user entities.User
	result := s.GetDB().WithContext(ctx).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Claims", func(db *gorm.DB) *gorm.DB {
				return db.Order("claims.created_at DESC")
			}).Order("roles.created_at DESC")
		}).
		Where("users.id = ?", userID).
		Find(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user_roles", fmt.Sprintf("failed to get user roles: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return nil, diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return nil, diag
	}

	return user.Roles, diag
}

func (s *RoleDataStore) AddUserToRole(ctx basecontext.BaseContext, tenantID string, userID string, roleIdOrSlug string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("store_add_user_to_role")
	var user entities.User
	result := s.GetDB().WithContext(ctx).
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		diag.AddError("failed_to_get_user", fmt.Sprintf("failed to get user: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}

	// check if the roles exist
	existingRole, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, roleIdOrSlug)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return diag
	}
	if existingRole == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleIdOrSlug,
		})
		return diag
	}

	// check if the user is already in the role
	var userRole entities.UserRoles
	result = s.GetDB().WithContext(ctx).Where("user_id = ? AND role_id = ?", user.ID, existingRole.ID).First(&userRole)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user_role", fmt.Sprintf("failed to get user role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
				"user_id": userID,
				"role_id": existingRole.ID,
			})
			return diag
		}
	}
	if userRole.RoleID != "" {
		diag.AddError("user_already_in_role", "user already in role", "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": existingRole.ID,
		})
		return diag
	}

	// add the role to the user
	userRole.UserID = user.ID
	userRole.RoleID = existingRole.ID
	result = s.GetDB().WithContext(ctx).Create(&userRole)
	if result.Error != nil {
		diag.AddError("failed_to_add_user_to_role", fmt.Sprintf("failed to add user to role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": existingRole.ID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) RemoveUserFromRole(ctx basecontext.BaseContext, tenantID string, userID string, roleIdOrSlug string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("store_remove_user_from_role")
	var user entities.User
	result := s.GetDB().WithContext(ctx).
		Preload("Roles").
		Where("users.id = ?", userID).
		Find(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		diag.AddError("failed_to_get_user", fmt.Sprintf("failed to get user: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}

	// check if the role exists
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, roleIdOrSlug)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return diag
	}
	if role == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleIdOrSlug,
		})
		return diag
	}

	// remove the role from the user
	var userRole entities.UserRoles
	result = s.GetDB().WithContext(ctx).Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRole)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("user_role_not_found", "user role not found", "role_data_store", map[string]interface{}{
				"user_id": userID,
				"role_id": role.ID,
			})
			return diag
		}
		diag.AddError("failed_to_get_user_role", fmt.Sprintf("failed to get user role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": role.ID,
		})
		return diag
	}

	result = s.GetDB().WithContext(ctx).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Delete(&userRole)
	if result.Error != nil {
		diag.AddError("failed_to_remove_user_from_role", fmt.Sprintf("failed to remove user from role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": role.ID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) AddClaimToRole(ctx basecontext.BaseContext, tenantID string, roleID string, claimID string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("store_add_claim_to_role")
	var role entities.Role
	result := s.GetDB().WithContext(ctx).
		Where("id = ? AND tenant_id = ?", roleID, tenantID).
		First(&role)
	if result.Error != nil {
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return diag
	}
	if role.ID == "" {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return diag
	}

	// check if the claim exists
	var claim entities.Claim
	result = s.GetDB().WithContext(ctx).Where("id = ?", claimID).First(&claim)
	if result.Error != nil {
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}
	if claim.ID == "" {
		diag.AddError("claim_not_found", "claim not found", "role_data_store", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}

	// check if the claim is already in the role
	var roleClaim entities.RoleClaims
	result = s.GetDB().WithContext(ctx).Where("role_id = ? AND claim_id = ?", role.ID, claim.ID).First(&roleClaim)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_role_claim", fmt.Sprintf("failed to get role_claim: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
				"role_id":  roleID,
				"claim_id": claimID,
			})
			return diag
		}
	}
	if roleClaim.RoleID != "" {
		diag.AddError("claim_already_in_role", "claim already in role", "role_data_store", map[string]interface{}{
			"role_id":  roleID,
			"claim_id": claimID,
		})
		return diag
	}

	// add the claim to the role
	roleClaim.RoleID = role.ID
	roleClaim.ClaimID = claim.ID
	result = s.GetDB().WithContext(ctx).Create(&roleClaim)
	if result.Error != nil {
		diag.AddError("failed_to_add_claim_to_role", fmt.Sprintf("failed to add claim to role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id":  roleID,
			"claim_id": claimID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) RemoveClaimFromRole(ctx basecontext.BaseContext, tenantID string, roleID string, claimID string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("store_remove_claim_from_role")
	var role entities.Role
	result := s.GetDB().WithContext(ctx).
		Where("id = ? AND tenant_id = ?", roleID, tenantID).
		First(&role)
	if result.Error != nil {
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return diag
	}
	if role.ID == "" {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return diag
	}

	// check if the claim exists
	var claim entities.Claim
	result = s.GetDB().WithContext(ctx).Where("id = ?", claimID).First(&claim)
	if result.Error != nil {
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}
	if claim.ID == "" {
		diag.AddError("claim_not_found", "claim not found", "role_data_store", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}

	// check if the claim is in the role
	var roleClaim entities.RoleClaims
	result = s.GetDB().WithContext(ctx).Where("role_id = ? AND claim_id = ?", role.ID, claim.ID).First(&roleClaim)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("role_claim_not_found", "role claim not found", "role_data_store", map[string]interface{}{
				"role_id":  roleID,
				"claim_id": claimID,
			})
			return diag
		}
		diag.AddError("failed_to_get_role_claim", fmt.Sprintf("failed to get role claim: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id":  roleID,
			"claim_id": claimID,
		})
		return diag
	}

	// remove the claim from the role
	result = s.GetDB().WithContext(ctx).Where("role_id = ? AND claim_id = ?", role.ID, claim.ID).Delete(&roleClaim)
	if result.Error != nil {
		diag.AddError("failed_to_remove_claim_from_role", fmt.Sprintf("failed to remove claim from role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id":  roleID,
			"claim_id": claimID,
		})
		return diag
	}

	return diag
}
