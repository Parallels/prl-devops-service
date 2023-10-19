package startup

import (
	"Parallels/pd-api-service/controllers"

	"github.com/gorilla/mux"
)

func InitControllers() *mux.Router {
	// Create a new router
	r := mux.NewRouter()

	// Define a handler function for the /users endpoint
	r.HandleFunc("/users", controllers.ListUsers()).Methods("GET")
	r.HandleFunc("/users", controllers.CreateUser()).Methods("POST")
	r.HandleFunc("/users/token", controllers.LoginUser()).Methods("POST")
	r.HandleFunc("/users/token/validate", controllers.ValidateTokenHandler()).Methods("POST")
	r.HandleFunc("/users/{id}", controllers.GetUserByID()).Methods("GET")

	r.HandleFunc("/machines", controllers.GetMachinesController()).Methods("GET")
	r.HandleFunc("/machines", controllers.CreateMachine()).Methods("POST")
	r.HandleFunc("/machines/{id}", controllers.DeleteMachineController()).Methods("DELETE")
	r.HandleFunc("/machines/{id}", controllers.GetMachineController()).Methods("GET")
	r.HandleFunc("/machines/{id}/start", controllers.StartMachineController()).Methods("GET")
	r.HandleFunc("/machines/{id}/stop", controllers.StopMachineController()).Methods("GET")
	r.HandleFunc("/machines/{id}/restart", controllers.RestartMachineController()).Methods("GET")
	r.HandleFunc("/machines/{id}/suspend", controllers.SuspendMachineController()).Methods("GET")
	r.HandleFunc("/machines/{id}/reset", controllers.ResetMachineController()).Methods("GET")
	r.HandleFunc("/machines/{id}/pause", controllers.PauseMachineController()).Methods("GET")
	r.HandleFunc("/machines/{id}/status", controllers.StatusMachineController()).Methods("GET")

	r.HandleFunc("/templates/virtual_machines", controllers.GetVirtualMachinesTemplatesController()).Methods("GET")
	r.HandleFunc("/templates/virtual_machines/{name}", controllers.GetVirtualMachineTemplateController()).Methods("GET")
	return r
}
