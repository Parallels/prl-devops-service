package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/restapi"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"github.com/Parallels/prl-devops-service/serviceprovider/parallelsdesktop"
	"github.com/Parallels/prl-devops-service/serviceprovider/system"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func registerVirtualMachinesHandlers(ctx basecontext.ApiContext, version string) {
	// provider := serviceprovider.Get()
	// Validation is now done at startup via 'host' module check
	ctx.LogInfof("Registering version %s virtual machine handlers", version)
	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines").
		WithRequiredClaim(constants.LIST_VM_CLAIM).
		WithHandler(GetVirtualMachinesHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines/{id}").
		WithRequiredClaim(constants.LIST_VM_CLAIM).
		WithHandler(GetVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/machines").
		WithRequiredClaim(constants.CREATE_VM_CLAIM).
		WithHandler(CreateVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/machines/{id}/snapshots").
		WithRequiredClaim(constants.CREATE_SNAPSHOT_VM_CLAIM).
		WithHandler(CreateSnapshot()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/machines/{id}/snapshots/{snapshot_id}").
		WithRequiredClaim(constants.DELETE_SNAPSHOT_VM_CLAIM).
		WithHandler(DeleteSnapshot()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines/{id}/snapshots").
		WithRequiredClaim(constants.LIST_SNAPSHOT_VM_CLAIM).
		WithHandler(ListSnapshot()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/machines/{id}/snapshots/{snapshot_id}/revert").
		WithRequiredClaim(constants.REVERT_SNAPSHOT_VM_CLAIM).
		WithHandler(RevertSnapshot()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/machines/{id}/snapshots/{snapshot_id}/switch").
		WithRequiredClaim(constants.SWITCH_SNAPSHOT_VM_CLAIM).
		WithHandler(SwitchSnapshot()).
		Register()

	restapi.NewController().
		WithMethod(restapi.DELETE).
		WithVersion(version).
		WithPath("/machines/{id}").
		WithRequiredClaim(constants.DELETE_VM_CLAIM).
		WithHandler(DeleteVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/machines/register").
		WithRequiredClaim(constants.CREATE_VM_CLAIM).
		WithHandler(RegisterVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/machines/{id}/unregister").
		WithRequiredClaim(constants.UPDATE_VM_CLAIM).
		WithHandler(UnregisterVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines/{id}/start").
		WithRequiredClaim(constants.UPDATE_VM_STATES_CLAIM).
		WithHandler(StartVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines/{id}/stop").
		WithRequiredClaim(constants.UPDATE_VM_STATES_CLAIM).
		WithHandler(StopVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines/{id}/restart").
		WithRequiredClaim(constants.UPDATE_VM_STATES_CLAIM).
		WithHandler(RestartVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines/{id}/pause").
		WithRequiredClaim(constants.UPDATE_VM_STATES_CLAIM).
		WithHandler(PauseVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines/{id}/resume").
		WithRequiredClaim(constants.UPDATE_VM_STATES_CLAIM).
		WithHandler(ResumeMachineController()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines/{id}/reset").
		WithRequiredClaim(constants.UPDATE_VM_STATES_CLAIM).
		WithHandler(ResetMachineController()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines/{id}/suspend").
		WithRequiredClaim(constants.UPDATE_VM_STATES_CLAIM).
		WithHandler(SuspendVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.GET).
		WithVersion(version).
		WithPath("/machines/{id}/status").
		WithRequiredClaim(constants.LIST_VM_CLAIM).
		WithHandler(GetVirtualMachineStatusHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/machines/{id}/set").
		WithRequiredClaim(constants.UPDATE_VM_CLAIM).
		WithHandler(SetVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/machines/{id}/execute").
		WithRequiredClaim(constants.EXECUTE_COMMAND_VM_CLAIM).
		WithHandler(ExecuteCommandOnVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/machines/{id}/upload").
		WithRequiredClaim(constants.UPDATE_VM_CLAIM).
		WithHandler(UploadFileToVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/machines/{id}/rename").
		WithRequiredClaim(constants.UPDATE_VM_CLAIM).
		WithHandler(RenameVirtualMachineHandler()).
		Register()

	restapi.NewController().
		WithMethod(restapi.PUT).
		WithVersion(version).
		WithPath("/machines/{id}/clone").
		WithRequiredClaim(constants.CREATE_VM_CLAIM).
		WithHandler(CloneVirtualMachineHandler()).
		Register()
}

// @Summary		Gets all the virtual machines
// @Description	This endpoint returns all the virtual machines
// @Tags			Machines
// @Produce		json
// @Param			filter	header		string	false	"X-Filter"
// @Success		200		{object}	[]models.ParallelsVM
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines [get]
func GetVirtualMachinesHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		vms, err := svc.GetCachedVms(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(vms) == 0 {
			w.WriteHeader(http.StatusOK)
			vms = make([]models.ParallelsVM, 0)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(vms)
			ctx.LogInfof("No machines found")
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(vms)
		ctx.LogInfof("Machines returned: %v", len(vms))
	}
}

// @Summary		Gets a virtual machine
// @Description	This endpoint returns a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id	path		string	true	"Machine ID"
// @Success		200	{object}	models.ParallelsVM
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id} [get]
func GetVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		vm, err := svc.GetVmSync(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if vm == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Machine %v not found", id),
				Code:    http.StatusNotFound,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(vm)
		ctx.LogInfof("Machine returned: %v", vm.ID)
	}
}

// @Summary		Starts a virtual machine
// @Description	This endpoint starts a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id	path		string	true	"Machine ID"
// @Success		200	{object}	models.VirtualMachineOperationResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/start [get]
func StartVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.StartVm(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Start",
			Status:    "Success",
		}

		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Machine started: %v", id)
	}
}

// @Summary		Stops a virtual machine
// @Description	This endpoint stops a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id		path		string	true	"Machine ID"
// @Param			force	query		bool	false	"Force stop the virtual machine"
// @Success		200		{object}	models.VirtualMachineOperationResponse
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/stop [get]
func StopVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		// Get force parameter from query string
		force := r.URL.Query().Get("force") == "true"

		var flags parallelsdesktop.DesiredStateFlags
		if force {
			flags = parallelsdesktop.NewDesiredStateFlags("--force")
		} else {
			flags = parallelsdesktop.NewDesiredStateFlags()
		}

		err := svc.StopVm(ctx, id, flags)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Stop",
			Status:    "Success",
		}

		_ = json.NewEncoder(w).Encode(result)
		if force {
			ctx.LogInfof("Machine force stopped: %v", id)
		} else {
			ctx.LogInfof("Machine stopped: %v", id)
		}
	}
}

// @Summary		Restarts a virtual machine
// @Description	This endpoint restarts a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id	path		string	true	"Machine ID"
// @Success		200	{object}	models.VirtualMachineOperationResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/restart [get]
func RestartVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.RestartVm(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Restart",
			Status:    "Success",
		}
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Machine restarted: %v", id)
	}
}

// @Summary		Suspends a virtual machine
// @Description	This endpoint suspends a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id	path		string	true	"Machine ID"
// @Success		200	{object}	models.VirtualMachineOperationResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/suspend [get]
func SuspendVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.SuspendVm(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Suspend",
			Status:    "Success",
		}
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Machine suspended: %v", id)
	}
}

// @Summary		Resumes a virtual machine
// @Description	This endpoint resumes a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id	path		string	true	"Machine ID"
// @Success		200	{object}	models.VirtualMachineOperationResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/resume [get]
func ResumeMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.ResumeVm(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Resume",
			Status:    "Success",
		}
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Machine resumed: %v", id)
	}
}

// @Summary		Reset a virtual machine
// @Description	This endpoint reset a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id	path		string	true	"Machine ID"
// @Success		200	{object}	models.VirtualMachineOperationResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/reset [get]
func ResetMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.ResetVm(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Reset",
			Status:    "Success",
		}
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Machine reset: %v", id)
	}
}

// @Summary		Pauses a virtual machine
// @Description	This endpoint pauses a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id	path		string	true	"Machine ID"
// @Success		200	{object}	models.VirtualMachineOperationResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/pause [get]
func PauseVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.PauseVm(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Pause",
			Status:    "Success",
		}

		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Machine paused: %v", id)
	}
}

// @Summary		Deletes a virtual machine
// @Description	This endpoint deletes a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id	path	string	true	"Machine ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id} [delete]
func DeleteVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.DeleteVm(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfof("Machine deleted: %v", id)
	}
}

// @Summary		Get the current state of a virtual machine
// @Description	This endpoint returns the current state of a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id	path		string	true	"Machine ID"
// @Success		200	{object}	models.VirtualMachineStatusResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/status [get]
func GetVirtualMachineStatusHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]
		sps, err := svc.SwitchSnapshot(ctx, id, &models.SnapshotSwitchRequest{
			SnapshotId: "2f69cff8-7fdc-4e9a-bb26-adffacc53440",
			SkipResume: false,
		})
		// response, err := svc.VmStatus(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}
		ctx.LogInfof("Snapshots for machine %v: %v", id, sps)
		result := models.VirtualMachineStatusResponse{
			ID:           "id",
			Status:       "test",
			IpConfigured: "",
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Machine status returned: %v", id)
	}
}

