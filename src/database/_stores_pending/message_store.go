package stores

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/database/interfaces"
	"github.com/cjlapao/common-go-logger"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

var (
	messageDataStoreInstance *MessageDataStore
	messageDataStoreOnce     sync.Once
)

type MessageDataStoreInterface interface {
	interfaces.Store
	CreateMessage(ctx context.Context, message *entities.Message) error

	GetPendingMessages(ctx context.Context, limit int) ([]*entities.Message, error)
	GetScheduledMessages(ctx context.Context) ([]*entities.Message, error)
	UpdateMessageStatus(ctx context.Context, messageID string, status entities.MessageStatus, errorMsg string) error
	MarkMessageProcessing(ctx context.Context, messageID string) error
	CompleteMessage(ctx context.Context, messageID string) error
	FailMessage(ctx context.Context, messageID string, errorMsg string) error
	DeleteMessage(ctx context.Context, messageID string) error
	GetMessageStats(ctx context.Context) (*entities.MessageStats, error)
	RecoverOrphanedMessages(ctx context.Context, maxProcessingAge time.Duration) (int, error)
	GetStuckRetryingMessages(ctx context.Context, maxRetryAge time.Duration) ([]*entities.Message, error)
	ResetStuckRetryingMessages(ctx context.Context, maxRetryAge time.Duration) (int, error)
	CleanupOldMessages(ctx context.Context, maxAge time.Duration) (int, error)
	CleanupOldAbandonedMessages(ctx context.Context, maxAge time.Duration) (int, error)
	CleanupOldEvents(ctx context.Context, maxAge time.Duration) (int, error)
	CreateMessageEvent(ctx context.Context, event *entities.MessageEvent) error
	CreateWorker(ctx context.Context, worker *entities.Worker) error
	DeleteWorker(ctx context.Context, workerName string) error
	GetWorkerByName(ctx context.Context, workerName string) (*entities.Worker, error)
	GetAllWorkers(ctx context.Context) ([]*entities.Worker, error)
	UpdateWorkerStatus(ctx context.Context, workerName string, isRunning bool) error
	PerformStartupRecovery(ctx context.Context) error
}

type MessageDataStore struct {
	common.BaseDataStore
}

func GetMessageDataStoreInstance() MessageDataStoreInterface {
	if messageDataStoreInstance == nil {
		return NewMessageStore()
	}
	return messageDataStoreInstance
}

func NewMessageStore() *MessageDataStore {
	return &MessageDataStore{}
}

func (s *MessageDataStore) Name() string {
	return "message_store"
}

func (s *MessageDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	messageDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *MessageDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *MessageDataStore) IsEnabled() bool {
	return true
}

func (s *MessageDataStore) Dependencies() []string {
	return []string{}
}

func (s *MessageDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.GetInstance().Get()
	logging.Info("Initializing message store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get(config.DatabaseMigrateKey).GetBool() {
		logging.Info("Running message migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to run message migrations: %v", err)
		}
		logging.Info("Message migrations completed")
	}

	messageDataStoreInstance = s
	return nil
}

// Kept for backward compatibility
func InitializeMessageDataStore(db *gorm.DB) (MessageDataStoreInterface, *apperrors.Diagnostics) {
	if messageDataStoreInstance != nil {
		return messageDataStoreInstance, nil
	}
	s := NewMessageStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := errors.NewDiagnostics("initialize_message_data_store")
		diag.AddError("failed_to_initialize_message_store", err.Error(), "message_data_store", err)
		return nil, diag
	}
	return messageDataStoreInstance, nil
}

// Migrate implements the DataStore interface
func (s *MessageDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.Message{}); err != nil {
		return fmt.Errorf("failed to migrate message table: %w", err)
	}
	if err := s.GetDB().AutoMigrate(&entities.MessageEvent{}); err != nil {
		return fmt.Errorf("failed to migrate message event table: %w", err)
	}
	if err := s.GetDB().AutoMigrate(&entities.Worker{}); err != nil {
		return fmt.Errorf("failed to migrate worker table: %w", err)
	}
	return nil
}

