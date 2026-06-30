package stores

import (
	"github.com/Parallels/prl-devops-service/data/models"
	"context"
	goerrors "errors"
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/database/interfaces"

	logging "github.com/cjlapao/common-go-logger"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	claimDataStoreInstance *ClaimDataStore
	claimDataStoreOnce     sync.Once
)

type ClaimDataStoreInterface interface {
	interfaces.Store
	GetClaims(ctx basecontext.BaseContext, tenantID string) ([]models.Claim, *apperrors.Diagnostics)

	GetClaimsByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.Claim], *apperrors.Diagnostics)
	GetClaimBySlugOrID(ctx basecontext.BaseContext, tenantID string, slugOrID string) (*models.Claim, *apperrors.Diagnostics)
	GetClaimUsers(ctx basecontext.BaseContext, tenantID string, claimID string) ([]models.User, *apperrors.Diagnostics)
	GetClaimUsersByQuery(ctx basecontext.BaseContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.User], *apperrors.Diagnostics)
	CreateClaim(ctx basecontext.BaseContext, tenantID string, claim *models.Claim) (*models.Claim, *apperrors.Diagnostics)
	UpdateClaim(ctx basecontext.BaseContext, tenantID string, claim *models.Claim) *apperrors.Diagnostics
	DeleteClaim(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics
	GetClaimsByLevel(ctx basecontext.BaseContext, tenantID string, level models.SecurityLevel) ([]models.Claim, *apperrors.Diagnostics)
	AddClaimToUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIdOrSlug string) *apperrors.Diagnostics
	RemoveClaimFromUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIdOrSlug string) *apperrors.Diagnostics
	GetClaimApiKeys(ctx basecontext.BaseContext, tenantID string, claimID string) ([]models.ApiKey, *apperrors.Diagnostics)
	GetClaimApiKeysByQuery(ctx basecontext.BaseContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.ApiKey], *apperrors.Diagnostics)
	AddClaimToApiKey(ctx basecontext.BaseContext, tenantID string, claimID string, apiKeyID string) *apperrors.Diagnostics
	RemoveClaimFromApiKey(ctx basecontext.BaseContext, tenantID string, claimID string, apiKeyID string) *apperrors.Diagnostics
	GetClaimRoles(ctx basecontext.BaseContext, tenantID string, claimID string) ([]models.Role, *apperrors.Diagnostics)
	GetClaimRolesByQuery(ctx basecontext.BaseContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.Role], *apperrors.Diagnostics)
	AddClaimToRole(ctx basecontext.BaseContext, tenantID string, claimID string, roleID string) *apperrors.Diagnostics
	RemoveClaimFromRole(ctx basecontext.BaseContext, tenantID string, claimID string, roleID string) *apperrors.Diagnostics
}

type ClaimDataStore struct {
	common.BaseDataStore
}

func GetClaimDataStoreInstance() ClaimDataStoreInterface {
	if claimDataStoreInstance == nil {
		return NewClaimStore()
	}
	return claimDataStoreInstance
}

func NewClaimStore() *ClaimDataStore {
	return &ClaimDataStore{}
}

func (s *ClaimDataStore) Name() string {
	return "claim_store"
}

func (s *ClaimDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	claimDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *ClaimDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *ClaimDataStore) IsEnabled() bool {
	return true
}

func (s *ClaimDataStore) Dependencies() []string {
	return []string{}
}

func (s *ClaimDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.Get().Get()
	logger := logging.Get(); logger.Info("Initializing claim store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get("database_migrate").GetBool() {
		logger := logging.Get(); logger.Info("Running claim migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate claim store: %v", err)
		}
		logger := logging.Get(); logger.Info("Claim migrations completed")
	}

	claimDataStoreInstance = s
	return nil
}