// @Summary		Configures a virtual machine
// @Description	This endpoint configures a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id				path		string								true	"Machine ID"
// @Param			configRequest	body		models.VirtualMachineConfigRequest	true	"Machine Set Request"
// @Success		200				{object}	models.VirtualMachineConfigResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/set [put]
func SetVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.VirtualMachineConfigRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		params := mux.Vars(r)
		id := params["id"]

		if err := svc.ConfigureVm(ctx, id, &request); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		result := models.VirtualMachineConfigResponse{
			Operations: make([]models.VirtualMachineConfigResponseOperation, 0),
		}

		for _, op := range request.Operations {
			rOp := models.VirtualMachineConfigResponseOperation{
				Group:     op.Group,
				Operation: op.Operation,
			}
			if op.Error != nil {
				rOp.Status = "Error"
				rOp.Error = op.Error.Error()
			} else {
				rOp.Status = "Success"
			}

			result.Operations = append(result.Operations, rOp)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Machine configured: %v", id)
	}
}

// @Summary		Clones a virtual machine
// @Description	This endpoint clones a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id				path		string										true	"Machine ID"
// @Param			configRequest	body		models.VirtualMachineCloneCommandRequest	true	"Machine Clone Request"
// @Success		200				{object}	models.VirtualMachineCloneCommandResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/clone [put]
func CloneVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.VirtualMachineCloneCommandRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		params := mux.Vars(r)
		id := params["id"]
		configure := models.VirtualMachineConfigRequest{
			Operations: []*models.VirtualMachineConfigRequestOperation{
				{
					Group:     "machine",
					Operation: "clone",
					Options: []*models.VirtualMachineConfigRequestOperationOption{
						{
							Flag:  "name",
							Value: request.CloneName,
						},
					},
				},
			},
		}

		if err := svc.ConfigureVm(ctx, id, &configure); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		result := models.VirtualMachineCloneCommandResponse{}

		vmId, err := svc.GetVmSync(ctx, request.CloneName)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		result.Id = vmId.ID
		result.Status = "Success"

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
		ctx.LogInfof("Machine %v cloned successfully to %v with id %v", id, request.CloneName, result.Id)
	}
}

