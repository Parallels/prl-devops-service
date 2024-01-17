package jwt

import (
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
	TokenDuration string
}

func NewDefaultOptions(ctx basecontext.ApiContext) *JwtOptions {
	if ctx == nil {
		ctx = basecontext.NewRootBaseContext()
	}

	return &JwtOptions{
		ctx:           ctx,
		Algorithm:     JwtSigningAlgorithmHS256,
		TokenDuration: "15m",
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

func (o *JwtOptions) WithTokenDuration(duration string) *JwtOptions {
	o.TokenDuration = duration
	return o
}
