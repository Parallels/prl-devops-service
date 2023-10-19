package controllers

import (
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/services"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// GetUsers is a public function that returns all users
func ListUsers() http.HandlerFunc {
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

		users, err := dbService.GetUsers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var jsonData []byte

		if users == nil {
			// Marshal an empty slice to JSON
			jsonData, err = json.Marshal(make([]models.User, 0))
		} else {
			// Marshal the users slice to JSON
			jsonData, err = json.Marshal(users)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header to application/json
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON data to the response
		w.Write(jsonData)
	}
}

// GetUserByID is a public function that returns a user by ID
func GetUserByID() http.HandlerFunc {
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
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Query the users table for the user with the given ID
		user, err := dbService.GetUser(id)

		if err != nil || user == nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Marshal the user struct to JSON
		jsonData, err := json.Marshal(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header to application/json
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON data to the response
		w.Write(jsonData)
	}
}

func CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Hash the password with SHA-256
		hashedPassword := sha256.Sum256([]byte(user.Password))
		user.Password = hex.EncodeToString(hashedPassword[:])

		// Set the ID to 0 to ensure that a new ID is generated
		user.ID = helpers.GenerateId()

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

		err = dbService.CreateUser(&user)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		dbService.Disconnect()

		// Return the created user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}
