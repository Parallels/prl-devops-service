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
	registrar   interfaces.HostRegistrar
	vmRefresher VMRefresher
}

// ResourceUpdater interface for updating host resources (used by HostStatsHandler)
type ResourceUpdater interface {
	UpdateHostResourcesForEvent(ctx basecontext.ApiContext, hostID string) error
}

// VMRefresher fetches a single VM from the host and syncs it to the DB.
// Called after VM_STATE_CHANGED to capture config that may have changed with the transition.
type VMRefresher interface {
	UpdateHostVMForEvent(ctx basecontext.ApiContext, hostID string, vmID string) error
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

// SetVMRefresher sets the VM refresher dependency
func (h *PDfMEventHandler) SetVMRefresher(refresher VMRefresher) {
	h.vmRefresher = refresher
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

// getHostFromDatabase retrieves host from database with error handling
func (h *PDfMEventHandler) getHostFromDatabase(ctx basecontext.ApiContext, hostID string) (*data_models.OrchestratorHost, error) {
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

// findVMIndex finds the index of a VM in the host's VMs slice
func findVMIndex(vms []data_models.VirtualMachine, vmID string) int {
	for i, vm := range vms {
		if vm.ID == vmID {
			return i
		}
	}
	return -1
}

// updateHostInDatabase updates host in database with error handling and logging
func (h *PDfMEventHandler) updateHostInDatabase(ctx basecontext.ApiContext, host *data_models.OrchestratorHost, vmID, operation string) error {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error getting database service: %v", err)
		return err
	}

	if _, err := dbService.UpdateOrchestratorHost(ctx, host); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] [orchestrator] Error updating VM %s %s in DB: %v", vmID, operation, err)
		return err
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM %s %s", vmID, operation)
	return nil
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

	host, err := h.getHostFromDatabase(ctx, hostID)
	if err != nil {
		return
	}

	vmIndex := findVMIndex(host.VirtualMachines, stateChange.VmID)
	if vmIndex == -1 {
		ctx.LogWarnf("[PDfMEventHandler] [orchestrator] VM %s not found in host %s", stateChange.VmID, hostID)
		return
	}

	// Fast path: update state immediately so the UI sees it right away.
	host.VirtualMachines[vmIndex].State = stateChange.CurrentState

	if err := h.updateHostInDatabase(ctx, host, stateChange.VmID, fmt.Sprintf("state updated to %s", stateChange.CurrentState)); err != nil {
		return
	}

	h.emitHostVMEvent(ctx, hostID, "HOST_VM_STATE_CHANGED", *stateChange)

	// Async full-VM sync: fetch the current VM config from the host to capture any
	// settings that changed alongside the state transition (IP, RAM, CPU, etc.).
	if h.vmRefresher != nil {
		go func() {
			if err := h.vmRefresher.UpdateHostVMForEvent(ctx, hostID, stateChange.VmID); err != nil {
				ctx.LogWarnf("[PDfMEventHandler] [orchestrator] Could not sync VM %s config after state change: %v", stateChange.VmID, err)
			}
		}()
	}
}

func (h *PDfMEventHandler) handleVmAdded(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	vmAdded, err := unmarshalEventBody[models.VmAdded](ctx, event, "VM added event")
	if err != nil {
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM added: %s (VM: %s, Host: %s)", vmAdded.NewVm.Name, vmAdded.VmID, hostID)

	host, err := h.getHostFromDatabase(ctx, hostID)
	if err != nil {
		return
	}

	dtoVm := mappers.MapDtoVirtualMachineFromApi(vmAdded.NewVm)
	dtoVm.HostId = host.ID
	dtoVm.HostName = getHostName(*host)
	dtoVm.Host = host.GetHost()
	dtoVm.HostUrl = host.GetHostUrl()
	host.VirtualMachines = append(host.VirtualMachines, dtoVm)

	if err := h.updateHostInDatabase(ctx, host, vmAdded.VmID, "added"); err != nil {
		return
	}

	h.emitHostVMEvent(ctx, hostID, "HOST_VM_ADDED", *vmAdded)
}

func (h *PDfMEventHandler) handleVmRemoved(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	vmRemoved, err := unmarshalEventBody[models.VmRemoved](ctx, event, "VM removed event")
	if err != nil {
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM removed: %s (VM: %s, Host: %s)", vmRemoved.VmID, vmRemoved.VmID, hostID)

	host, err := h.getHostFromDatabase(ctx, hostID)
	if err != nil {
		return
	}

	vmIndex := findVMIndex(host.VirtualMachines, vmRemoved.VmID)
	if vmIndex != -1 {
		host.VirtualMachines = append(host.VirtualMachines[:vmIndex], host.VirtualMachines[vmIndex+1:]...)
	}

	if err := h.updateHostInDatabase(ctx, host, vmRemoved.VmID, "removed"); err != nil {
		return
	}

	h.emitHostVMEvent(ctx, hostID, "HOST_VM_REMOVED", *vmRemoved)
}

func (h *PDfMEventHandler) handleVmUpdated(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	vmUpdated, err := unmarshalEventBody[models.VmUpdated](ctx, event, "VM updated event")
	if err != nil {
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] VM updated: %s (VM: %s, Host: %s)", vmUpdated.NewVm.Name, vmUpdated.VmID, hostID)

	host, err := h.getHostFromDatabase(ctx, hostID)
	if err != nil {
		return
	}

	vmIndex := findVMIndex(host.VirtualMachines, vmUpdated.VmID)
	dtoVm := mappers.MapDtoVirtualMachineFromApi(vmUpdated.NewVm)
	dtoVm.HostId = host.ID
	dtoVm.HostName = getHostName(*host)
	dtoVm.Host = host.GetHost()
	dtoVm.HostUrl = host.GetHostUrl()

	if vmIndex != -1 {
		host.VirtualMachines[vmIndex] = dtoVm
	} else {
		ctx.LogWarnf("[PDfMEventHandler] [orchestrator] VM %s not found in host %s for update", vmUpdated.VmID, hostID)
		host.VirtualMachines = append(host.VirtualMachines, dtoVm)
	}

	if err := h.updateHostInDatabase(ctx, host, vmUpdated.VmID, "updated"); err != nil {
		return
	}

	h.emitHostVMEvent(ctx, hostID, "HOST_VM_UPDATED", *vmUpdated)
}

func (h *PDfMEventHandler) handleVmUptimeChanged(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	uptimeChanged, err := unmarshalEventBody[models.VmUptimeChanged](ctx, event, "VM uptime changed event")
	if err != nil {
		return
	}

	ctx.LogInfof("[PDfMEventHandler] [orchestrator] [uptime] VM uptime changed: %s (VM: %s, Host: %s)",
		uptimeChanged.Uptime, uptimeChanged.VmID, hostID)

	host, err := h.getHostFromDatabase(ctx, hostID)
	if err != nil {
		return
	}

	vmIndex := findVMIndex(host.VirtualMachines, uptimeChanged.VmID)
	if vmIndex == -1 {
		ctx.LogWarnf("[PDfMEventHandler] [orchestrator] [uptime] VM %s not found in host %s for uptime update",
			uptimeChanged.VmID, hostID)
		return
	}

	host.VirtualMachines[vmIndex].Uptime = uptimeChanged.Uptime

	if err := h.updateHostInDatabase(ctx, host, uptimeChanged.VmID, fmt.Sprintf("uptime updated: %s", uptimeChanged.Uptime)); err != nil {
		return
	}

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
}


func getHostName(host data_models.OrchestratorHost) string {
	if host.Description != "" {
		return host.Description
	}
	return host.Host
}