// @Summary		Executes a command on a virtual machine
// @Description	This endpoint executes a command on a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id				path		string										true	"Machine ID"
// @Param			executeRequest	body		models.VirtualMachineExecuteCommandRequest	true	"Machine Execute Command Request"
// @Success		200				{object}	models.VirtualMachineExecuteCommandResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/execute [put]
func ExecuteCommandOnVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.VirtualMachineExecuteCommandRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		params := mux.Vars(r)
		id := params["id"]

		if response, err := svc.ExecuteCommandOnVm(ctx, id, &request); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		} else {
			w.WriteHeader(http.StatusOK)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Command executed on machine: %v", id)
		}
	}
}

// @Summary		Uploads a file to a virtual machine
// @Description	This endpoint executes a command on a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id				path		string								true	"Machine ID"
// @Param			executeRequest	body		models.VirtualMachineUploadRequest	true	"Machine Upload file Command Request"
// @Success		200				{object}	models.VirtualMachineUploadResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/upload [post]
func UploadFileToVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.VirtualMachineUploadRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		params := mux.Vars(r)
		id := params["id"]

		if response, err := svc.LocalUploadToVm(ctx, id, &request); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		} else {
			w.WriteHeader(http.StatusOK)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Command executed on machine: %v", id)
		}
	}
}

