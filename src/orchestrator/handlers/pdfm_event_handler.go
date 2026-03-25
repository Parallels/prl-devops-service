package handlers

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

type PDfMEventHandler struct {
	registrar  interfaces.HostRegistrar
	hwEnqueuer interfaces.HardwareEnqueuer
}

// ResourceUpdater interface for updating host resources (used by HostStatsHandler)
type ResourceUpdater interface {
	UpdateHostResourcesForEvent(ctx basecontext.ApiContext, hostID string) error
}

var (
	pdfmInstance *PDfMEventHandler
	pdfmOnce     sync.Once
)

func NewPDfMEventHandler(registrar interfaces.HostRegistrar) *PDfMEventHandler {
	pdfmOnce.Do(func() {
		pdfmInstance = &PDfMEventHandler{
			registrar: registrar,
		}
		registrar.RegisterHandler([]constants.EventType{constants.EventTypePDFM}, pdfmInstance)
	})
	return pdfmInstance
}

// SetHardwareEnqueuer injects the hardware update queue after construction.
// Must be called before any events are dispatched (i.e. immediately after
// NewPDfMEventHandler in OrchestratorService.Start).
func (h *PDfMEventHandler) SetHardwareEnqueuer(enqueuer interfaces.HardwareEnqueuer) {
	h.hwEnqueuer = enqueuer
}

// enqueueHardwareUpdate requests a hardware refresh for the host. Called at the
// end of every VM event handler so that Resources (disk space, CPU/memory,
// MacVmsRunning) are updated immediately after each VM state change.
func (h *PDfMEventHandler) enqueueHardwareUpdate(ctx basecontext.ApiContext, hostID string) {
	if h.hwEnqueuer == nil {
		ctx.LogWarnf("[PDfMEventHandler] No hardware enqueuer configured — skipping for host %s", hostID)
		return
	}
	h.hwEnqueuer.Enqueue(hostID)
}

func (h *PDfMEventHandler) Handle(ctx basecontext.ApiContext, hostID string, eventType constants.EventType, payload []byte) {
	if eventType != constants.EventTypePDFM {
		return
	}

	var event models.EventMessage
	if err := json.Unmarshal(payload, &event); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error unmarshalling event message: %v", err)
		return
	}

	switch event.Message {
	case "VM_STATE_CHANGED":
		h.handleVmStateChange(ctx, hostID, event)
	case "VM_ADDED":
		h.handleVmAdded(ctx, hostID, event)
	case "VM_REMOVED":
		h.handleVmRemoved(ctx, hostID, event)
	case "VM_UPDATED":
		h.handleVmUpdated(ctx, hostID, event)
	case "VM_UPTIME_CHANGED":
		h.handleVmUptimeChanged(ctx, hostID, event)
	case "VM_SNAPSHOTS_UPDATED":
		h.handleVMSnapshotsUpdated(ctx, hostID, event)
	case "MAC_VMS_RUNNING_NOW":
		h.handleMacVmsRunningNow(ctx, hostID, event)
	default:
		ctx.LogWarnf("[PDfMEventHandler] Unknown event message : %s", event.Message)
	}
}

// unmarshalEventBody is a generic helper to unmarshal event body
func unmarshalEventBody[T any](ctx basecontext.ApiContext, event models.EventMessage, eventType string) (*T, error) {
	bodyBytes, err := json.Marshal(event.Body)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error marshalling event body: %v", err)
		return nil, err
	}

	var result T
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error unmarshalling %s: %v", eventType, err)
		return nil, err
	}

	return &result, nil
}

// getHostConnectionInfo retrieves the host record to populate VM metadata fields.
// Returns an error if the host is not found or the DB is unavailable.
func (h *PDfMEventHandler) getHostConnectionInfo(ctx basecontext.ApiContext, hostID string) (*data_models.OrchestratorHost, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error getting database service: %v", err)
		return nil, err
	}

	host, err := dbService.GetOrchestratorHost(ctx, hostID)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error getting host %s from DB: %v", hostID, err)
		return nil, err
	}

	if host == nil {
		ctx.LogWarnf("[PDfMEventHandler] [orchestrator] Host %s not found in DB", hostID)
		return nil, fmt.Errorf("host not found")
	}

	return host, nil
}

// emitHostVMEvent emits orchestrator events with standardized error handling
func (h *PDfMEventHandler) emitHostVMEvent(ctx basecontext.ApiContext, hostID, eventName string, eventData interface{}) {
	if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
		msg := models.NewEventMessage(constants.EventTypeOrchestrator, eventName, models.HostVmEvent{
			HostID: hostID,
			Event:  eventData,
		})
		go func() {
			if err := emitter.Broadcast(msg); err != nil {
				ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Failed to broadcast event %s: %v", eventName, err)
			}
		}()
	}
}

func (h *PDfMEventHandler) handleVmStateChange(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	stateChange, err := unmarshalEventBody[models.VmStateChange](ctx, event, "VM state change")
	if err != nil {
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM state changed: %s -> %s (VM: %s, Host: %s)",
		stateChange.PreviousState, stateChange.CurrentState, stateChange.VmID, hostID)

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error getting database service: %v", err)
		return
	}

	if err := dbService.UpdateOrchestratorHostVMState(ctx, hostID, stateChange.VmID, stateChange.CurrentState); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error updating VM %s state: %v", stateChange.VmID, err)
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM %s state updated to %s", stateChange.VmID, stateChange.CurrentState)
	h.emitHostVMEvent(ctx, hostID, "HOST_VM_STATE_CHANGED", *stateChange)
	h.enqueueHardwareUpdate(ctx, hostID)
}

