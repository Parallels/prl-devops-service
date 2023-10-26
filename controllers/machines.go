package controllers

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/restapi"
	"Parallels/pd-api-service/service_provider"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

// LoginUser is a public function that logs in a user
func GetMachinesController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService
		filter := ""
		var err error
		var result []models.ParallelsVM
		if r.Header.Get("X-Filter") != "" {
			filter = r.Header.Get("X-Filter")
		}
		if filter == "" {
			common.Logger.Info("Getting unfiltered machines")
			result, err = svc.GetVms()
			if err != nil {
				ReturnApiError(w, models.NewFromError(err))
				return
			}
		} else {
			common.Logger.Info("Getting filtered machines")
			result, err = svc.GetFilteredVm(filter)
			if err != nil {
				ReturnApiError(w, models.NewFromError(err))
				return
			}
		}

		if result == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			result = make([]models.ParallelsVM, 0)
			json.NewEncoder(w).Encode(result)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func GetMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService

		params := mux.Vars(r)
		id := params["id"]

		result, err := svc.GetVm(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		if result == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Machine %v not found", id),
				Code:    http.StatusNotFound,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func StartMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.StartVm(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Start",
			Status:    "Success",
		}
		json.NewEncoder(w).Encode(result)
	}
}

func StopMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.StopVm(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Stop",
			Status:    "Success",
		}

		json.NewEncoder(w).Encode(result)
	}
}

func RestartMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.RestartVm(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Restart",
			Status:    "Success",
		}
		json.NewEncoder(w).Encode(result)
	}
}

func SuspendMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.SuspendVm(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Suspend",
			Status:    "Success",
		}
		json.NewEncoder(w).Encode(result)
	}
}

func ResumeMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.ResumeVm(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func ResetMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.ResetVm(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Reset",
			Status:    "Success",
		}
		json.NewEncoder(w).Encode(result)
	}
}

func PauseMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.ResetVm(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		result := models.VirtualMachineOperationResponse{
			ID:        id,
			Operation: "Pause",
			Status:    "Success",
		}

		json.NewEncoder(w).Encode(result)
	}
}

func DeleteMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService

		params := mux.Vars(r)
		id := params["id"]

		err := svc.DeleteVm(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func StatusMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().ParallelsService

		params := mux.Vars(r)
		id := params["id"]

		status, err := svc.VmStatus(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		result := models.VirtualMachineStatusResponse{
			ID:     id,
			Status: status,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func SetMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models.VirtualMachineConfigRequest
		svc := service_provider.Get().ParallelsService
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		params := mux.Vars(r)
		id := params["id"]

		if err := svc.ConfigureVm(id, &request); err != nil {
			ReturnApiError(w, models.NewFromError(err))
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
	}
}

func ExecuteCommandOnMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models.VirtualMachineExecuteCommandRequest
		svc := service_provider.Get().ParallelsService
		http_helper.MapRequestBody(r, &request)
		if request.Command == "" {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Invalid request body: command cannot be empty",
				Code:    http.StatusBadRequest,
			})
			return
		}

		params := mux.Vars(r)
		id := params["id"]

		if response, err := svc.ExecuteCommandOnVm(id, &request); err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}
}

func RenameMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models.RenameVirtualMachineRequest
		svc := service_provider.Get().ParallelsService

		http_helper.MapRequestBody(r, &request)
		params := mux.Vars(r)
		id := params["id"]
		request.ID = id

		if err := request.Validate(); err != nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := svc.RenameVm(request); err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		vm, err := svc.GetVm(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		if vm == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Machine %v not found", id),
				Code:    http.StatusNotFound,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vm)
	}
}

func RegisterMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models.RegisterVirtualMachineRequest
		svc := service_provider.Get().ParallelsService
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := svc.RegisterVm(request); err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		filter := fmt.Sprintf("Home=%s/", request.Path)
		vms, err := svc.GetFilteredVm(filter)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}
		if len(vms) == 0 {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Machine %v not found", request.Path),
				Code:    http.StatusNotFound,
			})
			return
		}
		if len(vms) != 1 {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Multiple machines found for %v", request.Path),
				Code:    http.StatusInternalServerError,
			})
			return
		}
		if request.MachineName != "" {
			if err := svc.RenameVm(models.RenameVirtualMachineRequest{
				ID:      vms[0].ID,
				NewName: request.MachineName,
			}); err != nil {
				ReturnApiError(w, models.NewFromError(err))
				return
			}

			vms[0].Name = request.MachineName
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(vms[0])
	}
}

func UnregisterMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models.UnregisterVirtualMachineRequest
		svc := service_provider.Get().ParallelsService

		http_helper.MapRequestBody(r, &request)
		params := mux.Vars(r)
		id := params["id"]
		request.ID = id

		if err := request.Validate(); err != nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := svc.UnregisterVm(request); err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		ReturnApiCommonResponse(w)
	}
}

func CreateMachine() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// svc := services.ParallelsService{}
		// downloadSvc := services.DownloadService{}
		// url := "https://releases.ubuntu.com/jammy/ubuntu-22.04.3-desktop-amd64.iso"
		// filename := "ubuntu-22.04.3-desktop-amd64.iso"

		// err := downloadSvc.DownloadFile(url, filename)
		// if err != nil {
		// 	http.Error(w, "Machine not found", http.StatusInternalServerError)
		// 	return
		// }

		// params := mux.Vars(r)
		// id := params["id"]

		// result := svc.StopVm(id)
		// if !result {
		//   http.Error(w, "Machine not found", http.StatusInternalServerError)
		//   return
		// }

		var request models.CreateVirtualMachineRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			http.Error(w, "No database connection", http.StatusInternalServerError)
			return
		}

		if err := dbService.Connect(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer dbService.Disconnect()

		template, err := dbService.GetVirtualMachineTemplate(request.Template)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		if template == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Template %v not found", request.Template),
				Code:    http.StatusNotFound,
			})
			return
		}

		template.Name = request.Name
		template.Owner = request.Owner
		template.Specs["memory"] = request.Memory
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		svc := service_provider.Get().ParallelsService
		vm, err := svc.CreateVm(*template, request.DesiredState)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
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
	}
}
