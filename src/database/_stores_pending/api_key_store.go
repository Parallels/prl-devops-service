package stores

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	authDataStoreInstance *ApiKeyDataStore
	authDataStoreOnce     sync.Once
)

type ApiKeyStoreInterface interface {
	interfaces.Store
	CreateApiKey(ctx basecontext.BaseContext, tenantID string, apiKey *entities.ApiKey) (*entities.ApiKey, *apperrors.Diagnostics)

	GetApiKeyByHash(ctx basecontext.BaseContext, tenantID string, keyHash string) (*entities.ApiKey, *apperrors.Diagnostics)
	GetApiKeyByDigest(ctx basecontext.BaseContext, tenantID string, digest string) (*entities.ApiKey, *apperrors.Diagnostics)
	GetApiKeyByPrefix(ctx basecontext.BaseContext, tenantID string, keyPrefix string) (*entities.ApiKey, *apperrors.Diagnostics)
	GetApiKeyByName(ctx basecontext.BaseContext, tenantID string, name string) (*entities.ApiKey, *apperrors.Diagnostics)
	GetApiKeys(ctx basecontext.BaseContext, tenantID string) ([]entities.ApiKey, *apperrors.Diagnostics)
	GetApiKeysByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.ApiKey], *apperrors.Diagnostics)
	GetApiKeyByIDOrSlug(ctx basecontext.BaseContext, tenantID string, id string) (*entities.ApiKey, *apperrors.Diagnostics)
	RevokeApiKey(ctx basecontext.BaseContext, tenantID string, id string, revokedBy string, reason string) *apperrors.Diagnostics
	DeleteApiKey(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics
	AddClaimToApiKey(ctx basecontext.BaseContext, tenantID string, id string, claimID string) *apperrors.Diagnostics
	RemoveClaimFromApiKey(ctx basecontext.BaseContext, tenantID string, id string, claimID string) *apperrors.Diagnostics
	UpdateApiKeyLastUsed(ctx basecontext.BaseContext, tenantID string, apiKeyID string) *apperrors.Diagnostics
	UpdateApiKeyClaims(ctx basecontext.BaseContext, tenantID string, apiKey *entities.ApiKey, claims []entities.Claim) *apperrors.Diagnostics
	GetDB() *gorm.DB
}

// ApiKeyDataStore handles auth-specific database operations
type ApiKeyDataStore struct {
	common.BaseDataStore
}

// GetApiKeyDataStoreInstance returns the singleton instance of the auth store
func GetApiKeyDataStoreInstance() ApiKeyStoreInterface {
	if authDataStoreInstance == nil {
		return NewApiKeyStore()
	}
	return authDataStoreInstance
}

func NewApiKeyStore() *ApiKeyDataStore {
	return &ApiKeyDataStore{}
}

func (s *ApiKeyDataStore) Name() string {
	return "api_key_store"
}

func (s *ApiKeyDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	authDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *ApiKeyDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *ApiKeyDataStore) IsEnabled() bool {
	return true
}

func (s *ApiKeyDataStore) Dependencies() []string {
	return []string{}
}

func (s *ApiKeyDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.GetInstance().Get()
	logging.Info("Initializing api key store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get(config.DatabaseMigrateKey).GetBool() {
		logging.Info("Running api key migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate api key store: %v", err)
		}
		logging.Info("Api key migrations completed")
	}

	authDataStoreInstance = s
	return nil
}

// InitializeApiKeyDataStore initializes the api key store singleton
// Kept for backward compatibility
func InitializeApiKeyDataStore(db *gorm.DB) (ApiKeyStoreInterface, *apperrors.Diagnostics) {
	if authDataStoreInstance != nil {
		return authDataStoreInstance, nil
	}
	s := NewApiKeyStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := errors.NewDiagnostics("initialize_api_key_data_store")
		diag.AddError("failed_to_initialize_api_key_store", err.Error(), "api_key_data_store", err)
		return nil, diag
	}
	return authDataStoreInstance, nil
}

// Migrate implements the DataStore interface
func (s *ApiKeyDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.ApiKey{}); err != nil {
		return fmt.Errorf("failed to migrate api_keys table: %v", err)
	}

	if err := s.GetDB().AutoMigrate(&entities.ApiKeyClaims{}); err != nil {
		return fmt.Errorf("failed to migrate api_key_claims table: %v", err)
	}

	// add unique index to the api_key_claims table
	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_api_key_claims ON api_key_claims (api_key_id, claim_id);").Error; err != nil {
		return fmt.Errorf("failed to create unique index on api_key_claims table: %v", err)
	}

	return nil
}

