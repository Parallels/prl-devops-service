package jwt

import (
	"time"

	"github.com/Parallels/prl-devops-service/errors"
	"github.com/golang-jwt/jwt/v4"
)

type JwtSystemToken struct {
	token    string
	tokenObj *jwt.Token
	Claims   map[string]interface{}
}

func (s *JwtSystemToken) Valid() (bool, error) {
	if s.tokenObj == nil {
		return false, errors.New("tokenObj is nil")
	}

	if _, ok := s.tokenObj.Claims.(jwt.MapClaims); !ok {
		return false, errors.New("invalid claims")
	}

	if err := s.tokenObj.Claims.Valid(); err != nil {
		return false, err
	}

	return s.tokenObj.Valid, nil
}

func (s *JwtSystemToken) GetTokenClaims() (map[string]interface{}, error) {
	claims, ok := s.tokenObj.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	s.Claims = claims
	return claims, nil
}

func (s *JwtSystemToken) GetEmail() (string, error) {
	if s.Claims == nil {
		_, err := s.GetTokenClaims()
		if err != nil {
			return "", err
		}
	}

	email, ok := s.Claims["email"].(string)
	if !ok {
		return "", errors.New("invalid email")
	}

	return email, nil
}

func (s *JwtSystemToken) GetExpiresAt() (time.Time, error) {
	if s.Claims == nil {
		_, err := s.GetTokenClaims()
		if err != nil {
			return time.Time{}, err
		}
	}

	expiresAt, ok := s.Claims["exp"].(float64)
	if !ok {
		return time.Time{}, errors.New("invalid expiresAt")
	}

	parsedTime := time.Unix(int64(expiresAt), 0)
	return parsedTime, nil
}

func (s *JwtSystemToken) GetClaim(key string) (interface{}, error) {
	if s.Claims == nil {
		_, err := s.GetTokenClaims()
		if err != nil {
			return nil, err
		}
	}

	claim, ok := s.Claims[key]
	if !ok {
		return nil, errors.New("invalid claim")
	}

	return claim, nil
}
