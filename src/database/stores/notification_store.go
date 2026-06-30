package stores

import (
	"github.com/Parallels/prl-devops-service/data/models"
	"context"
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
	notificationDataStoreInstance *NotificationDataStore
	notificationDataStoreOnce     sync.Once
)

type NotificationDataStoreInterface interface {
	interfaces.Store
	CreateNotification(ctx basecontext.BaseContext, tenantID string, notification *models.Notification) (*models.Notification, *apperrors.Diagnostics)

	GetNotifications(ctx basecontext.BaseContext, tenantID string, userID string, filterObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.Notification], *apperrors.Diagnostics)
	MarkAsRead(ctx basecontext.BaseContext, tenantID string, notificationID string, userID string) *apperrors.Diagnostics
	MarkAllAsRead(ctx basecontext.BaseContext, tenantID string, userID string) *apperrors.Diagnostics
	GetUnreadCount(ctx basecontext.BaseContext, tenantID string, userID string) (int64, *apperrors.Diagnostics)

	// Channels & Configuration
	CreateChannel(ctx basecontext.BaseContext, tenantID string, channel *models.NotificationChannel) (*models.NotificationChannel, *apperrors.Diagnostics)
	GetChannelByName(ctx basecontext.BaseContext, tenantID string, name string) (*models.NotificationChannel, *apperrors.Diagnostics)
	GetChannels(ctx basecontext.BaseContext, tenantID string) ([]models.NotificationChannel, *apperrors.Diagnostics)
	AddConfigToChannel(ctx basecontext.BaseContext, tenantID string, config *models.NotificationChannelConfig) (*models.NotificationChannelConfig, *apperrors.Diagnostics)
	GetConfigsForChannel(ctx basecontext.BaseContext, tenantID string, channelName string) ([]models.NotificationChannelConfig, *apperrors.Diagnostics)
}

type NotificationDataStore struct {
	common.BaseDataStore
}

func GetNotificationDataStoreInstance() NotificationDataStoreInterface {
	if notificationDataStoreInstance == nil {
		return NewNotificationStore()
	}
	return notificationDataStoreInstance
}

func NewNotificationStore() *NotificationDataStore {
	return &NotificationDataStore{}
}

func (s *NotificationDataStore) Name() string {
	return "notification_store"
}

func (s *NotificationDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	notificationDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *NotificationDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *NotificationDataStore) IsEnabled() bool {
	return true
}

func (s *NotificationDataStore) Dependencies() []string {
	return []string{}
}

func (s *NotificationDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.Get().Get()
	logger := logging.Get(); logger.Info("Initializing notification store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get("database_migrate").GetBool() {
		logger := logging.Get(); logger.Info("Running notification migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate notification store: %v", err)
		}
		logger := logging.Get(); logger.Info("Notification migrations completed")
	}

	notificationDataStoreInstance = s
	return nil
}

// Kept for backward compatibility
func InitializeNotificationDataStore(db *gorm.DB) (NotificationDataStoreInterface, *apperrors.Diagnostics) {
	if notificationDataStoreInstance != nil {
		return notificationDataStoreInstance, nil
	}
	s := NewNotificationStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := apperrors.NewDiagnostics("initialize_notification_data_store")
		diag.AddError("failed_to_initialize_notification_store", err.Error(), "notification_data_store", nil)
		return nil, diag
	}
	return notificationDataStoreInstance, nil
}

func (s *NotificationDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&models.Notification{}); err != nil {
		return fmt.Errorf("failed to migrate notification table: %s", err.Error())
	}
	if err := s.GetDB().AutoMigrate(&models.NotificationChannel{}); err != nil {
		return fmt.Errorf("failed to migrate notification channel table: %s", err.Error())
	}
	if err := s.GetDB().AutoMigrate(&models.NotificationChannelConfig{}); err != nil {
		return fmt.Errorf("failed to migrate notification channel config table: %s", err.Error())
	}
	return nil
}

func (s *NotificationDataStore) CreateNotification(ctx basecontext.BaseContext, tenantID string, notification *models.Notification) (*models.Notification, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("create_notification")
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	if err := s.GetDB().WithContext(ctx.Context()).Create(notification).Error; err != nil {
		diag.AddError("failed_to_create_notification", fmt.Sprintf("failed to create notification: %s", common.MapError(err).Error()), "notification_data_store", nil)
		return nil, diag
	}

	return notification, diag
}

