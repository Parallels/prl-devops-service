package mocks

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/errors"
)

type MockClaimStore struct {
	BaseMockStore
}

func NewMockClaimStore() *MockClaimStore {
	return &MockClaimStore{
		*NewBaseMockStore(),
	}
}

func (m *MockClaimStore) GetClaims(ctx basecontext.BaseContext, tenantID string) ([]entities.Claim, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.Claim), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) GetClaimsByQuery(ctx basecontext.BaseContext, tenantID string, query *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Claim]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) GetClaimBySlugOrID(ctx basecontext.BaseContext, tenantID string, slugOrID string) (*entities.Claim, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, slugOrID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.Claim), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) CreateClaim(ctx basecontext.BaseContext, tenantID string, claim *entities.Claim) (*entities.Claim, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, claim)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.Claim), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) UpdateClaim(ctx basecontext.BaseContext, tenantID string, claim *entities.Claim) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, claim)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockClaimStore) DeleteClaim(ctx basecontext.BaseContext, tenantID string, id string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockClaimStore) GetClaimUsers(ctx basecontext.BaseContext, tenantID string, claimID string) ([]entities.User, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, claimID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.User), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) GetClaimUsersByQuery(ctx basecontext.BaseContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, claimID, queryBuilder)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.User]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) AddClaimToUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIdOrSlug string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, userID, claimIdOrSlug)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockClaimStore) RemoveClaimFromUser(ctx basecontext.BaseContext, tenantID string, userID string, claimIdOrSlug string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, userID, claimIdOrSlug)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockClaimStore) GetClaimsByLevel(ctx basecontext.BaseContext, tenantID string, level entities.SecurityLevel) ([]entities.Claim, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, level)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.Claim), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) GetClaimApiKeys(ctx basecontext.BaseContext, tenantID string, claimID string) ([]entities.ApiKey, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, claimID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.ApiKey), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) GetClaimApiKeysByQuery(ctx basecontext.BaseContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.ApiKey], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, claimID, queryBuilder)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.ApiKey]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) AddClaimToApiKey(ctx basecontext.BaseContext, tenantID string, claimID string, apiKeyID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, claimID, apiKeyID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockClaimStore) RemoveClaimFromApiKey(ctx basecontext.BaseContext, tenantID string, claimID string, apiKeyID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, claimID, apiKeyID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockClaimStore) GetClaimRoles(ctx basecontext.BaseContext, tenantID string, claimID string) ([]entities.Role, *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, claimID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.Role), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) GetClaimRolesByQuery(ctx basecontext.BaseContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *errors.Diagnostics) {
	args := m.Called(ctx, tenantID, claimID, queryBuilder)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*filters.QueryBuilderResponse[entities.Role]), args.Get(1).(*errors.Diagnostics)
}

func (m *MockClaimStore) AddClaimToRole(ctx basecontext.BaseContext, tenantID string, claimID string, roleID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, claimID, roleID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockClaimStore) RemoveClaimFromRole(ctx basecontext.BaseContext, tenantID string, claimID string, roleID string) *errors.Diagnostics {
	args := m.Called(ctx, tenantID, claimID, roleID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}
