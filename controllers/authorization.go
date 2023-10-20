package controllers

import (
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/restapi"
	"Parallels/pd-api-service/services"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GetTokenController is a public function that logs in a user
func GetTokenController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var login models.ApiLogin
		err := json.NewDecoder(r.Body).Decode(&login)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		// Connect to the SQL server
		dbService := services.GetServices().JsonDatabase
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

		defer dbService.Disconnect()

		user, err := dbService.GetUser(login.Email)

		if err != nil || user == nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		if user == nil {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "User not found",
				Code:    http.StatusUnauthorized,
			})
			return
		}

		// Hash the password with SHA-256
		hashedPassword := helpers.Sha256Hash(login.Password)
		if hashedPassword != user.Password {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Invalid Password",
				Code:    http.StatusUnauthorized,
			})
			return
		}
		roles := make([]string, 0)
		claims := make([]string, 0)
		for _, role := range user.Roles {
			roles = append(roles, role.Name)
		}
		for _, claim := range user.Claims {
			claims = append(claims, claim.Name)
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email":  login.Email,
			"roles":  roles,
			"claims": claims,
			"exp":    time.Now().Add(time.Minute * constants.TOKEN_DURATION_MINUTES).Unix(),
		})

		// Sign the token with HMAC
		key := []byte(services.GetServices().HardwareSecret)
		tokenString, err := token.SignedString(key)
		if err != nil {
			ReturnApiError(w, models.NewFromErrorWithCode(err, 401))
			return
		}

		// Return the JWT token
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	}
}

func ValidateTokenController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var tokenRequest models.ApiValidateToken
		err := json.NewDecoder(r.Body).Decode(&tokenRequest)
		if err != nil {
			ReturnApiError(w, models.NewFromError(err))
			return
		}

		if tokenRequest.Token == "" {
			ReturnApiError(w, models.ApiErrorResponse{
				Message: "Token is required",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		token, err := jwt.Parse(tokenRequest.Token, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			// Return the secret key used to sign the token
			return []byte(services.GetServices().HardwareSecret), nil
		})

		if err != nil {
			ReturnApiError(w, models.NewFromErrorWithCode(err, 401))
			return
		}

		// Check if the token is valid
		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Token is valid"))
		} else {
			http.Error(w, "Token is invalid", http.StatusUnauthorized)
		}
	}
}
