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

	"github.com/Parallels/prl-devops-service/security"
	"github.com/cjlapao/common-go-logger"

	pkg_utils "github.com/Parallels/prl-devops-service/helpers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	userDataStoreInstance *UserDataStore
	userDataStoreOnce     sync.Once
)

type UserDataStoreInterface interface {
	interfaces.Store
	GetUserByID(ctx basecontext.BaseContext, tenantID string, id string) (*entities.User, *apperrors.Diagnostics)

	GetUserByUsername(ctx basecontext.BaseContext, tenantID string, username string) (*entities.User, *apperrors.Diagnostics)
	GetUsers(ctx basecontext.BaseContext, tenantID string) ([]entities.User, *apperrors.Diagnostics)
	GetUsersByQuery(ctx basecontext.BaseContext, tenantID string, filterObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *apperrors.Diagnostics)
	CreateUser(ctx basecontext.BaseContext, tenantID string, user *entities.User) (*entities.User, *apperrors.Diagnostics)
	UpdateUser(ctx basecontext.BaseContext, tenantID string, user *entities.User) *apperrors.Diagnostics
	UpdateUserPassword(ctx basecontext.BaseContext, tenantID string, id string, password string) *apperrors.Diagnostics
	BlockUser(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics
	SetRefreshToken(ctx basecontext.BaseContext, tenantID string, id string, refreshToken string) *apperrors.Diagnostics
	DeleteUser(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics
	GetUserClaims(ctx basecontext.BaseContext, tenantID string, userID string) ([]entities.Claim, *apperrors.Diagnostics)
	GetUserClaimsByQuery(ctx basecontext.BaseContext, tenantID string, userID string, filterObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *apperrors.Diagnostics)
	AddClaimToUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIdOrSlug string) *apperrors.Diagnostics
	RemoveClaimFromUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIdOrSlug string) *apperrors.Diagnostics
	GetUserRoles(ctx basecontext.BaseContext, tenantID string, userID string) ([]entities.Role, *apperrors.Diagnostics)
	GetUserRolesByQuery(ctx basecontext.BaseContext, tenantID string, userID string, filterObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *apperrors.Diagnostics)
	AddUserToRole(ctx basecontext.BaseContext, tenantID string, userID string, roleId string) *apperrors.Diagnostics
	RemoveUserFromRole(ctx basecontext.BaseContext, tenantID string, userID string, roleId string) *apperrors.Diagnostics
}

type UserDataStore struct {
	common.BaseDataStore
}

func GetUserDataStoreInstance() UserDataStoreInterface {
	if userDataStoreInstance == nil {
		return NewUserStore()
	}
	return userDataStoreInstance
}

func NewUserStore() *UserDataStore {
	return &UserDataStore{}
}

func (s *UserDataStore) Name() string {
	return "user_store"
}

func (s *UserDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	userDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *UserDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *UserDataStore) IsEnabled() bool {
	return true
}

func (s *UserDataStore) Dependencies() []string {
	return []string{}
}

func (s *UserDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.GetInstance().Get()
	logging.Info("Initializing user store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get(config.DatabaseMigrateKey).GetBool() {
		logging.Info("Running user migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate user store: %v", err)
		}
		logging.Info("User migrations completed")
	}

	userDataStoreInstance = s
	return nil
}

// Kept for backward compatibility if needed, but Init should be preferred
// Kept for backward compatibility if needed, but Init should be preferred
func InitializeUserDataStore(db *gorm.DB) (UserDataStoreInterface, *apperrors.Diagnostics) {
	if userDataStoreInstance != nil {
		return userDataStoreInstance, nil
	}
	s := NewUserStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := errors.NewDiagnostics("initialize_user_data_store")
		diag.AddError("failed_to_initialize_user_store", err.Error(), "user_data_store", err)
		return nil, diag
	}
	return userDataStoreInstance, nil
}

func (s *UserDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.UserRoles{}); err != nil {
		return fmt.Errorf("failed to migrate user role table: %s", err.Error())
	}

	if err := s.GetDB().AutoMigrate(&entities.UserClaims{}); err != nil {
		return fmt.Errorf("failed to migrate user claim table: %s", err.Error())
	}

	if err := s.GetDB().AutoMigrate(&entities.User{}); err != nil {
		return fmt.Errorf("failed to migrate user table: %s", err.Error())
	}

	// Add unique constraints to prevent duplicates
	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_unique ON user_roles(user_id, role_id);").Error; err != nil {
		return fmt.Errorf("failed to create unique index on user roles: %s", err.Error())
	}

	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_claims_unique ON user_claims(user_id, claim_id);").Error; err != nil {
		return fmt.Errorf("failed to create unique index on user claims: %s", err.Error())
	}

	return nil
}

