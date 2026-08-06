package stores

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/basecontext"
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
	packerTemplateDataStoreInstance *PackerTemplateDataStore
	packerTemplateDataStoreOnce     sync.Once
)

type PackerTemplateDataStoreInterface interface {
	interfaces.Store
	Get(ctx basecontext.BaseContext, id string) (*models.PackerTemplate, *apperrors.Diagnostics)
	Find(ctx basecontext.BaseContext, filter *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.PackerTemplate], *apperrors.Diagnostics)
	Create(ctx basecontext.BaseContext, template *models.PackerTemplate) (*models.PackerTemplate, *apperrors.Diagnostics)
	Update(ctx basecontext.BaseContext, template *models.PackerTemplate) *apperrors.Diagnostics
	Delete(ctx basecontext.BaseContext, id string) *apperrors.Diagnostics
}

type PackerTemplateDataStore struct {
	common.BaseDataStore
}

func GetPackerTemplateDataStoreInstance() PackerTemplateDataStoreInterface {
	if packerTemplateDataStoreInstance == nil {
		return NewPackerTemplateStore()
	}
	return packerTemplateDataStoreInstance
}

func NewPackerTemplateStore() *PackerTemplateDataStore {
	return &PackerTemplateDataStore{}
}

func (s *PackerTemplateDataStore) Name() string {
	return "packer_template_store"
}

func (s *PackerTemplateDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	packerTemplateDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *PackerTemplateDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *PackerTemplateDataStore) IsEnabled() bool {
	return true
}

func (s *PackerTemplateDataStore) Dependencies() []string {
	return []string{}
}

func (s *PackerTemplateDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.Get()
	logger := logging.Get()
	logger.Info("Initializing packer template store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.IsDatabaseAutoMigrateEnabled() {
		logger.Info("Running packer template migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate packer template store: %v", err)
		}
		logger.Info("Packer template migrations completed")
	}

	packerTemplateDataStoreInstance = s
	return nil
}

func (s *PackerTemplateDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&models.PackerTemplate{}); err != nil {
		return fmt.Errorf("failed to migrate packer_templates table: %v", err)
	}

	return nil
}

// Get retrieves a packer template by ID
func (s *PackerTemplateDataStore) Get(ctx basecontext.BaseContext, id string) (*models.PackerTemplate, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_get_packer_template")

	if id == "" {
		diag.AddError("id_required", "id is required", "packer_template_store", nil)
		return nil, diag
	}

	var template models.PackerTemplate
	err := s.GetDB().WithContext(ctx.Context()).
		Where("id = ?", id).
		First(&template).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			diag.AddError("packer_template_not_found", fmt.Sprintf("packer template not found: %s", id), "packer_template_store", nil)
			return nil, diag
		}
		diag.AddError("failed_to_get_packer_template", fmt.Sprintf("failed to get packer template: %s", common.MapError(err).Error()), "packer_template_store", nil)
		return nil, diag
	}

	return &template, nil
}

// Find retrieves packer templates with filtering
func (s *PackerTemplateDataStore) Find(ctx basecontext.BaseContext, filter *filters.QueryBuilder) (*filters.QueryBuilderResponse[models.PackerTemplate], *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_find_packer_templates")

	query := s.GetDB().WithContext(ctx.Context())

	if filter != nil {
		result, err := filters.QueryDatabase[models.PackerTemplate](query, "", filter)
		if err != nil {
			diag.AddError("failed_to_apply_filter", fmt.Sprintf("failed to apply filter: %s", err.Error()), "packer_template_store", nil)
			return nil, diag
		}
		return result, nil
	}

	// No filter - return all templates
	var templates []models.PackerTemplate
	if err := query.Find(&templates).Error; err != nil {
		diag.AddError("failed_to_find_packer_templates", fmt.Sprintf("failed to find packer templates: %s", common.MapError(err).Error()), "packer_template_store", nil)
		return nil, diag
	}

	return &filters.QueryBuilderResponse[models.PackerTemplate]{
		Items: templates,
		Total: int64(len(templates)),
	}, nil
}

// Create creates a new packer template
func (s *PackerTemplateDataStore) Create(ctx basecontext.BaseContext, template *models.PackerTemplate) (*models.PackerTemplate, *apperrors.Diagnostics) {
	diag := apperrors.NewDiagnostics("store_create_packer_template")

	if template.Name == "" {
		diag.AddError("name_required", "name is required", "packer_template_store", nil)
		return nil, diag
	}

	if template.ID == "" {
		template.ID = uuid.New().String()
	}

	now := time.Now().UTC().Format(time.RFC3339)
	template.CreatedAt = now
	template.UpdatedAt = now

	if err := s.GetDB().WithContext(ctx.Context()).Create(template).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			diag.AddError("packer_template_already_exists", fmt.Sprintf("packer template with id '%s' already exists", template.ID), "packer_template_store", nil)
			return nil, diag
		}
		diag.AddError("failed_to_create_packer_template", fmt.Sprintf("failed to create packer template: %s", common.MapError(err).Error()), "packer_template_store", nil)
		return nil, diag
	}

	return template, nil
}

// Update updates an existing packer template
func (s *PackerTemplateDataStore) Update(ctx basecontext.BaseContext, template *models.PackerTemplate) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_update_packer_template")

	if template.ID == "" {
		diag.AddError("id_required", "id is required", "packer_template_store", nil)
		return diag
	}

	template.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	result := s.GetDB().WithContext(ctx.Context()).
		Where("id = ?", template.ID).
		Select("*").
		Omit("created_at").
		Updates(template)

	if result.Error != nil {
		diag.AddError("failed_to_update_packer_template", fmt.Sprintf("failed to update packer template: %s", common.MapError(result.Error).Error()), "packer_template_store", nil)
		return diag
	}

	if result.RowsAffected == 0 {
		diag.AddError("packer_template_not_found", "packer template not found", "packer_template_store", nil)
		return diag
	}

	return nil
}

// Delete removes a packer template
func (s *PackerTemplateDataStore) Delete(ctx basecontext.BaseContext, id string) *apperrors.Diagnostics {
	diag := apperrors.NewDiagnostics("store_delete_packer_template")

	if id == "" {
		diag.AddError("id_required", "id is required", "packer_template_store", nil)
		return diag
	}

	result := s.GetDB().WithContext(ctx.Context()).
		Where("id = ?", id).
		Delete(&models.PackerTemplate{})

	if result.Error != nil {
		diag.AddError("failed_to_delete_packer_template", fmt.Sprintf("failed to delete packer template: %s", common.MapError(result.Error).Error()), "packer_template_store", nil)
		return diag
	}

	if result.RowsAffected == 0 {
		diag.AddError("packer_template_not_found", "packer template not found", "packer_template_store", nil)
		return diag
	}

	return nil
}
