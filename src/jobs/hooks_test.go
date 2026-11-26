package jobs

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/stretchr/testify/require"
)

func newJobTestManager(t *testing.T) (*JobManagerService, *data.JsonDatabase, basecontext.ApiContext) {
	t.Helper()

	ctx := basecontext.NewRootBaseContext()
	ctx.DisableLog()
	db := data.NewJsonDatabase(ctx, filepath.Join(t.TempDir(), "jobs.json"))
	// Drain the save/backup daemons before TempDir cleanup runs (cleanups are
	// LIFO, so this fires first).
	//nolint:staticcheck // legacy engine drain by design; the whole engine retires in v1.3.0 (#426)
	t.Cleanup(func() { _ = db.Disconnect(ctx) })
	serviceprovider.NewMockProvider().JsonDatabase = db

	manager := New(ctx)
	require.NotNil(t, manager)

	return manager, db, ctx
}

type terminalCall struct {
	jobID     string
	succeeded bool
}

func recordTerminalCalls(t *testing.T, hook func(call terminalCall)) *[]terminalCall {
	t.Helper()

	calls := &[]terminalCall{}
	SetJobTerminalHook(func(jobID string, succeeded bool) {
		call := terminalCall{jobID: jobID, succeeded: succeeded}
		if hook != nil {
			hook(call)
		}
		*calls = append(*calls, call)
	})
	t.Cleanup(func() { SetJobTerminalHook(nil) })

	return calls
}

// The hook exists so a create's capacity is handed back when its work is
// finished. These three properties are the whole contract: it fires on every
// route into a terminal state, only once the state is actually stored, and
// never in a way a job can be blamed for.
func TestJobTerminalHookFiresOnEveryTerminalRoute(t *testing.T) {
	manager, _, _ := newJobTestManager(t)
	calls := recordTerminalCalls(t, nil)

	pending, err := manager.CreateNewJob("root", "orchestrator", "create", "Creating VM")
	require.NoError(t, err)

	completed, err := manager.CreateNewJob("root", "orchestrator", "create", "Creating VM")
	require.NoError(t, err)
	require.NoError(t, manager.MarkJobComplete(completed.ID, "done"))

	withRecord, err := manager.CreateNewJob("root", "orchestrator", "create", "Creating VM")
	require.NoError(t, err)
	require.NoError(t, manager.MarkJobCompleteWithRecord(withRecord.ID, "done", "rec-1", "vm", "vm", "link"))

	failed, err := manager.CreateNewJob("root", "orchestrator", "create", "Creating VM")
	require.NoError(t, err)
	require.NoError(t, manager.MarkJobError(failed.ID, errors.New("host said no")))

	require.Equal(t, []terminalCall{
		{jobID: completed.ID, succeeded: true},
		{jobID: withRecord.ID, succeeded: true},
		{jobID: failed.ID, succeeded: false},
	}, *calls)

	// A job that is merely still running owes nobody a release. Releasing on
	// anything short of terminal would hand back capacity for a VM that is
	// still being created — the original bug, arriving from the other side.
	require.NotContains(t, *calls, terminalCall{jobID: pending.ID})
}

func TestJobTerminalHookSeesThePersistedState(t *testing.T) {
	manager, db, ctx := newJobTestManager(t)

	// What the hook observes is the property that matters: capacity may only be
	// released for a job whose terminal state is on disk. Released a moment
	// earlier, a create whose completion then failed to save would free room for
	// a VM that is still being made, and a second create would be admitted into
	// it.
	states := map[string]constants.JobState{}
	recordTerminalCalls(t, func(call terminalCall) {
		stored, err := db.GetJob(ctx, call.jobID)
		require.NoError(t, err)
		states[call.jobID] = stored.State
	})

	job, err := manager.CreateNewJob("root", "orchestrator", "create", "Creating VM")
	require.NoError(t, err)
	require.NoError(t, manager.MarkJobComplete(job.ID, "done"))
	require.Equal(t, constants.JobStateCompleted, states[job.ID])

	failed, err := manager.CreateNewJob("root", "orchestrator", "create", "Creating VM")
	require.NoError(t, err)
	require.NoError(t, manager.MarkJobError(failed.ID, errors.New("host said no")))
	require.Equal(t, constants.JobStateFailed, states[failed.ID])
}

func TestJobTerminalHookCannotBreakAJob(t *testing.T) {
	manager, db, ctx := newJobTestManager(t)
	SetJobTerminalHook(func(_ string, _ bool) {
		panic("the ledger exploded")
	})
	t.Cleanup(func() { SetJobTerminalHook(nil) })

	job, err := manager.CreateNewJob("root", "orchestrator", "create", "Creating VM")
	require.NoError(t, err)

	// A reservation has a lease that will reclaim anything left dangling; a job
	// must not fail because the bookkeeping on top of it did.
	require.NotPanics(t, func() {
		require.NoError(t, manager.MarkJobComplete(job.ID, "done"))
	})

	stored, err := db.GetJob(ctx, job.ID)
	require.NoError(t, err)
	require.Equal(t, constants.JobStateCompleted, stored.State)
}

func TestJobTerminalHookIsOptional(t *testing.T) {
	manager, _, _ := newJobTestManager(t)
	SetJobTerminalHook(nil)

	job, err := manager.CreateNewJob("root", "orchestrator", "create", "Creating VM")
	require.NoError(t, err)
	require.NoError(t, manager.MarkJobError(job.ID, errors.New("host said no")))

	// Nothing to assert beyond "the job manager works with no consumer wired",
	// which is the state a deployment without placement is in.
	require.NotNil(t, job)
}
