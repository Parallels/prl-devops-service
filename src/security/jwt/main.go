package jwt

import (
	"errors"
	"strconv"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/security"
	"github.com/golang-jwt/jwt/v4"
	"gopkg.in/square/go-jose.v2"
)

var globalJwtService *JwtService

type JwtService struct {
	ctx     basecontext.ApiContext
	Options *JwtOptions
}

func New(ctx basecontext.ApiContext) *JwtService {
	globalJwtService = &JwtService{
		ctx:     ctx,
		Options: NewDefaultOptions(ctx),
	}

	err := globalJwtService.processEnvironmentVariables()
	if err != nil {
		ctx.LogError("Error processing environment variables for jwt options: %s", err.Error())
	}

	return globalJwtService
}

func Get() *JwtService {
	if globalJwtService == nil {
		ctx := basecontext.NewRootBaseContext()
		return New(ctx)
	}

	return globalJwtService
}

func (s *JwtService) WithTokenDuration(durationInMinutes float64) *JwtService {
	s.Options.WithTokenDuration(durationInMinutes)
	return s
}

func (s *JwtService) WithSecret(secret string) *JwtService {
	s.Options.WithSecret(secret)
	return s
}

func (s *JwtService) WithPrivateKey(privateKey string) *JwtService {
	s.Options.WithPrivateKey(privateKey)
	return s
}

func (s *JwtService) WithAlgorithm(algorithm JwtSigningAlgorithm) *JwtService {
	s.Options.WithAlgorithm(algorithm)
	return s
}