func (h *PDfMEventHandler) handleVmAdded(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	vmAdded, err := unmarshalEventBody[models.VmAdded](ctx, event, "VM added event")
	if err != nil {
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM added: %s (VM: %s, Host: %s)", vmAdded.NewVm.Name, vmAdded.VmID, hostID)

	// Read host connection info to populate VM metadata fields.
	host, err := h.getHostConnectionInfo(ctx, hostID)
	if err != nil {
		return
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error getting database service: %v", err)
		return
	}

	dtoVm := mappers.MapDtoVirtualMachineFromApi(vmAdded.NewVm)
	dtoVm.HostId = host.ID
	dtoVm.HostName = getHostName(*host)
	dtoVm.Host = host.GetHost()
	dtoVm.HostUrl = host.GetHostUrl()

	if err := dbService.UpsertOrchestratorHostVM(ctx, hostID, dtoVm); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error upserting VM %s: %v", vmAdded.VmID, err)
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM %s added", vmAdded.VmID)
	h.emitHostVMEvent(ctx, hostID, "HOST_VM_ADDED", *vmAdded)
	h.enqueueHardwareUpdate(ctx, hostID)
}

func (h *PDfMEventHandler) handleVmRemoved(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	vmRemoved, err := unmarshalEventBody[models.VmRemoved](ctx, event, "VM removed event")
	if err != nil {
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM removed: %s (Host: %s)", vmRemoved.VmID, hostID)

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error getting database service: %v", err)
		return
	}

	if err := dbService.RemoveOrchestratorHostVM(ctx, hostID, vmRemoved.VmID); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error removing VM %s: %v", vmRemoved.VmID, err)
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM %s removed", vmRemoved.VmID)
	h.emitHostVMEvent(ctx, hostID, "HOST_VM_REMOVED", *vmRemoved)
	h.enqueueHardwareUpdate(ctx, hostID)
}

func (h *PDfMEventHandler) handleVmUpdated(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	vmUpdated, err := unmarshalEventBody[models.VmUpdated](ctx, event, "VM updated event")
	if err != nil {
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM updated: %s (VM: %s, Host: %s)", vmUpdated.NewVm.Name, vmUpdated.VmID, hostID)

	// Read host connection info to populate VM metadata fields.
	host, err := h.getHostConnectionInfo(ctx, hostID)
	if err != nil {
		return
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error getting database service: %v", err)
		return
	}

	dtoVm := mappers.MapDtoVirtualMachineFromApi(vmUpdated.NewVm)
	dtoVm.HostId = host.ID
	dtoVm.HostName = getHostName(*host)
	dtoVm.Host = host.GetHost()
	dtoVm.HostUrl = host.GetHostUrl()

	if err := dbService.UpsertOrchestratorHostVM(ctx, hostID, dtoVm); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error upserting VM %s: %v", vmUpdated.VmID, err)
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM %s updated", vmUpdated.VmID)
	h.emitHostVMEvent(ctx, hostID, "HOST_VM_UPDATED", *vmUpdated)
	h.enqueueHardwareUpdate(ctx, hostID)
}

func (h *PDfMEventHandler) handleVmUptimeChanged(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	uptimeChanged, err := unmarshalEventBody[models.VmUptimeChanged](ctx, event, "VM uptime changed event")
	if err != nil {
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] [uptime] VM uptime changed: %s (VM: %s, Host: %s)",
		uptimeChanged.Uptime, uptimeChanged.VmID, hostID)

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error getting database service: %v", err)
		return
	}

	if err := dbService.UpdateOrchestratorHostVMUptime(ctx, hostID, uptimeChanged.VmID, uptimeChanged.Uptime); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error updating VM %s uptime: %v", uptimeChanged.VmID, err)
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] [uptime] VM %s uptime updated: %s", uptimeChanged.VmID, uptimeChanged.Uptime)
	h.emitHostVMEvent(ctx, hostID, "HOST_VM_UPTIME_CHANGED", *uptimeChanged)
}

func (h *PDfMEventHandler) handleVMSnapshotsUpdated(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	snapshotsUpdated, err := unmarshalEventBody[models.VmSnapshotsUpdated](ctx, event, "VM snapshots updated event")
	if err != nil {
		return
	}

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error getting database service: %v", err)
		return
	}

	dbSnapshots := mappers.VMSnapshotsApiToDto(snapshotsUpdated.VMSnapshots)

	err = dbService.SetHostVMSnapshots(ctx, hostID, data_models.VMSnapshots{
		VMId:       snapshotsUpdated.VmID,
		VMSnapshot: dbSnapshots,
	})
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] [snapshots] Error updating snapshots in DB for VM %s: %v", snapshotsUpdated.VmID, err)
		return
	}
	ctx.LogInfof("[PDfMEventHandler] [orchestrator] [snapshots] VM snapshots updated:(VM: %s, Host: %s)", snapshotsUpdated.VmID, hostID)
	h.emitHostVMEvent(ctx, hostID, "HOST_VM_SNAPSHOTS_UPDATED", *snapshotsUpdated)
}

func (h *PDfMEventHandler) handleMacVmsRunningNow(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	macVmsRunningNow, err := unmarshalEventBody[models.MacVMsRunningNowEvent](ctx, event, "MAC VMs running now event")
	if err != nil {
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] MAC VMs running now: %v (Host: %s)", macVmsRunningNow, hostID)
	h.emitHostVMEvent(ctx, hostID, "HOST_MAC_VMS_RUNNING_NOW", *macVmsRunningNow)
	h.enqueueHardwareUpdate(ctx, hostID)
}

func getHostName(host data_models.OrchestratorHost) string {
	if host.Description != "" {
		return host.Description
	}
	return host.Host
}
