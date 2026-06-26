package mocks

import (
	"context"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// BaseMockStore provides a reusable mock for any store interface
// It embeds mock.Mock and provides common method implementations
type BaseMockStore struct {
	mock.Mock
}

// NewBaseMockStore creates a new BaseMockStore instance
func NewBaseMockStore() *BaseMockStore {
	return &BaseMockStore{}
}

// Name returns the name of the service
func (m *BaseMockStore) Name() string {
	args := m.Called()
	if args.Get(0) == nil {
		return "mock_store"
	}
	return args.String(0)
}

// Init initializes the service
// Init initializes the service
func (m *BaseMockStore) Init(ctx context.Context, db *gorm.DB) error {
	args := m.Called(ctx, db)
	return args.Error(0)
}

// Health checks the service health
func (m *BaseMockStore) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// IsEnabled checks if the service is enabled
func (m *BaseMockStore) IsEnabled() bool {
	args := m.Called()
	if len(args) == 0 {
		return true
	}
	return args.Bool(0)
}

// Dependencies returns the service dependencies
func (m *BaseMockStore) Dependencies() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return []string{}
	}
	return args.Get(0).([]string)
}

// ============================================================================
// Tenant Store Methods
// ============================================================================

func (m *BaseMockStore) GetTenantBySlug(ctx basecontext.BaseContext, slug string) (*entities.Tenant, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *BaseMockStore) GetTenantByID(ctx basecontext.BaseContext, id string) (*entities.Tenant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *BaseMockStore) GetTenantByIDOrSlug(ctx basecontext.BaseContext, idOrSlug string) (*entities.Tenant, error) {
	args := m.Called(ctx, idOrSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *BaseMockStore) GetTenants(ctx basecontext.BaseContext) ([]entities.Tenant, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Tenant), args.Error(1)
}

func (m *BaseMockStore) GetTenantsByQuery(ctx basecontext.BaseContext, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Tenant], *errors.Diagnostics) {
	args := m.Called(ctx, queryBuilder)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Tenant]), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) CreateTenant(ctx basecontext.BaseContext, tenant *entities.Tenant) (*entities.Tenant, error) {
	args := m.Called(ctx, tenant)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *BaseMockStore) UpdateTenant(ctx basecontext.BaseContext, tenant *entities.Tenant) error {
	args := m.Called(ctx, tenant)
	return args.Error(0)
}

func (m *BaseMockStore) DeleteTenant(ctx basecontext.BaseContext, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *BaseMockStore) Migrate() error {
	args := m.Called()
	return args.Error(0)
}

// ============================================================================
// User Store Methods
// ============================================================================

func (m *BaseMockStore) GetUsersByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, queryBuilder)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.User]), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) CreateUser(ctx basecontext.BaseContext, tenantID string, user *entities.User) (*entities.User, error) {
	args := m.Called(ctx, tenantID, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *BaseMockStore) GetUserByID(ctx basecontext.BaseContext, tenantID string, id string) (*entities.User, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *BaseMockStore) GetUserByUsername(ctx basecontext.BaseContext, tenantID string, username string) (*entities.User, error) {
	args := m.Called(ctx, tenantID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *BaseMockStore) UpdateUser(ctx basecontext.BaseContext, tenantID string, user *entities.User) error {
	args := m.Called(ctx, tenantID, user)
	return args.Error(0)
}

func (m *BaseMockStore) UpdateUserPassword(ctx basecontext.BaseContext, tenantID string, id string, password string) error {
	args := m.Called(ctx, tenantID, id, password)
	return args.Error(0)
}

func (m *BaseMockStore) BlockUser(ctx basecontext.BaseContext, tenantID string, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *BaseMockStore) SetRefreshToken(ctx basecontext.BaseContext, tenantID string, id string, refreshToken string) error {
	args := m.Called(ctx, tenantID, id, refreshToken)
	return args.Error(0)
}

func (m *BaseMockStore) DeleteUser(ctx basecontext.BaseContext, tenantID string, id string) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *BaseMockStore) GetRolesByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, queryBuilder)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Role]), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) GetClaimsByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, queryBuilder)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Claim]), args.Get(1).(*errors.Diagnostics)
}

