package controllers

import (
	data_modules "Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/restapi"
	"Parallels/pd-api-service/service_provider"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// LoginUser is a public function that logs in a user
func GetVirtualMachinesTemplatesController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().JsonDatabase

		err := svc.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer svc.Disconnect()

		result, err := svc.GetVirtualMachineTemplates()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		if result == nil {
			w.WriteHeader(http.StatusOK)
			result = make([]data_modules.VirtualMachineTemplate, 0)
			json.NewEncoder(w).Encode(result)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func GetVirtualMachineTemplateController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := service_provider.Get().JsonDatabase

		err := svc.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer svc.Disconnect()

		params := mux.Vars(r)
		name := params["name"]

		result, err := svc.GetVirtualMachineTemplate(name)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		if result == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: fmt.Sprintf("Machine %v not found", name),
				Code:    http.StatusNotFound,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}
