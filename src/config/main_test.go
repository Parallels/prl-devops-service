package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	str "github.com/stretchr/testify/assert"
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	originalValue, existed := os.LookupEnv(key)
	str.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, originalValue)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

func TestGetKeyUsesConfigFileValueWhenEnvIsUnset(t *testing.T) {
	unsetEnv(t, constants.API_PORT_ENV_VAR)
	unsetEnv(t, constants.DATABASE_FOLDER_ENV_VAR)

	cfg := New(basecontext.NewBaseContext())
	cfg.config.Environment = map[string]string{
		constants.API_PORT_ENV_VAR:        "8080",
		constants.DATABASE_FOLDER_ENV_VAR: "/config/data",
	}

	str.Equal(t, "8080", cfg.GetKey(constants.API_PORT_ENV_VAR))
	str.Equal(t, "/config/data", cfg.GetKey(constants.DATABASE_FOLDER_ENV_VAR))
}

func TestGetKeyEnvironmentOverridesConfigFileValue(t *testing.T) {
	t.Setenv(constants.API_PORT_ENV_VAR, "9090")
	t.Setenv(constants.DATABASE_FOLDER_ENV_VAR, "/env/data")

	cfg := New(basecontext.NewBaseContext())
	cfg.config.Environment = map[string]string{
		constants.API_PORT_ENV_VAR:        "8080",
		constants.DATABASE_FOLDER_ENV_VAR: "/config/data",
	}

	str.Equal(t, "9090", cfg.GetKey(constants.API_PORT_ENV_VAR))
	str.Equal(t, "/env/data", cfg.GetKey(constants.DATABASE_FOLDER_ENV_VAR))
}

func TestGetKeyMatchesConfigFileKeysCaseInsensitively(t *testing.T) {
	unsetEnv(t, constants.API_PORT_ENV_VAR)

	cfg := New(basecontext.NewBaseContext())
	cfg.config.Environment = map[string]string{
		"api_port": "8080",
	}

	str.Equal(t, "8080", cfg.GetKey(constants.API_PORT_ENV_VAR))
}

func TestGetKeyEmptyEnvironmentValueOverridesConfigFileValue(t *testing.T) {
	t.Setenv(constants.API_PORT_ENV_VAR, "")

	cfg := New(basecontext.NewBaseContext())
	cfg.config.Environment = map[string]string{
		constants.API_PORT_ENV_VAR: "8080",
	}

	str.Equal(t, "", cfg.GetKey(constants.API_PORT_ENV_VAR))
}

// generateTestCertificate creates a self-signed certificate for testing
// Returns (certPEM, keyPEM string, error)
func generateTestCertificate(notBefore, notAfter time.Time) (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
			CommonName:   "localhost",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", "", err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return string(certPEM), string(keyPEM), nil
}

func TestValidateTlsConfiguration_TlsDisabled(t *testing.T) {
	unsetEnv(t, constants.TLS_ENABLED_ENV_VAR)
	unsetEnv(t, constants.TLS_CERTIFICATE_ENV_VAR)
	unsetEnv(t, constants.TLS_PRIVATE_KEY_ENV_VAR)

	cfg := New(basecontext.NewBaseContext())
	isValid, errMsg := cfg.ValidateTlsConfiguration()

	str.False(t, isValid)
	str.Contains(t, errMsg, "TLS is not enabled")
}

func TestValidateTlsConfiguration_ValidCertificate(t *testing.T) {
	// Generate a valid certificate (valid for 1 year)
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	certPEM, keyPEM, err := generateTestCertificate(notBefore, notAfter)
	str.NoError(t, err)

	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "true")
	t.Setenv(constants.TLS_CERTIFICATE_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(certPEM)))
	t.Setenv(constants.TLS_PRIVATE_KEY_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(keyPEM)))

	cfg := New(basecontext.NewBaseContext())
	isValid, errMsg := cfg.ValidateTlsConfiguration()

	str.True(t, isValid)
	str.Empty(t, errMsg)
}

func TestValidateTlsConfiguration_ExpiredCertificate(t *testing.T) {
	// Generate an expired certificate (expired 1 day ago)
	notBefore := time.Now().Add(-365 * 24 * time.Hour)
	notAfter := time.Now().Add(-24 * time.Hour)
	certPEM, keyPEM, err := generateTestCertificate(notBefore, notAfter)
	str.NoError(t, err)

	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "true")
	t.Setenv(constants.TLS_CERTIFICATE_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(certPEM)))
	t.Setenv(constants.TLS_PRIVATE_KEY_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(keyPEM)))

	cfg := New(basecontext.NewBaseContext())
	isValid, errMsg := cfg.ValidateTlsConfiguration()

	str.False(t, isValid)
	str.Contains(t, errMsg, "Certificate expired")
}

func TestValidateTlsConfiguration_NotYetValidCertificate(t *testing.T) {
	// Generate a certificate that's not yet valid (valid from tomorrow)
	notBefore := time.Now().Add(24 * time.Hour)
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	certPEM, keyPEM, err := generateTestCertificate(notBefore, notAfter)
	str.NoError(t, err)

	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "true")
	t.Setenv(constants.TLS_CERTIFICATE_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(certPEM)))
	t.Setenv(constants.TLS_PRIVATE_KEY_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(keyPEM)))

	cfg := New(basecontext.NewBaseContext())
	isValid, errMsg := cfg.ValidateTlsConfiguration()

	str.False(t, isValid)
	str.Contains(t, errMsg, "Certificate not yet valid")
}

