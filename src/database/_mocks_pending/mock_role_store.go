package mocks

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/errors"
)

type MockRoleStore struct {
	BaseMockStore
}

func NewMockRoleStore() *MockRoleStore {
	return &MockRoleStore{
		BaseMockStore: *NewBaseMockStore(),
	}
}

func (m *MockRoleStore) GetRoles(ctx basecontext.BaseContext, tenantID string) ([]entities.Role, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.Role), args.Get(1).(*errors.Diagnostics)
}

func (m *MockRoleStore) GetRolesByQuery(ctx basecontext.BaseContext, tenantID string, query *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Role]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockRoleStore) GetRoleBySlugOrID(ctx basecontext.BaseContext, tenantID string, slugOrID string) (*entities.Role, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, slugOrID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.Role), args.Get(1).(*errors.Diagnostics)
}

func (m *MockRoleStore) GetRoleUsers(ctx basecontext.BaseContext, tenantID string, roleID string) ([]entities.User, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, roleID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.User), args.Get(1).(*errors.Diagnostics)
}

func (m *MockRoleStore) GetRoleUsersByQuery(ctx basecontext.BaseContext, tenantID string, roleID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, roleID, queryObj)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.User]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockRoleStore) CreateRole(ctx basecontext.BaseContext, tenantID string, role *entities.Role) (*entities.Role, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, role)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.Role), args.Get(1).(*errors.Diagnostics)
}

func (m *MockRoleStore) UpdateRole(ctx basecontext.BaseContext, tenantID string, role *entities.Role) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, role)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockRoleStore) DeleteRole(ctx basecontext.BaseContext, tenantID string, id string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockRoleStore) GetRoleClaims(ctx basecontext.BaseContext, tenantID string, roleID string) ([]entities.Claim, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, roleID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.Claim), args.Get(1).(*errors.Diagnostics)
}

func (m *MockRoleStore) GetRoleClaimsByQuery(ctx basecontext.BaseContext, tenantID string, roleID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, roleID, queryObj)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Claim]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockRoleStore) GetUserRoles(ctx basecontext.BaseContext, tenantID string, userID string) ([]entities.Role, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.Role), args.Get(1).(*errors.Diagnostics)
}

func (m *MockRoleStore) AddUserToRole(ctx basecontext.BaseContext, tenantID string, userID string, roleIdOrSlug string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, userID, roleIdOrSlug)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockRoleStore) RemoveUserFromRole(ctx basecontext.BaseContext, tenantID string, userID string, roleIdOrSlug string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, userID, roleIdOrSlug)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockRoleStore) AddClaimToRole(ctx basecontext.BaseContext, tenantID string, roleID string, claimID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, roleID, claimID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockRoleStore) RemoveClaimFromRole(ctx basecontext.BaseContext, tenantID string, roleID string, claimID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, roleID, claimID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}
