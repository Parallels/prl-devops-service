package stores

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/data/models"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/interfaces"
	logging "github.com/cjlapao/common-go-logger"

	apperrors "github.com/Parallels/prl-devops-service/errors"
	"gorm.io/gorm"
)

var (
	emailDataStoreInstance *EmailDataStore
	emailDataStoreOnce     sync.Once
)

type EmailDataStoreInterface interface {
	interfaces.Store
	GetTemplateBySlug(ctx basecontext.BaseContext, slug string) (*models.EmailTemplate, *apperrors.Diagnostics)

	CreateTemplate(ctx basecontext.BaseContext, template *models.EmailTemplate) (*models.EmailTemplate, *apperrors.Diagnostics)
	GetTemplatesByTenant(ctx basecontext.BaseContext, tenantID string) ([]models.EmailTemplate, *apperrors.Diagnostics)
}

type EmailDataStore struct {
	common.BaseDataStore
}

func GetEmailDataStoreInstance() EmailDataStoreInterface {
	if emailDataStoreInstance == nil {
		return NewEmailStore()
	}
	return emailDataStoreInstance
}

func NewEmailStore() *EmailDataStore {
	return &EmailDataStore{}
}

func (s *EmailDataStore) Name() string {
	return "email_store"
}

func (s *EmailDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	emailDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *EmailDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *EmailDataStore) IsEnabled() bool {
	return true
}

func (s *EmailDataStore) Dependencies() []string {
	return []string{}
}

func (s *EmailDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.Get()
	logger := logging.Get()
	logger.Info("Initializing email store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.IsDatabaseAutoMigrateEnabled() {
		logger.Info("Running email migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate email store: %v", err)
		}
		logger.Info("Email migrations completed")
	}

	emailDataStoreInstance = s
	return nil
}

// Kept for backward compatibility
func InitializeEmailDataStore(db *gorm.DB) (EmailDataStoreInterface, *apperrors.Diagnostics) {
	if emailDataStoreInstance != nil {
		return emailDataStoreInstance, nil
	}
	s := NewEmailStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := apperrors.NewDiagnostics("initialize_email_data_store")
		diag.AddError("failed_to_initialize_email_store", err.Error(), "email_data_store", nil)
		return nil, diag
	}
	return emailDataStoreInstance, nil
}

func (s *EmailDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&models.EmailTemplate{}); err != nil {
		return fmt.Errorf("failed to migrate email template table: %s", err.Error())
	}
	return nil
}

func (s *EmailDataStore) GetTemplateBySlug(ctx basecontext.BaseContext, slug string) (*models.EmailTemplate, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("get_template_by_slug")
	var template models.EmailTemplate
	result := s.GetDB().WithContext(ctx.Context()).Where("slug = ?", slug).First(&template)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_template_by_slug", fmt.Sprintf("failed to get template by slug: %s", common.MapError(result.Error).Error()), "email_data_store", nil)
		return nil, diag
	}
	return &template, diag
}

func (s *EmailDataStore) CreateTemplate(ctx basecontext.BaseContext, template *models.EmailTemplate) (*models.EmailTemplate, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("create_template")
	if template.CreatedAt.IsZero() {
		template.CreatedAt = time.Now()
	}
	template.UpdatedAt = time.Now()

	if err := s.GetDB().WithContext(ctx.Context()).Create(template).Error; err != nil {
		diag.AddError("failed_to_create_template", fmt.Sprintf("failed to create template: %s", common.MapError(err).Error()), "email_data_store", nil)
		return nil, diag
	}
	return template, diag
}

func (s *EmailDataStore) GetTemplatesByTenant(ctx basecontext.BaseContext, tenantID string) ([]models.EmailTemplate, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("get_templates_by_tenant")
	var templates []models.EmailTemplate

	err := s.GetDB().WithContext(ctx.Context()).Where("tenant_id = ?", tenantID).Find(&templates).Error
	if err != nil {
		diag.AddError("failed_to_get_templates_by_tenant", fmt.Sprintf("failed to get templates by tenant: %s", common.MapError(err).Error()), "email_data_store", nil)
		return nil, diag
	}

	return templates, diag
}
