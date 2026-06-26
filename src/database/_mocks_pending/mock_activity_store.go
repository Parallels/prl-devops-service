package mocks

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/database/filters"
	"github.com/Parallels/prl-devops-service/errors"
)

type MockActivityStore struct {
	BaseMockStore
	Activities             []entities.Activity
	Activity               *entities.Activity
	CreateActivityResponse *entities.Activity
	CreateActivityErr      *errors.Diagnostics
}

func NewMockActivityStore() *MockActivityStore {
	return &MockActivityStore{
		BaseMockStore: *NewBaseMockStore(),
	}
}

func (m *MockActivityStore) CreateActivity(ctx basecontext.BaseContext, tenantID string, activity *entities.Activity) (*entities.Activity, *errors.Diagnostics) {
	if m.CreateActivityErr != nil {
		return nil, m.CreateActivityErr
	}
	if m.CreateActivityResponse != nil {
		return m.CreateActivityResponse, nil
	}
	return activity, nil
}

func (m *MockActivityStore) GetActivityByID(ctx basecontext.BaseContext, tenantID string, id string) (*entities.Activity, *errors.Diagnostics) {
	return m.Activity, nil
}

func (m *MockActivityStore) GetActivities(ctx basecontext.BaseContext, tenantID string) ([]entities.Activity, *errors.Diagnostics) {
	return m.Activities, nil
}

func (m *MockActivityStore) GetActivitiesByQuery(ctx basecontext.BaseContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Activity], *errors.Diagnostics) {
	return &filters.QueryBuilderResponse[entities.Activity]{
		Items: m.Activities,
		Total: int64(len(m.Activities)),
	}, nil
}

func (m *MockActivityStore) UpdateActivity(ctx basecontext.BaseContext, tenantID string, activity *entities.Activity) *errors.Diagnostics {
	return nil
}

func (m *MockActivityStore) DeleteActivity(ctx basecontext.BaseContext, tenantID string, id string) *errors.Diagnostics {
	return nil
}

func (m *MockActivityStore) GetActivitiesByFilterAdvanced(ctx basecontext.BaseContext, tenantID string, filter *entities.ActivityFilter, page, pageSize int) (*filters.FilterResponse[entities.Activity], *errors.Diagnostics) {
	return &filters.FilterResponse[entities.Activity]{
		Items: m.Activities,
		Total: int64(len(m.Activities)),
	}, nil
}

func (m *MockActivityStore) GetActivityStats(ctx basecontext.BaseContext, tenantID string, filter *entities.ActivityFilter) (map[string]interface{}, *errors.Diagnostics) {
	return nil, nil
}

func (m *MockActivityStore) GetTopActors(ctx basecontext.BaseContext, tenantID string, limit int, filter *entities.ActivityFilter) ([]map[string]interface{}, *errors.Diagnostics) {
	return nil, nil
}

func (m *MockActivityStore) GetActivityTrends(ctx basecontext.BaseContext, tenantID string, days int, filter *entities.ActivityFilter) ([]map[string]interface{}, *errors.Diagnostics) {
	return nil, nil
}

func (m *MockActivityStore) CreateActivitySummary(ctx basecontext.BaseContext, tenantID string, summary *entities.ActivitySummary) (*entities.ActivitySummary, *errors.Diagnostics) {
	return summary, nil
}

func (m *MockActivityStore) GetActivitySummaryByID(ctx basecontext.BaseContext, tenantID string, id string) (*entities.ActivitySummary, *errors.Diagnostics) {
	return nil, nil
}

func (m *MockActivityStore) UpdateActivitySummary(ctx basecontext.BaseContext, tenantID string, summary *entities.ActivitySummary) *errors.Diagnostics {
	return nil
}

func (m *MockActivityStore) DeleteActivitySummary(ctx basecontext.BaseContext, tenantID string, id string) *errors.Diagnostics {
	return nil
}

func (m *MockActivityStore) CleanupOldActivities(ctx basecontext.BaseContext, tenantID string, retentionDays int) *errors.Diagnostics {
	return nil
}

func (m *MockActivityStore) ArchiveActivities(ctx basecontext.BaseContext, tenantID string, beforeDate time.Time) *errors.Diagnostics {
	return nil
}
