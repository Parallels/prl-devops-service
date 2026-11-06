package orchestrator

import (
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/data/models"
)

// mergePreservedIp protects the orchestrator's view of a VM from "prlctl silence":
// when freshly fetched data (event, full refresh or live refetch) carries no
// internal IP but the stored record knows one for a still-running VM, the stored
// IP and its resolution status are kept instead of being wiped.
func mergePreservedIp(stored, incoming models.VirtualMachine) models.VirtualMachine {
	if incoming.InternalIpAddress == "" {
		if stored.InternalIpAddress != "" &&
			strings.EqualFold(incoming.State, "running") &&
			strings.EqualFold(stored.State, "running") {
			incoming.InternalIpAddress = stored.InternalIpAddress
		}
		if incoming.IpStatus == "" {
			incoming.IpStatus = stored.IpStatus
			incoming.IpStatusDetail = stored.IpStatusDetail
		}
	}
	return incoming
}

// runVmRefetchLoop periodically live-fetches VM data from every enabled+healthy
// host (GET /v1/machines?noCache=true) and upserts it into the orchestrator DB.
// This is the configured refetch cadence (ORCHESTRATOR_VM_REFETCH_INTERVAL_SECONDS,
// default 5 minutes) and complements the PDFM event stream and the heavier full
// refresh with authoritative host-side data.
func (s *OrchestratorService) runVmRefetchLoop() {
	interval := config.Get().OrchestratorVmRefetchInterval()
	s.ctx.LogInfof("[Orchestrator] Starting VM refetch loop (interval: %v)", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.syncContext.Done():
			s.ctx.LogInfof("[Orchestrator] Stopping VM refetch loop")
			return
		case <-ticker.C:
			hosts, err := s.db.GetOrchestratorHosts(s.ctx, "")
			if err != nil {
				continue
			}
			refetched := 0
			for _, host := range hosts {
				if !host.Enabled || host.State != HealthyState {
					continue
				}
				refetched++
				go s.refetchHostVmsLive(s.ctx, s.db, host.ID)
			}
			s.ctx.LogDebugf("[Orchestrator] VM refetch tick dispatched for %d hosts", refetched)
		}
	}
}
