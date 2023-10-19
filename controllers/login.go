package controllers

import (
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/services"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// LoginUser is a public function that logs in a user
func LoginUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var login models.ApiLogin
		err := json.NewDecoder(r.Body).Decode(&login)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Connect to the SQL server
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

		defer dbService.Disconnect()

		user, err := dbService.GetUser(login.Email)

		if err != nil || user == nil {
			http.Error(w, "Invalid email or password", http.StatusForbidden)
			return
		}

		// Hash the password with SHA-256
		hashedPassword := sha256.Sum256([]byte(login.Password))
		if hex.EncodeToString(hashedPassword[:]) != user.Password {
			http.Error(w, "Invalid email or password", http.StatusForbidden)
			return
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email": login.Email,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})

		// Sign the token with HMAC
		key := []byte(os.Getenv("JWT_SECRET"))
		tokenString, err := token.SignedString(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return the JWT token
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	}
}
