package config

import (
	"os"

	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/constants"

	log "github.com/cjlapao/common-go-logger"
	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/security"
)

type Config struct{}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) GetApiPort() string {
	port := helper.GetFlagValue(constants.API_PORT_FLAG, "")

	if port == "" {
		port = os.Getenv(constants.API_PORT_ENV_VAR)
	}

	if port == "" {
		port = constants.DEFAULT_API_PORT
	}

	return port
}

func (c *Config) GetApiPrefix() string {
	apiPrefix := os.Getenv(constants.API_PREFIX_ENV_VAR)
	if apiPrefix == "" {
		apiPrefix = constants.DEFAULT_API_PREFIX
	}

	return apiPrefix
}

func (c *Config) GetHmacSecret() string {
	hmacSecret := os.Getenv(constants.HMAC_SECRET_ENV_VAR)
	return hmacSecret
}

func (c *Config) GetLogLevel() string {
	logLevel := os.Getenv(constants.LOG_LEVEL_ENV_VAR)
	switch logLevel {
	case "DEBUG":
		logLevel = "DEBUG"
		common.Logger.LogLevel = log.Debug
	case "INFO":
		logLevel = "INFO"
		common.Logger.LogLevel = log.Info
	case "WARN":
		logLevel = "WARN"
		common.Logger.LogLevel = log.Warning
	case "ERROR":
		logLevel = "ERROR"
		common.Logger.LogLevel = log.Error
	}

	return logLevel
}

func (c *Config) GetSecurityKey() string {
	securityKey := os.Getenv(constants.SECURITY_KEY_ENV_VAR)
	if securityKey == "" {
		return ""
	}

	decoded, err := security.DecodeBase64String(securityKey)
	if err != nil {
		common.Logger.Error("Error decoding TLS Private Key: %v", err.Error())
		return ""
	}
	securityKey = decoded
	return securityKey
}

func (c *Config) GetTlsCertificate() string {
	tlsCertificate := os.Getenv(constants.TLS_CERTIFICATE_ENV_VAR)
	decoded, err := security.DecodeBase64String(tlsCertificate)
	if err != nil {
		common.Logger.Error("Error decoding TLS Private Key: %v", err.Error())
		return ""
	}
	tlsCertificate = string(decoded)
	return tlsCertificate
}

func (c *Config) GetTlsPrivateKey() string {
	tlsPrivateKey := os.Getenv(constants.TLS_PRIVATE_KEY_ENV_VAR)
	decoded, err := security.DecodeBase64String(tlsPrivateKey)
	if err != nil {
		common.Logger.Error("Error decoding TLS Private Key: %v", err.Error())
		return ""
	}

	tlsPrivateKey = string(decoded)
	return tlsPrivateKey
}

func (c *Config) GetTLSPort() string {
	tlsPort := os.Getenv(constants.TLS_PORT_ENV_VAR)
	if tlsPort == "" {
		tlsPort = constants.DEFAULT_API_TLS_PORT
	}

	return tlsPort
}

func (c *Config) TLSEnabled() bool {
	TLSEnabled := os.Getenv(constants.TLS_ENABLED_ENV_VAR)
	if TLSEnabled == "" || TLSEnabled == "false" {
		return false
	}
	if c.GetTlsCertificate() == "" || c.GetTlsPrivateKey() == "" {
		return false
	}
	return true
}

func (c *Config) GetTokenDurationMinutes() int {
	return constants.TOKEN_DURATION_MINUTES
}