func TestValidateTlsConfiguration_InvalidCertKeyPair(t *testing.T) {
	// Generate two different certificates to create a mismatched pair
	certPEM1, _, err1 := generateTestCertificate(time.Now(), time.Now().Add(365*24*time.Hour))
	str.NoError(t, err1)

	_, keyPEM2, err2 := generateTestCertificate(time.Now(), time.Now().Add(365*24*time.Hour))
	str.NoError(t, err2)

	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "true")
	t.Setenv(constants.TLS_CERTIFICATE_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(certPEM1)))
	t.Setenv(constants.TLS_PRIVATE_KEY_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(keyPEM2)))

	cfg := New(basecontext.NewBaseContext())
	isValid, errMsg := cfg.ValidateTlsConfiguration()

	str.False(t, isValid)
	str.Contains(t, errMsg, "Invalid TLS certificate/key pair")
}

func TestValidateTlsConfiguration_InvalidPEM(t *testing.T) {
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "true")
	t.Setenv(constants.TLS_CERTIFICATE_ENV_VAR, base64.StdEncoding.EncodeToString([]byte("invalid pem data")))
	t.Setenv(constants.TLS_PRIVATE_KEY_ENV_VAR, base64.StdEncoding.EncodeToString([]byte("invalid pem data")))

	cfg := New(basecontext.NewBaseContext())
	isValid, errMsg := cfg.ValidateTlsConfiguration()

	str.False(t, isValid)
	str.Contains(t, errMsg, "Invalid TLS certificate/key pair")
}

func TestShouldDisableHttpWhenTls_DisabledByDefault(t *testing.T) {
	// Generate a valid certificate
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	certPEM, keyPEM, err := generateTestCertificate(notBefore, notAfter)
	str.NoError(t, err)

	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "true")
	t.Setenv(constants.TLS_CERTIFICATE_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(certPEM)))
	t.Setenv(constants.TLS_PRIVATE_KEY_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(keyPEM)))
	unsetEnv(t, constants.DISABLE_HTTP_WHEN_TLS_ENV_VAR)

	cfg := New(basecontext.NewBaseContext())
	shouldDisable := cfg.ShouldDisableHttpWhenTls()

	str.False(t, shouldDisable, "HTTP should not be disabled by default (backward compatibility)")
}

func TestShouldDisableHttpWhenTls_EnabledWithValidTls(t *testing.T) {
	// Generate a valid certificate
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	certPEM, keyPEM, err := generateTestCertificate(notBefore, notAfter)
	str.NoError(t, err)

	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "true")
	t.Setenv(constants.TLS_CERTIFICATE_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(certPEM)))
	t.Setenv(constants.TLS_PRIVATE_KEY_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(keyPEM)))
	t.Setenv(constants.DISABLE_HTTP_WHEN_TLS_ENV_VAR, "true")

	cfg := New(basecontext.NewBaseContext())
	shouldDisable := cfg.ShouldDisableHttpWhenTls()

	str.True(t, shouldDisable, "HTTP should be disabled when DISABLE_HTTP_WHEN_TLS=true and TLS is valid")
}

func TestShouldDisableHttpWhenTls_EnabledWithInvalidTls(t *testing.T) {
	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "true")
	t.Setenv(constants.TLS_CERTIFICATE_ENV_VAR, base64.StdEncoding.EncodeToString([]byte("invalid")))
	t.Setenv(constants.TLS_PRIVATE_KEY_ENV_VAR, base64.StdEncoding.EncodeToString([]byte("invalid")))
	t.Setenv(constants.DISABLE_HTTP_WHEN_TLS_ENV_VAR, "true")

	cfg := New(basecontext.NewBaseContext())
	shouldDisable := cfg.ShouldDisableHttpWhenTls()

	str.False(t, shouldDisable, "HTTP should not be disabled when TLS validation fails (safety fallback)")
}

func TestShouldDisableHttpWhenTls_ExplicitlyDisabled(t *testing.T) {
	// Generate a valid certificate
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	certPEM, keyPEM, err := generateTestCertificate(notBefore, notAfter)
	str.NoError(t, err)

	t.Setenv(constants.TLS_ENABLED_ENV_VAR, "true")
	t.Setenv(constants.TLS_CERTIFICATE_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(certPEM)))
	t.Setenv(constants.TLS_PRIVATE_KEY_ENV_VAR, base64.StdEncoding.EncodeToString([]byte(keyPEM)))
	t.Setenv(constants.DISABLE_HTTP_WHEN_TLS_ENV_VAR, "false")

	cfg := New(basecontext.NewBaseContext())
	shouldDisable := cfg.ShouldDisableHttpWhenTls()

	str.False(t, shouldDisable, "HTTP should not be disabled when DISABLE_HTTP_WHEN_TLS=false")
}
