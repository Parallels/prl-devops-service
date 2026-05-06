package handlers

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/jobs"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
	"github.com/Parallels/prl-devops-service/orchestrator/registry"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

// HostJobEventHandler listens for job_manager events arriving from a connected
// orchestrator host over WebSocket.
//
// For host jobs that are linked to an orchestrator job (created via the async
// dispatch path), it translates progress/completion/failure into updates on the
// orchestrator job so the UI sees a single coherent job stream.
//
// For all other host job events it forwards them to the local UI event emitter
// unchanged (preserving the existing behaviour for direct host jobs).
type HostJobEventHandler struct {
	registrar interfaces.HostRegistrar
}

var (
	hostJobInstance *HostJobEventHandler
	hostJobOnce     sync.Once
)

func NewHostJobEventHandler(registrar interfaces.HostRegistrar) *HostJobEventHandler {
	hostJobOnce.Do(func() {
		hostJobInstance = &HostJobEventHandler{registrar: registrar}
		registrar.RegisterHandler([]constants.EventType{constants.EventTypeJobManager}, hostJobInstance)
	})
	return hostJobInstance
}

func (h *HostJobEventHandler) Handle(ctx basecontext.ApiContext, hostID string, eventType constants.EventType, payload []byte) {
	// Log first line to catch EVERY call regardless of type filter
	ctx.LogDebugf("[Orchestrator] [HS Event] Handler entered: hostID=%s eventType=%q (expected=%s)", hostID, eventType, constants.EventTypeJobManager)
	if eventType != constants.EventTypeJobManager {
		return
	}

	var event models.EventMessage
	if err := json.Unmarshal(payload, &event); err != nil {
		ctx.LogErrorf("[HostJobEventHandler] Error unmarshalling job event from host %s: %v", hostID, err)
		return
	}

	ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] Received job_manager event from host %s: message=%s", hostID, event.Message)

	emitter := serviceprovider.GetEventEmitter()
	if emitter == nil || !emitter.IsRunning() {
		ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] Event emitter not running, skipping event from host %s", hostID)
		return
	}

	// Re-marshal the body so we can decode it as a JobResponse.
	bodyBytes, err := json.Marshal(event.Body)
	if err != nil {
		ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] Failed to marshal event body from host %s: %v, forwarding raw", hostID, err)
		h.forwardRaw(ctx, hostID, event, emitter)
		return
	}
	ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] Re-marshaled body: %s", string(bodyBytes))
	var hostJob models.JobResponse
	if err := json.Unmarshal(bodyBytes, &hostJob); err != nil || hostJob.ID == "" {
		ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] Failed to unmarshal JobResponse from host %s (err=%v, id_empty=%v), forwarding raw", hostID, err, hostJob.ID == "")
		h.forwardRaw(ctx, hostID, event, emitter)
		return
	}

	ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] Parsed hostJob ID=%s state=%s progress=%d message=%q from host %s", hostJob.ID, hostJob.State, hostJob.Progress, hostJob.Message, hostID)

	// Check whether this host job is linked to an orchestrator job.
	reg := registry.Get()
	link, linked := reg.Lookup(hostJob.ID)
	if !linked {
		ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] NO LINK FOUND for hostJobID=%s on host %s — forwarding raw to UI (not dispatched by this orchestrator or link expired)", hostJob.ID, hostID)
		// Not dispatched by this orchestrator — forward as-is.
		h.forwardRaw(ctx, hostID, event, emitter)
		return
	}

	ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] LINK FOUND: hostJobID=%s -> orchestratorJobID=%s for host %s", hostJob.ID, link.OrchestratorJobID, hostID)

	// Translate the host job event into an orchestrator job update.
	jobManager := jobs.Get(ctx)
	if jobManager == nil {
		ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] JobManager is nil, cannot translate event for hostJobID=%s", hostJob.ID)
		return
	}

	switch {
	case event.Message == "JOB_COMPLETED":
		vmID := hostJob.ResultRecordId
		vmName := hostJob.ResultRecordName
		vmLinkID := hostID
		msg := fmt.Sprintf("Virtual machine created on host %s", hostID)
		if vmID != "" {
			msg = fmt.Sprintf("Virtual machine %s created on host %s", vmID, hostID)
		}
		ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] JOB_COMPLETED: hostJobID=%s -> orchestratorJobID=%s, vmID=%s vmName=%s", hostJob.ID, link.OrchestratorJobID, vmID, vmName)
		_ = jobManager.MarkJobCompleteWithRecord(link.OrchestratorJobID, msg, vmID, vmName, "virtual_machine", vmLinkID)
		reg.Remove(hostJob.ID)

	case hostJob.State == constants.JobStateFailed:
		errMsg := hostJob.Error
		if errMsg == "" {
			errMsg = fmt.Sprintf("host job %s on host %s failed", hostJob.ID, hostID)
		}
		ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] JOB_FAILED: hostJobID=%s -> orchestratorJobID=%s, error=%s", hostJob.ID, link.OrchestratorJobID, errMsg)
		_ = jobManager.MarkJobError(link.OrchestratorJobID, fmt.Errorf("%s", errMsg))
		reg.Remove(hostJob.ID)

	default:
		// Intermediate progress update — mirror the full host job state (progress,
		// steps, message) into the orchestrator job in one atomic write so the UI
		// sees the rich step data instead of a plain percentage.
		steps := mapJobStepResponsesToDataSteps(hostJob.Steps)
		ctx.LogDebugf("[Orchestrator] [HostJobEventHandler] PROGRESS_UPDATE: hostJobID=%s -> orchestratorJobID=%s, progress=%d, steps=%d, message=%q", hostJob.ID, link.OrchestratorJobID, hostJob.Progress, len(steps), hostJob.Message)
		_, _ = jobManager.UpdateJobProgressStepsAndMessage(
			link.OrchestratorJobID,
			hostJob.Progress,
			constants.JobStateRunning,
			steps,
			hostJob.Message,
		)
	}
	// Do NOT forward the raw event — the job manager already emitted a
	// translated JOB_UPDATED / JOB_COMPLETED event for the orchestrator job.
}

