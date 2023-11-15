package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Parallels/pd-api-service/models"
	"github.com/Parallels/pd-api-service/restapi"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

//	@Summary		Gets all the virtual machines
//	@Description	This endpoint returns all the virtual machines
//	@Tags			Machines
//	@Produce		json
//	@Param			filter	header		string	false	"X-Filter"
//	@Success		200		{object}	[]models.ParallelsVM
//	@Failure		400		{object}	models.ApiErrorResponse
//	@Failure		401		{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines [get]
func GetMachinesController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		vms, err := svc.GetVms(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(vms) == 0 {
			w.WriteHeader(http.StatusOK)
			vms = make([]models.ParallelsVM, 0)
			json.NewEncoder(w).Encode(vms)
			ctx.LogInfo("No machines found")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vms)
		ctx.LogInfo("Machines returned: %v", len(vms))
	}
}

//	@Summary		Gets a virtual machine
//	@Description	This endpoint returns a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id	path		string	true	"Machine ID"
//	@Success		200	{object}	models.ParallelsVM
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id} [get]
func GetMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		vm, err := svc.GetVm(ctx, id)
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
		json.NewEncoder(w).Encode(vm)
		ctx.LogInfo("Machine returned: %v", vm.ID)
	}
}

//	@Summary		Starts a virtual machine
//	@Description	This endpoint starts a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id	path		string	true	"Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/start [get]
func StartMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
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

		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Machine started: %v", id)
	}
}

//	@Summary		Stops a virtual machine
//	@Description	This endpoint stops a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id	path		string	true	"Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/stop [get]
func StopMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.StopVm(ctx, id)
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

		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Machine stopped: %v", id)
	}
}

//	@Summary		Restarts a virtual machine
//	@Description	This endpoint restarts a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id	path		string	true	"Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/restart [get]
func RestartMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
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
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Machine restarted: %v", id)
	}
}

//	@Summary		Suspends a virtual machine
//	@Description	This endpoint suspends a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id	path		string	true	"Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/suspend [get]
func SuspendMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
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
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Machine suspended: %v", id)
	}
}

//	@Summary		Resumes a virtual machine
//	@Description	This endpoint resumes a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id	path		string	true	"Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/resume [get]
func ResumeMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
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
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Machine resumed: %v", id)
	}
}

//	@Summary		Reset a virtual machine
//	@Description	This endpoint reset a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id	path		string	true	"Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/reset [get]
func ResetMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
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
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Machine reset: %v", id)
	}
}

//	@Summary		Pauses a virtual machine
//	@Description	This endpoint pauses a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id	path		string	true	"Machine ID"
//	@Success		200	{object}	models.VirtualMachineOperationResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/pause [get]
func PauseMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
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
			Operation: "Pause",
			Status:    "Success",
		}

		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Machine paused: %v", id)
	}
}

//	@Summary		Deletes a virtual machine
//	@Description	This endpoint deletes a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id	path	string	true	"Machine ID"
//	@Success		202
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id} [delete]
func DeleteMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
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
		ctx.LogInfo("Machine deleted: %v", id)
	}
}

//	@Summary		Get the current state of a virtual machine
//	@Description	This endpoint returns the current state of a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id	path		string	true	"Machine ID"
//	@Success		200	{object}	models.VirtualMachineStatusResponse
//	@Failure		400	{object}	models.ApiErrorResponse
//	@Failure		401	{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/status [post]
func StatusMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		params := mux.Vars(r)
		id := params["id"]

		status, err := svc.VmStatus(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		result := models.VirtualMachineStatusResponse{
			ID:     id,
			Status: status,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Machine status returned: %v", id)
	}
}

//	@Summary		Configures a virtual machine
//	@Description	This endpoint configures a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id				path		string								true	"Machine ID"
//	@Param			configRequest	body		models.VirtualMachineConfigRequest	true	"Machine Set Request"
//	@Success		200				{object}	models.VirtualMachineConfigResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/set [post]
func SetMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.VirtualMachineConfigRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		http_helper.MapRequestBody(r, &request)
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
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Machine configured: %v", id)
	}
}

//	@Summary		Executes a command on a virtual machine
//	@Description	This endpoint executes a command on a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id				path		string										true	"Machine ID"
//	@Param			executeRequest	body		models.VirtualMachineExecuteCommandRequest	true	"Machine Execute Command Request"
//	@Success		200				{object}	models.VirtualMachineExecuteCommandResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/execute [post]
func ExecuteCommandOnMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.VirtualMachineExecuteCommandRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService
		http_helper.MapRequestBody(r, &request)
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
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Command executed on machine: %v", id)
		}
	}
}

