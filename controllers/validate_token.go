package controllers

import (
	"Parallels/pd-api-service/models"
	"encoding/json"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
)

func ValidateTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tokenRequest models.ApiValidateToken
		err := json.NewDecoder(r.Body).Decode(&tokenRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if tokenRequest.Token == "" {
			http.Error(w, "Token is required", http.StatusBadRequest)
			return
		}

		token, err := jwt.Parse(tokenRequest.Token, func(token *jwt.Token) (interface{}, error) {
			// Validate the algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			// Return the secret key used to sign the token
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
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
