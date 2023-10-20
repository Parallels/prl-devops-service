package controllers

import (
	data_models "Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/mappers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/restapi"
	"Parallels/pd-api-service/services"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// GetUsers is a public function that returns all users
func GetApiKeysController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := services.GetServices().JsonDatabase
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

		apiKeys, err := dbService.GetApiKeys()
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		result := make([]models.ApiKey, 0)
		for _, apiKey := range apiKeys {
			result = append(result, mappers.ApiKeyFromDTO(apiKey))
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func DeleteApiKeyController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := services.GetServices().JsonDatabase
		if dbService == nil {
			http.Error(w, "No database connection", http.StatusInternalServerError)
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		defer dbService.Disconnect()

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.RemoveKey(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusAccepted)
	}
}

// GetUserByID is a public function that returns a user by ID
func GetApiKeyByIdOrNameController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := services.GetServices().JsonDatabase
		if dbService == nil {
			http.Error(w, "No database connection", http.StatusInternalServerError)
			return
		}

		err := dbService.Connect()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		// Query the users table for the apiKey with the given ID
		apiKey, err := dbService.GetApiKey(id)

		if err != nil || apiKey == nil {
			http.Error(w, "Api Key not found", http.StatusNotFound)
			return
		}

		resultUser := mappers.ApiKeyFromDTO(*apiKey)

		// Marshal the user struct to JSON
		jsonData, err := json.Marshal(resultUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func CreateApiKeyController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var apiKey data_models.ApiKey
		err := json.NewDecoder(r.Body).Decode(&apiKey)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Set the ID to 0 to ensure that a new ID is generated
		apiKey.ID = helpers.GenerateId()
		if apiKey.Secret == "" {
			http.Error(w, "Api Key Secret cannot be null", http.StatusInternalServerError)
			return
		}
		if apiKey.Key == "" {
			http.Error(w, "Api Key cannot be null", http.StatusInternalServerError)
			return
		}
		apiKey.Key = strings.ToUpper(strings.ReplaceAll(apiKey.Key, " ", "_"))

		if apiKey.Name == "" {
			http.Error(w, "Api Key Name cannot be null", http.StatusInternalServerError)
			return
		}

		dbService := services.GetServices().JsonDatabase
		if dbService == nil {
			http.Error(w, "No database connection", http.StatusInternalServerError)
			return
		}
		err = dbService.Connect()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = dbService.CreateApiKey(&apiKey)
		response := mappers.ApiKeyFromDTO(apiKey)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		dbService.Disconnect()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

func RevokeApiKeyController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		// Connect to the SQL server
		dbService := services.GetServices().JsonDatabase
		if dbService == nil {
			http.Error(w, "No database connection", http.StatusInternalServerError)
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		err := dbService.Connect()
		if err != nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "No database connection",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		defer dbService.Disconnect()

		// Get the user ID from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.RevokeKey(id)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Write the JSON data to the response
		w.WriteHeader(http.StatusAccepted)
	}
}
