package config

// Cohesive getter group moved from main.go (pure move, no behaviour
// change) - see .ai-agents/features/417-gorm-database-layer/420-database-configuration/plan.md D-6

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/Parallels/prl-devops-service/common"
	"github.com/Parallels/prl-devops-service/constants"

	"github.com/cjlapao/common-go/security"
)

func (c *Config) ApiPort() string {
	port := c.GetKey(constants.API_PORT_ENV_VAR)

	if port == "" {
		port = constants.DEFAULT_API_PORT
	}

	return port
}

func (c *Config) ApiPrefix() string {
	apiPrefix := c.GetKey(constants.API_PREFIX_ENV_VAR)
	if apiPrefix == "" {
		apiPrefix = constants.DEFAULT_API_PREFIX
	}

	return apiPrefix
}

func (c *Config) IsCorsEnabled() bool {
	// Module takes precedence; fall back to the direct env var for backward
	// compatibility with plists that set ENABLE_CORS=true directly.
	return c.IsModuleEnabled(constants.CORS_MODE) || c.GetBoolKey(constants.ENABLE_CORS_ENV_VAR)
}

func (c *Config) TlsCertificate() string {
	tlsCertificate := c.GetKey(constants.TLS_CERTIFICATE_ENV_VAR)
	decoded, err := security.DecodeBase64String(tlsCertificate)
	if err != nil {
		common.Logger.Error("Error decoding TLS Private Key: %v", err.Error())
		return ""
	}
	tlsCertificate = string(decoded)
	return tlsCertificate
}

func (c *Config) TlsPrivateKey() string {
	tlsPrivateKey := c.GetKey(constants.TLS_PRIVATE_KEY_ENV_VAR)
	decoded, err := security.DecodeBase64String(tlsPrivateKey)
	if err != nil {
		common.Logger.Error("Error decoding TLS Private Key: %v", err.Error())
		return ""
	}

	tlsPrivateKey = string(decoded)
	return tlsPrivateKey
}

func (c *Config) TlsPort() string {
	tlsPort := c.GetKey(constants.TLS_PORT_ENV_VAR)
	if tlsPort == "" {
		tlsPort = constants.DEFAULT_API_TLS_PORT
	}

	return tlsPort
}

func (c *Config) TlsEnabled() bool {
	TLSEnabled := c.GetKey(constants.TLS_ENABLED_ENV_VAR)
	if TLSEnabled == "" || TLSEnabled == "false" {
		return false
	}
	if c.TlsCertificate() == "" || c.TlsPrivateKey() == "" {
		return false
	}
	return true
}

// ValidateTlsConfiguration performs comprehensive validation of TLS configuration
// Returns (isValid bool, errorMessage string)
func (c *Config) ValidateTlsConfiguration() (bool, string) {
	// Check if TLS is enabled
	if !c.TlsEnabled() {
		return false, "TLS is not enabled or missing certificate/key"
	}

	cert := c.TlsCertificate()
	key := c.TlsPrivateKey()

	// Validate certificate and key can form a valid keypair
	_, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return false, fmt.Sprintf("Invalid TLS certificate/key pair: %v", err)
	}

	// Check certificate expiration
	block, _ := pem.Decode([]byte(cert))
	if block == nil {
		return false, "Failed to parse certificate PEM"
	}

	certObj, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Sprintf("Failed to parse certificate: %v", err)
	}

	// Check if certificate is expired or not yet valid
	now := time.Now()
	if now.Before(certObj.NotBefore) {
		return false, fmt.Sprintf("Certificate not yet valid (valid from: %v)", certObj.NotBefore)
	}
	if now.After(certObj.NotAfter) {
		return false, fmt.Sprintf("Certificate expired on: %v", certObj.NotAfter)
	}

	// Warn if certificate expires soon (30 days)
	daysUntilExpiry := certObj.NotAfter.Sub(now).Hours() / 24
	if daysUntilExpiry < 30 {
		common.Logger.Warn("TLS certificate expires in %.0f days", daysUntilExpiry)
	}

	return true, ""
}

// ShouldDisableHttpWhenTls checks if HTTP should be disabled when TLS is enabled
func (c *Config) ShouldDisableHttpWhenTls() bool {
	// Check environment variable (default: false for backward compatibility)
	disableHttp := c.GetBoolKey(constants.DISABLE_HTTP_WHEN_TLS_ENV_VAR)
	if !disableHttp {
		return false
	}

	// Only disable HTTP if TLS is properly validated
	isValid, errMsg := c.ValidateTlsConfiguration()
	if !isValid {
		common.Logger.Warn("⚠ DISABLE_HTTP_WHEN_TLS ignored: %s", errMsg)
		common.Logger.Warn("⚠ Falling back to dual-protocol mode (HTTP + HTTPS)")
		return false
	}

	return true
}

func (c *Config) Localhost() string {
	schema := "http"
	host := "localhost"
	port := c.ApiPort()
	if c.TlsEnabled() {
		schema = "https"
		port = c.TlsPort()
	}

	return schema + "://" + host + ":" + port
}

func (c *Config) DisableTlsValidation() bool {
	val := c.GetBoolKey(constants.TLS_DISABLE_VALIDATION_ENV_VAR)

	return val
}

func (c *Config) IsReverseProxyEnable() bool {
	envVar := c.GetKey(constants.DISABLE_REVERSE_PROXY_ENV_VAR)
	if envVar == "true" || envVar == "1" {
		return false
	}

	return true
}

func (c *Config) IsReverseProxyEnabled() bool {
	return c.IsModuleEnabled(constants.REVERSE_PROXY_MODE)
}

func (c *Config) ReverseProxyHost() string {
	host := c.GetKey(constants.REVERSE_PROXY_HOST_ENV_VAR)
	if host == "" {
		host = constants.DEFAULT_REVERSE_PROXY_HOST
	}

	return host
}

func (c *Config) ReverseProxyPort() string {
	port := c.GetKey(constants.REVERSE_PROXY_PORT_ENV_VAR)
	if port == "" {
		port = constants.DEFAULT_REVERSE_PROXY_PORT
	}

	return port
}

func (c *Config) GetReverseProxyConfig() *ReverseProxyConfig {
	return c.config.ReverseProxy
}
