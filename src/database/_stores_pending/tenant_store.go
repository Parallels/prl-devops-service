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
	tenantDataStoreInstance *TenantDataStore
	tenantDataStoreOnce     sync.Once
)

type TenantDataStoreInterface interface {
	interfaces.Store
	GetTenantByIDOrSlug(ctx basecontext.BaseContext, idOrSlug string) (*entities.Tenant, *apperrors.Diagnostics)

	GetTenants(ctx basecontext.BaseContext) ([]entities.Tenant, *apperrors.Diagnostics)
	GetTenantsByQuery(ctx basecontext.BaseContext, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Tenant], *apperrors.Diagnostics)
	CreateTenant(ctx basecontext.BaseContext, tenant *entities.Tenant) (*entities.Tenant, *apperrors.Diagnostics)
	UpdateTenant(ctx basecontext.BaseContext, tenant *entities.Tenant) *apperrors.Diagnostics
	DeleteTenant(ctx basecontext.BaseContext, id string) *apperrors.Diagnostics
}

type TenantDataStore struct {
	common.BaseDataStore
}

func GetTenantDataStoreInstance() TenantDataStoreInterface {
	if tenantDataStoreInstance == nil {
		return NewTenantStore()
	}
	return tenantDataStoreInstance
}

func NewTenantStore() *TenantDataStore {
	return &TenantDataStore{}
}

func (s *TenantDataStore) Name() string {
	return "tenant_store"
}

func (s *TenantDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	tenantDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *TenantDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *TenantDataStore) IsEnabled() bool {
	return true
}

func (s *TenantDataStore) Dependencies() []string {
	return []string{}
}

func (s *TenantDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.GetInstance().Get()
	logging.Info("Initializing tenant store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get(config.DatabaseMigrateKey).GetBool() {
		logging.Info("Running tenant migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to run tenant migrations: %v", err)
		}
		logging.Info("Tenant migrations completed")
	}

	tenantDataStoreInstance = s
	return nil
}

// Kept for backward compatibility
func InitializeTenantDataStore(db *gorm.DB) (TenantDataStoreInterface, *apperrors.Diagnostics) {
	if tenantDataStoreInstance != nil {
		return tenantDataStoreInstance, nil
	}
	s := NewTenantStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := errors.NewDiagnostics("initialize_tenant_data_store")
		diag.AddError("failed_to_initialize_tenant_store", err.Error(), "tenant_data_store", err)
		return nil, diag
	}
	return tenantDataStoreInstance, nil
}

func (s *TenantDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.Tenant{}); err != nil {
		return fmt.Errorf("failed to migrate tenant table: %w", err)
	}

	return nil
}

func (s *TenantDataStore) GetTenantByIDOrSlug(ctx basecontext.BaseContext, idOrSlug string) (*entities.Tenant, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_tenant_by_id_or_slug")
	var tenant entities.Tenant
	if err := s.GetDB().WithContext(ctx).Where("id = ? OR slug = ?", idOrSlug, idOrSlug).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_tenant", fmt.Sprintf("failed to get tenant: %s", common.MapError(err).Error()), "tenant_data_store", nil)
		return nil, diag
	}
	return &tenant, diag
}

func (s *TenantDataStore) GetTenants(ctx basecontext.BaseContext) ([]entities.Tenant, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_tenants")
	var tenants []entities.Tenant
	if err := s.GetDB().WithContext(ctx).Find(&tenants).Error; err != nil {
		diag.AddError("failed_to_get_tenants", fmt.Sprintf("failed to get tenants: %s", common.MapError(err).Error()), "tenant_data_store", nil)
		return nil, diag
	}
	return tenants, diag
}

func (s *TenantDataStore) GetTenantsByQuery(ctx basecontext.BaseContext, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Tenant], *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_get_tenants_by_query")
	result, err := filters.QueryDatabase[entities.Tenant](s.GetDB().WithContext(ctx), "", queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_tenants", fmt.Sprintf("failed to get tenants: %s", common.MapError(err).Error()), "tenant_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *TenantDataStore) CreateTenant(ctx basecontext.BaseContext, tenant *entities.Tenant) (*entities.Tenant, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("store_create_tenant")
	if tenant.ID == "" {
		tenant.ID = uuid.New().String()
	}
	tenant.Slug = pkg_utils.Slugify(tenant.Name)
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	if err := s.GetDB().WithContext(ctx).Create(tenant).Error; err != nil {
		diag.AddError("failed_to_create_tenant", fmt.Sprintf("failed to create tenant: %s", common.MapError(err).Error()), "tenant_data_store", nil)
		return nil, diag
	}
	return tenant, diag
}

func (s *TenantDataStore) UpdateTenant(ctx basecontext.BaseContext, tenant *entities.Tenant) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("store_update_tenant")
	// Get the original tenant from the database
	originalTenant, diagErr := s.GetTenantByIDOrSlug(ctx, tenant.ID)
	if diagErr.HasErrors() {
		return diagErr
	}
	if originalTenant == nil {
		diag.AddError("tenant_not_found", "tenant not found", "tenant_data_store", nil)
		return diag
	}

	if tenant.Name != "" {
		tenant.Slug = pkg_utils.Slugify(tenant.Name)
	}

	// Generate partial update map by comparing original with updated
	updates := common.PartialUpdateMap(originalTenant, tenant, "updated_at", "slug")
	if err := s.GetDB().WithContext(ctx).Model(&entities.Tenant{}).Where("id = ?", tenant.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_tenant", fmt.Sprintf("failed to update tenant: %s", common.MapError(err).Error()), "tenant_data_store", nil)
		return diag
	}
	return diag
}

func (s *TenantDataStore) DeleteTenant(ctx basecontext.BaseContext, id string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("store_delete_tenant")
	// deleting all claims for the tenant
	if err := s.GetDB().WithContext(ctx).Delete(&entities.Claim{}, "tenant_id = ?", id).Error; err != nil {
		diag.AddError("failed_to_delete_claims", fmt.Sprintf("failed to delete claims: %s", common.MapError(err).Error()), "tenant_data_store", nil)
		return diag
	}
	// deleting all roles for the tenant
	if err := s.GetDB().WithContext(ctx).Delete(&entities.Role{}, "tenant_id = ?", id).Error; err != nil {
		diag.AddError("failed_to_delete_roles", fmt.Sprintf("failed to delete roles: %s", common.MapError(err).Error()), "tenant_data_store", nil)
		return diag
	}
	// deleting all users for the tenant
	if err := s.GetDB().WithContext(ctx).Delete(&entities.User{}, "tenant_id = ?", id).Error; err != nil {
		diag.AddError("failed_to_delete_users", fmt.Sprintf("failed to delete users: %s", common.MapError(err).Error()), "tenant_data_store", nil)
		return diag
	}

	if err := s.GetDB().WithContext(ctx).Delete(&entities.Tenant{}, "id = ?", id).Error; err != nil {
		diag.AddError("failed_to_delete_tenant", fmt.Sprintf("failed to delete tenant: %s", common.MapError(err).Error()), "tenant_data_store", nil)
		return diag
	}
	return diag
}
