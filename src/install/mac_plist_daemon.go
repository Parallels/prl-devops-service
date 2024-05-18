package install

import (
	"bytes"
	"text/template"

	"github.com/Parallels/prl-devops-service/constants"
)

type PlistTemplateData struct {
	Path                     string
	ExecutableName           string
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
	Mode                     string
	UseOrchestratorResources string
}

var plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>UserName</key>
  <string>root</string>
  <key>Label</key>
  <string>com.parallels.devops-service</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{ .Path }}/{{ .ExecutableName }}</string>
  </array>
  <key>EnvironmentVariables</key>
  <dict>
    {{- if .Port }}
    <key>` + constants.API_PORT_ENV_VAR + `</key>
    <string>{{ .Port }}</string>
    {{- end }}
    {{- if .Prefix }}
    <key>` + constants.API_PREFIX_ENV_VAR + `</key>
    <string>{{ .Prefix }}</string>
    {{- end }}
    {{- if .RootPassword }}
    <key>` + constants.ROOT_PASSWORD_ENV_VAR + `</key>
    <string>{{ .RootPassword }}</string>
    {{- end }}
    {{- if .EncryptionRsaKey }}
    <key>` + constants.ENCRYPTION_SECURITY_KEY_ENV_VAR + `</key>
    <string>{{ .EncryptionRsaKey }}</string>
    {{- end }}
    {{- if .HmacSecret }}
    <key>` + constants.HMAC_SECRET_ENV_VAR + `</key>
    <string>{{ .HmacSecret }}</string>
    {{- end }}
    {{- if .LogLevel }}
    <key>` + constants.LOG_LEVEL_ENV_VAR + `</key>
    <string>{{ .LogLevel }}</string>
    {{- end }}
    {{- if .EnableTLS }}
    <key>` + constants.TLS_ENABLED_ENV_VAR + `</key>
    <string>{{ .EnableTLS }}</string>
    {{- end }}
    {{- if .HostTLSPort }}
    <key>` + constants.TLS_PORT_ENV_VAR + `</key>
    <string>{{ .HostTLSPort }}</string>
    {{- end }}
    {{- if .TlsCertificate }}
    <key>` + constants.TLS_CERTIFICATE_ENV_VAR + `</key>
    <string>{{ .TlsCertificate }}</string>
    {{- end }}
    {{- if .TlsPrivateKey }}
    <key>` + constants.TLS_PRIVATE_KEY_ENV_VAR + `</key>
    <string>{{ .TlsPrivateKey }}</string>
    {{- end }}
    {{- if .DisableCatalogCaching }}
    <key>` + constants.DISABLE_CATALOG_CACHING_ENV_VAR + `</key>
    <string>{{ .DisableCatalogCaching }}</string>
    {{- end }}
    {{- if .TokenDurationMinutes }}
    <key>` + constants.TOKEN_DURATION_MINUTES_ENV_VAR + `</key>
    <string>{{ .TokenDurationMinutes }}</string>
    {{- end }}
    {{- if .Mode }}
    <key>` + constants.MODE_ENV_VAR + `</key>
    <string>{{ .Mode }}</string>
    {{- end }}
    {{- if .UseOrchestratorResources }}
    <key>` + constants.USE_ORCHESTRATOR_RESOURCES_ENV_VAR + `</key>
    <string>{{ .UseOrchestratorResources }}</string>
    {{- end }}
  </dict>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <dict>
    <key>SuccessfulExit</key>
    <false/>
  </dict>
  <key>StandardErrorPath</key>
  <string>/tmp/api-service.job.err</string>
  <key>StandardOutPath</key>
  <key>RunAtLoad</key>
  <true/>
  <string>/tmp/api-service.job.out</string> 
</dict>
</plist>"`

func generatePlist(path string, config ApiServiceConfig) (string, error) {
	// Define the text template
	tmpl, err := template.New("parallels-devops").Parse(plistTemplate)
	if err != nil {
		return "", err
	}

	// Execute the template with a value
	var tpl bytes.Buffer
	templateData := PlistTemplateData{
		Path:                 path,
		ExecutableName:       constants.ExecutableName,
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
		Mode:                 config.Mode,
	}
	if config.EnableTLS {
		templateData.EnableTLS = "true"
	}
	if config.DisableCatalogCaching {
		templateData.DisableCatalogCaching = "true"
	}
	if config.UseOrchestratorResources {
		templateData.UseOrchestratorResources = "true"
	}

	err = tmpl.Execute(&tpl, templateData)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}
