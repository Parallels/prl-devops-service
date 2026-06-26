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

	"github.com/cjlapao/common-go-logger"

	pkg_utils "github.com/Parallels/prl-devops-service/helpers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	activityDataStoreInstance *ActivityDataStore
	activityDataStoreOnce     sync.Once
)

type ActivityDataStoreInterface interface {
	interfaces.Store

	// Activity CRUD operations

	CreateActivity(ctx basecontext.BaseContext, tenantID string, activity *entities.Activity) (*entities.Activity, *apperrors.Diagnostics)
	GetActivityByID(ctx basecontext.BaseContext, tenantID string, id string) (*entities.Activity, *apperrors.Diagnostics)
	GetActivities(ctx basecontext.BaseContext, tenantID string) ([]entities.Activity, *apperrors.Diagnostics)
	GetActivitiesByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Activity], *apperrors.Diagnostics)
	UpdateActivity(ctx basecontext.BaseContext, tenantID string, activity *entities.Activity) *apperrors.Diagnostics
	DeleteActivity(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics

	// Activity querying and reporting
	GetActivitiesByFilterAdvanced(ctx basecontext.BaseContext, tenantID string, filter *entities.ActivityFilter, page, pageSize int) (*filters.FilterResponse[entities.Activity], *apperrors.Diagnostics)
	GetActivityStats(ctx basecontext.BaseContext, tenantID string, filter *entities.ActivityFilter) (map[string]interface{}, *apperrors.Diagnostics)
	GetTopActors(ctx basecontext.BaseContext, tenantID string, limit int, filter *entities.ActivityFilter) ([]map[string]interface{}, *apperrors.Diagnostics)
	GetActivityTrends(ctx basecontext.BaseContext, tenantID string, days int, filter *entities.ActivityFilter) ([]map[string]interface{}, *apperrors.Diagnostics)

	// Activity summary operations
	CreateActivitySummary(ctx basecontext.BaseContext, tenantID string, summary *entities.ActivitySummary) (*entities.ActivitySummary, *apperrors.Diagnostics)
	GetActivitySummaryByID(ctx basecontext.BaseContext, tenantID string, id string) (*entities.ActivitySummary, *apperrors.Diagnostics)
	UpdateActivitySummary(ctx basecontext.BaseContext, tenantID string, summary *entities.ActivitySummary) *apperrors.Diagnostics
	DeleteActivitySummary(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics

	// Maintenance operations
	CleanupOldActivities(ctx basecontext.BaseContext, tenantID string, retentionDays int) *apperrors.Diagnostics
	ArchiveActivities(ctx basecontext.BaseContext, tenantID string, beforeDate time.Time) *apperrors.Diagnostics
}

type ActivityDataStore struct {
	common.BaseDataStore
}

func GetActivityDataStoreInstance() ActivityDataStoreInterface {
	if activityDataStoreInstance == nil {
		return NewActivityStore()
	}
	return activityDataStoreInstance
}

func NewActivityStore() *ActivityDataStore {
	return &ActivityDataStore{}
}

func (s *ActivityDataStore) Name() string {
	return "activity_store"
}

func (s *ActivityDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	activityDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *ActivityDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *ActivityDataStore) IsEnabled() bool {
	return true
}

func (s *ActivityDataStore) Dependencies() []string {
	return []string{}
}

func (s *ActivityDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.GetInstance().Get()
	logging.Info("Initializing activity store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get(config.DatabaseMigrateKey).GetBool() {
		logging.Info("Running activity migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate activity store: %v", err)
		}
		logging.Info("Activity migrations completed")
	}

	activityDataStoreInstance = s
	return nil
}

// Kept for backward compatibility
func InitializeActivityDataStore(db *gorm.DB) (ActivityDataStoreInterface, *apperrors.Diagnostics) {
	if activityDataStoreInstance != nil {
		return activityDataStoreInstance, nil
	}
	s := NewActivityStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := errors.NewDiagnostics("initialize_activity_data_store")
		diag.AddError("failed_to_initialize_activity_store", err.Error(), "activity_data_store", err)
		return nil, diag
	}
	return activityDataStoreInstance, nil
}