// CreateApiKey creates a new API key for a user
func (s *ApiKeyDataStore) CreateApiKey(ctx basecontext.BaseContext, tenantID string, apiKey *entities.ApiKey) (*entities.ApiKey, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("create_api_key")
	defer diag.Complete()

	if apiKey == nil {
		diag.AddError("api_key_is_nil", "api key is nil", "api_key_data_store", nil)
		return nil, diag
	}

	if apiKey.KeyHash == "" {
		diag.AddError("api_key_hash_is_empty", "api key hash is empty", "api_key_data_store", nil)
		return nil, diag
	}

	if apiKey.KeyPrefix == "" {
		diag.AddError("api_key_prefix_is_empty", "api key prefix is empty", "api_key_data_store", nil)
		return nil, diag
	}

	if apiKey.Name == "" {
		diag.AddError("api_key_name_is_empty", "api key name is empty", "api_key_data_store", nil)
		return nil, diag
	}

	if apiKey.Claims == nil {
		diag.AddError("api_key_claims_are_empty", "api key claims are empty", "api_key_data_store", nil)
		return nil, diag
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		diag.AddError("api_key_expires_at_is_in_the_past", "api key expires at is in the past", "api_key_data_store", nil)
		return nil, diag
	}

	if apiKey.ID == "" {
		apiKey.ID = uuid.New().String()
	}
	if apiKey.Slug == "" {
		apiKey.Slug = pkg_utils.Slugify(apiKey.Name)
	}

	apiKey.CreatedAt = time.Now()
	apiKey.UpdatedAt = time.Now()
	baseModel := common.GetTenantBaseModelFromContext(ctx, &apiKey.BaseModelWithTenant)
	apiKey.ID = baseModel.ID
	apiKey.CreatedAt = baseModel.CreatedAt
	apiKey.UpdatedAt = baseModel.UpdatedAt
	apiKey.CreatedBy = baseModel.CreatedBy
	apiKey.UpdatedBy = baseModel.UpdatedBy
	apiKey.TenantID = baseModel.TenantID
	if apiKey.TenantID == "" {
		apiKey.TenantID = tenantID
	}

	// Hash the API key before storing
	encryptionService := encryption.GetInstance()
	keyHash, err := encryptionService.HashPassword(apiKey.KeyHash)
	if err != nil {
		diag.AddError("failed_to_hash_api_key", fmt.Sprintf("failed to hash API key: %s", err.Error()), "api_key_data_store", nil)
		return nil, diag
	}
	apiKey.KeyHash = keyHash

	// Create deterministic digest (SHA-256 of the full key)
	sha := sha256.Sum256([]byte(apiKey.KeyDigest))
	apiKey.KeyDigest = hex.EncodeToString(sha[:])

	result := s.GetDB().WithContext(ctx).Create(apiKey)
	if result.Error != nil {
		diag.AddError("failed_to_create_api_key", fmt.Sprintf("failed to create API key: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return nil, diag
	}

	return apiKey, diag
}

// GetApiKeyByDigest retrieves an API key by its deterministic digest
func (s *ApiKeyDataStore) GetApiKeyByDigest(ctx basecontext.BaseContext, tenantID string, digest string) (*entities.ApiKey, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_api_key_by_digest")
	defer diag.Complete()

	var apiKey entities.ApiKey
	result := s.GetDB().WithContext(ctx).Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "key_digest = ? AND tenant_id = ?", digest, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_api_key_by_digest", fmt.Sprintf("failed to get API key by digest: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return nil, diag
	}
	return &apiKey, diag
}

// GetApiKeyByHash retrieves an API key by its hash
func (s *ApiKeyDataStore) GetApiKeyByHash(ctx basecontext.BaseContext, tenantID string, keyHash string) (*entities.ApiKey, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_api_key_by_hash")
	defer diag.Complete()

	var apiKey entities.ApiKey
	result := s.GetDB().WithContext(ctx).Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "key_hash = ? AND tenant_id = ?", keyHash, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_api_key_by_hash", fmt.Sprintf("failed to get API key by hash: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return nil, diag
	}
	return &apiKey, diag
}