// @Summary		Renames a virtual machine
// @Description	This endpoint Renames a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id				path		string								true	"Machine ID"
// @Param			renameRequest	body		models.RenameVirtualMachineRequest	true	"Machine Rename Request"
// @Success		200				{object}	models.ParallelsVM
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/rename [put]
func RenameVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.RenameVirtualMachineRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]
		request.ID = id

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := svc.RenameVm(ctx, request); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		vm, err := svc.GetVmSync(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if vm == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Machine %v not found", id),
				Code:    http.StatusNotFound,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(vm)
		ctx.LogInfof("Machine renamed: %v", id)
	}
}

// @Summary		Registers a virtual machine
// @Description	This endpoint registers a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id				path		string									true	"Machine ID"
// @Param			registerRequest	body		models.RegisterVirtualMachineRequest	true	"Machine Register Request"
// @Success		200				{object}	models.ParallelsVM
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/register [post]
func RegisterVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.RegisterVirtualMachineRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := svc.RegisterVm(ctx, request); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		filter := fmt.Sprintf("Home=%s/,i", request.Path)
		vms := []models.ParallelsVM{}
		var err error
		// To avoid race condition, retry fetching the VM from cache for up to 30 seconds
		retryTime := 30 * time.Second
		endTime := time.Now().Add(retryTime)
		for time.Now().Before(endTime) {
			vms, err = svc.GetCachedVms(ctx, filter)
			if err == nil && len(vms) > 0 {
				ctx.LogInfof("Vm found in cache after register: %v", vms[0].ID)
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		if len(vms) == 0 {
			ctx.LogInfof("Cache vm timeout reached, trying direct fetch")
			vms, err = svc.GetVms(ctx, filter)
		}
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(vms) == 0 {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Machine %v not found", request.Path),
				Code:    http.StatusNotFound,
			})
			return
		}

		if len(vms) != 1 {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Multiple machines found for %v", request.Path),
				Code:    http.StatusInternalServerError,
			})
			return
		}

		if request.MachineName != "" {
			if err := svc.RenameVm(ctx, models.RenameVirtualMachineRequest{
				ID:      vms[0].ID,
				NewName: request.MachineName,
			}); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}

			vms[0].Name = request.MachineName
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(vms[0])
		ctx.LogInfof("Machine registered: %v", vms[0].ID)
	}
}

// @Summary		Unregister a virtual machine
// @Description	This endpoint unregister a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id					path		string									true	"Machine ID"
// @Param			unregisterRequest	body		models.UnregisterVirtualMachineRequest	true	"Machine Unregister Request"
// @Success		200					{object}	models.ApiCommonResponse
// @Failure		400					{object}	models.ApiErrorResponse
// @Failure		401					{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/unregister [post]
func UnregisterVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.UnregisterVirtualMachineRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		params := mux.Vars(r)
		id := params["id"]
		request.ID = id

		if err := svc.UnregisterVm(ctx, request); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		ReturnApiCommonResponse(w)
		ctx.LogInfof("Machine unregistered: %v", id)
	}
}