// CreateMessage creates a new message in the queue
func (s *MessageDataStore) CreateMessage(ctx context.Context, message *entities.Message) error {
	message.ID = uuid.New().String()
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	if err := s.GetDB().WithContext(ctx).Create(message).Error; err != nil {
		return fmt.Errorf("failed to create message: %w", common.MapError(err))
	}
	return nil
}

// GetPendingMessages retrieves pending messages, ordered by priority and creation time
func (s *MessageDataStore) GetPendingMessages(ctx context.Context, limit int) ([]*entities.Message, error) {
	var messages []*entities.Message

	query := s.GetDB().WithContext(ctx).
		Where("status = ? AND (scheduled_at IS NULL OR scheduled_at <= ?)", entities.MessageStatusPending, time.Now()).
		Order("priority DESC, created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending messages: %w", common.MapError(err))
	}

	return messages, nil
}

// GetScheduledMessages retrieves messages that are scheduled to run now
func (s *MessageDataStore) GetScheduledMessages(ctx context.Context) ([]*entities.Message, error) {
	var messages []*entities.Message

	if err := s.GetDB().WithContext(ctx).
		Where("status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?", entities.MessageStatusPending, time.Now()).
		Order("priority DESC, scheduled_at ASC").
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get scheduled messages: %w", common.MapError(err))
	}

	return messages, nil
}

// UpdateMessageStatus updates the status of a message
func (s *MessageDataStore) UpdateMessageStatus(ctx context.Context, messageID string, status entities.MessageStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if errorMsg != "" {
		updates["error"] = errorMsg
	}

	if status == entities.MessageStatusCompleted || status == entities.MessageStatusFailed {
		updates["processed_at"] = time.Now()
	}

	if err := s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("id = ?", messageID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update message status: %w", common.MapError(err))
	}

	return nil
}

// MarkMessageProcessing marks a message as being processed
func (s *MessageDataStore) MarkMessageProcessing(ctx context.Context, messageID string) error {
	return s.UpdateMessageStatus(ctx, messageID, entities.MessageStatusProcessing, "")
}

// CompleteMessage marks a message as completed
func (s *MessageDataStore) CompleteMessage(ctx context.Context, messageID string) error {
	return s.UpdateMessageStatus(ctx, messageID, entities.MessageStatusCompleted, "")
}

// FailMessage marks a message as failed and increments retry count
func (s *MessageDataStore) FailMessage(ctx context.Context, messageID string, errorMsg string) error {
	// First, increment the retry count and check if we should retry or mark as failed
	var message entities.Message
	if err := s.GetDB().WithContext(ctx).First(&message, "id = ?", messageID).Error; err != nil {
		return fmt.Errorf("failed to get message for retry: %w", common.MapError(err))
	}

	message.RetryCount++
	message.Error = errorMsg
	message.UpdatedAt = time.Now()

	// If we've exceeded max retries, mark as failed, otherwise mark for retry
	if message.RetryCount >= message.MaxRetries {
		message.Status = entities.MessageStatusFailed
		message.ProcessedAt = &message.UpdatedAt
	} else {
		message.Status = entities.MessageStatusRetrying
		// Schedule for retry (could add exponential backoff here)
		retryAt := time.Now().Add(time.Duration(message.RetryCount) * time.Minute)
		message.ScheduledAt = &retryAt
	}

	if err := s.GetDB().WithContext(ctx).Save(&message).Error; err != nil {
		return fmt.Errorf("failed to update message for retry: %w", common.MapError(err))
	}

	return nil
}

// DeleteMessage removes a message from the queue
func (s *MessageDataStore) DeleteMessage(ctx context.Context, messageID string) error {
	if err := s.GetDB().WithContext(ctx).Delete(&entities.Message{}, "id = ?", messageID).Error; err != nil {
		return fmt.Errorf("failed to delete message: %w", common.MapError(err))
	}
	return nil
}

