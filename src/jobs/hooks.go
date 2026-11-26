package jobs

import "sync"

// Terminal-state hook.
//
// A create holds capacity on its host from the moment placement decides to put
// it there until the work actually finishes, and only the job manager knows
// when that is. jobs deliberately does not learn what capacity is, who keeps
// the ledger, or how to reach a host — importing the orchestrator from here
// would be a cycle, and importing placement would make a bookkeeping concern a
// dependency of every job in the service. So the consumer is injected.

var (
	jobTerminalMu   sync.RWMutex
	jobTerminalHook func(jobID string, succeeded bool)
)

// SetJobTerminalHook registers the terminal-state consumer. Last writer wins,
// nil clears it. Startup sets it before any job can run; tests may set it at
// any time, which is why the access is guarded rather than documented as
// set-once.
func SetJobTerminalHook(hook func(jobID string, succeeded bool)) {
	jobTerminalMu.Lock()
	defer jobTerminalMu.Unlock()
	jobTerminalHook = hook
}

// notifyJobTerminal hands a finished job to whoever is waiting for it.
//
// It is called only after the terminal state has been persisted: releasing
// capacity for a job whose completion was not stored would hand back room for a
// VM that is still being created, which is the bug this whole feature closes.
//
// The hook's failures are swallowed on purpose. The reservation carries a lease
// that reclaims anything left behind, and a job must not fail because
// bookkeeping did.
func (jms *JobManagerService) notifyJobTerminal(jobID string, succeeded bool) {
	jobTerminalMu.RLock()
	hook := jobTerminalHook
	jobTerminalMu.RUnlock()

	if hook == nil || jobID == "" {
		return
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			jms.apiCtx.LogWarnf("[Jobs] Terminal hook panicked for job %s: %v", jobID, recovered)
		}
	}()

	hook(jobID, succeeded)
}