// Kept for backward compatibility
func InitializeClaimDataStore(db *gorm.DB) (ClaimDataStoreInterface, *apperrors.Diagnostics) {
	if claimDataStoreInstance != nil {
		return claimDataStoreInstance, nil
	}
	s := NewClaimStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := apperrors.NewDiagnostics("initialize_claim_data_store")
		diag.AddError("failed_to_initialize_claim_store", err.Error(), "claim_data_store", nil)
		return nil, diag
	}
	return claimDataStoreInstance, nil
}

func (s *ClaimDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&models.Claim{}); err != nil {
		return fmt.Errorf("failed to migrate claim table: %v", err)
	}

	return nil
}

func (s *ClaimDataStore) GetClaims(ctx basecontext.BaseContext, tenantID string) ([]models.Claim, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_claims")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}
	db := s.GetDB()

	var claims []models.Claim
	result := db.WithContext(ctx.Context()).Where("tenant_id = ?", tenantID).Find(&claims)
	if result.Error != nil {
		diag.AddError("failed_to_get_claims", fmt.Sprintf("failed to get claims: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return nil, diag
	}
	return claims, diag
}

func (s *ClaimDataStore) GetClaimsByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.Claim], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_claims_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}

	db := s.GetDB().WithContext(ctx.Context())

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := filters.QueryDatabase[models.Claim](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_claims_by_query", fmt.Sprintf("failed to get claims by query: %s", common.MapError(err).Error()), "claim_data_store", nil)
		return nil, diag
	}

	return result, diag
}

func (s *ClaimDataStore) GetClaimBySlugOrID(ctx basecontext.BaseContext, tenantID string, slugOrID string) (*models.Claim, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_claim_by_slug_or_id")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}

	db := s.GetDB().WithContext(ctx.Context())
	db = db.Where("tenant_id = ?", tenantID)

	var claim models.Claim
	result := db.First(&claim, "(slug = ? OR id = ?)", slugOrID, slugOrID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return nil, diag
	}
	return &claim, diag
}

func (s *ClaimDataStore) GetClaimUsers(ctx basecontext.BaseContext, tenantID string, claimID string) ([]models.User, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_claim_users")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}

	var users []models.User
	claim, userDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if userDiag.HasErrors() {
		diag.Append(userDiag)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}
	query := s.GetDB().WithContext(ctx.Context()).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN user_claims ON users.id = user_claims.user_id").
		Where("user_claims.claim_id = ?", claim.ID)
	if tenantID != "" {
		query = query.Where("users.tenant_id = ?", tenantID)
	}
	result := query.Find(&users)
	if result.Error != nil {
		diag.AddError("failed_to_get_claim_users", fmt.Sprintf("failed to get claim users: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return nil, diag
	}
	return users, diag
}

func (s *ClaimDataStore) GetClaimUsersByQuery(ctx basecontext.BaseContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.User], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_claim_users_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}

	claim, userDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if userDiag.HasErrors() {
		diag.Append(userDiag)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}
	db := s.GetDB().WithContext(ctx.Context()).
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN user_claims ON users.id = user_claims.user_id").
		Where("user_claims.claim_id = ?", claim.ID)

	result, err := filters.QueryDatabase[models.User](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_claim_users_by_query", fmt.Sprintf("failed to get claim users by query: %s", common.MapError(err).Error()), "claim_data_store", nil)
		return nil, diag
	}

	return result, diag
}

