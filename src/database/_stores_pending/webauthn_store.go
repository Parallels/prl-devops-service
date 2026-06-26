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

	"gorm.io/gorm"
)

var (
	webAuthnDataStoreInstance *WebAuthnDataStore
	webAuthnDataStoreOnce     sync.Once
)

type WebAuthnStoreInterface interface {
	interfaces.Store
	GetCredentialsByUser(ctx basecontext.BaseContext, userID string) ([]entities.WebAuthnCredential, *apperrors.Diagnostics)

	GetCredentialByCredentialID(ctx basecontext.BaseContext, credentialID []byte) (*entities.WebAuthnCredential, *apperrors.Diagnostics)
	SaveCredential(ctx basecontext.BaseContext, cred *entities.WebAuthnCredential) *apperrors.Diagnostics
	UpdateCredential(ctx basecontext.BaseContext, cred *entities.WebAuthnCredential) *apperrors.Diagnostics
}

type WebAuthnDataStore struct {
	common.BaseDataStore
}

func GetWebAuthnDataStoreInstance() WebAuthnStoreInterface {
	if webAuthnDataStoreInstance == nil {
		return NewWebAuthnStore()
	}
	return webAuthnDataStoreInstance
}

func NewWebAuthnStore() *WebAuthnDataStore {
	return &WebAuthnDataStore{}
}

func (s *WebAuthnDataStore) Name() string {
	return "webauthn_store"
}

func (s *WebAuthnDataStore) Init(ctx context.Context, db *gorm.DB) error {
	var err error
	webAuthnDataStoreOnce.Do(func() {
		initErr := s.initialize(ctx, db)
		if initErr != nil {
			err = initErr
			return
		}
	})
	return err
}

func (s *WebAuthnDataStore) Health(ctx context.Context) error {
	return nil
}

func (s *WebAuthnDataStore) IsEnabled() bool {
	return true
}

func (s *WebAuthnDataStore) Dependencies() []string {
	return []string{}
}

func (s *WebAuthnDataStore) initialize(ctx context.Context, db *gorm.DB) error {
	cfg := config.GetInstance().Get()
	logging.Info("Initializing WebAuthn store...")

	s.BaseDataStore = *common.NewBaseDataStore(db)

	if cfg.Get(config.DatabaseMigrateKey).GetBool() {
		logging.Info("Running WebAuthn migrations")
		if err := s.Migrate(); err != nil {
			return fmt.Errorf("failed to migrate WebAuthn store: %v", err)
		}
		logging.Info("WebAuthn migrations completed")
	}

	webAuthnDataStoreInstance = s
	return nil
}

// Kept for backward compatibility
func InitializeWebAuthnDataStore(db *gorm.DB) (WebAuthnStoreInterface, *apperrors.Diagnostics) {
	if webAuthnDataStoreInstance != nil {
		return webAuthnDataStoreInstance, nil
	}
	s := NewWebAuthnStore()
	err := s.Init(context.Background(), db)
	if err != nil {
		diag := errors.NewDiagnostics("initialize_webauthn_data_store")
		diag.AddError("failed_to_initialize_webauthn_store", err.Error(), "webauthn_data_store", err)
		return nil, diag
	}
	return webAuthnDataStoreInstance, nil
}

func (s *WebAuthnDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.WebAuthnCredential{}); err != nil {
		return fmt.Errorf("failed to migrate webauthn credential table: %v", err)
	}
	return nil
}

func (s *WebAuthnDataStore) GetCredentialsByUser(ctx basecontext.BaseContext, userID string) ([]entities.WebAuthnCredential, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_credentials_by_user")
	var credentials []entities.WebAuthnCredential
	if err := s.GetDB().WithContext(ctx).Where("user_id = ?", userID).Find(&credentials).Error; err != nil {
		diag.AddError("failed_to_get_credentials", fmt.Sprintf("failed to get credentials for user %s: %s", userID, common.MapError(err).Error()), "webauthn_data_store", nil)
		return nil, diag
	}
	return credentials, nil // Return empty list if none found, diag is nil (clean)
}

func (s *WebAuthnDataStore) GetCredentialByCredentialID(ctx basecontext.BaseContext, credentialID []byte) (*entities.WebAuthnCredential, *apperrors.Diagnostics) {
	diag := errors.NewDiagnostics("get_credential_by_id")
	var cred entities.WebAuthnCredential
	result := s.GetDB().WithContext(ctx).Where("credential_id = ?", credentialID).First(&cred)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if record not found
		}
		diag.AddError("failed_to_get_credential", fmt.Sprintf("failed to get credential: %s", common.MapError(result.Error).Error()), "webauthn_data_store", nil)
		return nil, diag
	}
	return &cred, nil
}

func (s *WebAuthnDataStore) SaveCredential(ctx basecontext.BaseContext, cred *entities.WebAuthnCredential) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("save_credential")
	if cred.CreatedAt.IsZero() {
		cred.CreatedAt = time.Now()
	}
	cred.UpdatedAt = time.Now()

	if err := s.GetDB().WithContext(ctx).Create(cred).Error; err != nil {
		diag.AddError("failed_to_create_credential", fmt.Sprintf("failed to create credential: %s", common.MapError(err).Error()), "webauthn_data_store", nil)
		return diag
	}
	return nil
}

func (s *WebAuthnDataStore) UpdateCredential(ctx basecontext.BaseContext, cred *entities.WebAuthnCredential) *apperrors.Diagnostics {
	diag := errors.NewDiagnostics("update_credential")
	cred.UpdatedAt = time.Now()
	// Usually WebAuthn updates are for SignCount
	if err := s.GetDB().WithContext(ctx).Save(cred).Error; err != nil {
		diag.AddError("failed_to_update_credential", fmt.Sprintf("failed to update credential: %s", common.MapError(err).Error()), "webauthn_data_store", nil)
		return diag
	}
	return nil
}