func (s *JwtService) Sign(claims map[string]interface{}) (string, error) {
	if claims["email"] == "" {
		return "", errors.New("email cannot be empty")
	}

	expiresAt := time.Now().Add(s.Options.TokenDuration).Unix()
	var method jwt.SigningMethod

	switch s.Options.Algorithm {
	case JwtSigningAlgorithmHS256:
		method = jwt.SigningMethodHS256
	case JwtSigningAlgorithmHS384:
		method = jwt.SigningMethodHS384
	case JwtSigningAlgorithmHS512:
		method = jwt.SigningMethodHS512
	case JwtSigningAlgorithmRS256:
		method = jwt.SigningMethodRS256
	case JwtSigningAlgorithmRS384:
		method = jwt.SigningMethodRS384
	case JwtSigningAlgorithmRS512:
		method = jwt.SigningMethodRS512
	default:
		method = jwt.SigningMethodHS256
		s.Options.Algorithm = JwtSigningAlgorithmHS256
	}
	if claims["roles"] == nil {
		claims["roles"] = []string{}
	}
	if claims["claims"] == nil {
		claims["claims"] = map[string]interface{}{}
	}

	defaultClaims := jwt.MapClaims{
		"exp": expiresAt,
	}

	for k, v := range claims {
		defaultClaims[k] = v
	}

	token := jwt.NewWithClaims(method, defaultClaims)

	var key interface{}

	switch s.Options.Algorithm {
	case JwtSigningAlgorithmHS256, JwtSigningAlgorithmHS384, JwtSigningAlgorithmHS512:
		if s.Options.Secret != "" {
			key = []byte(s.Options.Secret)
		} else {
			return "", errors.New("secret cannot be empty")
		}
	case JwtSigningAlgorithmRS256, JwtSigningAlgorithmRS384, JwtSigningAlgorithmRS512:
		if s.Options.PrivateKey != "" {
			decodedKey, err := security.Base64Decode(s.Options.PrivateKey)
			if err != nil {
				return "", err
			}
			privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(decodedKey)
			if err != nil {
				return "", err
			}

			key = privateKey
		} else {
			return "", errors.New("private key cannot be empty")
		}
	}

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *JwtService) GenerateJWKS() (string, error) {
	if s.Options.PrivateKey == "" {
		return "", errors.New("private key cannot be empty")
	}

	decodedKey, err := security.Base64Decode(s.Options.PrivateKey)
	if err != nil {
		return "", err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(decodedKey)
	if err != nil {
		return "", err
	}

	var algorithm string
	switch s.Options.Algorithm {
	case JwtSigningAlgorithmRS256:
		algorithm = "RS256"
	case JwtSigningAlgorithmRS384:
		algorithm = "RS384"
	case JwtSigningAlgorithmRS512:
		algorithm = "RS512"
	default:
		algorithm = "RS256"
	}

	thumbprint, err := security.CalculatePrivateKeyThumbprint(privateKey)
	if err != nil {
		return "", err
	}

	jwk := jose.JSONWebKey{Key: privateKey, KeyID: thumbprint, Algorithm: algorithm}

	jwkBytes, err := jwk.MarshalJSON()
	if err != nil {
		return "", err
	}

	return string(jwkBytes), nil
}

func (s *JwtService) Parse(token string) (*JwtSystemToken, error) {
	var key interface{}

	switch s.Options.Algorithm {
	case JwtSigningAlgorithmHS256, JwtSigningAlgorithmHS384, JwtSigningAlgorithmHS512:
		if s.Options.Secret != "" {
			key = []byte(s.Options.Secret)
		} else {
			return nil, errors.New("secret cannot be empty")
		}
	case JwtSigningAlgorithmRS256, JwtSigningAlgorithmRS384, JwtSigningAlgorithmRS512:
		if s.Options.PrivateKey != "" {
			decodedKey, err := security.Base64Decode(s.Options.PrivateKey)
			if err != nil {
				return nil, err
			}
			privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(decodedKey)
			if err != nil {
				return nil, err
			}
			publicKey := privateKey.Public()
			if err != nil {
				return nil, err
			}
			key = publicKey
		} else {
			return nil, errors.New("private key cannot be empty")
		}
	}

	tokenObj, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	systemToken := &JwtSystemToken{
		token:    token,
		tokenObj: tokenObj,
	}
	systemToken.GetTokenClaims()

	return systemToken, nil
}

func (s *JwtService) processEnvironmentVariables() error {
	cfg := config.NewConfig()
	if cfg.GetKey(constants.JWT_SIGN_ALGORITHM_ENV_VAR) != "" {
		algorithm := JwtSigningAlgorithm(cfg.GetKey(constants.JWT_SIGN_ALGORITHM_ENV_VAR))
		switch algorithm {
		case JwtSigningAlgorithmHS256, JwtSigningAlgorithmHS384, JwtSigningAlgorithmHS512,
			JwtSigningAlgorithmRS256, JwtSigningAlgorithmRS384, JwtSigningAlgorithmRS512:
		default:
			return errors.New("invalid signing algorithm")
		}

		s.Options.WithAlgorithm(algorithm)
	}

	if cfg.GetKey(constants.JWT_HMACS_SECRET_ENV_VAR) != "" {
		s.Options.WithSecret(cfg.GetKey(constants.JWT_HMACS_SECRET_ENV_VAR))
	}

	if cfg.GetKey(constants.JWT_PRIVATE_KEY_ENV_VAR) != "" {
		s.Options.WithPrivateKey(cfg.GetKey(constants.JWT_PRIVATE_KEY_ENV_VAR))
	}

	if cfg.GetKey(constants.JWT_DURATION_ENV_VAR) != "" {
		durationInMinutes, err := strconv.ParseFloat(cfg.GetKey(constants.JWT_DURATION_ENV_VAR), 64)
		if err != nil {
			return err
		}
		s.Options.WithTokenDuration(durationInMinutes)
	}

	// generating a default secret if none is provided
	if s.Options.Algorithm == JwtSigningAlgorithmHS256 || s.Options.Algorithm == JwtSigningAlgorithmHS384 || s.Options.Algorithm == JwtSigningAlgorithmHS512 {
		if s.Options.Secret == "" {
			randStr, err := security.GenerateCryptoRandomString(80)
			if err != nil {
				s.ctx.LogError("Error generating random string: %s", err.Error())
				return err
			}
			s.Options.WithSecret(randStr)
		}
	}

	return nil
}
