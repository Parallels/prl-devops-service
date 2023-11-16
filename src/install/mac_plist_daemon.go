package install

import (
	"bytes"
	"text/template"
)

type PlistTemplateData struct {
	Path             string
	Port             string
	RootPassword     string
	EncryptionRsaKey string
	HmacSecret       string
	LogLevel         string
	EnableTLS        string
	HostTLSPort      string
	TlsCertificate   string
	TlsPrivateKey    string
}

var plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>UserName</key>
  <string>root</string>
  <key>Label</key>
  <string>com.parallels.api-service</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{ .Path }}/pd-api-service</string>
    <string>--port={{ .Port }}</string>
  </array>
  <key>EnvironmentVariables</key>
  <dict>
    <key>ROOT_PASSWORD</key>
    <string>{{ .RootPassword }}</string>
    <key>SECURITY_PRIVATE_KEY</key>
    <string>{{ .EncryptionRsaKey }}</string>
    <key>HMAC_SECRET</key>
    <string>{{ .HmacSecret }}</string>
    <key>LOG_LEVEL</key>
    <string>{{ .LogLevel }}</string>
    <key>TLS_ENABLED</key>
    <string>{{ .EnableTLS }}</string>
    <key>TLS_PORT</key>
    <string>{{ .HostTLSPort }}</string>
    <key>TLS_CERTIFICATE</key>
    <string>{{ .TlsCertificate }}</string>
    <key>TLS_PRIVATE_KEY</key>
    <string>{{ .TlsPrivateKey }}</string>
  </dict>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardErrorPath</key>
  <string>/tmp/api-service.job.err</string>
  <key>StandardOutPath</key>
  <string>/tmp/api-service.job.out</string> 
</dict>
</plist>"`

func generatePlist(path string, config ApiServiceConfig) (string, error) {
	// Define the text template
	tmpl, err := template.New("parallels-api").Parse(plistTemplate)
	if err != nil {
		return "", err
	}

	// Execute the template with a value
	var tpl bytes.Buffer
	templateData := PlistTemplateData{
		Path:             path,
		Port:             config.Port,
		RootPassword:     config.RootPassword,
		EncryptionRsaKey: config.EncryptionRsaKey,
		HmacSecret:       config.HmacSecret,
		LogLevel:         config.LogLevel,
		HostTLSPort:      config.TLSPort,
		TlsCertificate:   config.TLSCertificate,
		TlsPrivateKey:    config.TLSPrivateKey,
	}
	if config.EnableTLS {
		templateData.EnableTLS = "true"
	}

	err = tmpl.Execute(&tpl, templateData)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}
