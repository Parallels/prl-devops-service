package stores

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/database/models"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/database/interfaces"
	"github.com/Parallels/prl-devops-service/security/password"

	apperrors "github.com/Parallels/prl-devops-service/errors"
	logging "github.com/cjlapao/common-go-logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	authDataStoreInstance *ApiKeyDataStore
	authDataStoreOnce     sync.Once
)

type ApiKeyStoreInterface interface {
	interfaces.Store
	CreateApiKey(ctx basecontext.BaseContext, apiKey *models.ApiKey) (*models.ApiKey, *apperrors.Diagnostics)
	GetApiKey(ctx basecontext.BaseContext, idOrName string) (*models.ApiKey, *apperrors.Diagnostics)
	GetApiKeys(ctx basecontext.BaseContext) ([]models.ApiKey, *apperrors.Diagnostics)
	GetApiKeysByQuery(ctx basecontext.BaseContext, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.ApiKey], *apperrors.Diagnostics)
	RevokeApiKey(ctx basecontext.BaseContext, id string) *apperrors.Diagnostics
	DeleteApiKey(ctx basecontext.BaseContext, id string) *apperrors.Diagnostics
	UpdateApiKey(ctx basecontext.BaseContext, apiKey *models.ApiKey) *apperrors.Diagnostics
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
	cfg := config.Get()
	logger := logging.Get()
	logger.Info("Initializing api key store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.IsDatabaseAutoMigrateEnabled() {
		logger.Info("Running api key migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate api key store: %v", err)
		}
		logger.Info("Api key migrations completed")
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
		diag := apperrors.NewDiagnostics("initialize_api_key_data_store")
		diag.AddError("failed_to_initialize_api_key_store", err.Error(), "api_key_data_store", nil)
		return nil, diag
	}
	return authDataStoreInstance, nil
}

// Migrate implements the DataStore interface
func (s *ApiKeyDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&models.ApiKey{}); err != nil {
		return fmt.Errorf("failed to migrate api_keys table: %v", err)
	}

	if err := s.GetDB().AutoMigrate(&models.RoleClaims{}); err != nil {
		return fmt.Errorf("failed to migrate role_claims table: %v", err)
	}

	// add unique index to the role_claims table
	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_role_claims ON role_claims (role_id, claim_id);").Error; err != nil {
		return fmt.Errorf("failed to create unique index on role_claims table: %v", err)
	}

	return nil
}

// CreateApiKey creates a new API key
func (s *ApiKeyDataStore) CreateApiKey(ctx basecontext.BaseContext, apiKey *models.ApiKey) (*models.ApiKey, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("create_api_key")
	defer diag.Complete()

	if apiKey == nil {
		diag.AddError("api_key_is_nil", "api key is nil", "api_key_data_store", nil)
		return nil, diag
	}

	if apiKey.Name == "" {
		diag.AddError("api_key_name_is_empty", "api key name is empty", "api_key_data_store", nil)
		return nil, diag
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		diag.AddError("api_key_expires_at_is_in_the_past", "api key expires at is in the past", "api_key_data_store", nil)
		return nil, diag
	}

	if apiKey.ID == "" {
		apiKey.ID = uuid.New().String()
	}
	apiKey.CreatedAt = time.Now()
	apiKey.UpdatedAt = time.Now()

	// Hash the secret before storing
	if apiKey.Secret != "" {
		passwdSvc := password.Get()
		hashSecret, err := passwdSvc.Hash(apiKey.Secret, apiKey.ID)
		if err != nil {
			diag.AddError("failed_to_hash_secret", fmt.Sprintf("failed to hash secret: %s", err.Error()), "api_key_data_store", nil)
			return nil, diag
		}
		apiKey.Secret = hashSecret
	}

	result := s.GetDB().WithContext(ctx.Context()).Create(apiKey)
	if result.Error != nil {
		diag.AddError("failed_to_create_api_key", fmt.Sprintf("failed to create API key: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return nil, diag
	}

	return apiKey, diag
}

// GetApiKey retrieves an API key by ID, name, or key
func (s *ApiKeyDataStore) GetApiKey(ctx basecontext.BaseContext, idOrName string) (*models.ApiKey, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("get_api_key")
	defer diag.Complete()

	var apiKey models.ApiKey
	result := s.GetDB().WithContext(ctx.Context()).Where("id = ? OR name = ? OR key = ?", idOrName, idOrName, idOrName).First(&apiKey)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_api_key", fmt.Sprintf("failed to get API key: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return nil, diag
	}
	return &apiKey, diag
}

// GetApiKeys retrieves all API keys
func (s *ApiKeyDataStore) GetApiKeys(ctx basecontext.BaseContext) ([]models.ApiKey, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("get_api_keys")
	defer diag.Complete()

	var apiKeys []models.ApiKey
	result := s.GetDB().WithContext(ctx.Context()).Find(&apiKeys)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_keys", fmt.Sprintf("failed to get API keys: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return nil, diag
	}

	return apiKeys, diag
}

// GetApiKeysByQuery retrieves paginated API keys based on a query
func (s *ApiKeyDataStore) GetApiKeysByQuery(ctx basecontext.BaseContext, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.ApiKey], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("get_api_keys_by_query")
	defer diag.Complete()

	db := s.GetDB().WithContext(ctx.Context())

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := filters.QueryDatabase[models.ApiKey](db, "", queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_api_keys", fmt.Sprintf("failed to get API keys: %s", common.MapError(err).Error()), "api_key_data_store", nil)
		return nil, diag
	}

	return result, diag
}

// RevokeApiKey revokes an API key
func (s *ApiKeyDataStore) RevokeApiKey(ctx basecontext.BaseContext, id string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("revoke_api_key")
	defer diag.Complete()

	now := time.Now()
	result := s.GetDB().WithContext(ctx.Context()).Model(&models.ApiKey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"revoked":    true,
		"revoked_at": now,
		"updated_at": now,
	})
	if result.Error != nil {
		diag.AddError("failed_to_revoke_api_key", fmt.Sprintf("failed to revoke API key: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}
	return diag
}

// DeleteApiKey permanently deletes an API key
func (s *ApiKeyDataStore) DeleteApiKey(ctx basecontext.BaseContext, id string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("delete_api_key")
	defer diag.Complete()

	result := s.GetDB().WithContext(ctx.Context()).Where("id = ? OR name = ? OR key = ?", id, id, id).Delete(&models.ApiKey{})
	if result.Error != nil {
		diag.AddError("failed_to_delete_api_key", fmt.Sprintf("failed to delete API key: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}
	return diag
}

// UpdateApiKey updates an API key
func (s *ApiKeyDataStore) UpdateApiKey(ctx basecontext.BaseContext, apiKey *models.ApiKey) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("update_api_key")
	defer diag.Complete()

	if apiKey == nil {
		diag.AddError("api_key_is_nil", "api key is nil", "api_key_data_store", nil)
		return diag
	}

	apiKey.UpdatedAt = time.Now()
	result := s.GetDB().WithContext(ctx.Context()).Save(apiKey)
	if result.Error != nil {
		diag.AddError("failed_to_update_api_key", fmt.Sprintf("failed to update API key: %s", common.MapError(result.Error).Error()), "api_key_data_store", nil)
		return diag
	}
	return diag
}