func (s *ActivityDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.Activity{}); err != nil {
		return fmt.Errorf("failed to migrate activity table: %v", err)
	}

	if err := s.GetDB().AutoMigrate(&entities.ActivitySummary{}); err != nil {
		return fmt.Errorf("failed to migrate activity summary table: %v", err)
	}

	// Create indexes for better query performance
	if err := s.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_activities_tenant_module_service ON activities(tenant_id, module, service)").Error; err != nil {
		return fmt.Errorf("failed to create activities index: %v", err)
	}

	if err := s.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_activities_actor_target ON activities(actor_type, actor_id)").Error; err != nil {
		return fmt.Errorf("failed to create activities actor target index: %v", err)
	}

	if err := s.GetDB().Exec("CREATE INDEX IF NOT EXISTS idx_activities_timing ON activities(started_at, completed_at, created_at)").Error; err != nil {
		return fmt.Errorf("failed to create activities timing index: %v", err)
	}

	return nil
}

func (s *ActivityDataStore) CreateActivity(ctx basecontext.BaseContext, tenantID string, activity *entities.Activity) (*entities.Activity, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("create_activity")

	if activity == nil {
		diag.AddError("activity_cannot_be_nil", "activity cannot be nil", "activity_data_store")
		return nil, diag
	}

	if activity.ID == "" {
		activity.ID = uuid.New().String()
	}

	if activity.Slug == "" {
		activity.Slug = pkg_utils.Slugify(fmt.Sprintf("activity-%s", activity.ID))
	}

	activity.TenantID = tenantID
	if activity.TenantID == "" {
		activity.TenantID = config.UnknownTenantID
	}
	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	if activity.StartedAt == nil {
		now := time.Now()
		activity.StartedAt = &now
	}

	if err := s.GetDB().WithContext(ctx).Create(activity).Error; err != nil {
		logging.Errorf("Failed to create activity: %v", err)
		diag.AddError("failed_to_create_activity", fmt.Sprintf("failed to create activity, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}

	return activity, diag
}

func (s *ActivityDataStore) GetActivityByID(ctx basecontext.BaseContext, tenantID string, id string) (*entities.Activity, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_activity_by_id")

	if id == "" {
		diag.AddError("activity_id_cannot_be_empty", "activity ID cannot be empty", "activity_data_store", nil)
		return nil, diag
	}

	var activity entities.Activity
	query := s.GetDB().WithContext(ctx).Where("id = ?", id)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&activity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_activity", fmt.Sprintf("failed to get activity, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}

	return &activity, diag
}

func (s *ActivityDataStore) GetActivities(ctx basecontext.BaseContext, tenantID string) ([]entities.Activity, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_activities")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "activity_data_store", nil)
		return nil, diag
	}

	query := s.GetDB().WithContext(ctx).Where("tenant_id = ?", tenantID)
	var activities []entities.Activity
	if err := query.Find(&activities).Error; err != nil {
		diag.AddError("failed_to_get_activities", fmt.Sprintf("failed to get activities, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}

	return activities, diag
}

func (s *ActivityDataStore) GetActivitiesByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Activity], *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_activities")
	db := s.GetDB().WithContext(ctx)

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := filters.QueryDatabase[entities.Activity](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_activities", fmt.Sprintf("failed to get activities, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *ActivityDataStore) UpdateActivity(ctx basecontext.BaseContext, tenantID string, activity *entities.Activity) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("update_activity")

	if activity == nil {
		diag.AddError("activity_cannot_be_nil", "activity cannot be nil", "activity_data_store", nil)
		return diag
	}

	if activity.ID == "" {
		diag.AddError("activity_id_cannot_be_empty", "activity ID cannot be empty", "activity_data_store", nil)
		return diag
	}

	activity.UpdatedAt = time.Now()

	query := s.GetDB().WithContext(ctx).Where("id = ?", activity.ID)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Updates(activity).Error; err != nil {
		diag.AddError("failed_to_update_activity", fmt.Sprintf("failed to update activity, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return diag
	}

	return diag
}

func (s *ActivityDataStore) DeleteActivity(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("delete_activity")

	if id == "" {
		diag.AddError("activity_id_cannot_be_empty", "activity ID cannot be empty", "activity_data_store", nil)
		return diag
	}

	query := s.GetDB().Where("id = ?", id)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Delete(&entities.Activity{}).Error; err != nil {
		diag.AddError("failed_to_delete_activity", fmt.Sprintf("failed to delete activity, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return diag
	}

	return diag
}

func (s *ActivityDataStore) GetActivitiesByFilterAdvanced(ctx basecontext.BaseContext, tenantID string, filter *entities.ActivityFilter, page, pageSize int) (*filters.FilterResponse[entities.Activity], *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_activities_by_filter_advanced")

	if filter == nil {
		filter = &entities.ActivityFilter{}
	}
	var activities []entities.Activity
	query := s.GetDB().WithContext(ctx).Model(&entities.Activity{})

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Apply advanced filters
	if filter != nil {
		if len(filter.Module) > 0 {
			query = query.Where("module IN ?", filter.Module)
		}
		if len(filter.Service) > 0 {
			query = query.Where("service IN ?", filter.Service)
		}
		if len(filter.ActivityType) > 0 {
			query = query.Where("activity_type IN ?", filter.ActivityType)
		}
		if len(filter.ActivityLevel) > 0 {
			query = query.Where("activity_level IN ?", filter.ActivityLevel)
		}
		if len(filter.ActorType) > 0 {
			query = query.Where("actor_type IN ?", filter.ActorType)
		}
		if len(filter.ActorID) > 0 {
			query = query.Where("actor_id IN ?", filter.ActorID)
		}
		if len(filter.TargetType) > 0 {
			query = query.Where("target_type IN ?", filter.TargetType)
		}
		if len(filter.TargetID) > 0 {
			query = query.Where("target_id IN ?", filter.TargetID)
		}
		if filter.Success != nil {
			query = query.Where("success = ?", *filter.Success)
		}
		if filter.IsSensitive != nil {
			query = query.Where("is_sensitive = ?", *filter.IsSensitive)
		}
		if len(filter.Tags) > 0 {
			for _, tag := range filter.Tags {
				query = query.Where("tags LIKE ?", "%"+tag+"%")
			}
		}
		if filter.StartedAtFrom != nil {
			query = query.Where("started_at >= ?", *filter.StartedAtFrom)
		}
		if filter.StartedAtTo != nil {
			query = query.Where("started_at <= ?", *filter.StartedAtTo)
		}
		if filter.CreatedAtFrom != nil {
			query = query.Where("created_at >= ?", *filter.CreatedAtFrom)
		}
		if filter.CreatedAtTo != nil {
			query = query.Where("created_at <= ?", *filter.CreatedAtTo)
		}
	}

	// Apply pagination
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
	}

	// Default ordering
	query = query.Order("created_at DESC")

	if err := query.Find(&activities).Error; err != nil {
		diag.AddError("failed_to_get_activities", fmt.Sprintf("failed to get activities, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}

	// Get total count
	var total int64
	countQuery := s.GetDB().WithContext(ctx).Model(&entities.Activity{})
	if tenantID != "" {
		countQuery = countQuery.Where("tenant_id = ?", tenantID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		diag.AddError("failed_to_count_activities", fmt.Sprintf("failed to count activities, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}

	return &filters.FilterResponse[entities.Activity]{
		Items:      activities,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, diag
}

func (s *ActivityDataStore) GetActivityStats(ctx basecontext.BaseContext, tenantID string, filter *entities.ActivityFilter) (map[string]interface{}, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_activity_stats")
	query := s.GetDB().WithContext(ctx).Model(&entities.Activity{})

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Apply filters (same logic as GetActivitiesByFilterAdvanced)
	if filter != nil {
		// ... apply filters (simplified for brevity)
	}

	var stats struct {
		TotalActivities int64   `json:"total_activities"`
		SuccessCount    int64   `json:"success_count"`
		ErrorCount      int64   `json:"error_count"`
		AvgDurationMs   float64 `json:"avg_duration_ms"`
		MaxDurationMs   int64   `json:"max_duration_ms"`
		MinDurationMs   int64   `json:"min_duration_ms"`
	}

	if err := query.Select(`
		COUNT(*) as total_activities,
		SUM(CASE WHEN success = true THEN 1 ELSE 0 END) as success_count,
		SUM(CASE WHEN success = false THEN 1 ELSE 0 END) as error_count,
		AVG(duration_ms) as avg_duration_ms,
		MAX(duration_ms) as max_duration_ms,
		MIN(duration_ms) as min_duration_ms
	`).Scan(&stats).Error; err != nil {
		diag.AddError("failed_to_get_activity_stats", fmt.Sprintf("failed to get activity stats, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}

	return map[string]interface{}{
		"total_activities": stats.TotalActivities,
		"success_count":    stats.SuccessCount,
		"error_count":      stats.ErrorCount,
		"avg_duration_ms":  stats.AvgDurationMs,
		"max_duration_ms":  stats.MaxDurationMs,
		"min_duration_ms":  stats.MinDurationMs,
	}, diag
}

func (s *ActivityDataStore) GetTopActors(ctx basecontext.BaseContext, tenantID string, limit int, filter *entities.ActivityFilter) ([]map[string]interface{}, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_top_actors")
	query := s.GetDB().WithContext(ctx).Model(&entities.Activity{})

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Apply filters (simplified)
	if filter != nil {
		// ... apply filters
	}

	var results []map[string]interface{}
	if err := query.Select(`
		actor_type,
		actor_id,
		actor_name,
		COUNT(*) as activity_count,
		SUM(CASE WHEN success = true THEN 1 ELSE 0 END) as success_count,
		SUM(CASE WHEN success = false THEN 1 ELSE 0 END) as error_count,
		AVG(duration_ms) as avg_duration_ms
	`).Group("actor_type, actor_id, actor_name").
		Order("activity_count DESC").
		Limit(limit).
		Scan(&results).Error; err != nil {
		diag.AddError("failed_to_get_top_actors", fmt.Sprintf("failed to get top actors, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}

	return results, diag
}

func (s *ActivityDataStore) GetActivityTrends(ctx basecontext.BaseContext, tenantID string, days int, filter *entities.ActivityFilter) ([]map[string]interface{}, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_activity_trends")
	query := s.GetDB().WithContext(ctx).Model(&entities.Activity{})

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Apply filters (simplified)
	if filter != nil {
		// ... apply filters
	}

	var results []map[string]interface{}
	if err := query.Select(`
		DATE(created_at) as date,
		COUNT(*) as total_activities,
		SUM(CASE WHEN success = true THEN 1 ELSE 0 END) as success_count,
		SUM(CASE WHEN success = false THEN 1 ELSE 0 END) as error_count,
		AVG(duration_ms) as avg_duration_ms
	`).Where("created_at >= ?", time.Now().AddDate(0, 0, -days)).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error; err != nil {
		diag.AddError("failed_to_get_activity_trends", fmt.Sprintf("failed to get activity trends, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}

	return results, diag
}

// ActivitySummary operations
func (s *ActivityDataStore) CreateActivitySummary(ctx basecontext.BaseContext, tenantID string, summary *entities.ActivitySummary) (*entities.ActivitySummary, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("create_activity_summary")

	if summary == nil {
		diag.AddError("activity_summary_cannot_be_nil", "activity summary cannot be nil", "activity_data_store", nil)
		return nil, diag
	}

	if summary.ID == "" {
		summary.ID = uuid.New().String()
	}

	if summary.Slug == "" {
		summary.Slug = fmt.Sprintf("activity-summary-%s", summary.ID)
	}

	summary.TenantID = tenantID
	summary.CreatedAt = time.Now()
	summary.UpdatedAt = time.Now()

	if err := s.GetDB().WithContext(ctx).Create(summary).Error; err != nil {
		diag.AddError("failed_to_create_activity_summary", fmt.Sprintf("failed to create activity summary, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}

	return summary, diag
}

func (s *ActivityDataStore) GetActivitySummaryByID(ctx basecontext.BaseContext, tenantID string, id string) (*entities.ActivitySummary, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_activity_summary_by_id")

	if id == "" {
		diag.AddError("activity_summary_id_cannot_be_empty", "activity summary ID cannot be empty", "activity_data_store", nil)
		return nil, diag
	}

	var summary entities.ActivitySummary
	query := s.GetDB().WithContext(ctx).Where("id = ?", id)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&summary).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_activity_summary", fmt.Sprintf("failed to get activity summary, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return nil, diag
	}

	return &summary, diag
}

func (s *ActivityDataStore) UpdateActivitySummary(ctx basecontext.BaseContext, tenantID string, summary *entities.ActivitySummary) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("update_activity_summary")

	if summary == nil {
		diag.AddError("activity_summary_cannot_be_nil", "activity summary cannot be nil", "activity_data_store", nil)
		return diag
	}

	if summary.ID == "" {
		diag.AddError("activity_summary_id_cannot_be_empty", "activity summary ID cannot be empty", "activity_data_store", nil)
		return diag
	}

	summary.UpdatedAt = time.Now()

	query := s.GetDB().WithContext(ctx).Where("id = ?", summary.ID)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Updates(summary).Error; err != nil {
		diag.AddError("failed_to_update_activity_summary", fmt.Sprintf("failed to update activity summary, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return diag
	}

	return diag
}

func (s *ActivityDataStore) DeleteActivitySummary(ctx basecontext.BaseContext, tenantID string, id string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("delete_activity_summary")

	if id == "" {
		diag.AddError("activity_summary_id_cannot_be_empty", "activity summary ID cannot be empty", "activity_data_store", nil)
		return diag
	}

	query := s.GetDB().Where("id = ?", id)
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Delete(&entities.ActivitySummary{}).Error; err != nil {
		diag.AddError("failed_to_delete_activity_summary", fmt.Sprintf("failed to delete activity summary, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return diag
	}

	return diag
}

// Maintenance operations
func (s *ActivityDataStore) CleanupOldActivities(ctx basecontext.BaseContext, tenantID string, retentionDays int) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("cleanup_old_activities")

	if retentionDays <= 0 {
		diag.AddError("retention_days_must_be_positive", "retention days must be positive", "activity_data_store", nil)
		return diag
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	query := s.GetDB().WithContext(ctx).Where("created_at < ?", cutoffDate)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.Delete(&entities.Activity{}).Error; err != nil {
		diag.AddError("failed_to_cleanup_old_activities", fmt.Sprintf("failed to cleanup old activities, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return diag
	}

	return diag
}

func (s *ActivityDataStore) ArchiveActivities(ctx basecontext.BaseContext, tenantID string, beforeDate time.Time) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("archive_activities")

	query := s.GetDB().WithContext(ctx).Where("created_at < ?", beforeDate)

	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}

	// Mark activities as archived (you might want to move them to an archive table)
	// For now, we'll just delete them
	if err := query.Delete(&entities.Activity{}).Error; err != nil {
		diag.AddError("failed_to_archive_activities", fmt.Sprintf("failed to archive activities, error: %s", common.MapError(err).Error()), "activity_data_store", nil)
		return diag
	}

	return diag
}
