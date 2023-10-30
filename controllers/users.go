package controllers

import (
	data_models "Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/mappers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/restapi"
	"Parallels/pd-api-service/service_provider"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// GetUsers is a public function that returns all users
func GetUsersController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer dbService.Disconnect()

		users, err := dbService.GetUsers()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		responseUsers := make([]models.User, 0)
		for _, user := range users {
			responseUsers = append(responseUsers, mappers.UserFromDTO(user))
		}

		jsonData, err := json.Marshal(responseUsers)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

// GetUserController is a public function that returns a user by ID
func GetUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer dbService.Disconnect()

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Query the users table for the user with the given ID
		user, err := dbService.GetUser(id)

		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}
		if user == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "User not found",
				Code:    http.StatusNotFound,
			})
			return
		}

		resultUser := mappers.UserFromDTO(*user)

		// Marshal the user struct to JSON
		jsonData, err := json.Marshal(resultUser)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func CreateUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var user data_models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Set the ID to 0 to ensure that a new ID is generated
		user.ID = helpers.GenerateId()

		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}
		err = dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		dbUser, err := dbService.CreateUser(&user)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}
		if dbUser == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "User not found",
				Code:    http.StatusNotFound,
			})
			return
		}

		response := mappers.UserFromDTO(*dbUser)
		dbService.Disconnect()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetUserByID is a public function that returns a user by ID
func DeleteUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer dbService.Disconnect()

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		err = dbService.RemoveUser(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusAccepted)
	}
}

func UpdateUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var user data_models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err = dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		user.ID = id
		err = dbService.UpdateUser(user)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		dbService.Disconnect()

		w.WriteHeader(http.StatusAccepted)
	}
}

func GetUserRolesController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer dbService.Disconnect()

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Query the users table for the user with the given ID
		user, err := dbService.GetUser(id)

		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}
		if user == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "User not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		roles := user.Roles
		if roles == nil {
			roles = make([]data_models.Role, 0)
		}

		// Marshal the user struct to JSON
		jsonData, err := json.Marshal(roles)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func AddRoleToUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var role data_models.Role
		err := json.NewDecoder(r.Body).Decode(&role)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}
		err = dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		if err := dbService.AddRoleToUser(id, role.Name); err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		response := mappers.RoleFromDTO(role)
		dbService.Disconnect()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetUserByID is a public function that returns a user by ID
func RemoveRoleFromUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer dbService.Disconnect()

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]
		roleId := vars["role_id"]

		if err = dbService.RemoveRoleFromUser(id, roleId); err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusAccepted)
	}
}

func GetUserClaimsController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer dbService.Disconnect()

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Query the users table for the user with the given ID
		user, err := dbService.GetUser(id)

		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}
		if user == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "User not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		claims := user.Claims
		if claims == nil {
			claims = make([]data_models.Claim, 0)
		}

		// Marshal the user struct to JSON
		jsonData, err := json.Marshal(claims)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func AddClaimToUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var claim data_models.Claim
		err := json.NewDecoder(r.Body).Decode(&claim)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}
		err = dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		if err := dbService.AddClaimToUser(id, claim.Name); err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		response := mappers.ClaimFromDTO(claim)
		dbService.Disconnect()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetUserByID is a public function that returns a user by ID
func RemoveClaimFromUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := service_provider.Get().JsonDatabase
		if dbService == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		defer dbService.Disconnect()

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]
		roleId := vars["role_id"]

		if err = dbService.RemoveRoleFromUser(id, roleId); err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusAccepted)
	}
}