func (s *ClaimDataStore) CreateClaim(ctx basecontext.BaseContext, tenantID string, claim *models.Claim) (*models.Claim, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_create_claim")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}

	if claim.ID == "" {
		claim.ID = uuid.New().String()
	}
	claim.CreatedAt = time.Now()
	claim.UpdatedAt = time.Now()

	result := s.GetDB().WithContext(ctx.Context()).Create(claim)
	if result.Error != nil {
		diag.AddError("failed_to_create_claim", fmt.Sprintf("failed to create claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return nil, diag
	}
	return claim, diag
}

func (s *ClaimDataStore) UpdateClaim(ctx basecontext.BaseContext, tenantID string, claim *models.Claim) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_update_claim")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return diag
	}

	claim.UpdatedAt = time.Now()
	if claim.Slug != "" {
	}

	// check if the claim exists in the database
	claim.UpdatedAt = time.Now()
	if claim.Slug != "" {
	}

	// check if the claim exists in the database
	existingClaim, getClaimDiag := s.GetClaimBySlugOrID(ctx, tenantID, claim.ID)
	if getClaimDiag.HasErrors() {
		diag.Append(getClaimDiag)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// using the partial update map to update the claim
	updates := common.PartialUpdateMap(existingClaim, claim, "updated_at", "slug")
	if err := s.GetDB().WithContext(ctx.Context()).
		Model(&models.Claim{}).
		Where("id = ?", claim.ID).
		Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_claim", fmt.Sprintf("failed to update claim: %s", common.MapError(err).Error()), "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) DeleteClaim(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_delete_claim")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return diag
	}

	err := s.GetDB().WithContext(ctx.Context()).
		Where("tenant_id = ?", tenantID).
		Where("id = ? OR slug = ?", id, id).
		Delete(&models.Claim{}).Error
	if err != nil {
		diag.AddError("failed_to_delete_claim", fmt.Sprintf("failed to delete claim: %s", common.MapError(err).Error()), "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) GetClaimsByLevel(ctx basecontext.BaseContext, tenantID string, level models.SecurityLevel) ([]models.Claim, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_claims_by_level")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}

	var claims []models.Claim
	result := s.GetDB().WithContext(ctx.Context()).
		Where("tenant_id = ?", tenantID).
		Where("security_level = ?", level).
		Find(&claims)
	if result.Error != nil {
		diag.AddError("failed_to_get_claims_by_level", fmt.Sprintf("failed to get claims by level: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return nil, diag
	}
	return claims, diag
}

func (s *ClaimDataStore) AddClaimToUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIdOrSlug string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_add_claim_to_user")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return diag
	}

	var user models.User
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, userID).First(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user", fmt.Sprintf("failed to get user: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimIdOrSlug)
	if err != nil && err.HasErrors() {
		diag.Append(err)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is already assigned to the user
	var userClaims models.UserClaims
	result = s.GetDB().WithContext(ctx.Context()).Where("user_id = ? AND claim_id = ?", user.ID, existingClaim.ID).First(&userClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user_claim", fmt.Sprintf("failed to get user claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
			return diag
		}
	}
	if userClaims.ClaimID != "" {
		diag.AddError("claim_already_assigned_to_user", "claim already assigned to user", "claim_data_store", nil)
		return diag
	}

	// create the user claim
	userClaims.UserID = user.ID
	userClaims.ClaimID = existingClaim.ID
	result = s.GetDB().WithContext(ctx.Context()).Create(&userClaims)
	if result.Error != nil {
		diag.AddError("failed_to_create_user_claim", fmt.Sprintf("failed to create user claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) RemoveClaimFromUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIdOrSlug string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_remove_claim_from_user")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return diag
	}

	var user models.User
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, userID).First(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user", fmt.Sprintf("failed to get user: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimIdOrSlug)
	if err != nil && err.HasErrors() {
		diag.Append(err)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is assigned to the user
	var userClaims models.UserClaims
	result = s.GetDB().WithContext(ctx.Context()).Where("user_id = ? AND claim_id = ?", user.ID, existingClaim.ID).First(&userClaims)
	if result.Error != nil {
		diag.AddError("failed_to_get_user_claim", fmt.Sprintf("failed to get user claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}
	if userClaims.ClaimID == "" {
		diag.AddError("claim_not_assigned_to_user", "claim not assigned to user", "claim_data_store", nil)
		return diag
	}

	// delete the user claim
	result = s.GetDB().WithContext(ctx.Context()).Where("user_id = ? AND claim_id = ?", user.ID, existingClaim.ID).Delete(&userClaims)
	if result.Error != nil {
		diag.AddError("failed_to_delete_user_claim", fmt.Sprintf("failed to delete user claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) GetClaimApiKeys(ctx basecontext.BaseContext, tenantID string, claimID string) ([]models.ApiKey, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_claim_api_keys")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}

	var apiKeys []models.ApiKey
	claim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil && err.HasErrors() {
		diag.Append(err)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}

	query := s.GetDB().WithContext(ctx.Context()).
		Preload("Claims").
		Joins("JOIN api_key_claims ON api_keys.id = api_key_claims.api_key_id").
		Where("api_key_claims.claim_id = ?", claim.ID)
	if tenantID != "" {
		query = query.Where("api_keys.tenant_id = ?", tenantID)
	}

	result := query.Find(&apiKeys)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_keys", fmt.Sprintf("failed to get API keys: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return nil, diag
	}

	return apiKeys, diag
}

func (s *ClaimDataStore) GetClaimApiKeysByQuery(ctx basecontext.BaseContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.ApiKey], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_claim_api_keys_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}

	claim, getClaimDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if getClaimDiag.HasErrors() {
		diag.Append(getClaimDiag)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}

	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN api_key_claims ON api_keys.id = api_key_claims.api_key_id").
		Where("api_key_claims.claim_id = ?", claim.ID)

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := filters.QueryDatabase[models.ApiKey](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_api_keys", fmt.Sprintf("failed to get API keys: %s", common.MapError(err).Error()), "claim_data_store", nil)
		return nil, diag
	}

	return result, diag
}

func (s *ClaimDataStore) AddClaimToApiKey(ctx basecontext.BaseContext, tenantID string, claimID string, apiKeyID string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_add_claim_to_api_key")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return diag
	}

	var apiKey models.ApiKey
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, apiKeyID).First(&apiKey)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_key", fmt.Sprintf("failed to get API key: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}
	if apiKey.ID == "" {
		diag.AddError("api_key_not_found", "API key not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil && err.HasErrors() {
		diag.Append(err)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is already assigned to the api key
	var apiKeyClaims models.ApiKeyClaims
	result = s.GetDB().WithContext(ctx.Context()).Where("api_key_id = ? AND claim_id = ?", apiKey.ID, existingClaim.ID).First(&apiKeyClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_api_key_claim", fmt.Sprintf("failed to get API key claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
			return diag
		}
	}
	if apiKeyClaims.ClaimID != "" {
		diag.AddError("claim_already_assigned_to_api_key", "claim already assigned to API key", "claim_data_store", nil)
		return diag
	}

	// create the api key claim
	apiKeyClaims.ApiKeyID = apiKey.ID
	apiKeyClaims.ClaimID = existingClaim.ID
	result = s.GetDB().WithContext(ctx.Context()).Create(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_create_api_key_claim", fmt.Sprintf("failed to create API key claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) RemoveClaimFromApiKey(ctx basecontext.BaseContext, tenantID string, claimID string, apiKeyID string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_remove_claim_from_api_key")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return diag
	}

	var apiKey models.ApiKey
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, apiKeyID).First(&apiKey)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_key", fmt.Sprintf("failed to get API key: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}
	if apiKey.ID == "" {
		diag.AddError("api_key_not_found", "API key not found", "claim_data_store", nil)
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil && err.HasErrors() {
		diag.Append(err)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is assigned to the api key
	var apiKeyClaims models.ApiKeyClaims
	result = s.GetDB().WithContext(ctx.Context()).Where("api_key_id = ? AND claim_id = ?", apiKey.ID, existingClaim.ID).First(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_key_claim", fmt.Sprintf("failed to get API key claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}
	if apiKeyClaims.ClaimID == "" {
		diag.AddError("claim_not_assigned_to_api_key", "claim not assigned to API key", "claim_data_store", nil)
		return diag
	}

	// delete the api key claim
	result = s.GetDB().WithContext(ctx.Context()).Where("api_key_id = ? AND claim_id = ?", apiKey.ID, existingClaim.ID).Delete(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_delete_api_key_claim", fmt.Sprintf("failed to delete API key claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) GetClaimRolesByQuery(ctx basecontext.BaseContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.Role], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_claim_roles_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}

	claim, getDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if getDiag.HasErrors() {
		diag.Append(getDiag)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}

	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN role_claims ON roles.id = role_claims.role_id").
		Where("role_claims.claim_id = ?", claim.ID)
	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := filters.QueryDatabase[models.Role](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_roles", fmt.Sprintf("failed to get roles: %s", common.MapError(err).Error()), "claim_data_store", nil)
		return nil, diag
	}

	return result, diag
}

