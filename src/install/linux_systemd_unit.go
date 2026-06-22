package install

import (
	"bytes"
	"text/template"
)

const (
	LINUX_SYSTEMD_UNIT_DIR  = "/etc/systemd/system"
	LINUX_SYSTEMD_UNIT_NAME = "prl-devops-service.service"
)

type SystemdTemplateData struct {
	ExecutablePath     string
	DisableFileLogging bool
}

var systemdUnitTemplate = `[Unit]
Description=Parallels DevOps Service
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart={{ .ExecutablePath }} --config /etc/prl-devops-service/prldevops_config.yaml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`

func generateSystemdUnit(executablePath string, config ApiServiceConfig) (string, error) {
	tmpl, err := template.New("prl-devops-service").Parse(systemdUnitTemplate)
	if err != nil {
		return "", err
	}

	data := SystemdTemplateData{
		ExecutablePath:     executablePath,
		DisableFileLogging: config.DisableFileLogging,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