// @Summary		Creates a virtual machine
// @Description	This endpoint creates a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			createRequest	body		models.CreateVirtualMachineRequest	true	"New Machine Request"
// @Success		200				{object}	models.CreateVirtualMachineResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines [post]
func CreateVirtualMachineHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request models.CreateVirtualMachineRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Attempt to get the architecture from the system
		if request.Architecture == "" {
			svcCtl := system.Get()
			arch, err := svcCtl.GetArchitecture(ctx)
			if err != nil {
				ReturnApiError(ctx, w, models.ApiErrorResponse{
					Message: "Failed to get architecture and none was provided",
					Code:    400,
				})
				return
			}
			request.Architecture = arch
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Decide which service to use to create the VM
		if request.PackerTemplate != nil {
			response, err := createPackerTemplate(ctx, request)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}

			w.WriteHeader(http.StatusOK)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Machine created using packer template: %v", response.ID)

		} else if request.VagrantBox != nil {
			response, err := createVagrantBox(ctx, request)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}

			w.WriteHeader(http.StatusOK)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Machine created using vagrant box: %v", response.ID)
			return
		} else if request.CatalogManifest != nil {
			response, err := createCatalogMachine(ctx, request)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}

			w.WriteHeader(http.StatusOK)
			defer r.Body.Close()
			_ = json.NewEncoder(w).Encode(response)
			ctx.LogInfof("Machine created using catalog: %v", response.ID)
			return
		} else {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: no template was specified",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}
}

// @Summary		Creates a snapshot for a virtual machine
// @Description	This endpoint creates a snapshot for a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id				path		string							true	"Machine ID"
// @Param			createRequest	body		models.CreateSnapShotRequest	true	"Create Snapshot Request"
// @Success		200				{object}	models.ApiCommonResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/snapshots [post]
func CreateSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.CreateSnapShotRequest
		ctx.LogInfof("Creating snapshot")

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			ctx.LogErrorf("Error decoding JSON: %v", err)
			return
		}

		// Extract ID from path
		params := mux.Vars(r)
		request.VMId = params["id"]

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// TODO: Call the service provider to actually create the snapshot using request

		ctx.LogDebugf("User: %v", request.VMId)
		ctx.LogDebugf("User: %v", request.VMName)
		ctx.LogDebugf("User: %v", request.SnapshotName)
		ctx.LogDebugf("User: %v", request.SnapshotDescription)
	}
}

// @Summary		Deletes a snapshot of a virtual machine
// @Description	This endpoint deletes a snapshot of a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id			path	string	true	"Machine ID"
// @Param			snapshot_id	path	string	true	"Snapshot ID"
// @Success		202
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/snapshots/{snapshot_id} [delete]
func DeleteSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.DeleteSnapshotRequest
		ctx.LogInfof("Deleting snapshot")

		// Extract ID from path (Standard for DELETE requests)
		params := mux.Vars(r)
		request.VMId = params["id"]
		request.ChildName = params["snapshot_id"]

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid path parameters: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// TODO: Call the service provider to actually delete the snapshot using request

		ctx.LogDebugf("VM ID: %v", request.VMId)
		ctx.LogDebugf("Snapshot to delete: %v", request.ChildName)
	}
}

