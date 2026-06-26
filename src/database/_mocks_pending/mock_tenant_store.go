package mocks

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/errors"
)

type MockTenantStore struct {
	BaseMockStore
}

func (m *MockTenantStore) GetTenants(ctx basecontext.BaseContext) ([]entities.Tenant, *errors.Diagnostics) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.Tenant), args.Get(1).(*errors.Diagnostics)
}

func (m *MockTenantStore) GetTenantsByQuery(ctx basecontext.BaseContext, query *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Tenant], *errors.Diagnostics) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Tenant]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockTenantStore) GetTenantByIDOrSlug(ctx basecontext.BaseContext, idOrSlug string) (*entities.Tenant, *errors.Diagnostics) {
	args := m.Called(ctx, idOrSlug)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.Tenant), args.Get(1).(*errors.Diagnostics)
}

func (m *MockTenantStore) CreateTenant(ctx basecontext.BaseContext, tenant *entities.Tenant) (*entities.Tenant, *errors.Diagnostics) {
	args := m.Called(ctx, tenant)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.Tenant), args.Get(1).(*errors.Diagnostics)
}

func (m *MockTenantStore) UpdateTenant(ctx basecontext.BaseContext, tenant *entities.Tenant) *errors.Diagnostics {
	args := m.Called(ctx, tenant)
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockTenantStore) DeleteTenant(ctx basecontext.BaseContext, id string) *errors.Diagnostics {
	args := m.Called(ctx, id)
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockTenantStore) Migrate() error {
	args := m.Called()
	return args.Error(0)
}

func NewMockTenantStore() *MockTenantStore {
	return &MockTenantStore{
		BaseMockStore: *NewBaseMockStore(),
	}
}
