package orchestrator

import (
	"testing"
	"time"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
)

func TestFilterAndSortHosts_EmptyTags(t *testing.T) {
	// Scenario: No selection tags provided, should return all hosts, sorted by ping
	validHosts := []data_models.OrchestratorHost{
		{ID: "host1", Tags: []string{"prod", "web"}},
		{ID: "host2", Tags: []string{"dev", "db"}},
	}
	req := models.CreateVirtualMachineRequest{}

	mockPing := func(host data_models.OrchestratorHost) time.Duration {
		if host.ID == "host1" {
			return 50 * time.Millisecond
		}
		return 10 * time.Millisecond
	}

	result, err := filterAndSortHosts(validHosts, req, mockPing)

	assert.Nil(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "host2", result[0].ID) // Lowest ping first
	assert.Equal(t, "host1", result[1].ID)
}

func TestFilterAndSortHosts_SelectionTagsMatch(t *testing.T) {
	// Scenario: Selection tags provided, some match
	validHosts := []data_models.OrchestratorHost{
		{ID: "host1", Tags: []string{"prod", "web"}},
		{ID: "host2", Tags: []string{"dev", "db"}},
		{ID: "host3", Tags: []string{"prod", "db"}},
	}
	req := models.CreateVirtualMachineRequest{
		SelectionTags: []string{"prod"},
	}

	mockPing := func(host data_models.OrchestratorHost) time.Duration {
		return 10 * time.Millisecond
	}

	result, err := filterAndSortHosts(validHosts, req, mockPing)

	assert.Nil(t, err)
	assert.Len(t, result, 2)
	assert.True(t, result[0].ID == "host1" || result[0].ID == "host3")
}

func TestFilterAndSortHosts_SelectionTagsNoMatch(t *testing.T) {
	// Scenario: Selection tags provided, none match
	validHosts := []data_models.OrchestratorHost{
		{ID: "host1", Tags: []string{"prod", "web"}},
	}
	req := models.CreateVirtualMachineRequest{
		SelectionTags: []string{"eu-west-1"},
	}

	mockPing := func(h data_models.OrchestratorHost) time.Duration { return 0 }

	result, err := filterAndSortHosts(validHosts, req, mockPing)

	assert.NotNil(t, err)
	assert.Equal(t, 400, err.Code)
	assert.Contains(t, err.Message, "tag condition")
	assert.Nil(t, result)
}

func TestFilterAndSortHosts_CacheLocality(t *testing.T) {
	// Scenario: Request is a catalog manifest. Only one host has the exact cache.
	validHosts := []data_models.OrchestratorHost{
		{
			ID: "host1",
			CacheItems: []models.HostCatalogCacheItem{
				{CatalogId: "cat1", Version: "v1", Architecture: "arm64"},
			},
		},
		{
			ID: "host2",
			CacheItems: []models.HostCatalogCacheItem{
				{CatalogId: "cat1", Version: "v2", Architecture: "arm64"},
			},
		},
		{
			ID:         "host3",
			CacheItems: nil, // No cache
		},
	}

	req := models.CreateVirtualMachineRequest{
		Architecture: "arm64",
		CatalogManifest: &models.CreateCatalogVirtualMachineRequest{
			CatalogId: "cat1",
			Version:   "v1",
		},
	}

	mockPing := func(h data_models.OrchestratorHost) time.Duration { return 10 * time.Millisecond }

	result, err := filterAndSortHosts(validHosts, req, mockPing)

	assert.Nil(t, err)
	// It should exclusively select host1, discarding host2 and host3 since host1 has the exact cache
	assert.Len(t, result, 1)
	assert.Equal(t, "host1", result[0].ID)
}

func TestFilterAndSortHosts_CacheLocalityFallback(t *testing.T) {
	// Scenario: Request is a catalog manifest, but NO host has it cached. It should keep all valid hosts.
	validHosts := []data_models.OrchestratorHost{
		{ID: "host1", CacheItems: nil},
		{ID: "host2", CacheItems: nil},
	}

	req := models.CreateVirtualMachineRequest{
		Architecture: "arm64",
		CatalogManifest: &models.CreateCatalogVirtualMachineRequest{
			CatalogId: "missing-cat",
			Version:   "v1",
		},
	}

	mockPing := func(h data_models.OrchestratorHost) time.Duration { return 10 * time.Millisecond }

	result, err := filterAndSortHosts(validHosts, req, mockPing)

	assert.Nil(t, err)
	// Both hosts survive because caching is opportunistic
	assert.Len(t, result, 2)
}

func TestFilterAndSortHosts_PingLatencySort(t *testing.T) {
	// Scenario: Ensure lowest ping is always placed at index 0.
	validHosts := []data_models.OrchestratorHost{
		{ID: "host_slow"},
		{ID: "host_fast"},
		{ID: "host_medium"},
	}

	req := models.CreateVirtualMachineRequest{}

	mockPing := func(h data_models.OrchestratorHost) time.Duration {
		switch h.ID {
		case "host_slow":
			return 500 * time.Millisecond
		case "host_fast":
			return 10 * time.Millisecond
		case "host_medium":
			return 100 * time.Millisecond
		}
		return 0
	}

	result, err := filterAndSortHosts(validHosts, req, mockPing)

	assert.Nil(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "host_fast", result[0].ID)
	assert.Equal(t, "host_medium", result[1].ID)
	assert.Equal(t, "host_slow", result[2].ID)
}
