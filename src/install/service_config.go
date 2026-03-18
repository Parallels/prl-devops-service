package install

type ApiServiceConfig struct {
	Port                  string `json:"port,omitempty"`
	Prefix                string `json:"prefix,omitempty"`
	InstallVersion        string `json:"install_version,omitempty"`
	RootPassword          string `json:"root_password,omitempty"`
	HmacSecret            string `json:"hmac_secret,omitempty"`
	EncryptionRsaKey      string `json:"encryption_rsa_key,omitempty"`
	LogLevel              string `json:"log_level,omitempty"`
	EnableTLS             bool   `json:"enable_tls,omitempty"`
	TLSPort               string `json:"tls_port,omitempty"`
	TLSCertificate        string `json:"tls_certificate,omitempty"`
	TLSPrivateKey         string `json:"tls_private_key,omitempty"`
	DisableCatalogCaching bool   `json:"disable_catalog_caching,omitempty"`
	TokenDurationMinutes  string `json:"token_duration_minutes,omitempty"`
	// EnabledModules is the comma-separated list of modules to enable
	// (e.g. "api,host,catalog,orchestrator"). "api" is always included.
	// Replaces the older Mode field; Mode is still accepted for backward
	// compatibility with existing config files.
	EnabledModules string `json:"enabled_modules,omitempty"`
	// Mode is kept for backward compatibility with existing config files.
	// Prefer EnabledModules for new installations.
	Mode                     string `json:"mode,omitempty"`
	UseOrchestratorResources bool   `json:"use_orchestrator_resources,omitempty"`
	LogOutput                bool   `json:"log_output,omitempty"`
	DisableFileLogging       bool   `json:"disable_file_logging,omitempty"`
	EnableCors               bool   `json:"enable_cors,omitempty"`
	CorsAllowedOrigins       string `json:"cors_allowed_origins,omitempty"`
	CorsAllowedMethods       string `json:"cors_allowed_methods,omitempty"`
	CorsAllowedHeaders       string `json:"cors_allowed_headers,omitempty"`
}
