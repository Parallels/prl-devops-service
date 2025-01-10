package install

type ApiServiceConfig struct {
	Port                     string `json:"port,omitempty"`
	Prefix                   string `json:"prefix,omitempty"`
	InstallVersion           string `json:"install_version,omitempty"`
	RootPassword             string `json:"root_password,omitempty"`
	HmacSecret               string `json:"hmac_secret,omitempty"`
	EncryptionRsaKey         string `json:"encryption_rsa_key,omitempty"`
	LogLevel                 string `json:"log_level,omitempty"`
	EnableTLS                bool   `json:"enable_tls,omitempty"`
	TLSPort                  string `json:"tls_port,omitempty"`
	TLSCertificate           string `json:"tls_certificate,omitempty"`
	TLSPrivateKey            string `json:"tls_private_key,omitempty"`
	DisableCatalogCaching    bool   `json:"disable_catalog_caching,omitempty"`
	TokenDurationMinutes     string `json:"token_duration_minutes,omitempty"`
	Mode                     string `json:"mode,omitempty"`
	UseOrchestratorResources bool   `json:"use_orchestrator_resources,omitempty"`
	LogOutput                bool   `json:"log_output,omitempty"`
	DisableFileLogging       bool   `json:"disable_file_logging,omitempty"`
}
