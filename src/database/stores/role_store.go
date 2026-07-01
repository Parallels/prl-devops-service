package stores

import (
	"context"
	goerrors "errors"
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/database/interfaces"

	apperrors "github.com/Parallels/prl-devops-service/errors"
	logging "github.com/cjlapao/common-go-logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	roleDataStoreInstance *RoleDataStore
	roleDataStoreOnce     sync.Once
)

type RoleDataStoreInterface interface {
	interfaces.Store
	GetRoles(ctx basecontext.BaseContext) ([]models.Role, *apperrors.Diagnostics)
	GetRolesByQuery(ctx basecontext.BaseContext, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.Role], *apperrors.Diagnostics)
	GetRoleBySlugOrID(ctx basecontext.BaseContext, slugOrID string) (*models.Role, *apperrors.Diagnostics)
	GetRoleUsers(ctx basecontext.BaseContext, roleID string) ([]models.User, *apperrors.Diagnostics)
	GetRoleUsersByQuery(ctx basecontext.BaseContext, roleID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.User], *apperrors.Diagnostics)
	CreateRole(ctx basecontext.BaseContext, role *models.Role) (*models.Role, *apperrors.Diagnostics)
	UpdateRole(ctx basecontext.BaseContext, role *models.Role) *apperrors.Diagnostics
	DeleteRole(ctx basecontext.BaseContext, id string) *apperrors.Diagnostics
	GetRoleClaims(ctx basecontext.BaseContext, roleID string) ([]models.Claim, *apperrors.Diagnostics)
	GetRoleClaimsByQuery(ctx basecontext.BaseContext, roleID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.Claim], *apperrors.Diagnostics)
	GetUserRoles(ctx basecontext.BaseContext, userID string) ([]models.Role, *apperrors.Diagnostics)
	AddUserToRole(ctx basecontext.BaseContext, userID string, roleIdOrSlug string) *apperrors.Diagnostics
	RemoveUserFromRole(ctx basecontext.BaseContext, userID string, roleIdOrSlug string) *apperrors.Diagnostics
	AddClaimToRole(ctx basecontext.BaseContext, roleID string, claimID string) *apperrors.Diagnostics
	RemoveClaimFromRole(ctx basecontext.BaseContext, roleID string, claimID string) *apperrors.Diagnostics
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
	cfg := config.Get()
	logger := logging.Get()
	logger.Info("Initializing role store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.IsDatabaseAutoMigrateEnabled() {
		logger.Info("Running role migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate role store: %v", err)
		}
		logger.Info("Role migrations completed")
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
		diag := apperrors.NewDiagnostics("initialize_role_data_store")
		diag.AddError("failed_to_initialize_role_store", err.Error(), "role_data_store", nil)
		return nil, diag
	}
	return roleDataStoreInstance, nil
}

func (s *RoleDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&models.RoleClaims{}); err != nil {
		return fmt.Errorf("failed to migrate role claim table: %v", err)
	}

	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_role_claims_unique ON role_claims(role_id, claim_id);").Error; err != nil {
		return fmt.Errorf("failed to create unique index on role claims: %v", err)
	}

	if err := s.GetDB().AutoMigrate(&models.Role{}); err != nil {
		return fmt.Errorf("failed to migrate role table: %v", err)
	}

	return nil
}

func (s *RoleDataStore) GetRoles(ctx basecontext.BaseContext) ([]models.Role, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_roles")

	var roles []models.Role
	result := s.GetDB().WithContext(ctx.Context()).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Find(&roles)
	if result.Error != nil {
		diag.AddError("failed_to_get_roles", fmt.Sprintf("failed to get roles: %s", common.MapError(result.Error).Error()), "role_data_store", nil)
		return nil, diag
	}
	return roles, diag
}

func (s *RoleDataStore) GetRolesByQuery(ctx basecontext.BaseContext, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.Role], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_roles_by_query")

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	db := s.GetDB().WithContext(ctx.Context())
	db = db.Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	})

	result, err := filters.QueryDatabase[models.Role](db, "", queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_roles", fmt.Sprintf("failed to get roles: %s", common.MapError(err).Error()), "role_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *RoleDataStore) GetRoleBySlugOrID(ctx basecontext.BaseContext, slugOrID string) (*models.Role, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_role_by_slug_or_id")

	var role models.Role
	db := s.GetDB().WithContext(ctx.Context()).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Where("name = ? OR id = ?", slugOrID, slugOrID).
		First(&role)
	if db.Error != nil {
		if goerrors.Is(db.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %s", common.MapError(db.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id": slugOrID,
		})
		return nil, diag
	}

	return &role, diag
}