// GetApiKeyByPrefix retrieves an API key by its prefix (for validation)
func (s *ApiKeyDataStore) GetApiKeyByPrefix(ctx basecontext.BaseContext, tenantID string, keyPrefix string) (*entities.ApiKey, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_api_key_by_prefix")
	defer diag.Complete()

	var apiKey entities.ApiKey
	result := s.GetDB().WithContext(ctx).Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "key_prefix = ? AND tenant_id = ?", keyPrefix, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_api_key_by_prefix", fmt.Sprintf("failed to get API key by prefix: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return nil, diag
	}
	return &apiKey, diag
}

// GetApiKeyByIDOrSlug retrieves an API key by ID
func (s *ApiKeyDataStore) GetApiKeyByIDOrSlug(ctx basecontext.BaseContext, tenantID string, id string) (*entities.ApiKey, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_api_key_by_id_or_slug")
	defer diag.Complete()

	var apiKey entities.ApiKey
	result := s.GetDB().WithContext(ctx).Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "id = ? or slug = ? AND tenant_id = ?", id, id, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_api_key_by_id_or_slug", fmt.Sprintf("failed to get API key by ID: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return nil, diag
	}
	return &apiKey, diag
}

// RevokeApiKey revokes an API key
func (s *ApiKeyDataStore) RevokeApiKey(ctx basecontext.BaseContext, tenantID string, id string, revokedBy string, reason string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("revoke_api_key")
	defer diag.Complete()

	now := time.Now()
	result := s.GetDB().WithContext(ctx).Model(&entities.ApiKey{}).Where("id = ? AND tenant_id = ?", id, tenantID).Updates(map[string]interface{}{
		"is_active":         false,
		"revoked_at":        now,
		"revoked_by":        revokedBy,
		"revocation_reason": reason,
		"updated_at":        now,
	})
	if result.Error != nil {
		diag.AddError("failed_to_revoke_api_key", fmt.Sprintf("failed to revoke API key: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}
	return diag
}

// DeleteApiKey permanently deletes an API key
func (s *ApiKeyDataStore) DeleteApiKey(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("delete_api_key")
	defer diag.Complete()

	// deleting the api key relationships first
	err := s.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("api_key_id = ?", id).Delete(&entities.ApiKeyClaims{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&entities.ApiKey{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		diag.AddError("failed_to_delete_api_key", fmt.Sprintf("failed to delete API key: %s", common.MapError(err).Error()), "api_key_data_store", nil)
		return diag
	}
	return diag
}

// CleanupExpiredAPIKeys removes expired API keys
func (s *ApiKeyDataStore) CleanupExpiredAPIKeys(ctx basecontext.BaseContext) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("cleanup_expired_api_keys")
	defer diag.Complete()

	now := time.Now()
	result := s.GetDB().WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at < ?", now).Delete(&entities.ApiKey{})
	if result.Error != nil {
		diag.AddError("failed_to_cleanup_expired_api_keys", fmt.Sprintf("failed to cleanup expired API keys: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}
	return diag
}

func (s *ApiKeyDataStore) GetApiKeys(ctx basecontext.BaseContext, tenantID string) ([]entities.ApiKey, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_api_keys")
	defer diag.Complete()

	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "api_key_data_store", nil)
		return nil, diag
	}

	var apiKeys []entities.ApiKey
	db := s.GetDB().WithContext(ctx).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		})

	if tenantID != "" {
		db = db.Where("tenant_id = ?", tenantID)
	}

	result := db.Find(&apiKeys)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_keys", fmt.Sprintf("failed to get API keys: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return nil, diag
	}

	return apiKeys, diag
}

// GetApiKeysByQuery retrieves paginated API keys based on a pagination and filter
func (s *ApiKeyDataStore) GetApiKeysByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.ApiKey], *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_paginated_api_keys")
	defer diag.Complete()

	db := s.GetDB().WithContext(ctx)
	db = db.Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	})

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := filters.QueryDatabase[entities.ApiKey](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_api_keys", fmt.Sprintf("failed to get API keys: %s", common.MapError(err).Error()), "api_key_data_store", nil)
		return nil, diag
	}

	return result, diag
}