// mapJobStepResponsesToDataSteps converts the API-level step slice received
// from the host into the DB-level slice expected by the job manager.
func mapJobStepResponsesToDataSteps(src []models.JobStepResponse) []data_models.JobStep {
	steps := make([]data_models.JobStep, 0, len(src))
	for _, s := range src {
		steps = append(steps, data_models.JobStep{
			Name:              s.Name,
			DisplayName:       s.DisplayName,
			Weight:            s.Weight,
			Parallel:          s.Parallel,
			HasPercentage:     s.HasPercentage,
			State:             s.State,
			CurrentPercentage: s.CurrentPercentage,
			Value:             s.Value,
			Total:             s.Total,
			ETA:               s.ETA,
			Message:           s.Message,
			Error:             s.Error,
			Filename:          s.Filename,
			Unit:              s.Unit,
		})
	}
	return steps
}

// forwardRaw broadcasts the host job event to the local UI unchanged.
func (h *HostJobEventHandler) forwardRaw(ctx basecontext.ApiContext, hostID string, event models.EventMessage, emitter interface {
	Broadcast(*models.EventMessage) error
	IsRunning() bool
},
) {
	msg := models.NewEventMessage(constants.EventTypeJobManager, event.Message, models.HostJobEvent{
		HostID:            hostID,
		Event:             event.Body,
		IsOrchestratorJob: h.extractIsOrchestratorJob(event.Body),
	})
	go func() {
		if err := emitter.Broadcast(msg); err != nil {
			ctx.LogErrorf("[HostJobEventHandler] Failed to broadcast job event from host %s: %v", hostID, err)
		}
	}()
}

func (h *HostJobEventHandler) extractIsOrchestratorJob(body interface{}) bool {
	if body == nil {
		return false
	}
	if m, ok := body.(map[string]interface{}); ok {
		if v, ok := m["is_orchestrator_job"]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
	}
	return false
}
