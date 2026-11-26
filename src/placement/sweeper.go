package placement

// Lease and orphan sweeper. Mirrors the stale-job watcher body
// (database/jobs_watcher.go): ticker + cancel channel, blocking body, and the
// goroutine started by the Start hook that startup owns. The hooks cover the
// normal path; this covers what a hook physically cannot — a process killed
// mid-create, and jobs transitioned straight in the database by
// DetectStaleJobs / RecoverOngoingJobs.
//
// Both sweeps are set-based deletes, so two replicas sweeping concurrently is
// harmless (plan D8, multi-replica requirement).

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
)

// ReconcileOnce runs one sweep. Exported for startup wiring and for tests that
// must not wait on a ticker.
func (s *Service) ReconcileOnce(ctx basecontext.ApiContext) {
	if s == nil || s.store == nil || !s.Enabled() {
		return
	}

	if _, err := s.store.ExpireLeases(ctx, time.Now().UTC()); err != nil {
		ctx.LogErrorf("[Placement] Lease expiry sweep failed: %v", err)
	}
	if _, err := s.store.ReconcileOrphans(ctx); err != nil {
		ctx.LogErrorf("[Placement] Orphan reconcile failed: %v", err)
	}
}

// StartReconciler launches the sweeper goroutine. Cancel channel is created
// synchronously, for the reason jobs_watcher.go records: a Stop racing the
// goroutine's lazy init would otherwise no-op and leak the goroutine.
func (s *Service) StartReconciler(ctx basecontext.ApiContext) {
	if s == nil {
		return
	}
	if !s.Enabled() {
		ctx.LogInfof("[Placement] Reconciler not started (reservations disabled)")
		return
	}

	s.startMu.Lock()
	defer s.startMu.Unlock()
	if s.stopReconciler != nil {
		return
	}
	stop := make(chan bool)
	s.stopReconciler = stop
	go s.runReconciler(ctx, stop)
}

// StopReconciler signals the sweeper to exit. Nil-safe before start; blocking
// send mirrors the stale-job watcher idiom.
func (s *Service) StopReconciler() {
	if s == nil {
		return
	}

	s.startMu.Lock()
	stop := s.stopReconciler
	s.stopReconciler = nil
	s.startMu.Unlock()

	if stop != nil {
		stop <- true
	}
}

// runReconciler takes the cancel channel as a parameter on purpose: Stop
// nils the field, and selecting on the field would then block on a nil channel
// forever while the sender waited for a receiver that never arrives.
func (s *Service) runReconciler(ctx basecontext.ApiContext, stop chan bool) {
	interval := time.Duration(config.Get().PlacementReconcileIntervalSeconds()) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.ReconcileOnce(ctx) // sweep on boot: leases may have expired while we were down
	for {
		select {
		case <-stop:
			ctx.LogInfof("[Placement] Reconciler stopped")
			return
		case <-ticker.C:
			s.ReconcileOnce(ctx)
		}
	}
}