func (s *NotificationDataStore) GetNotifications(ctx basecontext.BaseContext, tenantID string, userID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.Notification], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("get_notifications")
	db := s.GetDB().WithContext(ctx.Context()).Where("tenant_id = ? AND user_id = ?", tenantID, userID)

	result, err := filters.QueryDatabase[models.Notification](db, tenantID, queryObj)
	if err != nil {
		diag.AddError("failed_to_get_notifications", fmt.Sprintf("failed to get notifications: %s", common.MapError(err).Error()), "notification_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *NotificationDataStore) MarkAsRead(ctx basecontext.BaseContext, tenantID string, notificationID string, userID string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("mark_notification_as_read")

	now := time.Now()
	updates := map[string]interface{}{
		"read":       true,
		"read_at":    now,
		"updated_at": now,
	}

	result := s.GetDB().WithContext(ctx.Context()).Model(&models.Notification{}).
		Where("tenant_id = ? AND user_id = ? AND id = ?", tenantID, userID, notificationID).
		Updates(updates)

	if result.Error != nil {
		diag.AddError("failed_to_mark_notification_as_read", fmt.Sprintf("failed to mark notification as read: %s", common.MapError(result.Error).Error()), "notification_data_store", nil)
		return diag
	}

	if result.RowsAffected == 0 {
		diag.AddError("notification_not_found", "notification not found or already read", "notification_data_store", map[string]interface{}{
			"id": notificationID,
		})
	}

	return diag
}

func (s *NotificationDataStore) MarkAllAsRead(ctx basecontext.BaseContext, tenantID string, userID string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("mark_all_notifications_as_read")

	now := time.Now()
	updates := map[string]interface{}{
		"read":       true,
		"read_at":    now,
		"updated_at": now,
	}

	result := s.GetDB().WithContext(ctx.Context()).Model(&models.Notification{}).
		Where("tenant_id = ? AND user_id = ? AND read = ?", tenantID, userID, false).
		Updates(updates)

	if result.Error != nil {
		diag.AddError("failed_to_mark_all_notifications_as_read", fmt.Sprintf("failed to mark all notifications as read: %s", common.MapError(result.Error).Error()), "notification_data_store", nil)
		return diag
	}

	return diag
}

func (s *NotificationDataStore) GetUnreadCount(ctx basecontext.BaseContext, tenantID string, userID string) (int64, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("get_unread_notification_count")
	var count int64

	result := s.GetDB().WithContext(ctx.Context()).Model(&models.Notification{}).
		Where("tenant_id = ? AND user_id = ? AND read = ?", tenantID, userID, false).
		Count(&count)

	if result.Error != nil {
		diag.AddError("failed_to_get_unread_count", fmt.Sprintf("failed to get unread count: %s", common.MapError(result.Error).Error()), "notification_data_store", nil)
		return 0, diag
	}

	return count, diag
}

// Channels & Configuration

func (s *NotificationDataStore) CreateChannel(ctx basecontext.BaseContext, tenantID string, channel *models.NotificationChannel) (*models.NotificationChannel, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("create_notification_channel")
	if channel.ID == "" {
		channel.ID = uuid.New().String()
	}
	channel.CreatedAt = time.Now()
	channel.UpdatedAt = time.Now()

	if err := s.GetDB().WithContext(ctx.Context()).Create(channel).Error; err != nil {
		diag.AddError("failed_to_create_notification_channel", fmt.Sprintf("failed to create notification channel: %s", common.MapError(err).Error()), "notification_data_store", nil)
		return nil, diag
	}

	return channel, diag
}

func (s *NotificationDataStore) GetChannelByName(ctx basecontext.BaseContext, tenantID string, name string) (*models.NotificationChannel, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("get_notification_channel_by_name")
	var channel models.NotificationChannel

	result := s.GetDB().WithContext(ctx.Context()).Where("tenant_id = ? AND name = ?", tenantID, name).First(&channel)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Return nil if not found, let caller handle
		}
		diag.AddError("failed_to_get_notification_channel", fmt.Sprintf("failed to get notification channel: %s", common.MapError(result.Error).Error()), "notification_data_store", nil)
		return nil, diag
	}

	return &channel, diag
}

func (s *NotificationDataStore) GetChannels(ctx basecontext.BaseContext, tenantID string) ([]models.NotificationChannel, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("get_notification_channels")
	var channels []models.NotificationChannel

	result := s.GetDB().WithContext(ctx.Context()).Where("tenant_id = ?", tenantID).Find(&channels)
	if result.Error != nil {
		diag.AddError("failed_to_get_notification_channels", fmt.Sprintf("failed to get notification channels: %s", common.MapError(result.Error).Error()), "notification_data_store", nil)
		return nil, diag
	}

	return channels, diag
}

func (s *NotificationDataStore) AddConfigToChannel(ctx basecontext.BaseContext, tenantID string, config *models.NotificationChannelConfig) (*models.NotificationChannelConfig, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("add_config_to_notification_channel")
	if config.ID == "" {
		config.ID = uuid.New().String()
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	if err := s.GetDB().WithContext(ctx.Context()).Create(config).Error; err != nil {
		diag.AddError("failed_to_add_config_to_channel", fmt.Sprintf("failed to add config to channel: %s", common.MapError(err).Error()), "notification_data_store", nil)
		return nil, diag
	}

	return config, diag
}

func (s *NotificationDataStore) GetConfigsForChannel(ctx basecontext.BaseContext, tenantID string, channelName string) ([]models.NotificationChannelConfig, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("get_configs_for_channel")
	var configs []models.NotificationChannelConfig

	// Join with Channel to filter by channel name
	result := s.GetDB().WithContext(ctx.Context()).
		Joins("JOIN notification_channels ON notification_channels.id = notification_channel_configs.channel_id").
		Where("notification_channel_configs.tenant_id = ? AND notification_channels.name = ?", tenantID, channelName).
		Find(&configs)

	if result.Error != nil {
		diag.AddError("failed_to_get_configs_for_channel", fmt.Sprintf("failed to get configs for channel: %s", common.MapError(result.Error).Error()), "notification_data_store", nil)
		return nil, diag
	}

	return configs, diag
}