func (s *RoleDataStore) GetRoleUsersByQuery(ctx basecontext.BaseContext, roleID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.User], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_role_users_by_query")
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, roleID)
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

	db := s.GetDB().WithContext(ctx.Context()).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", role.ID)
	result, err := filters.QueryDatabase[models.User](db, "", queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_role_users", fmt.Sprintf("failed to get role users: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}

	return result, diag
}

func (s *RoleDataStore) GetRoleUsers(ctx basecontext.BaseContext, roleID string) ([]models.User, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_role_users")
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, roleID)
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

	var users []models.User
	result := s.GetDB().WithContext(ctx.Context()).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", role.ID).
		Find(&users)
	if result.Error != nil {
		diag.AddError("failed_to_get_role_users", fmt.Sprintf("failed to get role users: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}

	return users, diag
}

func (s *RoleDataStore) GetRoleClaims(ctx basecontext.BaseContext, roleID string) ([]models.Claim, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_role_claims")
	var role models.Role
	db := s.GetDB().WithContext(ctx.Context()).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Where("id = ?", roleID).
		First(&role)
	if db.Error != nil {
		diag.AddError("failed_to_get_role_claims", fmt.Sprintf("failed to get role claims: %s", common.MapError(db.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}
	return role.Claims, diag
}

func (s *RoleDataStore) GetRoleClaimsByQuery(ctx basecontext.BaseContext, roleID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.Claim], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_role_claims_by_query")
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, roleID)
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
	db := s.GetDB().WithContext(ctx.Context()).
		Joins("JOIN role_claims ON claims.id = role_claims.claim_id").
		Where("role_claims.role_id = ?", role.ID)
	result, err := filters.QueryDatabase[models.Claim](db, "", queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_role_claims", fmt.Sprintf("failed to get role claims: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}
	return result, diag
}

func (s *RoleDataStore) CreateRole(ctx basecontext.BaseContext, role *models.Role) (*models.Role, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_create_role")
	if role.ID == "" {
		role.ID = uuid.New().String()
	}
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	// Store the claims to associate after creating the role
	claimsToAssociate := role.Claims
	role.Claims = nil // Clear claims to avoid GORM trying to create them

	result := s.GetDB().WithContext(ctx.Context()).Create(role)
	if result.Error != nil {
		diag.AddError("failed_to_create_role", fmt.Sprintf("failed to create role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role": role,
		})
		return nil, diag
	}

	// Associate claims if any were provided
	if len(claimsToAssociate) > 0 {
		// Get the full claim entities from the database
		var dbClaims []models.Claim
		for _, claim := range claimsToAssociate {
			var dbClaim models.Claim
			if result := s.GetDB().WithContext(ctx.Context()).Where("id = ?", claim.ID).First(&dbClaim); result.Error != nil {
				diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim with id %s: %s", claim.ID, common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
					"claim_id": claim.ID,
				})
				return nil, diag
			}
			dbClaims = append(dbClaims, dbClaim)
		}

		// Associate the claims with the role using Append
		if err := s.GetDB().WithContext(ctx.Context()).Model(role).Association("Claims").Append(dbClaims); err != nil {
			diag.AddError("failed_to_associate_claims_with_role", fmt.Sprintf("failed to associate claims with role: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
				"role_id": role.ID,
			})
			return nil, diag
		}
	}

	return role, diag
}

func (s *RoleDataStore) UpdateRole(ctx basecontext.BaseContext, role *models.Role) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_update_role")
	role.UpdatedAt = time.Now()

	// check if the role exists in the database
	existingRole, getRoleDiag := s.GetRoleBySlugOrID(ctx, role.ID)
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
	updates := common.PartialUpdateMap(existingRole, role, "updated_at", "name")
	if err := s.GetDB().WithContext(ctx.Context()).Model(&models.Role{}).Where("id = ?", role.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_role", fmt.Sprintf("failed to update role: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"role_id": role.ID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) DeleteRole(ctx basecontext.BaseContext, id string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_delete_role")
	result := s.GetDB().WithContext(ctx.Context()).Delete(&models.Role{}, "id = ?", id)
	if result.Error != nil {
		diag.AddError("failed_to_delete_role", fmt.Sprintf("failed to delete role: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"role_id": id,
		})
		return diag
	}
	return diag
}

func (s *RoleDataStore) GetUserRoles(ctx basecontext.BaseContext, userID string) ([]models.Role, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_user_roles")
	var user models.User
	result := s.GetDB().WithContext(ctx.Context()).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Claims", func(db *gorm.DB) *gorm.DB {
				return db.Order("claims.created_at DESC")
			}).Order("roles.created_at DESC")
		}).
		Where("users.id = ?", userID).
		First(&user)
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

