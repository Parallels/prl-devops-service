package stores

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/database/interfaces"
	"github.com/cjlapao/common-go-logger"


	"gorm.io/gorm"
)

var (
	configurationDataStoreInstance *ConfigurationDataStore
	configurationDataStoreOnce     sync.Once
)

type ConfigurationDataStoreInterface interface {
	interfaces.Store
	GetConfigurationValue(ctx context.Context, tenantID string, key string, value interface{}) (interface{}, *apperrors.Diagnostics)
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
	cfg := config.GetInstance().Get()
	logging.Info("Initializing configuration store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get(config.DatabaseMigrateKey).GetBool() {
		logging.Info("Running configuration migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate configuration store: %v", err)
		}
		logging.Info("Configuration migrations completed")
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
		diag := errors.NewDiagnostics("initialize_configuration_data_store")
		diag.AddError("failed_to_initialize_configuration_store", err.Error(), "configuration_data_store", err)
		return nil, diag
	}
	return configurationDataStoreInstance, nil
}

func (s *ConfigurationDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.Configuration{}); err != nil {
		return fmt.Errorf("failed to migrate configuration table: %v", err)
	}
	return nil
}

func (s *ConfigurationDataStore) GetConfigurationValue(ctx context.Context, tenantID string, key string, value interface{}) (interface{}, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_configuration_value")
	db := s.GetDB()
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "configuration_data_store")
		return nil, diag
	}
	if key == "" {
		diag.AddError("key_cannot_be_empty", "key cannot be empty", "configuration_data_store")
		return nil, diag
	}

	err := db.Where("key = ?", key).
		Where("tenant_id = ?", tenantID).
		First(&entities.Configuration{}).
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
