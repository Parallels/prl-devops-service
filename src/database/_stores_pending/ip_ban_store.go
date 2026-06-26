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
	"github.com/Parallels/prl-devops-service/database/interfaces"
	"github.com/cjlapao/common-go-logger"


	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ipBanDataStoreInstance *IpBanDataStore
	ipBanDataStoreOnce     sync.Once
)

type IpBanStoreInterface interface {
	interfaces.Store
	GetIpBan(ctx basecontext.BaseContext, ip string) (*entities.IpBan, *apperrors.Diagnostics)

	CreateIpBan(ctx basecontext.BaseContext, ipBan *entities.IpBan) (*entities.IpBan, *apperrors.Diagnostics)
	RevokeIpBan(ctx basecontext.BaseContext, ip string) *apperrors.Diagnostics
	GetActiveBans(ctx basecontext.BaseContext) ([]entities.IpBan, *apperrors.Diagnostics)
	GetDB() *gorm.DB
}

type IpBanDataStore struct {
	common.BaseDataStore
}

func GetIpBanDataStoreInstance() IpBanStoreInterface {
	if ipBanDataStoreInstance == nil {
		return NewIpBanStore()
	}
	return ipBanDataStoreInstance
}

func NewIpBanStore() *IpBanDataStore {
	return &IpBanDataStore{}
}

func (s *IpBanDataStore) Name() string {
	return "ip_ban_store"
}

func (s *IpBanDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	ipBanDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *IpBanDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *IpBanDataStore) IsEnabled() bool {
	return true
}

func (s *IpBanDataStore) Dependencies() []string {
	return []string{}
}

func (s *IpBanDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.GetInstance().Get()
	logging.Info("Initializing ip ban store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get(config.DatabaseMigrateKey).GetBool() {
		logging.Info("Running ip ban migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate ip ban store: %v", err)
		}
		logging.Info("Ip ban migrations completed")
	}

	ipBanDataStoreInstance = s
	return nil
}

// Kept for backward compatibility
func InitializeIpBanDataStore(db *gorm.DB) (IpBanStoreInterface, *apperrors.Diagnostics) {
	if ipBanDataStoreInstance != nil {
		return ipBanDataStoreInstance, nil
	}
	s := NewIpBanStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := errors.NewDiagnostics("initialize_ip_ban_data_store")
		diag.AddError("failed_to_initialize_ip_ban_store", err.Error(), "ip_ban_data_store", err)
		return nil, diag
	}
	return ipBanDataStoreInstance, nil
}

func (s *IpBanDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.IpBan{}); err != nil {
		return fmt.Errorf("failed to migrate ip_bans table: %v", err)
	}
	return nil
}

func (s *IpBanDataStore) GetIpBan(ctx basecontext.BaseContext, ip string) (*entities.IpBan, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_ip_ban")
	var ipBan entities.IpBan

	// Check for active ban: Enabled AND (No Expiry OR Expiry in Future)
	err := s.GetDB().WithContext(ctx).
		Where("ip = ? AND enabled = ?", ip, true).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		First(&ipBan).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_ip_ban", fmt.Sprintf("failed to get ip ban: %s", common.MapError(err).Error()), "ip_ban_data_store", nil)
		return nil, diag
	}
	return &ipBan, diag
}

func (s *IpBanDataStore) CreateIpBan(ctx basecontext.BaseContext, ipBan *entities.IpBan) (*entities.IpBan, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("create_ip_ban")
	if ipBan.ID == "" {
		ipBan.ID = uuid.New().String()
	}
	ipBan.CreatedAt = time.Now()
	ipBan.UpdatedAt = time.Now()
	ipBan.BannedAt = time.Now()
	ipBan.Enabled = true

	// Upsert: If IP exists, update it
	// We use Save/Create with OnConflict depending on GORM version or DB, but simpler to just query first or use Clauses
	// Here we'll try to find existing first to update or create new
	// Actually, easier to use Clauses.OnConflict for Postgres if we were sure of DB type, but let's do find-update manually for safety across DBs if needed, or standard GORM upsert.
	// Since we have UniqueIndex on IP, we can use that.

	var existing entities.IpBan
	err := s.GetDB().WithContext(ctx).Where("ip = ?", ipBan.IP).First(&existing).Error
	if err == nil {
		// Update existing
		existing.Reason = ipBan.Reason
		existing.BannedAt = ipBan.BannedAt
		existing.ExpiresAt = ipBan.ExpiresAt
		existing.Enabled = true
		existing.UpdatedAt = time.Now()
		existing.TenantID = ipBan.TenantID
		existing.BanLevel = ipBan.BanLevel

		if err := s.GetDB().WithContext(ctx).Save(&existing).Error; err != nil {
			diag.AddError("failed_to_update_ip_ban", fmt.Sprintf("failed to update ip ban: %s", common.MapError(err).Error()), "ip_ban_data_store", nil)
			return nil, diag
		}
		return &existing, diag
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		diag.AddError("failed_to_check_existing_ip_ban", fmt.Sprintf("failed to check existing ip ban: %s", common.MapError(err).Error()), "ip_ban_data_store", nil)
		return nil, diag
	}

	if err := s.GetDB().WithContext(ctx).Create(ipBan).Error; err != nil {
		diag.AddError("failed_to_create_ip_ban", fmt.Sprintf("failed to create ip ban: %s", common.MapError(err).Error()), "ip_ban_data_store", nil)
		return nil, diag
	}

	return ipBan, diag
}

func (s *IpBanDataStore) RevokeIpBan(ctx basecontext.BaseContext, ip string) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("revoke_ip_ban")

	// Soft delete mechanism via Enabled flag or just set Enabled=false
	// The requirement said "way to unblock", so setting Enabled=false is good.

	err := s.GetDB().WithContext(ctx).Model(&entities.IpBan{}).
		Where("ip = ?", ip).
		Updates(map[string]interface{}{
			"enabled":    false,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		diag.AddError("failed_to_revoke_ip_ban", fmt.Sprintf("failed to revoke ip ban: %s", common.MapError(err).Error()), "ip_ban_data_store", nil)
		return diag
	}
	return diag
}

func (s *IpBanDataStore) GetActiveBans(ctx basecontext.BaseContext) ([]entities.IpBan, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_active_bans")
	var bans []entities.IpBan

	if s.GetDB() == nil {
		diag.AddError("db_is_nil", "database connection is nil", "ip_ban_data_store", nil)
		return nil, diag
	}

	err := s.GetDB().WithContext(ctx).
		Where("enabled = ?", true).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Find(&bans).Error

	if err != nil {
		diag.AddError("failed_to_get_active_bans", fmt.Sprintf("failed to get active bans: %s", common.MapError(err).Error()), "ip_ban_data_store", nil)
		return nil, diag
	}

	return bans, diag
}
