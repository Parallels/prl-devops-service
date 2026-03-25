package mappers

import (
	"testing"

	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers

func toClaimIDSet(claims []models.UserClaimResponse) map[string]models.UserClaimResponse {
	m := make(map[string]models.UserClaimResponse, len(claims))
	for _, c := range claims {
		m[c.ID] = c
	}
	return m
}

// Tests

func TestComputeEffectiveClaims_DirectOnly(t *testing.T) {
	user := data_models.User{
		Claims: []data_models.Claim{
			{ID: "READ_ONLY", Name: "READ_ONLY"},
			{ID: "LIST_VM", Name: "LIST_VM"},
		},
		Roles: []data_models.Role{},
	}

	result := ComputeEffectiveClaims(user)

	require.Len(t, result, 2)
	for _, c := range result {
		assert.False(t, c.IsInherited, "direct claims should not be marked as inherited")
		assert.Empty(t, c.SourceRole)
	}
	m := toClaimIDSet(result)
	assert.Contains(t, m, "READ_ONLY")
	assert.Contains(t, m, "LIST_VM")
}

func TestComputeEffectiveClaims_InheritedOnly(t *testing.T) {
	user := data_models.User{
		Claims: []data_models.Claim{},
		Roles: []data_models.Role{
			{
				ID:   "USER",
				Name: "USER",
				Claims: []data_models.Claim{
					{ID: "READ_ONLY", Name: "READ_ONLY"},
					{ID: "LIST_VM", Name: "LIST_VM"},
				},
			},
		},
	}

	result := ComputeEffectiveClaims(user)

	require.Len(t, result, 2)
	for _, c := range result {
		assert.True(t, c.IsInherited)
		assert.Equal(t, "USER", c.SourceRole)
	}
}

func TestComputeEffectiveClaims_Mixed_DirectPreferred(t *testing.T) {
	user := data_models.User{
		Claims: []data_models.Claim{
			{ID: "READ_ONLY", Name: "READ_ONLY"}, // also in role
		},
		Roles: []data_models.Role{
			{
				ID:   "USER",
				Name: "USER",
				Claims: []data_models.Claim{
					{ID: "READ_ONLY", Name: "READ_ONLY"},
					{ID: "LIST_VM", Name: "LIST_VM"},
				},
			},
		},
	}

	result := ComputeEffectiveClaims(user)

	require.Len(t, result, 2, "READ_ONLY should appear only once")

	m := toClaimIDSet(result)
	assert.False(t, m["READ_ONLY"].IsInherited, "direct claim should win over inherited")
	assert.Empty(t, m["READ_ONLY"].SourceRole)
	assert.True(t, m["LIST_VM"].IsInherited)
	assert.Equal(t, "USER", m["LIST_VM"].SourceRole)
}

func TestComputeEffectiveClaims_MultipleRoles_Deduplicated(t *testing.T) {
	user := data_models.User{
		Claims: []data_models.Claim{},
		Roles: []data_models.Role{
			{
				ID:   "ROLE_A",
				Name: "ROLE_A",
				Claims: []data_models.Claim{
					{ID: "SHARED_CLAIM", Name: "SHARED_CLAIM"},
				},
			},
			{
				ID:   "ROLE_B",
				Name: "ROLE_B",
				Claims: []data_models.Claim{
					{ID: "SHARED_CLAIM", Name: "SHARED_CLAIM"},
					{ID: "UNIQUE_CLAIM", Name: "UNIQUE_CLAIM"},
				},
			},
		},
	}

	result := ComputeEffectiveClaims(user)

	require.Len(t, result, 2, "SHARED_CLAIM should appear only once")

	m := toClaimIDSet(result)
	assert.True(t, m["SHARED_CLAIM"].IsInherited)
	assert.Equal(t, "ROLE_A", m["SHARED_CLAIM"].SourceRole, "first role wins for SourceRole")
	assert.True(t, m["UNIQUE_CLAIM"].IsInherited)
	assert.Equal(t, "ROLE_B", m["UNIQUE_CLAIM"].SourceRole)
}

func TestComputeEffectiveClaims_EmptyUser(t *testing.T) {
	user := data_models.User{}

	result := ComputeEffectiveClaims(user)

	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestComputeEffectiveClaims_RoleWithNoClaims(t *testing.T) {
	user := data_models.User{
		Claims: []data_models.Claim{
			{ID: "DIRECT_CLAIM", Name: "DIRECT_CLAIM"},
		},
		Roles: []data_models.Role{
			{ID: "EMPTY_ROLE", Name: "EMPTY_ROLE", Claims: []data_models.Claim{}},
		},
	}

	result := ComputeEffectiveClaims(user)

	require.Len(t, result, 1)
	assert.Equal(t, "DIRECT_CLAIM", result[0].ID)
	assert.False(t, result[0].IsInherited)
}

func TestComputeEffectiveClaimIDs_ReturnsAllIDs(t *testing.T) {
	user := data_models.User{
		Claims: []data_models.Claim{
			{ID: "DIRECT", Name: "DIRECT"},
		},
		Roles: []data_models.Role{
			{
				ID:   "ROLE",
				Name: "ROLE",
				Claims: []data_models.Claim{
					{ID: "INHERITED", Name: "INHERITED"},
				},
			},
		},
	}

	ids := ComputeEffectiveClaimIDs(user)

	assert.Len(t, ids, 2)
	assert.Contains(t, ids, "DIRECT")
	assert.Contains(t, ids, "INHERITED")
}

func TestComputeEffectiveClaimIDs_NoDuplicates(t *testing.T) {
	user := data_models.User{
		Claims: []data_models.Claim{
			{ID: "CLAIM_X", Name: "CLAIM_X"},
		},
		Roles: []data_models.Role{
			{
				ID:   "ROLE",
				Name: "ROLE",
				Claims: []data_models.Claim{
					{ID: "CLAIM_X", Name: "CLAIM_X"},
					{ID: "CLAIM_Y", Name: "CLAIM_Y"},
				},
			},
		},
	}

	ids := ComputeEffectiveClaimIDs(user)

	assert.Len(t, ids, 2, "CLAIM_X must not appear twice")
	seen := make(map[string]int)
	for _, id := range ids {
		seen[id]++
	}
	assert.Equal(t, 1, seen["CLAIM_X"])
	assert.Equal(t, 1, seen["CLAIM_Y"])
}