// GetMessageStats returns statistics about messages
func (s *MessageDataStore) GetMessageStats(ctx context.Context) (*entities.MessageStats, error) {
	stats := &entities.MessageStats{}

	// Count pending messages
	if err := s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ?", entities.MessageStatusPending).
		Count(&stats.TotalPending).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending messages: %w", common.MapError(err))
	}

	// Count processing messages
	if err := s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ?", entities.MessageStatusProcessing).
		Count(&stats.TotalProcessing).Error; err != nil {
		return nil, fmt.Errorf("failed to count processing messages: %w", common.MapError(err))
	}

	// Count completed messages
	if err := s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ?", entities.MessageStatusCompleted).
		Count(&stats.TotalCompleted).Error; err != nil {
		return nil, fmt.Errorf("failed to count completed messages: %w", common.MapError(err))
	}

	// Count failed messages
	if err := s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ?", entities.MessageStatusFailed).
		Count(&stats.TotalFailed).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed messages: %w", common.MapError(err))
	}

	// Count retrying messages
	if err := s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ?", entities.MessageStatusRetrying).
		Count(&stats.TotalRetrying).Error; err != nil {
		return nil, fmt.Errorf("failed to count retrying messages: %w", common.MapError(err))
	}

	// Count abandoned messages
	if err := s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ?", entities.MessageStatusAbandoned).
		Count(&stats.TotalAbandoned).Error; err != nil {
		return nil, fmt.Errorf("failed to count abandoned messages: %w", common.MapError(err))
	}

	return stats, nil
}

// RecoverOrphanedMessages finds messages that were stuck in processing state and resets them
func (s *MessageDataStore) RecoverOrphanedMessages(ctx context.Context, maxProcessingAge time.Duration) (int, error) {
	// Find messages that have been in processing state for too long
	cutoffTime := time.Now().Add(-maxProcessingAge)

	var orphanedMessages []*entities.Message
	if err := s.GetDB().WithContext(ctx).
		Where("status = ? AND updated_at < ?", entities.MessageStatusProcessing, cutoffTime).
		Find(&orphanedMessages).Error; err != nil {
		return 0, fmt.Errorf("failed to find orphaned messages: %w", common.MapError(err))
	}

	if len(orphanedMessages) == 0 {
		return 0, nil
	}

	// Reset orphaned messages to pending status
	result := s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ? AND updated_at < ?", entities.MessageStatusProcessing, cutoffTime).
		Updates(map[string]interface{}{
			"status":     entities.MessageStatusPending,
			"updated_at": time.Now(),
			"error":      "Recovered from orphaned processing state",
		})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to recover orphaned messages: %w", common.MapError(result.Error))
	}

	return int(result.RowsAffected), nil
}

// GetStuckRetryingMessages finds messages that are stuck in retrying state and resets them to pending
func (s *MessageDataStore) GetStuckRetryingMessages(ctx context.Context, maxRetryAge time.Duration) ([]*entities.Message, error) {
	cutoffTime := time.Now().Add(-maxRetryAge)

	var messages []*entities.Message
	if err := s.GetDB().WithContext(ctx).
		Where("status = ? AND (scheduled_at IS NULL OR scheduled_at < ?)", entities.MessageStatusRetrying, cutoffTime).
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to find stuck retrying messages: %w", common.MapError(err))
	}

	return messages, nil
}

// ResetStuckRetryingMessages resets messages that are stuck in retrying state back to pending
func (s *MessageDataStore) ResetStuckRetryingMessages(ctx context.Context, maxRetryAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxRetryAge)

	result := s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ? AND (scheduled_at IS NULL OR scheduled_at < ?)", entities.MessageStatusRetrying, cutoffTime).
		Updates(map[string]interface{}{
			"status":       entities.MessageStatusPending,
			"updated_at":   time.Now(),
			"scheduled_at": nil,
		})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to reset stuck retrying messages: %w", common.MapError(result.Error))
	}

	return int(result.RowsAffected), nil
}

// CleanupOldMessages removes old completed/failed messages to prevent database bloat
func (s *MessageDataStore) CleanupOldMessages(ctx context.Context, maxAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxAge)

	result := s.GetDB().WithContext(ctx).
		Where("(status = ? OR status = ?) AND processed_at < ?", entities.MessageStatusCompleted, entities.MessageStatusFailed, cutoffTime).
		Delete(&entities.Message{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old messages: %w", common.MapError(result.Error))
	}

	return int(result.RowsAffected), nil
}