// @Summary		Lists snapshots of a virtual machine
// @Description	This endpoint lists snapshots of a virtual machine
// @Tags			Machines
// @Produce		json
// @Param			id	path		string	true	"Machine ID"
// @Success		200	{object}	models.ApiCommonResponse
// @Failure		400	{object}	models.ApiErrorResponse
// @Failure		401	{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/snapshots [get]
func ListSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.ListSnapshotRequest
		ctx.LogInfof("Listing snapshots")

		// Extract ID from path (Standard for GET requests)
		params := mux.Vars(r)
		request.VMId = params["id"]

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid path parameters: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// TODO: Call the service provider to actually fetch the snapshot list using request

		ctx.LogDebugf("List snapshots for VM ID: %v", request.VMId)
	}
}

// @Summary		Reverts a virtual machine to a snapshot
// @Description	This endpoint reverts a virtual machine to a snapshot
// @Tags			Machines
// @Produce		json
// @Param			id				path		string							true	"Machine ID"
// @Param			snapshot_id		path		string							true	"Snapshot ID"
// @Param			revertRequest	body		models.RevertSnapshotRequest	false	"Revert Snapshot Request"
// @Success		200				{object}	models.ApiCommonResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/snapshots/{snapshot_id}/revert [post]
func RevertSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.RevertSnapshotRequest
		ctx.LogInfof("Reverting snapshot")

		// Extract IDs from path
		params := mux.Vars(r)
		request.VMId = params["id"]

		// If you also want to pass snapshot ID in body, that's fine,
		// but typically we get it from path for POST to a specific snapshot action:
		snapshotId := params["snapshot_id"]

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			// Ignore EOF error for cases where body might be empty if all info is in path
			if err.Error() != "EOF" {
				ctx.LogErrorf("Error decoding JSON: %v", err)
				return
			}
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// TODO: Call the service provider to actually revert the snapshot using request

		ctx.LogDebugf("VM ID: %v", request.VMId)
		ctx.LogDebugf("Reverting to: %v", snapshotId)
	}
}

// @Summary		Switches a virtual machine to a snapshot
// @Description	This endpoint switches a virtual machine to a snapshot
// @Tags			Machines
// @Produce		json
// @Param			id				path		string							true	"Machine ID"
// @Param			snapshot_id		path		string							true	"Snapshot ID"
// @Param			switchRequest	body		models.SwitchSnapshotRequest	false	"Switch Snapshot Request"
// @Success		200				{object}	models.ApiCommonResponse
// @Failure		400				{object}	models.ApiErrorResponse
// @Failure		401				{object}	models.OAuthErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/machines/{id}/snapshots/{snapshot_id}/switch [post]
func SwitchSnapshot() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)
		var request models.SwitchSnapshotRequest
		ctx.LogInfof("Switching snapshot")

		// Extract IDs from path
		params := mux.Vars(r)
		request.VMId = params["id"]

		snapshotId := params["snapshot_id"]

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			// Ignore EOF error for cases where body might be empty if all info is in path
			if err.Error() != "EOF" {
				ctx.LogErrorf("Error decoding JSON: %v", err)
				return
			}
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// TODO: Call the service provider to actually switch the snapshot using request

		ctx.LogDebugf("VM ID: %v", request.VMId)
		ctx.LogDebugf("Switching to: %v", snapshotId)
		ctx.LogDebugf("Skip Resume: %v", request.SkipResume)
	}
}

func createPackerTemplate(ctx basecontext.ApiContext, request models.CreateVirtualMachineRequest) (*models.CreateVirtualMachineResponse, error) {
	provider := serviceprovider.Get()
	config := config.Get()
	if !config.GetBoolKey(constants.ENABLE_VAGRANT_PLUGIN_ENV_VAR) {
		return nil, errors.NewWithCode("Vagrant plugin is not enabled, please enable it before trying", 400)
	}

	parallelsDesktopService := provider.ParallelsDesktopService

	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	template, err := dbService.GetPackerTemplate(ctx, request.PackerTemplate.Template)
	if err != nil {
		return nil, err
	}

	if template == nil {
		return nil, errors.NewWithCodef(404, "Template %v not found", request.PackerTemplate.Template)
	}

	template.Name = request.Name
	template.Owner = request.Owner
	if request.PackerTemplate.Specs != nil {
		if request.PackerTemplate.Specs.Cpu != "" {
			template.Specs["cpu"] = request.PackerTemplate.Specs.Cpu
		}
		if request.PackerTemplate.Specs.Memory != "" {
			template.Specs["memory"] = request.PackerTemplate.Specs.Memory
		}
	}

	vm, err := parallelsDesktopService.CreateVm(ctx, *template, request.PackerTemplate.DesiredState)
	if err != nil {
		return nil, err
	}

	response := models.CreateVirtualMachineResponse{
		ID:           vm.ID,
		Name:         vm.Name,
		Owner:        template.Owner,
		CurrentState: vm.State,
	}

	return &response, nil
}

