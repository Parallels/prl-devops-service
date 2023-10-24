package controllers

import (
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/restapi"
	"Parallels/pd-api-service/services"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

// LoginUser is a public function that logs in a user
func GetMachinesController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := services.GetServices().ParallelsService
		filter := ""
		var err error
		result := make([]models.ParallelsVM, 0)
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
		svc := services.GetServices().ParallelsService

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
		svc := services.GetServices().ParallelsService

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
		svc := services.GetServices().ParallelsService

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
		svc := services.GetServices().ParallelsService

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
		svc := services.GetServices().ParallelsService

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
		svc := services.GetServices().ParallelsService

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
		svc := services.GetServices().ParallelsService

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
		svc := services.GetServices().ParallelsService

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
		svc := services.GetServices().ParallelsService

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
		svc := services.GetServices().ParallelsService

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
		var request models.VirtualMachineSetRequest
		svc := services.GetServices().ParallelsService
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

		if err := svc.ConfigVmSetRequest(id, &request); err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		result := models.VirtualMachineSetResponse{
			Operations: make([]models.VirtualMachineSetOperationResponse, 0),
		}
		for _, op := range request.Operations {
			rOp := models.VirtualMachineSetOperationResponse{
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
		svc := services.GetServices().ParallelsService
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

		dbService := services.GetServices().JsonDatabase
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

		svc := services.GetServices().ParallelsService
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

func PushRemoteMachineController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var request models.PushRemoteMachineRequest
		// svc := services.GetServices().ParallelsService
		http_helper.MapRequestBody(r, &request)
		if request.LocalPath == "" {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Invalid request body: local path cannot be empty",
				Code:    http.StatusBadRequest,
			})
			return
		}

		manifest := services.NewManifestService()
		result, err := manifest.GenerateManifest(request.Name, request.LocalPath)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// params := mux.Vars(r)
		// id := params["id"]

		// bucket := aws.S3Bucket{
		// 	Name:      "carlos-pd-machines",
		// 	Region:    "us-east-2",
		// 	AccessKey: "AKIA56JTTF6DLXWTRFUJ",
		// 	SecretKey: "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",
		// }
		// svc := aws.NewAwsS3Service(bucket)

		// if err := svc.UploadFile("/Users/cjlapao/poc-api-service"); err != nil {
		// 	ReturnApiError(w, models.NewFromError(err))
		// 	return
		// } else {
		// 	w.WriteHeader(http.StatusOK)
		// 	json.NewEncoder(w).Encode("ok")
		// }

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}
