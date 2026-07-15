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
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/database/interfaces"
	"github.com/Parallels/prl-devops-service/database/models"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	logging "github.com/cjlapao/common-go-logger"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	userConfigDataStoreInstance *UserConfigDataStore
	userConfigDataStoreOnce     sync.Once
)

type UserConfigDataStoreInterface interface {
	interfaces.Store
	Get(ctx basecontext.BaseContext, userID, idOrSlug string) (*models.UserConfig, *apperrors.Diagnostics)
	Find(ctx basecontext.BaseContext, userID string, filter *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.UserConfig], *apperrors.Diagnostics)
	Create(ctx basecontext.BaseContext, config *models.UserConfig) (*models.UserConfig, *apperrors.Diagnostics)
	Update(ctx basecontext.BaseContext, config *models.UserConfig) *apperrors.Diagnostics
	Delete(ctx basecontext.BaseContext, userID, idOrSlug string) *apperrors.Diagnostics
}

type UserConfigDataStore struct {
	common.BaseDataStore
}

func GetUserConfigDataStoreInstance() UserConfigDataStoreInterface {
	if userConfigDataStoreInstance == nil {
		return NewUserConfigStore()
	}
	return userConfigDataStoreInstance
}

func NewUserConfigStore() *UserConfigDataStore {
	return &UserConfigDataStore{}
}

func (s *UserConfigDataStore) Name() string {
	return "user_config_store"
}

func (s *UserConfigDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	userConfigDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *UserConfigDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *UserConfigDataStore) IsEnabled() bool {
	return true
}

func (s *UserConfigDataStore) Dependencies() []string {
	return []string{"user_store"}
}

func (s *UserConfigDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.Get()
	logger := logging.Get()
	logger.Info("Initializing user config store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.IsDatabaseAutoMigrateEnabled() {
		logger.Info("Running user config migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate user config store: %v", err)
		}
		logger.Info("User config migrations completed")
	}

	userConfigDataStoreInstance = s
	return nil
}

func (s *UserConfigDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&models.UserConfig{}); err != nil {
		return fmt.Errorf("failed to migrate user_configs table: %v", err)
	}

	// Create composite unique index on user_id + slug
	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_config_user_slug ON user_configs(user_id, slug);").Error; err != nil {
		return fmt.Errorf("failed to create unique index on user_configs: %v", err)
	}

	return nil
}

// Get retrieves a user config by ID or slug
func (s *UserConfigDataStore) Get(ctx basecontext.BaseContext, userID, idOrSlug string) (*models.UserConfig, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_user_config")

	if userID == "" {
		diag.AddError("user_id_required", "user_id is required", "user_config_store", nil)
		return nil, diag
	}

	if idOrSlug == "" {
		diag.AddError("id_or_slug_required", "id or slug is required", "user_config_store", nil)
		return nil, diag
	}

	var config models.UserConfig
	err := s.GetDB().WithContext(ctx.Context()).
		Where("user_id = ?", userID).
		Where("id = ? OR slug = ?", idOrSlug, idOrSlug).
		First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			diag.AddError("user_config_not_found", fmt.Sprintf("user config not found: %s", idOrSlug), "user_config_store", nil)
			return nil, diag
		}
		diag.AddError("failed_to_get_user_config", fmt.Sprintf("failed to get user config: %s", common.MapError(err).Error()), "user_config_store", nil)
		return nil, diag
	}

	return &config, nil
}

// Find retrieves user configs with filtering
func (s *UserConfigDataStore) Find(ctx basecontext.BaseContext, userID string, filter *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.UserConfig], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_find_user_configs")

	if userID == "" {
		diag.AddError("user_id_required", "user_id is required", "user_config_store", nil)
		return nil, diag
	}

	query := s.GetDB().WithContext(ctx.Context()).Where("user_id = ?", userID)

	if filter != nil {
		result, err := filters.QueryDatabase[models.UserConfig](query, "", filter)
		if err != nil {
			diag.AddError("failed_to_apply_filter", fmt.Sprintf("failed to apply filter: %s", err.Error()), "user_config_store", nil)
			return nil, diag
		}
		return result, nil
	}

	// No filter - return all configs for the user
	var configs []models.UserConfig
	if err := query.Find(&configs).Error; err != nil {
		diag.AddError("failed_to_find_user_configs", fmt.Sprintf("failed to find user configs: %s", common.MapError(err).Error()), "user_config_store", nil)
		return nil, diag
	}

	return &filters.QueryBuilderResponse[models.UserConfig]{
		Items: configs,
		Total: int64(len(configs)),
	}, nil
}

// Create creates a new user config
func (s *UserConfigDataStore) Create(ctx basecontext.BaseContext, config *models.UserConfig) (*models.UserConfig, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_create_user_config")

	if config.UserID == "" {
		diag.AddError("user_id_required", "user_id is required", "user_config_store", nil)
		return nil, diag
	}

	if config.Slug == "" {
		diag.AddError("slug_required", "slug is required", "user_config_store", nil)
		return nil, diag
	}

	if config.ID == "" {
		config.ID = uuid.New().String()
	}

	now := time.Now()
	config.CreatedAt = now
	config.UpdatedAt = now

	if err := s.GetDB().WithContext(ctx.Context()).Create(config).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			diag.AddError("user_config_already_exists", fmt.Sprintf("user config with slug '%s' already exists for this user", config.Slug), "user_config_store", nil)
			return nil, diag
		}
		diag.AddError("failed_to_create_user_config", fmt.Sprintf("failed to create user config: %s", common.MapError(err).Error()), "user_config_store", nil)
		return nil, diag
	}

	return config, nil
}

// Update updates an existing user config
func (s *UserConfigDataStore) Update(ctx basecontext.BaseContext, config *models.UserConfig) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_update_user_config")

	if config.ID == "" {
		diag.AddError("id_required", "id is required", "user_config_store", nil)
		return diag
	}

	if config.UserID == "" {
		diag.AddError("user_id_required", "user_id is required", "user_config_store", nil)
		return diag
	}

	config.UpdatedAt = time.Now()

	result := s.GetDB().WithContext(ctx.Context()).
		Where("id = ? AND user_id = ?", config.ID, config.UserID).
		Updates(config)

	if result.Error != nil {
		diag.AddError("failed_to_update_user_config", fmt.Sprintf("failed to update user config: %s", common.MapError(result.Error).Error()), "user_config_store", nil)
		return diag
	}

	if result.RowsAffected == 0 {
		diag.AddError("user_config_not_found", "user config not found or does not belong to this user", "user_config_store", nil)
		return diag
	}

	return nil
}

// Delete removes a user config
func (s *UserConfigDataStore) Delete(ctx basecontext.BaseContext, userID, idOrSlug string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_delete_user_config")

	if userID == "" {
		diag.AddError("user_id_required", "user_id is required", "user_config_store", nil)
		return diag
	}

	if idOrSlug == "" {
		diag.AddError("id_or_slug_required", "id or slug is required", "user_config_store", nil)
		return diag
	}

	result := s.GetDB().WithContext(ctx.Context()).
		Where("user_id = ?", userID).
		Where("id = ? OR slug = ?", idOrSlug, idOrSlug).
		Delete(&models.UserConfig{})

	if result.Error != nil {
		diag.AddError("failed_to_delete_user_config", fmt.Sprintf("failed to delete user config: %s", common.MapError(result.Error).Error()), "user_config_store", nil)
		return diag
	}

	if result.RowsAffected == 0 {
		diag.AddError("user_config_not_found", "user config not found", "user_config_store", nil)
		return diag
	}

	return nil
}
