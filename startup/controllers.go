package startup

import (
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/controllers"
	"fmt"

	"github.com/gorilla/mux"
)

func InitControllers() *mux.Router {
	// Create a new router
	r := mux.NewRouter()

	// Define a handler function for the /users endpoint
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/users"), controllers.ListUsers()).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/users"), controllers.CreateUser()).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/users/token"), controllers.LoginUser()).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/users/token/validate"), controllers.ValidateTokenHandler()).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/users/{id}"), controllers.GetUserByID()).Methods("GET")

	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines"), controllers.GetMachinesController()).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines"), controllers.CreateMachine()).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines/{id}"), controllers.DeleteMachineController()).Methods("DELETE")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines/{id}"), controllers.GetMachineController()).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines/{id}/start"), controllers.StartMachineController()).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines/{id}/stop"), controllers.StopMachineController()).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines/{id}/restart"), controllers.RestartMachineController()).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines/{id}/suspend"), controllers.SuspendMachineController()).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines/{id}/reset"), controllers.ResetMachineController()).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines/{id}/pause"), controllers.PauseMachineController()).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/machines/{id}/status"), controllers.StatusMachineController()).Methods("GET")

	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/templates/virtual_machines"), controllers.GetVirtualMachinesTemplatesController()).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s%s", constants.API_PREFIX, "/templates/virtual_machines/{name}"), controllers.GetVirtualMachineTemplateController()).Methods("GET")
	return r
}
