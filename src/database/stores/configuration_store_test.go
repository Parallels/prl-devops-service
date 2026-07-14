package stores_test

import (
	"context"
	"testing"

	"github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/stores"
	"github.com/Parallels/prl-devops-service/database/stores/testhelpers"
	logging "github.com/cjlapao/common-go-logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConfigurationDataStore(t *testing.T) {
	logger := logging.Get()
	logger.Info("Starting configuration store tests")
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	store := &stores.ConfigurationDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	ctx := context.Background()

	t.Run("CreateConfiguration", func(t *testing.T) {
		config := &models.Configuration{
			BaseModel: models.BaseModel{ID: uuid.New().String()},
			Vault:     "default",
			Key:       "app.name",
			Value:     "Test Application",
			Version:   1,
		}

		result := db.Create(config)
		assert.NoError(t, result.Error)
		assert.NotEmpty(t, config.ID)
		assert.Equal(t, "app.name", config.Key)
		assert.Equal(t, "Test Application", config.Value)
	})

	t.Run("GetConfiguration", func(t *testing.T) {
		config := &models.Configuration{
			BaseModel: models.BaseModel{ID: uuid.New().String()},
			Vault:     "default",
			Key:       "app.version",
			Value:     "1.0.0",
			Version:   1,
		}
		db.Create(config)

		var retrieved models.Configuration
		result := db.Where("key = ?", "app.version").First(&retrieved)
		assert.NoError(t, result.Error)
		assert.Equal(t, "1.0.0", retrieved.Value)
	})

	t.Run("UpdateConfiguration", func(t *testing.T) {
		config := &models.Configuration{
			BaseModel: models.BaseModel{ID: uuid.New().String()},
			Vault:     "default",
			Key:       "app.theme",
			Value:     "dark",
			Version:   1,
		}
		db.Create(config)

		config.Value = "light"
		config.Version = 2
		result := db.Save(config)
		assert.NoError(t, result.Error)

		var updated models.Configuration
		db.Where("key = ?", "app.theme").First(&updated)
		assert.Equal(t, "light", updated.Value)
		assert.Equal(t, 2, updated.Version)
	})

	t.Run("DeleteConfiguration", func(t *testing.T) {
		config := &models.Configuration{
			BaseModel: models.BaseModel{ID: uuid.New().String()},
			Vault:     "default",
			Key:       "temp.setting",
			Value:     "temporary",
			Version:   1,
		}
		db.Create(config)

		result := db.Delete(config)
		assert.NoError(t, result.Error)

		var deleted models.Configuration
		result = db.Where("key = ?", "temp.setting").First(&deleted)
		assert.Error(t, result.Error)
	})

	t.Run("ListConfigurations", func(t *testing.T) {
		var configs []models.Configuration
		result := db.Find(&configs)
		assert.NoError(t, result.Error)
		assert.True(t, len(configs) >= 1)
	})

	t.Run("GetConfigurationByVault", func(t *testing.T) {
		config := &models.Configuration{
			BaseModel: models.BaseModel{ID: uuid.New().String()},
			Vault:     "custom",
			Key:       "custom.key",
			Value:     "custom value",
			Version:   1,
		}
		db.Create(config)

		var configs []models.Configuration
		result := db.Where("vault = ?", "custom").Find(&configs)
		assert.NoError(t, result.Error)
		assert.True(t, len(configs) >= 1)
	})

	t.Run("VersionIncrement", func(t *testing.T) {
		config := &models.Configuration{
			BaseModel: models.BaseModel{ID: uuid.New().String()},
			Vault:     "default",
			Key:       "versioned.key",
			Value:     "v1",
			Version:   1,
		}
		db.Create(config)

		config.Value = "v2"
		config.Version++
		db.Save(config)

		var retrieved models.Configuration
		db.Where("key = ?", "versioned.key").First(&retrieved)
		assert.Equal(t, 2, retrieved.Version)
		assert.Equal(t, "v2", retrieved.Value)
	})

	_ = ctx
	_ = store
}