func (s *ApiKeyDataStore) AddClaimToApiKey(ctx basecontext.BaseContext, tenantID string, id string, claimID string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("add_claim_to_api_key")
	defer diag.Complete()

	var dbClaim entities.Claim
	result := s.GetDB().WithContext(ctx).First(&dbClaim, "id = ? AND tenant_id = ?", claimID, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("claim_not_found", "claim not found", "api_key_data_store", nil)
			return diag
		}
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}

	existingApiKey, getApiKeyDiag := s.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return diag
	}
	if existingApiKey == nil {
		diag.AddError("api_key_not_found", "API key not found", "api_key_data_store", nil)
		return diag
	}
	// checking if the claim is already assigned to the api key
	var apiKeyClaims entities.ApiKeyClaims
	result = s.GetDB().WithContext(ctx).Where("api_key_id = ? AND claim_id = ?", id, claimID).First(&apiKeyClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user_claim", fmt.Sprintf("failed to get user claim: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
			return diag
		}
	}
	if apiKeyClaims.ClaimID != "" {
		diag.AddError("claim_already_assigned_to_api_key", "claim already assigned to API key", "api_key_data_store", nil)
		return diag
	}

	// creating the api key claim
	apiKeyClaims.ApiKeyID = id
	apiKeyClaims.ClaimID = dbClaim.ID
	result = s.GetDB().WithContext(ctx).Create(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_create_api_key_claim", fmt.Sprintf("failed to create API key claim: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}
	return diag
}

func (s *ApiKeyDataStore) RemoveClaimFromApiKey(ctx basecontext.BaseContext, tenantID string, id string, claimID string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("remove_claim_from_api_key")
	defer diag.Complete()

	var dbClaim entities.Claim
	result := s.GetDB().WithContext(ctx).First(&dbClaim, "id = ? AND tenant_id = ?", claimID, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("claim_not_found", "claim not found", "api_key_data_store", nil)
			return diag
		}
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}

	existingApiKey, getApiKeyDiag := s.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return diag
	}
	if existingApiKey == nil {
		diag.AddError("api_key_not_found", "API key not found", "api_key_data_store", nil)
		return diag
	}

	// checking if the claim is assigned to the api key
	var apiKeyClaims entities.ApiKeyClaims
	result = s.GetDB().WithContext(ctx).Where("api_key_id = ? AND claim_id = ? AND tenant_id = ?", id, claimID, tenantID).First(&apiKeyClaims)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("claim_not_assigned_to_api_key", "claim not assigned to API key", "api_key_data_store", nil)
			return diag
		}
		diag.AddError("failed_to_get_api_key_claim", fmt.Sprintf("failed to get API key claim: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}

	// deleting the api key claim
	result = s.GetDB().WithContext(ctx).Where("api_key_id = ? AND claim_id = ? AND tenant_id = ?", id, claimID, tenantID).Delete(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_delete_api_key_claim", fmt.Sprintf("failed to delete API key claim: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}

	return diag
}

func (s *ApiKeyDataStore) GetApiKeyByName(ctx basecontext.BaseContext, tenantID string, name string) (*entities.ApiKey, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_api_key_by_name")
	defer diag.Complete()

	var apiKey entities.ApiKey
	result := s.GetDB().WithContext(ctx).Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "name = ? AND tenant_id = ?", name, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_api_key_by_name", fmt.Sprintf("failed to get API key by name: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return nil, diag
	}

	return &apiKey, diag
}

func (s *ApiKeyDataStore) UpdateApiKeyLastUsed(ctx basecontext.BaseContext, tenantID string, apiKeyID string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("update_api_key_last_used")
	defer diag.Complete()

	now := time.Now()
	result := s.GetDB().WithContext(ctx).
		Model(&entities.ApiKey{}).
		Where("id = ? AND tenant_id = ?", apiKeyID, tenantID).
		Update("last_used_at", now)
	if result.Error != nil {
		diag.AddError("failed_to_update_api_key_last_used", fmt.Sprintf("failed to update API key last used: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}
	return diag
}

func (s *ApiKeyDataStore) UpdateApiKeyClaims(ctx basecontext.BaseContext, tenantID string, apiKey *entities.ApiKey, claims []entities.Claim) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("update_api_key_claims")
	defer diag.Complete()

	err := s.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// First, clear any existing claim associations
		if err := tx.Model(apiKey).Association("Claims").Clear(); err != nil {
			return err
		}

		// Then add the new claim associations
		if len(claims) > 0 {
			if err := tx.Model(apiKey).Association("Claims").Append(claims); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		diag.AddError("failed_to_update_api_key_claims", fmt.Sprintf("failed to update API key claims: %s", common.MapError(err).Error()), "api_key_data_store", nil)
		return diag
	}

	return diag
}
