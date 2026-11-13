package orchestrator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/data"
	dbmodels "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/stretchr/testify/require"
)

// The UpdateHost tag override replaces the stored tags exactly — the legacy
// DB-tag merge must NOT resurrect them — and preserves the nil/empty
// distinction: omitted tags keep today's merge semantics, an explicit empty
// slice clears them.
func TestUpdateHostTagOverride(t *testing.T) {
	dir, err := os.MkdirTemp("", "devops-update-host-tags-*")
	require.NoError(t, err)
	ctx := newCheckTestContext()
	db := data.NewJsonDatabase(ctx, filepath.Join(dir, "update-host-tags.json"))
	// Cleanups are LIFO: register the daemon drain AFTER the dir removal so it
	// fires first and the save goroutine cannot write into the deleted dir
	// (T12 class).
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	//nolint:staticcheck // legacy engine drain by design; the whole engine retires in v1.3.0 (#426)
	t.Cleanup(func() { _ = db.Disconnect(ctx) })
	serviceprovider.NewMockProvider().JsonDatabase = db

	// State != healthy so the update skips the live hardware probe.
	host, err := db.CreateOrchestratorHost(ctx, dbmodels.OrchestratorHost{
		Host: "tags.local", Enabled: true, State: "unknown", Tags: []string{"old-tag"},
	})
	require.NoError(t, err)

	svc := &OrchestratorService{ctx: ctx}

	// Replace: the override wins over the stored tags.
	fetched, err := db.GetOrchestratorHost(ctx, host.ID)
	require.NoError(t, err)
	updated, err := svc.UpdateHost(ctx, fetched, []string{"pool=build", "os=sonoma"}, true)
	require.NoError(t, err)
	require.Equal(t, []string{"pool=build", "os=sonoma"}, updated.Tags)
	stored, err := db.GetOrchestratorHost(ctx, host.ID)
	require.NoError(t, err)
	require.Equal(t, []string{"pool=build", "os=sonoma"}, stored.Tags)

	// Omitted: the stored tags survive untouched (legacy merge semantics).
	fetched, err = db.GetOrchestratorHost(ctx, host.ID)
	require.NoError(t, err)
	updated, err = svc.UpdateHost(ctx, fetched, nil, false)
	require.NoError(t, err)
	require.Equal(t, []string{"pool=build", "os=sonoma"}, updated.Tags)

	// Clear: the explicit empty set is honored, not refilled from the DB.
	fetched, err = db.GetOrchestratorHost(ctx, host.ID)
	require.NoError(t, err)
	updated, err = svc.UpdateHost(ctx, fetched, []string{}, true)
	require.NoError(t, err)
	require.NotNil(t, updated.Tags, "cleared tags stay an explicit empty list, not nil")
	require.Empty(t, updated.Tags)
	stored, err = db.GetOrchestratorHost(ctx, host.ID)
	require.NoError(t, err)
	require.Empty(t, stored.Tags, "the merge must not resurrect the previous tags")
}
