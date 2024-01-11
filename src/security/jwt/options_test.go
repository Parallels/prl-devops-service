package jwt

import (
	"testing"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
)

func TestNewDefaultOptions(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	options := NewDefaultOptions(ctx)
	if options == nil {
		t.Errorf("NewDefaultOptions returned nil")
	}
}

func TestWithAlgorithm(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	options := NewDefaultOptions(ctx)
	options.WithAlgorithm(JwtSigningAlgorithmHS256)
	if options.Algorithm != JwtSigningAlgorithmHS256 {
		t.Errorf("WithAlgorithm did not set Algorithm")
	}
}

func TestWithSecret(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	options := NewDefaultOptions(ctx)
	options.WithSecret("secret")
	if options.Secret != "secret" {
		t.Errorf("WithSecret did not set Secret")
	}
}

func TestWithPrivateKey(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	options := NewDefaultOptions(ctx)
	options.WithPrivateKey("privateKey")
	if options.PrivateKey != "privateKey" {
		t.Errorf("WithPrivateKey did not set PrivateKey")
	}
}

func TestWithTokenDuration(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	options := NewDefaultOptions(ctx)
	options.WithTokenDuration(20)
	if options.TokenDuration != time.Duration(20)*time.Minute {
		t.Errorf("WithTokenDuration did not set TokenDuration")
	}
}
