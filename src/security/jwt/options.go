package jwt

import (
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
)

type JwtSigningAlgorithm string

const (
	JwtSigningAlgorithmHS256 JwtSigningAlgorithm = "HS256"
	JwtSigningAlgorithmHS384 JwtSigningAlgorithm = "HS384"
	JwtSigningAlgorithmHS512 JwtSigningAlgorithm = "HS512"
	JwtSigningAlgorithmRS256 JwtSigningAlgorithm = "RS256"
	JwtSigningAlgorithmRS384 JwtSigningAlgorithm = "RS384"
	JwtSigningAlgorithmRS512 JwtSigningAlgorithm = "RS512"
)

type JwtOptions struct {
	ctx           basecontext.ApiContext
	Algorithm     JwtSigningAlgorithm
	Secret        string
	PrivateKey    string
	TokenDuration time.Duration
}

func NewDefaultOptions(ctx basecontext.ApiContext) *JwtOptions {
	if ctx == nil {
		ctx = basecontext.NewRootBaseContext()
	}

	return &JwtOptions{
		ctx:           ctx,
		Algorithm:     JwtSigningAlgorithmHS256,
		TokenDuration: time.Duration(20) * time.Minute,
	}
}

func (o *JwtOptions) WithAlgorithm(algorithm JwtSigningAlgorithm) *JwtOptions {
	o.Algorithm = algorithm
	return o
}

func (o *JwtOptions) WithSecret(secret string) *JwtOptions {
	o.Secret = secret
	return o
}

func (o *JwtOptions) WithPrivateKey(privateKey string) *JwtOptions {
	o.PrivateKey = privateKey
	return o
}

func (o *JwtOptions) WithTokenDuration(durationInMinutes float64) *JwtOptions {
	o.TokenDuration = time.Duration(durationInMinutes) * time.Minute
	return o
}