// CleanupOldAbandonedMessages removes old abandoned messages
func (s *MessageDataStore) CleanupOldAbandonedMessages(ctx context.Context, maxAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxAge)

	result := s.GetDB().WithContext(ctx).
		Where("status = ? AND updated_at < ?", entities.MessageStatusAbandoned, cutoffTime).
		Delete(&entities.Message{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old abandoned messages: %w", common.MapError(result.Error))
	}

	return int(result.RowsAffected), nil
}

// CleanupOldEvents removes old message events
func (s *MessageDataStore) CleanupOldEvents(ctx context.Context, maxAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxAge)

	result := s.GetDB().WithContext(ctx).
		Where("timestamp < ?", cutoffTime).
		Delete(&entities.MessageEvent{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old events: %w", common.MapError(result.Error))
	}

	return int(result.RowsAffected), nil
}

// CreateMessageEvent creates a new message event
func (s *MessageDataStore) CreateMessageEvent(ctx context.Context, event *entities.MessageEvent) error {
	event.Timestamp = time.Now()
	if err := s.GetDB().WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("failed to create message event: %w", common.MapError(err))
	}
	return nil
}

// CreateWorker creates a new worker record
func (s *MessageDataStore) CreateWorker(ctx context.Context, worker *entities.Worker) error {
	worker.CreatedAt = time.Now()
	worker.UpdatedAt = time.Now()
	if err := s.GetDB().WithContext(ctx).Create(worker).Error; err != nil {
		return fmt.Errorf("failed to create worker: %w", common.MapError(err))
	}
	return nil
}

// DeleteWorker deletes a worker record
func (s *MessageDataStore) DeleteWorker(ctx context.Context, workerName string) error {
	if err := s.GetDB().WithContext(ctx).Where("name = ?", workerName).Delete(&entities.Worker{}).Error; err != nil {
		return fmt.Errorf("failed to delete worker: %w", common.MapError(err))
	}
	return nil
}

// GetWorkerByName retrieves a worker by name
func (s *MessageDataStore) GetWorkerByName(ctx context.Context, workerName string) (*entities.Worker, error) {
	var worker entities.Worker
	if err := s.GetDB().WithContext(ctx).Where("name = ?", workerName).First(&worker).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get worker: %w", common.MapError(err))
	}
	return &worker, nil
}

// GetAllWorkers retrieves all workers
func (s *MessageDataStore) GetAllWorkers(ctx context.Context) ([]*entities.Worker, error) {
	var workers []*entities.Worker
	if err := s.GetDB().WithContext(ctx).Find(&workers).Error; err != nil {
		return nil, fmt.Errorf("failed to get workers: %w", common.MapError(err))
	}
	return workers, nil
}

// UpdateWorkerStatus updates worker status
func (s *MessageDataStore) UpdateWorkerStatus(ctx context.Context, workerName string, isRunning bool) error {
	if err := s.GetDB().WithContext(ctx).
		Model(&entities.Worker{}).
		Where("name = ?", workerName).
		Updates(map[string]interface{}{
			"is_running": isRunning,
			"updated_at": time.Now(),
			"last_seen":  time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update worker status: %w", common.MapError(err))
	}
	return nil
}

// PerformStartupRecovery performs startup recovery operations
func (s *MessageDataStore) PerformStartupRecovery(ctx context.Context) error {
	// Reset processing messages to pending
	err := s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ?", entities.MessageStatusProcessing).
		Updates(map[string]interface{}{
			"status":      entities.MessageStatusPending,
			"worker_name": "",
			"updated_at":  time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to reset processing messages: %w", common.MapError(err))
	}

	// Reset retrying messages to pending if they haven't exceeded max retries
	err = s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ? AND retry_count < max_retries", entities.MessageStatusRetrying).
		Updates(map[string]interface{}{
			"status":     entities.MessageStatusPending,
			"updated_at": time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to reset retrying messages: %w", common.MapError(err))
	}

	// Abandon messages that have exceeded max retries
	err = s.GetDB().WithContext(ctx).
		Model(&entities.Message{}).
		Where("status = ? AND retry_count >= max_retries", entities.MessageStatusRetrying).
		Updates(map[string]interface{}{
			"status":     entities.MessageStatusAbandoned,
			"updated_at": time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to abandon exceeded retry messages: %w", common.MapError(err))
	}

	return nil
}
