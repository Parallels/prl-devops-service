package legacy_migrator

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The legacy store is a document store: a record can be embedded in more than
// one parent, and nothing stops the same VM appearing inside two hosts' slices.
// Flattening those copies into entity rows hands the same primary key to one
// INSERT twice — which SQLite absorbs (the later copy overwrites) and PostgreSQL
// rejects outright:
//
//	ERROR: ON CONFLICT DO UPDATE command cannot affect row a second time (SQLSTATE 21000)
//
// That abort was not a PostgreSQL quirk to work around, it was the honest
// report of data SQLite had been silently resolving on its own.

func specByName(t *testing.T, name string) CollectionSpec {
	t.Helper()

	for _, s := range collectionPlan() {
		if s.Name == name {
			return s
		}
	}

	t.Fatalf("no collection spec named %s", name)
	return CollectionSpec{}
}

func TestDuplicateEmbeddedVMMigratesOnce(t *testing.T) {
	db := migDB(t)
	ctx := basecontext.NewRootBaseContext()

	legacy := &data.Data{
		OrchestratorHosts: []models.OrchestratorHost{
			{ID: "h1", VirtualMachines: []models.VirtualMachine{{ID: "vm-1", Name: "shared"}}},
			{ID: "h2", VirtualMachines: []models.VirtualMachine{{ID: "vm-1", Name: "shared"}}},
		},
	}

	// Before the dedupe this returned the SQLSTATE 21000 error on PostgreSQL and
	// left the operator with "all changes rolled back".
	_, err := specByName(t, "VirtualMachines").Migrate(ctx, db, legacy, defaultChunkSize)
	require.NoError(t, err)

	var rows []models.VirtualMachine
	require.NoError(t, db.Find(&rows).Error)
	require.Len(t, rows, 1, "the two embedded copies are one entity row")

	// First-seen wins, the rule every rows() function already follows. On SQLite
	// this used to be whichever host happened to come last in the file, so the
	// VM's host assignment moved between runs on its own.
	assert.Equal(t, "h1", rows[0].HostId)
}

func TestDuplicateRowsAcrossChunksStillCollapse(t *testing.T) {
	db := migDB(t)
	ctx := basecontext.NewRootBaseContext()

	// Chunk size 2 puts the second copy of vm-dup in a different statement than
	// the first: batching must not be the reason a duplicate survives.
	legacy := &data.Data{
		OrchestratorHosts: []models.OrchestratorHost{
			{ID: "h1", VirtualMachines: []models.VirtualMachine{
				{ID: "vm-a", Name: "a"},
				{ID: "vm-dup", Name: "first"},
			}},
			{ID: "h2", VirtualMachines: []models.VirtualMachine{
				{ID: "vm-dup", Name: "second"},
				{ID: "vm-b", Name: "b"},
			}},
		},
	}

	_, err := specByName(t, "VirtualMachines").Migrate(ctx, db, legacy, 2)
	require.NoError(t, err)

	var rows []models.VirtualMachine
	require.NoError(t, db.Find(&rows).Error)
	require.Len(t, rows, 3)

	for _, r := range rows {
		if r.ID == "vm-dup" {
			assert.Equal(t, "first", r.Name, "first-seen copy wins across chunk boundaries")
		}
	}
}

func TestDedupeByPKKeepsFirstAndLeavesUnreadableTypesAlone(t *testing.T) {
	type row struct {
		ID string `gorm:"primaryKey"`
		N  int
	}

	out, dropped := dedupeByPK([]row{{ID: "a", N: 1}, {ID: "b", N: 1}, {ID: "a", N: 2}}, "id")
	require.Len(t, out, 2)
	assert.Equal(t, 1, out[0].N, "the first copy is the one written")
	assert.Equal(t, []string{"a"}, dropped)

	// A column named host_id is backed by a field HostID: matched without
	// underscores on either side.
	type edge struct {
		HostID string `gorm:"column:host_id"`
	}
	edges, dropped := dedupeByPK([]edge{{HostID: "h1"}, {HostID: "h1"}}, "host_id")
	assert.Len(t, edges, 1)
	assert.Equal(t, []string{"h1"}, dropped)

	// No key to read ⇒ pass through. Guessing would drop rows, and a genuine
	// conflict is the engine's to report.
	numbers, dropped := dedupeByPK([]int{1, 1, 2}, "id")
	assert.Equal(t, []int{1, 1, 2}, numbers)
	assert.Empty(t, dropped)
}