// CreateUser creates a new user
func (s *UserDataStore) CreateUser(ctx basecontext.BaseContext, tenantID string, user *entities.User) (*entities.User, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("create_user")
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.TenantID = tenantID
	if user.Username != "" {
		user.Slug = pkg_utils.Slugify(user.Username)
	}

	encryptionService := encryption.GetInstance()
	encryptedPassword, err := encryptionService.HashPassword(user.Password)
	if err != nil {
		diag.AddError("failed_to_encrypt_password", fmt.Sprintf("failed to encrypt password: %s", err.Error()), "user_data_store", nil)
		return nil, diag
	}
	user.Password = encryptedPassword

	// Store the roles and claims to associate after creating the user
	rolesToAssociate := user.Roles
	claimsToAssociate := user.Claims
	user.Roles = nil  // Clear roles to avoid GORM trying to create them
	user.Claims = nil // Clear claims to avoid GORM trying to create them

	err = s.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// Associate roles if any were provided
		if len(rolesToAssociate) > 0 {
			// Get the full role entities from the database
			var dbRoles []entities.Role
			for _, role := range rolesToAssociate {
				var dbRole entities.Role
				// Use the transaction DB (tx) for lookups too, though reading from main DB is fine, consistency usually prefers tx or main
				// But we are in transaction, so we should try to use tx if we want to see effects?
				// Reading existing roles relies on them being committed.
				if result := s.GetDB().WithContext(ctx).Where("id = ?", role.ID).First(&dbRole); result.Error != nil {
					return fmt.Errorf("failed to get role with id %s: %w", role.ID, result.Error)
				}
				dbRoles = append(dbRoles, dbRole)
			}

			// First, clear any existing role associations (for new user this is empty but safe)
			// Actually for Create, we don't need Clear, but we need Append
			if err := tx.Model(user).Association("Roles").Append(dbRoles); err != nil {
				return fmt.Errorf("failed to associate roles with user: %w", err)
			}
		}

		// Associate claims if any were provided
		if len(claimsToAssociate) > 0 {
			// Get the full claim entities from the database
			var dbClaims []entities.Claim
			for _, claim := range claimsToAssociate {
				var dbClaim entities.Claim
				if result := s.GetDB().WithContext(ctx).Where("id = ?", claim.ID).First(&dbClaim); result.Error != nil {
					return fmt.Errorf("failed to get claim with id %s: %w", claim.ID, result.Error)
				}
				dbClaims = append(dbClaims, dbClaim)
			}

			if err := tx.Model(user).Association("Claims").Append(dbClaims); err != nil {
				return fmt.Errorf("failed to associate claims with user: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		diag.AddError("failed_to_create_user", fmt.Sprintf("failed to create user: %s", err.Error()), "user_data_store", nil)
		return nil, diag
	}

	return user, diag
}

// GetUserByID retrieves a user by ID
func (s *UserDataStore) GetUserByID(ctx basecontext.BaseContext, tenantID string, id string) (*entities.User, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_user_by_id")
	var user entities.User
	result := s.GetDB().WithContext(ctx).Preload("Roles").Preload("Roles.Claims").Preload("Claims").First(&user, "tenant_id = ? AND (id = ? OR slug = ?)", tenantID, id, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found as per requirement
		}
		diag.AddError("failed_to_get_user_by_id", fmt.Sprintf("failed to get user by id: %s", result.Error.Error()), "user_data_store", nil)
		return nil, diag
	}
	return &user, diag
}

// GetUserByUsername retrieves a user by username
func (s *UserDataStore) GetUserByUsername(ctx basecontext.BaseContext, tenantID string, username string) (*entities.User, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_user_by_username")
	var user entities.User
	result := s.GetDB().WithContext(ctx).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Roles.Claims").
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		First(&user, "tenant_id = ? AND username = ?", tenantID, username)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_user_by_username", fmt.Sprintf("failed to get user by username: %s", result.Error.Error()), "user_data_store", nil)
		return nil, diag
	}
	return &user, diag
}