func createVagrantBox(ctx basecontext.ApiContext, request models.CreateVirtualMachineRequest) (*models.CreateVirtualMachineResponse, error) {
	provider := serviceprovider.Get()
	config := config.Get()
	if !config.GetBoolKey(constants.ENABLE_VAGRANT_PLUGIN_ENV_VAR) {
		return nil, errors.NewWithCode("Vagrant plugin is not enabled, please enable it before trying", 400)
	}

	vagrantService := provider.VagrantService
	parallelsDesktopService := provider.ParallelsDesktopService

	// Updating plugins
	// if err := vagrantService.UpdatePlugins(request.Owner); err != nil {
	// 	ReturnApiError(ctx, w, models.NewFromError(err))
	// 	return
	// }

	if request.VagrantBox.Box != "" {
		// Generating the vagrant file
		if content, err := vagrantService.GenerateVagrantFile(ctx, *request.VagrantBox); err != nil {
			ctx.LogErrorf("Error generating vagrant file: %v", err)
			ctx.LogErrorf("Vagrant file content: %v", content)
			return nil, err
		}
	} else {
		if !helper.FileExists(request.VagrantBox.VagrantFilePath) {
			return nil, errors.NewWithCodef(400, "Vagrant file %v not found", request.VagrantBox.VagrantFilePath)
		}
	}

	// Creating the box
	if err := vagrantService.Up(ctx, *request.VagrantBox); err != nil {
		return nil, err
	}

	var response models.CreateVirtualMachineResponse
	if request.VagrantBox.Box == "" {
		response = models.CreateVirtualMachineResponse{
			Name:         request.Name,
			ID:           "unknown",
			CurrentState: "unknown",
			Owner:        request.Owner,
		}

		vm, err := parallelsDesktopService.GetVmSync(ctx, request.Name)
		if err != nil {
			return nil, err
		}

		if vm != nil {
			response.ID = vm.ID
			response.CurrentState = vm.State
			response.Name = vm.Name
			response.Owner = vm.User
		}

	} else {
		vm, err := parallelsDesktopService.GetVmSync(ctx, request.Name)
		if err != nil {
			return nil, err
		}

		if vm == nil {
			return nil, errors.NewWithCode("The machine was not found", 404)
		}

		response = models.CreateVirtualMachineResponse{
			Name:         vm.Name,
			ID:           vm.ID,
			CurrentState: vm.State,
			Owner:        vm.User,
		}
	}

	if response.CurrentState == "running" || response.CurrentState == "unknown" {
		if err := parallelsDesktopService.StopVm(ctx, response.ID, parallelsdesktop.NewDesiredStateFlags()); err != nil {
			return nil, err
		}
		response.CurrentState = "stopped"
	}

	if request.VagrantBox.Specs != nil {
		configureRequest := models.VirtualMachineConfigRequest{
			Operations: make([]*models.VirtualMachineConfigRequestOperation, 0),
		}

		if request.VagrantBox.Specs.Cpu != "" {
			configureRequest.Operations = append(configureRequest.Operations, &models.VirtualMachineConfigRequestOperation{
				Group:     "cpu",
				Operation: "set",
				Value:     request.VagrantBox.Specs.Cpu,
			})
		}

		if request.VagrantBox.Specs.Memory != "" {
			configureRequest.Operations = append(configureRequest.Operations, &models.VirtualMachineConfigRequestOperation{
				Group:     "memory",
				Operation: "set",
				Value:     request.VagrantBox.Specs.Memory,
			})
		}

		if err := parallelsDesktopService.ConfigureVm(ctx, response.ID, &configureRequest); err != nil {
			return nil, err
		}
	}

	if request.StartOnCreate && response.CurrentState == "stopped" {
		err := parallelsDesktopService.StartVm(ctx, response.ID)
		if err != nil {
			return nil, err
		}
		response.CurrentState = "running"
	}

	return &response, nil
}

