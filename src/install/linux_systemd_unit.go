package install

import (
	"bytes"
	"text/template"

	"github.com/Parallels/prl-devops-service/constants"
)

const (
	LINUX_SYSTEMD_UNIT_DIR  = "/etc/systemd/system"
	LINUX_SYSTEMD_UNIT_NAME = "prl-devops-service.service"
)

type SystemdTemplateData struct {
	ExecutablePath           string
	Port                     string
	Prefix                   string
	RootPassword             string
	EncryptionRsaKey         string
	HmacSecret               string
	LogLevel                 string
	EnableTLS                string
	HostTLSPort              string
	TlsCertificate           string
	TlsPrivateKey            string
	DisableCatalogCaching    string
	TokenDurationMinutes     string
	EnabledModules           string
	UseOrchestratorResources string
	EnableCors               string
	CorsAllowedOrigins       string
	CorsAllowedMethods       string
	CorsAllowedHeaders       string
	DisableFileLogging       bool
}

// systemdUnitTemplate produces a systemd service unit file.
// Environment= lines are emitted only when the value is non-empty so the
// service inherits sensible defaults for anything that was not configured.
var systemdUnitTemplate = `[Unit]
Description=Parallels DevOps Service
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart={{ .ExecutablePath }}
Restart=on-failure
RestartSec=5s
{{- if not .DisableFileLogging }}
Environment="` + constants.LOG_TO_FILE_ENV_VAR + `=true"
{{- end }}
{{- if .Port }}
Environment="` + constants.API_PORT_ENV_VAR + `={{ .Port }}"
{{- end }}
{{- if .Prefix }}
Environment="` + constants.API_PREFIX_ENV_VAR + `={{ .Prefix }}"
{{- end }}
{{- if .RootPassword }}
Environment="` + constants.ROOT_PASSWORD_ENV_VAR + `={{ .RootPassword }}"
{{- end }}
{{- if .EncryptionRsaKey }}
Environment="` + constants.ENCRYPTION_SECURITY_KEY_ENV_VAR + `={{ .EncryptionRsaKey }}"
{{- end }}
{{- if .HmacSecret }}
Environment="` + constants.HMAC_SECRET_ENV_VAR + `={{ .HmacSecret }}"
{{- end }}
{{- if .LogLevel }}
Environment="` + constants.LOG_LEVEL_ENV_VAR + `={{ .LogLevel }}"
{{- end }}
{{- if .EnableTLS }}
Environment="` + constants.TLS_ENABLED_ENV_VAR + `={{ .EnableTLS }}"
{{- end }}
{{- if .HostTLSPort }}
Environment="` + constants.TLS_PORT_ENV_VAR + `={{ .HostTLSPort }}"
{{- end }}
{{- if .TlsCertificate }}
Environment="` + constants.TLS_CERTIFICATE_ENV_VAR + `={{ .TlsCertificate }}"
{{- end }}
{{- if .TlsPrivateKey }}
Environment="` + constants.TLS_PRIVATE_KEY_ENV_VAR + `={{ .TlsPrivateKey }}"
{{- end }}
{{- if .DisableCatalogCaching }}
Environment="` + constants.DISABLE_CATALOG_CACHING_ENV_VAR + `={{ .DisableCatalogCaching }}"
{{- end }}
{{- if .TokenDurationMinutes }}
Environment="` + constants.TOKEN_DURATION_MINUTES_ENV_VAR + `={{ .TokenDurationMinutes }}"
{{- end }}
{{- if .EnabledModules }}
Environment="` + constants.ENABLED_MODULES_ENV_VAR + `={{ .EnabledModules }}"
{{- end }}
{{- if .UseOrchestratorResources }}
Environment="` + constants.USE_ORCHESTRATOR_RESOURCES_ENV_VAR + `={{ .UseOrchestratorResources }}"
{{- end }}
{{- if .EnableCors }}
Environment="` + constants.ENABLE_CORS_ENV_VAR + `={{ .EnableCors }}"
{{- end }}
{{- if .CorsAllowedOrigins }}
Environment="` + constants.CORS_ALLOWED_ORIGINS_ENV_VAR + `={{ .CorsAllowedOrigins }}"
{{- end }}
{{- if .CorsAllowedMethods }}
Environment="` + constants.CORS_ALLOWED_METHODS_ENV_VAR + `={{ .CorsAllowedMethods }}"
{{- end }}
{{- if .CorsAllowedHeaders }}
Environment="` + constants.CORS_ALLOWED_HEADERS_ENV_VAR + `={{ .CorsAllowedHeaders }}"
{{- end }}

[Install]
WantedBy=multi-user.target
`

func generateSystemdUnit(executablePath string, config ApiServiceConfig) (string, error) {
	tmpl, err := template.New("prl-devops-service").Parse(systemdUnitTemplate)
	if err != nil {
		return "", err
	}

	data := SystemdTemplateData{
		ExecutablePath:       executablePath,
		Port:                 config.Port,
		Prefix:               config.Prefix,
		RootPassword:         config.RootPassword,
		EncryptionRsaKey:     config.EncryptionRsaKey,
		HmacSecret:           config.HmacSecret,
		LogLevel:             config.LogLevel,
		HostTLSPort:          config.TLSPort,
		TlsCertificate:       config.TLSCertificate,
		TlsPrivateKey:        config.TLSPrivateKey,
		TokenDurationMinutes: config.TokenDurationMinutes,
		EnabledModules:     config.EnabledModules,
		CorsAllowedOrigins: config.CorsAllowedOrigins,
		CorsAllowedMethods: config.CorsAllowedMethods,
		CorsAllowedHeaders: config.CorsAllowedHeaders,
		DisableFileLogging: config.DisableFileLogging,
	}
	if config.EnableTLS {
		data.EnableTLS = "true"
	}
	if config.DisableCatalogCaching {
		data.DisableCatalogCaching = "true"
	}
	if config.UseOrchestratorResources {
		data.UseOrchestratorResources = "true"
	}
	if config.EnableCors {
		data.EnableCors = "true"
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