func (s *RoleDataStore) AddUserToRole(ctx basecontext.BaseContext, userID string, roleIdOrSlug string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_add_user_to_role")
	var user models.User
	result := s.GetDB().WithContext(ctx.Context()).
		Where("id = ?", userID).
		First(&user)
	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("user_not_found", "user not found", "role_data_store", map[string]interface{}{
				"user_id": userID,
			})
			return diag
		}
		diag.AddError("failed_to_get_user", fmt.Sprintf("failed to get user: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}

	// check if the role exists
	existingRole, getRoleDiag := s.GetRoleBySlugOrID(ctx, roleIdOrSlug)
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

	// Use GORM's Association API to add the role to the user
	// GORM automatically handles duplicates in many-to-many relationships
	if err := s.GetDB().WithContext(ctx.Context()).Model(&user).Association("Roles").Append(existingRole); err != nil {
		diag.AddError("failed_to_add_user_to_role", fmt.Sprintf("failed to add user to role: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": existingRole.ID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) RemoveUserFromRole(ctx basecontext.BaseContext, userID string, roleIdOrSlug string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_remove_user_from_role")
	var user models.User
	result := s.GetDB().WithContext(ctx.Context()).
		Preload("Roles").
		Where("users.id = ?", userID).
		First(&user)
	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("user_not_found", "user not found", "role_data_store", map[string]interface{}{
				"user_id": userID,
			})
			return diag
		}
		diag.AddError("failed_to_get_user", fmt.Sprintf("failed to get user: %s", common.MapError(result.Error).Error()), "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}

	// check if the role exists
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, roleIdOrSlug)
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
	if err := s.GetDB().WithContext(ctx.Context()).Model(&user).Association("Roles").Delete(role); err != nil {
		diag.AddError("failed_to_remove_user_from_role", fmt.Sprintf("failed to remove user from role: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": role.ID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) AddClaimToRole(ctx basecontext.BaseContext, roleID string, claimID string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_add_claim_to_role")
	var role models.Role
	result := s.GetDB().WithContext(ctx.Context()).
		Where("id = ?", roleID).
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
	var claim models.Claim
	result = s.GetDB().WithContext(ctx.Context()).Where("id = ?", claimID).First(&claim)
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

	// Use GORM's Association API to add the claim to the role
	// GORM automatically handles duplicates in many-to-many relationships
	if err := s.GetDB().WithContext(ctx.Context()).Model(&role).Association("Claims").Append(&claim); err != nil {
		diag.AddError("failed_to_add_claim_to_role", fmt.Sprintf("failed to add claim to role: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"role_id":  roleID,
			"claim_id": claimID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) RemoveClaimFromRole(ctx basecontext.BaseContext, roleID string, claimID string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_remove_claim_from_role")
	var role models.Role
	result := s.GetDB().WithContext(ctx.Context()).
		Where("id = ?", roleID).
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
	var claim models.Claim
	result = s.GetDB().WithContext(ctx.Context()).Where("id = ?", claimID).First(&claim)
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
	var roleClaim models.RoleClaims
	result = s.GetDB().WithContext(ctx.Context()).Where("role_id = ? AND claim_id = ?", role.ID, claim.ID).First(&roleClaim)
	if result.Error != nil {
		if goerrors.Is(result.Error, gorm.ErrRecordNotFound) {
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

	// Use GORM's Association API to remove the claim from the role
	if err := s.GetDB().WithContext(ctx.Context()).Model(&role).Association("Claims").Delete(&claim); err != nil {
		diag.AddError("failed_to_remove_claim_from_role", fmt.Sprintf("failed to remove claim from role: %s", common.MapError(err).Error()), "role_data_store", map[string]interface{}{
			"role_id":  roleID,
			"claim_id": claimID,
		})
		return diag
	}

	return diag
}