func (s *ClaimDataStore) GetClaimRoles(ctx basecontext.BaseContext, tenantID string, claimID string) ([]models.Role, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_claim_roles")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return nil, diag
	}

	claim, getDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if getDiag.HasErrors() {
		diag.Append(getDiag)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}

	var roles []models.Role
	db := s.GetDB().WithContext(ctx.Context()).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN role_claims ON roles.id = role_claims.role_id").
		Where("role_claims.claim_id = ?", claim.ID)

	err := db.Find(&roles).Error
	if err != nil {
		diag.AddError("failed_to_get_roles", fmt.Sprintf("failed to get roles: %s", common.MapError(err).Error()), "claim_data_store", nil)
		return nil, diag
	}

	return roles, diag
}

func (s *ClaimDataStore) AddClaimToRole(ctx basecontext.BaseContext, tenantID string, claimID string, roleID string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_add_claim_to_role")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return diag
	}

	var role models.Role
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, roleID).First(&role)
	if result.Error != nil {
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}
	if role.ID == "" {
		diag.AddError("role_not_found", "role not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim exists
	existingClaim, existingClaimDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if existingClaimDiag.HasErrors() {
		diag.Append(existingClaimDiag)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is assigned to the role
	var roleClaims models.RoleClaims
	result = s.GetDB().WithContext(ctx.Context()).Where("role_id = ? AND claim_id = ?", role.ID, existingClaim.ID).First(&roleClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_role_claim", fmt.Sprintf("failed to get role claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
			return diag
		}
	}
	if roleClaims.ClaimID != "" {
		diag.AddError("claim_already_assigned_to_role", "claim already assigned to role", "claim_data_store", nil)
		return diag
	}

	// create the role claim

	roleClaims.RoleID = role.ID
	roleClaims.ClaimID = existingClaim.ID
	result = s.GetDB().WithContext(ctx.Context()).Create(&roleClaims)
	if result.Error != nil {
		diag.AddError("failed_to_create_role_claim", fmt.Sprintf("failed to create role claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) RemoveClaimFromRole(ctx basecontext.BaseContext, tenantID string, claimID string, roleID string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_remove_claim_from_role")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store", nil)
		return diag
	}

	var role models.Role
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, roleID).First(&role)
	if result.Error != nil {
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}
	if role.ID == "" {
		diag.AddError("role_not_found", "role not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim exists
	existingClaim, existingClaimDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if existingClaimDiag.HasErrors() {
		diag.Append(existingClaimDiag)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is assigned to the role
	var roleClaims models.RoleClaims
	result = s.GetDB().WithContext(ctx.Context()).Where("role_id = ? AND claim_id = ?", role.ID, existingClaim.ID).First(&roleClaims)
	if result.Error != nil {
		diag.AddError("failed_to_get_role_claim", fmt.Sprintf("failed to get role claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}
	if roleClaims.ClaimID == "" {
		diag.AddError("claim_not_assigned_to_role", "claim not assigned to role", "claim_data_store", nil)
		return diag
	}

	// delete the role claim
	result = s.GetDB().WithContext(ctx.Context()).Where("role_id = ? AND claim_id = ?", role.ID, existingClaim.ID).Delete(&roleClaims)
	if result.Error != nil {
		diag.AddError("failed_to_delete_role_claim", fmt.Sprintf("failed to delete role claim: %s", common.MapError(result.Error).Error()), "claim_data_store", nil)
		return diag
	}

	return diag
}