func (s *UserDataStore) GetUsers(ctx basecontext.BaseContext, tenantID string) ([]entities.User, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_users")
	var users []entities.User
	result := s.GetDB().WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&users)
	if result.Error != nil {
		diag.AddError("failed_to_get_users", fmt.Sprintf("failed to get users: %s", common.MapError(result.Error).Error()), "user_data_store", nil)
		return nil, diag
	}
	return users, diag
}

func (s *UserDataStore) GetUsersByQuery(ctx basecontext.BaseContext, tenantID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_users_by_query")
	db := s.GetDB().WithContext(ctx)
	db = db.Preload("Roles", func(db *gorm.DB) *gorm.DB {
		return db.Order("roles.created_at DESC")
	}).Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).Where("tenant_id = ?", tenantID)

	result, err := filters.QueryDatabase[entities.User](db, tenantID, queryObj)
	if err != nil {
		diag.AddError("failed_to_get_users_by_query", fmt.Sprintf("failed to get users by query: %s", common.MapError(err).Error()), "user_data_store", nil)
		return nil, diag
	}
	return result, diag
}

// UpdateUser updates an existing user
func (s *UserDataStore) UpdateUser(ctx basecontext.BaseContext, tenantID string, user *entities.User) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("update_user")
	user.UpdatedAt = time.Now()
	user.TenantID = tenantID
	currentUser, getUserDiag := s.GetUserByID(ctx, tenantID, user.ID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if currentUser == nil {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag // Or return specific NotFound error if diagnostic supports it
	}

	if user.Password != "" {
		encryptionService := encryption.GetInstance()
		encryptedPassword, err := encryptionService.HashPassword(user.Password)
		if err != nil {
			diag.AddError("failed_to_encrypt_password", fmt.Sprintf("failed to encrypt password: %s", err.Error()), "user_data_store", nil)
			return diag
		}
		user.Password = encryptedPassword
	}
	if user.Username != "" {
		user.Slug = pkg_utils.Slugify(user.Username)
	}

	updates := common.PartialUpdateMap(currentUser, user, "updated_at", "slug")
	if err := s.GetDB().WithContext(ctx).Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_user", fmt.Sprintf("failed to update user: %s", common.MapError(err).Error()), "user_data_store", nil)
		return diag
	}
	return diag
}

func (s *UserDataStore) UpdateUserPassword(ctx basecontext.BaseContext, tenantID string, id string, password string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("update_user_password")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, id)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag
	}

	encryptionService := encryption.GetInstance()
	encryptedPassword, err := encryptionService.HashPassword(password)
	if err != nil {
		diag.AddError("failed_to_encrypt_password", fmt.Sprintf("failed to encrypt password: %s", err.Error()), "user_data_store", nil)
		return diag
	}

	// Create a minimal user object with only the fields we want to update
	updatedUser := &entities.User{
		BaseModel: common.BaseModel{
			ID:        user.ID,
			UpdatedAt: time.Now(),
		},
		Password: encryptedPassword,
	}

	// Use PartialUpdateMap to only update the password and updated_at fields
	updates := common.PartialUpdateMap(user, updatedUser, "updated_at")
	if err := s.GetDB().WithContext(ctx).Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_user_password", fmt.Sprintf("failed to update user password: %s", common.MapError(err).Error()), "user_data_store", nil)
		return diag
	}
	return diag
}

func (s *UserDataStore) BlockUser(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("block_user")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, id)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag
	}

	// Create a minimal user object with only the fields we want to update
	updatedUser := &entities.User{
		BaseModel: common.BaseModel{
			ID:        user.ID,
			UpdatedAt: time.Now(),
		},
		Blocked: true,
	}

	// Use PartialUpdateMap to only update the blocked and updated_at fields
	updates := common.PartialUpdateMap(user, updatedUser, "updated_at")
	if err := s.GetDB().WithContext(ctx).Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_block_user", fmt.Sprintf("failed to block user: %s", common.MapError(err).Error()), "user_data_store", nil)
		return diag
	}
	return diag
}

func (s *UserDataStore) SetRefreshToken(ctx basecontext.BaseContext, tenantID string, id string, refreshToken string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("set_refresh_token")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, id)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag
	}

	// Create a minimal user object with only the fields we want to update
	updatedUser := &entities.User{
		BaseModel: common.BaseModel{
			ID:        user.ID,
			UpdatedAt: time.Now(),
		},
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Use PartialUpdateMap to only update the refresh token fields and updated_at
	updates := common.PartialUpdateMap(user, updatedUser, "updated_at")
	if err := s.GetDB().WithContext(ctx).Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_set_refresh_token", fmt.Sprintf("failed to set refresh token: %s", common.MapError(err).Error()), "user_data_store", nil)
		return diag
	}
	return diag
}

