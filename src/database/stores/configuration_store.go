package stores

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Parallels/prl-devops-service/database/models"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/interfaces"
	apperrors "github.com/Parallels/prl-devops-service/errors"
	logging "github.com/cjlapao/common-go-logger"

	"gorm.io/gorm"
)

var (
	configurationDataStoreInstance *ConfigurationDataStore
	configurationDataStoreOnce     sync.Once
)

type ConfigurationDataStoreInterface interface {
	interfaces.Store
	GetConfigurationValue(ctx context.Context, key string, value interface{}) (interface{}, *apperrors.Diagnostics)
}

type ConfigurationDataStore struct {
	common.BaseDataStore
}

func GetConfigurationDataStoreInstance() ConfigurationDataStoreInterface {
	if configurationDataStoreInstance == nil {
		return NewConfigurationStore()
	}
	return configurationDataStoreInstance
}

func NewConfigurationStore() *ConfigurationDataStore {
	return &ConfigurationDataStore{}
}

func (s *ConfigurationDataStore) Name() string {
	return "configuration_store"
}

func (s *ConfigurationDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	configurationDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *ConfigurationDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *ConfigurationDataStore) IsEnabled() bool {
	return true
}

func (s *ConfigurationDataStore) Dependencies() []string {
	return []string{}
}

func (s *ConfigurationDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.Get()
	logger := logging.Get()
	logger.Info("Initializing configuration store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.IsDatabaseAutoMigrateEnabled() {
		logger.Info("Running configuration migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate configuration store: %v", err)
		}
		logger.Info("Configuration migrations completed")
	}

	configurationDataStoreInstance = s
	return nil
}

// Kept for backward compatibility
func InitializeConfigurationDataStore(db *gorm.DB) (ConfigurationDataStoreInterface, *apperrors.Diagnostics) {
	if configurationDataStoreInstance != nil {
		return configurationDataStoreInstance, nil
	}
	s := NewConfigurationStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := apperrors.NewDiagnostics("initialize_configuration_data_store")
		diag.AddError("failed_to_initialize_configuration_store", err.Error(), "configuration_data_store", nil)
		return nil, diag
	}
	return configurationDataStoreInstance, nil
}

func (s *ConfigurationDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&models.Configuration{}); err != nil {
		return fmt.Errorf("failed to migrate configuration table: %v", err)
	}
	return nil
}

func (s *ConfigurationDataStore) GetConfigurationValue(ctx context.Context, key string, value interface{}) (interface{}, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_configuration_value")
	db := s.GetDB()
	if key == "" {
		diag.AddError("key_cannot_be_empty", "key cannot be empty", "configuration_data_store")
		return nil, diag
	}

	err := db.Where("key = ?", key).
		Where("tenant_id = ?").
		First(&models.Configuration{}).
		Scan(&value).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_configuration_value", fmt.Sprintf("failed to get configuration value: %s", common.MapError(err).Error()), "configuration_data_store", nil)
		return nil, diag
	}

	return value, diag
}
