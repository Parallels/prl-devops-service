package install

import (
	"bytes"
	"text/template"

	"github.com/Parallels/prl-devops-service/constants"
)

type PlistTemplateData struct {
	Path             string
	ExecutableName   string
	DisableFileLogging bool
	LogOutput        bool
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
    <string>--config</string>
    <string>/etc/prl-devops-service/prldevops_config.yaml</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <dict>
    <key>SuccessfulExit</key>
    <false/>
  </dict>
  {{- if .LogOutput }}
  <key>StandardErrorPath</key>
  <string>/tmp/devops-service.job.err</string>
  <key>StandardOutPath</key>
  <string>/tmp/devops-service.job.out</string> 
  {{- end }}
</dict>
</plist>`

func generatePlist(path string, config ApiServiceConfig) (string, error) {
	tmpl, err := template.New("parallels-devops").Parse(plistTemplate)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	templateData := PlistTemplateData{
		Path:             path,
		ExecutableName:   constants.ExecutableName,
		DisableFileLogging: config.DisableFileLogging,
		LogOutput:        config.LogOutput,
	}

	err = tmpl.Execute(&tpl, templateData)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}