func createCatalogMachine(ctx basecontext.ApiContext, request models.CreateVirtualMachineRequest) (*models.CreateVirtualMachineResponse, error) {
	provider := serviceprovider.Get()

	parallelsDesktopService := provider.ParallelsDesktopService

	var response models.CreateVirtualMachineResponse
	pullRequest := mappers.MapPullCatalogManifestRequestFromCreateCatalogVirtualMachineRequest(*request.CatalogManifest)
	if pullRequest.Architecture == "" {
		pullRequest.Architecture = request.Architecture
	}

	pullRequest.StartAfterPull = request.StartOnCreate

	if err := pullRequest.Validate(); err != nil {
		return nil, err
	}
	err := serviceprovider.TestCacheFolderAccess(ctx)
	if err != nil {
		ctx.LogErrorf("Cache folder access test failed: %v", err)
		return nil, err
	}

	manifest := catalog.NewManifestService(ctx)
	resultManifest := manifest.Pull(&pullRequest)
	if resultManifest.HasErrors() {
		errorMessage := "Error pulling manifest: \n"
		for _, err := range resultManifest.Errors {
			errorMessage += "\n" + err.Error() + " "
		}

		return nil, errors.NewWithCode(errorMessage, 400)
	}

	resultData := mappers.BasePullCatalogManifestResponseToApi(*resultManifest)
	resultData.ID = resultManifest.ID

	response = models.CreateVirtualMachineResponse{
		Name:  resultData.MachineName,
		ID:    resultData.MachineID,
		Owner: request.Owner,
	}

	vm, err := parallelsDesktopService.GetVmSync(ctx, response.ID)
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, errors.NewWithCode("The machine was not found", 404)
	}

	response.CurrentState = vm.State

	if request.CatalogManifest.Specs != nil {

		if response.CurrentState == "running" || response.CurrentState == "unknown" {
			if err := parallelsDesktopService.StopVm(ctx, response.ID, parallelsdesktop.NewDesiredStateFlags("--drop-state")); err != nil {
				return nil, err
			}
			response.CurrentState = "stopped"
		}

		configureRequest := models.VirtualMachineConfigRequest{
			Operations: make([]*models.VirtualMachineConfigRequestOperation, 0),
		}

		if request.CatalogManifest.Specs.Cpu != "" {
			configureRequest.Operations = append(configureRequest.Operations, &models.VirtualMachineConfigRequestOperation{
				Group:     "cpu",
				Operation: "set",
				Value:     request.CatalogManifest.Specs.Cpu,
			})
		}

		if request.CatalogManifest.Specs.Memory != "" {
			configureRequest.Operations = append(configureRequest.Operations, &models.VirtualMachineConfigRequestOperation{
				Group:     "memory",
				Operation: "set",
				Value:     request.CatalogManifest.Specs.Memory,
			})
		}

		if err := parallelsDesktopService.ConfigureVm(ctx, response.ID, &configureRequest); err != nil {
			return nil, err
		}
	}

	if request.StartOnCreate && response.CurrentState == "stopped" {
		err := parallelsDesktopService.StartVm(ctx, response.ID)
		if err != nil {
			return nil, err
		}
		response.CurrentState = "running"
	}

	return &response, nil
}
