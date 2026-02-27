package handlers

import (
	"encoding/json"
	"sync"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator/interfaces"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

type PDfMEventHandler struct {
	registrar interfaces.HostRegistrar
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

	default:
		ctx.LogWarnf("[PDfMEventHandler] Unknown event message : %s", event.Message)
	}
}

func (h *PDfMEventHandler) handleVmStateChange(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	// The body is interface{}, we need to marshal/unmarshal or type assert
	// Since we know the structure, let's try to marshal/unmarshal to VmStateChange
	bodyBytes, err := json.Marshal(event.Body)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error marshalling event body: %v", err)
		return
	}

	var stateChange models.VmStateChange
	if err := json.Unmarshal(bodyBytes, &stateChange); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error unmarshalling VM state change: %v", err)
		return
	}

	ctx.LogInfof("[PDfMEventHandler] VM state changed: %s -> %s (VM: %s, Host: %s)", stateChange.PreviousState, stateChange.CurrentState, stateChange.VmID, hostID)

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error getting database service: %v", err)
		return
	}

	// Update VM state in DB
	// We need to find the Host first
	host, err := dbService.GetOrchestratorHost(ctx, hostID)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error getting host %s from DB: %v", hostID, err)
		return
	}

	if host == nil {
		ctx.LogWarnf("[PDfMEventHandler] Host %s not found in DB", hostID)
		return
	}

	found := false
	for i, vm := range host.VirtualMachines {
		if vm.ID == stateChange.VmID {
			host.VirtualMachines[i].State = stateChange.CurrentState
			found = true
			break
		}
	}

	if !found {
		ctx.LogWarnf("[PDfMEventHandler] VM %s not found in host %s", stateChange.VmID, hostID)
		return
	}

	if _, err := dbService.UpdateOrchestratorHost(ctx, host); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error updating VM %s state in DB: %v", stateChange.VmID, err)
	} else {
		ctx.LogInfof("[PDfMEventHandler] Updated VM %s state to %s", stateChange.VmID, stateChange.CurrentState)

		// Emit VM state change event
		if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
			msg := models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_VM_STATE_CHANGED", models.HostVmEvent{
				HostID: hostID,
				Event:  stateChange,
			})
			go func() {
				if err := emitter.Broadcast(msg); err != nil {
					ctx.LogErrorf("[PDfMEventHandler] Failed to broadcast event HOST_VM_STATE_CHANGED: %v", err)
				}
			}()
		}
	}
}

func (h *PDfMEventHandler) handleVmAdded(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	// The body is interface{}, we need to marshal/unmarshal or type assert
	// Since we know the structure, let's try to marshal/unmarshal to VmAddedEvent
	bodyBytes, err := json.Marshal(event.Body)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error marshalling event body: %v", err)
		return
	}

	var vmAdded models.VmAdded
	if err := json.Unmarshal(bodyBytes, &vmAdded); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error unmarshalling VM added event: %v", err)
		return
	}

	ctx.LogInfof("[PDfMEventHandler] VM added: %s (VM: %s, Host: %s)", vmAdded.NewVm.Name, vmAdded.VmID, hostID)

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error getting database service: %v", err)
		return
	}

	// Update VM state in DB
	// We need to find the Host first
	host, err := dbService.GetOrchestratorHost(ctx, hostID)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error getting host %s from DB: %v", hostID, err)
		return
	}

	if host == nil {
		ctx.LogWarnf("[PDfMEventHandler] Host %s not found in DB", hostID)
		return
	}

	dtoVm := mappers.MapDtoVirtualMachineFromApi(vmAdded.NewVm)
	dtoVm.HostId = host.ID
	dtoVm.Host = host.GetHost()
	dtoVm.HostUrl = host.GetHostUrl()
	host.VirtualMachines = append(host.VirtualMachines, dtoVm)
	if _, err := dbService.UpdateOrchestratorHost(ctx, host); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error updating VM %s state in DB: %v", vmAdded.VmID, err)
	} else {
		ctx.LogInfof("[PDfMEventHandler] VM added %s", vmAdded.VmID)

		// Emit VM added event
		if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
			msg := models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_VM_ADDED", models.HostVmEvent{
				HostID: hostID,
				Event:  vmAdded,
			})
			go func() {
				if err := emitter.Broadcast(msg); err != nil {
					ctx.LogErrorf("[PDfMEventHandler] Failed to broadcast event HOST_VM_ADDED: %v", err)
				}
			}()
		}
	}
}