//	@Summary		Renames a virtual machine
//	@Description	This endpoint Renames a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id				path		string								true	"Machine ID"
//	@Param			renameRequest	body		models.RenameVirtualMachineRequest	true	"Machine Rename Request"
//	@Success		200				{object}	models.ParallelsVM
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/rename [post]
func RenameMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.RenameVirtualMachineRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		http_helper.MapRequestBody(r, &request)
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

		if err := svc.RenameVm(ctx, request); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		vm, err := svc.GetVm(ctx, id)
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
		json.NewEncoder(w).Encode(vm)
		ctx.LogInfo("Machine renamed: %v", id)
	}
}

//	@Summary		Registers a virtual machine
//	@Description	This endpoint registers a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id				path		string									true	"Machine ID"
//	@Param			registerRequest	body		models.RegisterVirtualMachineRequest	true	"Machine Register Request"
//	@Success		200				{object}	models.ParallelsVM
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/register [post]
func RegisterMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.RegisterVirtualMachineRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		http_helper.MapRequestBody(r, &request)
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
		vms, err := svc.GetVms(ctx, filter)
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
		json.NewEncoder(w).Encode(vms[0])
		ctx.LogInfo("Machine registered: %v", vms[0].ID)
	}
}

//	@Summary		Unregisters a virtual machine
//	@Description	This endpoint unregisters a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			id					path		string									true	"Machine ID"
//	@Param			unregisterRequest	body		models.UnregisterVirtualMachineRequest	true	"Machine Unregister Request"
//	@Success		200					{object}	models.ApiCommonResponse
//	@Failure		400					{object}	models.ApiErrorResponse
//	@Failure		401					{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines/{id}/unregister [post]
func UnregisterMachineController() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.UnregisterVirtualMachineRequest
		provider := serviceprovider.Get()
		svc := provider.ParallelsDesktopService

		http_helper.MapRequestBody(r, &request)
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
		ctx.LogInfo("Machine unregistered: %v", id)
	}
}

//	@Summary		Creates a virtual machine
//	@Description	This endpoint creates a virtual machine
//	@Tags			Machines
//	@Produce		json
//	@Param			createRequest	body		models.CreateVirtualMachineRequest	true	"New Machine Request"
//	@Success		200				{object}	models.CreateVirtualMachineResponse
//	@Failure		400				{object}	models.ApiErrorResponse
//	@Failure		401				{object}	models.OAuthErrorResponse
//	@Security		ApiKeyAuth
//	@Security		BearerAuth
//	@Router			/v1/machines [post]
func CreateMachine() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		provider := serviceprovider.Get()

		var request models.CreateVirtualMachineRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if request.PackerTemplate != nil {
			dbService, err := GetDatabaseService(ctx)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
				return
			}

			defer dbService.Disconnect(ctx)

			template, err := dbService.GetPackerTemplate(ctx, request.PackerTemplate.Template)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}

			if template == nil {
				ReturnApiError(ctx, w, models.ApiErrorResponse{
					Message: fmt.Sprintf("Template %v not found", request.PackerTemplate.Template),
					Code:    http.StatusNotFound,
				})
				return
			}

			template.Name = request.Name
			template.Owner = request.Owner
			if request.PackerTemplate.Cpu != "" {
				template.Specs["cpu"] = request.PackerTemplate.Cpu
			}
			if request.PackerTemplate.Memory != "" {
				template.Specs["memory"] = request.PackerTemplate.Memory
			}

			parallelsDesktopService := provider.ParallelsDesktopService
			vm, err := parallelsDesktopService.CreateVm(ctx, *template, request.PackerTemplate.DesiredState)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}

			response := models.CreateVirtualMachineResponse{
				ID:           vm.ID,
				Name:         vm.Name,
				Owner:        template.Owner,
				CurrentState: vm.State,
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Machine created using packer template: %v", vm.ID)
		} else if request.VagrantBox != nil {
			vagrantService := provider.VagrantService
			parallelsDesktopService := provider.ParallelsDesktopService

			// Updating plugins
			if err := vagrantService.UpdatePlugins(request.Owner); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}

			// Generating the vagrant file
			if content, err := vagrantService.GenerateVagrantFile(ctx, *request.VagrantBox); err != nil {
				ctx.LogError("Error generating vagrant file: %v", err)
				ctx.LogError("Vagrant file content: %v", content)
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}

			// Creating the box
			if err := vagrantService.Up(ctx, *request.VagrantBox); err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}

			vm, err := parallelsDesktopService.GetVm(ctx, request.Name)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}

			if vm == nil {
				ReturnApiError(ctx, w, models.ApiErrorResponse{
					Message: "The machine was not found",
					Code:    http.StatusNotFound,
				})
				return
			}

			response := models.CreateVirtualMachineResponse{
				Name:         vm.Name,
				ID:           vm.ID,
				CurrentState: vm.State,
				Owner:        vm.User,
			}

			// Write the JSON data to the response
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Machine created using vagrant box: %v", vm.ID)
		} else {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: no template was specified",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}
}
