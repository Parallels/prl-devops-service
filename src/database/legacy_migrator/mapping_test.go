package legacy_migrator

// Plan-contract units (task 03): the invariants the later tasks (04/05/07)
// rely on — determinism, uniqueness, FK-safe ordering, nil-safe counts —
// pinned so task 04 appending relationalPlan() cannot silently reorder the
// scalar half.

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectionPlanContract(t *testing.T) {
	plan := collectionPlan()
	names := make([]string, 0, len(plan))
	for _, s := range plan {
		names = append(names, s.Name)
	}
	assert.Equal(t, names, func() []string {
		again := collectionPlan() // deterministic: same order, same content
		n := make([]string, 0, len(again))
		for _, s := range again {
			n = append(n, s.Name)
		}
		return n
	}())

	seen := map[string]bool{}
	for _, n := range names {
		assert.False(t, seen[n], "duplicate collection %s", n)
		seen[n] = true
	}

	require.True(t, seen["OrchestratorHosts"] && seen["ReverseProxyHosts"])
	assert.Less(t, indexOf(names, "OrchestratorHosts"), indexOf(names, "ReverseProxyHosts"),
		"referenced before referrer (FK-safe)")

	for _, s := range plan {
		assert.NotNil(t, s.LegacyCount, s.Name)
		assert.NotNil(t, s.Migrate, s.Name)
		assert.NotNil(t, s.Validate, s.Name)
	}
}

func TestScalarPlanOnEmptyData(t *testing.T) {
	db := migDB(t)
	ctx := basecontext.NewRootBaseContext()
	empty := &data.Data{} // every collection nil, both pointers nil

	for _, s := range scalarPlan() {
		assert.Zero(t, s.LegacyCount(empty), s.Name)
		n, err := s.Migrate(ctx, db, empty, defaultChunkSize)
		require.NoError(t, err, s.Name)
		assert.Zero(t, n, s.Name)
		expected, actual, err := s.Validate(db, empty)
		require.NoError(t, err, s.Name)
		assert.Zero(t, expected)
		assert.Zero(t, actual)
	}
}

func indexOf(hay []string, needle string) int {
	for i, h := range hay {
		if h == needle {
			return i
		}
	}
	return -1
}