func (h *PDfMEventHandler) handleVmRemoved(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {
	// The body is interface{}, we need to marshal/unmarshal or type assert
	// Since we know the structure, let's try to marshal/unmarshal to VmRemovedEvent
	bodyBytes, err := json.Marshal(event.Body)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error marshalling event body: %v", err)
		return
	}

	var vmRemoved models.VmRemoved
	if err := json.Unmarshal(bodyBytes, &vmRemoved); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error unmarshalling VM removed event: %v", err)
		return
	}

	ctx.LogInfof("[PDfMEventHandler] VM removed: %s (VM: %s, Host: %s)", vmRemoved.VmID, vmRemoved.VmID, hostID)

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error getting database service: %v", err)
		return
	}

	// Update VM state in DB
	// We need to find the Host first
	host, err := dbService.GetOrchestratorHost(ctx, hostID)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error getting host %s from DB: %v", hostID, err)
		return
	}

	if host == nil {
		ctx.LogWarnf("[PDfMEventHandler] Host %s not found in DB", hostID)
		return
	}

	for i, vm := range host.VirtualMachines {
		if vm.ID == vmRemoved.VmID {
			host.VirtualMachines = append(host.VirtualMachines[:i], host.VirtualMachines[i+1:]...)
			break
		}
	}
	if _, err := dbService.UpdateOrchestratorHost(ctx, host); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error updating VM %s state in DB: %v", vmRemoved.VmID, err)
	} else {
		ctx.LogInfof("[PDfMEventHandler] Removed VM %s", vmRemoved.VmID)

		// Emit VM removed event
		if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
			msg := models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_VM_REMOVED", models.HostVmEvent{
				HostID: hostID,
				Event:  vmRemoved,
			})
			go func() {
				if err := emitter.Broadcast(msg); err != nil {
					ctx.LogErrorf("[PDfMEventHandler] Failed to broadcast event HOST_VM_REMOVED: %v", err)
				}
			}()
		}
	}
}

func (h *PDfMEventHandler) handleVmUpdated(ctx basecontext.ApiContext, hostID string, event models.EventMessage) {

	bodyBytes, err := json.Marshal(event.Body)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error marshalling event body: %v", err)
		return
	}

	var vmUpdated models.VmUpdated
	if err := json.Unmarshal(bodyBytes, &vmUpdated); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error unmarshalling VM updated event: %v", err)
		return
	}

	ctx.LogInfof("[PDfMEventHandler] VM updated: %s (VM: %s, Host: %s)", vmUpdated.NewVm.Name, vmUpdated.VmID, hostID)

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error getting database service: %v", err)
		return
	}

	// Update VM state in DB
	// We need to find the Host first
	host, err := dbService.GetOrchestratorHost(ctx, hostID)
	if err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error getting host %s from DB: %v", hostID, err)
		return
	}

	if host == nil {
		ctx.LogWarnf("[PDfMEventHandler] Host %s not found in DB", hostID)
		return
	}
	found := false
	for i, vm := range host.VirtualMachines {
		if vm.ID == vmUpdated.VmID {
			dtoVm := mappers.MapDtoVirtualMachineFromApi(vmUpdated.NewVm)
			dtoVm.HostId = host.ID
			dtoVm.Host = host.GetHost()
			dtoVm.HostUrl = host.GetHostUrl()
			host.VirtualMachines[i] = dtoVm
			found = true
			break
		}
	}
	if !found {
		ctx.LogWarnf("[PDfMEventHandler] VM %s not found in host %s for update", vmUpdated.VmID, hostID)
		dtoVm := mappers.MapDtoVirtualMachineFromApi(vmUpdated.NewVm)
		dtoVm.HostId = host.ID
		dtoVm.Host = host.GetHost()
		dtoVm.HostUrl = host.GetHostUrl()
		host.VirtualMachines = append(host.VirtualMachines, dtoVm)
	}
	if _, err := dbService.UpdateOrchestratorHost(ctx, host); err != nil {
		ctx.LogErrorf("[PDfMEventHandler] Error updating VM %s state in DB: %v", vmUpdated.VmID, err)
	} else {
		ctx.LogInfof("[PDfMEventHandler] VM updated %s", vmUpdated.VmID)
		if emitter := serviceprovider.GetEventEmitter(); emitter != nil && emitter.IsRunning() {
			msg := models.NewEventMessage(constants.EventTypeOrchestrator, "HOST_VM_UPDATED", models.HostVmEvent{
				HostID: hostID,
				Event:  vmUpdated,
			})
			go func() {
				if err := emitter.Broadcast(msg); err != nil {
					ctx.LogErrorf("[PDfMEventHandler] Failed to broadcast event HOST_VM_UPDATED: %v", err)
				}
			}()
		}
	}
}