// DeleteUser deletes a user
func (s *UserDataStore) DeleteUser(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("delete_user")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, id)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		// If user not found, strictly speaking deletion is a success (idempotency), or return not found error.
		// Standard pattern often returns not found or success. Given request, let's return success or specific error?
		// For Delete, usually we want to know if it existed.
		// But "user == nil" implies we adhered to "return nil if not found" in GetUserByID.
		// So we can return error "Record Not Found" here if we want strictness, or nil if idempotent.
		// Let's return error to be explicit it wasn't there to delete.
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag
	}
	if err := s.GetDB().WithContext(ctx).Delete(user).Error; err != nil {
		diag.AddError("failed_to_delete_user", fmt.Sprintf("failed to delete user: %s", common.MapError(err).Error()), "user_data_store", nil)
		return diag
	}
	return diag
}

func (s *UserDataStore) GetUserClaims(ctx basecontext.BaseContext, tenantID string, userID string) ([]entities.Claim, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_user_claims")
	var user entities.User
	result := s.GetDB().WithContext(ctx).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Where("users.id = ?", userID).
		Find(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user_claims", fmt.Sprintf("failed to get user claims: %s", common.MapError(result.Error).Error()), "user_data_store", nil)
		return nil, diag
	}
	// Note: Find might not return ErrRecordNotFound for struct query?
	// If user not found, user.ID will be empty.
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return nil, diag
	}

	return user.Claims, diag
}

func (s *UserDataStore) GetUserClaimsByQuery(ctx basecontext.BaseContext, tenantID string, userID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_user_claims_by_query")
	db := s.GetDB().WithContext(ctx)
	db = db.Table("user_claims").
		Joins("JOIN claims ON claims.id = user_claims.claim_id").
		Where("user_claims.user_id = ?", userID).
		Where("claims.tenant_id = ?", tenantID)

	result, err := filters.QueryDatabase[entities.Claim](db, tenantID, queryObj)
	if err != nil {
		diag.AddError("failed_to_get_user_claims_by_query", fmt.Sprintf("failed to get user claims by query: %s", common.MapError(err).Error()), "user_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *UserDataStore) GetUserRoles(ctx basecontext.BaseContext, tenantID string, userID string) ([]entities.Role, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_user_roles")
	var user entities.User
	result := s.GetDB().WithContext(ctx).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Where("users.id = ?", userID).
		Find(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user_roles", fmt.Sprintf("failed to get user roles: %s", common.MapError(result.Error).Error()), "user_data_store", nil)
		return nil, diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return nil, diag
	}

	return user.Roles, diag
}

func (s *UserDataStore) GetUserRolesByQuery(ctx basecontext.BaseContext, tenantID string, userID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_user_roles_by_query")
	db := s.GetDB().WithContext(ctx)

	// query the user_roles table and join the roles table and filter by the user_id
	// and apply the query object to the query
	db = db.Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Where("roles.tenant_id = ?", tenantID)

	result, err := filters.QueryDatabase[entities.Role](db, tenantID, queryObj)
	if err != nil {
		diag.AddError("failed_to_get_user_roles_by_query", fmt.Sprintf("failed to get user roles by query: %s", common.MapError(err).Error()), "user_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *UserDataStore) AddUserToRole(ctx basecontext.BaseContext, tenantID string, userID string, roleId string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("add_user_to_role")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag
	}

	// checking if the dbRole exists in the database
	var dbRole entities.Role
	roleDbResult := s.GetDB().WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, roleId).First(&dbRole)
	if roleDbResult.Error != nil {
		// TODO: Refactor this to use GetRoleByID from RoleStore theoretically, but simple check here
		if errors.Is(roleDbResult.Error, gorm.ErrRecordNotFound) {
			diag.AddError("role_not_found", "role not found", "user_data_store", nil)
			return diag
		}
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %s", common.MapError(roleDbResult.Error).Error()), "user_data_store", nil)
		return diag
	}

	// role exists in the relationship
	var userRole entities.UserRoles
	userRoleDbResult := s.GetDB().WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleId).First(&userRole)
	if userRoleDbResult.Error != nil {
		// We expect record not found here
		if !errors.Is(userRoleDbResult.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user_role", fmt.Sprintf("failed to get user role: %s", common.MapError(userRoleDbResult.Error).Error()), "user_data_store", nil)
			return diag
		}
	}
	if userRole.RoleID != "" {
		diag.AddError("role_already_associated_with_user", "role already associated with user", "user_data_store", nil)
		return diag
	}

	// create the user role association
	userRole = entities.UserRoles{
		UserID: userID,
		RoleID: roleId,
	}

	createUserRoleDbResult := s.GetDB().WithContext(ctx).Create(&userRole)
	if createUserRoleDbResult.Error != nil {
		diag.AddError("failed_to_create_user_role", fmt.Sprintf("failed to create user role: %s", common.MapError(createUserRoleDbResult.Error).Error()), "user_data_store", nil)
		return diag
	}

	return diag
}