// ============================================================================
// Auth Store Methods
// ============================================================================

func (m *BaseMockStore) CreateAPIKey(ctx basecontext.BaseContext, apiKey *entities.ApiKey) (*entities.ApiKey, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ApiKey), args.Error(1)
}

func (m *BaseMockStore) GetAPIKeyByHash(ctx basecontext.BaseContext, keyHash string) (*entities.ApiKey, error) {
	args := m.Called(ctx, keyHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ApiKey), args.Error(1)
}

func (m *BaseMockStore) GetAPIKeyByPrefix(ctx basecontext.BaseContext, keyPrefix string) (*entities.ApiKey, error) {
	args := m.Called(ctx, keyPrefix)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ApiKey), args.Error(1)
}

func (m *BaseMockStore) GetAPIKeyByID(ctx basecontext.BaseContext, id string) (*entities.ApiKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ApiKey), args.Error(1)
}

func (m *BaseMockStore) ListAPIKeysByUserID(ctx basecontext.BaseContext, userID string) ([]entities.ApiKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.ApiKey), args.Error(1)
}

func (m *BaseMockStore) UpdateAPIKeyLastUsed(ctx basecontext.BaseContext, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *BaseMockStore) RevokeAPIKey(ctx basecontext.BaseContext, id string, revokedBy string, reason string) error {
	args := m.Called(ctx, id, revokedBy, reason)
	return args.Error(0)
}

func (m *BaseMockStore) DeleteAPIKey(ctx basecontext.BaseContext, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *BaseMockStore) CleanupExpiredAPIKeys(ctx basecontext.BaseContext) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *BaseMockStore) GetApiKeyByDigest(ctx basecontext.BaseContext, tenantID string, digest string) (*entities.ApiKey, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, digest)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.ApiKey), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) GetApiKeyByName(ctx basecontext.BaseContext, tenantID string, name string) (*entities.ApiKey, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, name)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.ApiKey), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) GetApiKeys(ctx basecontext.BaseContext, tenantID string) ([]entities.ApiKey, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.ApiKey), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) GetApiKeysByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.ApiKey], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, queryBuilder)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.ApiKey]), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) GetApiKeyByIDOrSlug(ctx basecontext.BaseContext, tenantID string, id string) (*entities.ApiKey, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.ApiKey), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) AddClaimToApiKey(ctx basecontext.BaseContext, tenantID string, id string, claimID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, id, claimID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *BaseMockStore) RemoveClaimFromApiKey(ctx basecontext.BaseContext, tenantID string, id string, claimID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, id, claimID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *BaseMockStore) UpdateApiKeyClaims(ctx basecontext.BaseContext, tenantID string, apiKey *entities.ApiKey, claims []entities.Claim) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, apiKey, claims)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *BaseMockStore) UpdateApiKeyLastUsed(ctx basecontext.BaseContext, tenantID string, apiKeyID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, apiKeyID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *BaseMockStore) CreateApiKey(ctx basecontext.BaseContext, tenantID string, apiKey *entities.ApiKey) (*entities.ApiKey, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, apiKey)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.ApiKey), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) GetApiKeyByHash(ctx basecontext.BaseContext, tenantID string, keyHash string) (*entities.ApiKey, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, keyHash)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.ApiKey), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) GetApiKeyByPrefix(ctx basecontext.BaseContext, tenantID string, keyPrefix string) (*entities.ApiKey, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, keyPrefix)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.ApiKey), args.Get(1).(*errors.Diagnostics)
}

func (m *BaseMockStore) RevokeApiKey(ctx basecontext.BaseContext, tenantID string, id string, revokedBy string, reason string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, id, revokedBy, reason)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *BaseMockStore) DeleteApiKey(ctx basecontext.BaseContext, tenantID string, id string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *BaseMockStore) GetDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

// ============================================================================
// Configuration Store Methods
// ============================================================================

func (m *BaseMockStore) GetConfigurationValue(ctx interface{}, key string, value interface{}) (interface{}, error) {
	args := m.Called(ctx, key, value)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}
