package mocks

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/errors"
)

type MockUserStore struct {
	BaseMockStore
}

func NewMockUserStore() *MockUserStore {
	return &MockUserStore{
		BaseMockStore: *NewBaseMockStore(),
	}
}

func (m *MockUserStore) GetUserByID(ctx basecontext.BaseContext, tenantID string, id string) (*entities.User, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.User), args.Get(1).(*errors.Diagnostics)
}

func (m *MockUserStore) GetUserByUsername(ctx basecontext.BaseContext, tenantID string, username string) (*entities.User, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, username)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.User), args.Get(1).(*errors.Diagnostics)
}

func (m *MockUserStore) GetUsers(ctx basecontext.BaseContext, tenantID string) ([]entities.User, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.User), args.Get(1).(*errors.Diagnostics)
}

func (m *MockUserStore) GetUsersByQuery(ctx basecontext.BaseContext, tenantID string, filterObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, filterObj)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.User]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockUserStore) CreateUser(ctx basecontext.BaseContext, tenantID string, user *entities.User) (*entities.User, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, user)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.User), args.Get(1).(*errors.Diagnostics)
}

func (m *MockUserStore) UpdateUser(ctx basecontext.BaseContext, tenantID string, user *entities.User) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, user)
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockUserStore) UpdateUserPassword(ctx basecontext.BaseContext, tenantID string, id string, password string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, id, password)
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockUserStore) BlockUser(ctx basecontext.BaseContext, tenantID string, id string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, id)
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockUserStore) SetRefreshToken(ctx basecontext.BaseContext, tenantID string, id string, refreshToken string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, id, refreshToken)
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockUserStore) DeleteUser(ctx basecontext.BaseContext, tenantID string, id string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, id)
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockUserStore) GetUserClaims(ctx basecontext.BaseContext, tenantID string, userID string) ([]entities.Claim, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.Claim), args.Get(1).(*errors.Diagnostics)
}

func (m *MockUserStore) GetUserClaimsByQuery(ctx basecontext.BaseContext, tenantID string, userID string, filterObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, userID, filterObj)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Claim]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockUserStore) AddClaimToUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIDOrSlug string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, userID, claimIDOrSlug)
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockUserStore) RemoveClaimFromUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIDOrSlug string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, userID, claimIDOrSlug)
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockUserStore) GetUserRoles(ctx basecontext.BaseContext, tenantID string, userID string) ([]entities.Role, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.Role), args.Get(1).(*errors.Diagnostics)
}

func (m *MockUserStore) GetUserRolesByQuery(ctx basecontext.BaseContext, tenantID string, userID string, filterObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, userID, filterObj)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Role]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockUserStore) AddUserToRole(ctx basecontext.BaseContext, tenantID string, userID string, roleID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, userID, roleID)
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockUserStore) RemoveUserFromRole(ctx basecontext.BaseContext, tenantID string, userID string, roleID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, userID, roleID)
	return args.Get(0).(*errors.Diagnostics)
}