func (s *UserDataStore) RemoveUserFromRole(ctx basecontext.BaseContext, tenantID string, userID string, roleId string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("remove_user_from_role")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag
	}

	// checking if the user role exists in the database
	var userRole entities.UserRoles
	userRoleDbResult := s.GetDB().WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleId).First(&userRole)
	if userRoleDbResult.Error != nil {
		if errors.Is(userRoleDbResult.Error, gorm.ErrRecordNotFound) {
			diag.AddError("user_role_not_found", "user role not found", "user_data_store", nil)
			return diag
		}
		diag.AddError("failed_to_get_user_role", fmt.Sprintf("failed to get user role: %s", common.MapError(userRoleDbResult.Error).Error()), "user_data_store", nil)
		return diag
	}

	// delete the user role association
	userRoleDbResult = s.GetDB().WithContext(ctx).Delete(&userRole)
	if userRoleDbResult.Error != nil {
		diag.AddError("failed_to_delete_user_role", fmt.Sprintf("failed to delete user role: %s", common.MapError(userRoleDbResult.Error).Error()), "user_data_store", nil)
		return diag
	}

	return diag
}

func (s *UserDataStore) AddClaimToUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIdOrSlug string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("add_claim_to_user")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag
	}

	var claim entities.Claim
	result := s.GetDB().WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, claimIdOrSlug).First(&claim)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("claim_not_found", "claim not found", "user_data_store", nil)
			return diag
		}
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", common.MapError(result.Error).Error()), "user_data_store", nil)
		return diag
	}

	// Check if the claim is already associated with the user
	var userClaims entities.UserClaims
	result = s.GetDB().WithContext(ctx).Where("user_id = ? AND claim_id = ?", user.ID, claim.ID).First(&userClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user_claim", fmt.Sprintf("failed to get user claim: %s", common.MapError(result.Error).Error()), "user_data_store", nil)
			return diag
		}
	}
	if userClaims.ClaimID != "" {
		diag.AddError("claim_already_associated_with_user", "claim already associated with user", "user_data_store", nil)
		return diag
	}

	// Create the user claim association
	userClaim := entities.UserClaims{
		UserID:  user.ID,
		ClaimID: claim.ID,
	}

	result = s.GetDB().WithContext(ctx).Create(&userClaim)
	if result.Error != nil {
		diag.AddError("failed_to_create_user_claim", fmt.Sprintf("failed to create user claim: %s", common.MapError(result.Error).Error()), "user_data_store", nil)
		return diag
	}

	return diag
}

func (s *UserDataStore) RemoveClaimFromUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIdOrSlug string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("remove_claim_from_user")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag
	}

	var claim entities.Claim
	result := s.GetDB().WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, claimIdOrSlug).First(&claim)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("claim_not_found", "claim not found", "user_data_store", nil)
			return diag
		}
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", common.MapError(result.Error).Error()), "user_data_store", nil)
		return diag
	}

	// Check if the claim is associated with the user
	var userClaims entities.UserClaims
	result = s.GetDB().WithContext(ctx).Where("user_id = ? AND claim_id = ?", user.ID, claim.ID).First(&userClaims)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("claim_not_associated_with_user", "claim not associated with user", "user_data_store", nil)
			return diag
		}
		diag.AddError("failed_to_get_user_claim", fmt.Sprintf("failed to get user claim: %s", common.MapError(result.Error).Error()), "user_data_store", nil)
		return diag
	}

	// Delete the user claim association
	result = s.GetDB().WithContext(ctx).Delete(&userClaims)
	if result.Error != nil {
		diag.AddError("failed_to_delete_user_claim", fmt.Sprintf("failed to delete user claim: %s", common.MapError(result.Error).Error()), "user_data_store", nil)
		return diag
	}

	return diag
}
