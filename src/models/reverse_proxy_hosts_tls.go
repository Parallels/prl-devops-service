package models

import "github.com/Parallels/prl-devops-service/errors"

type ReverseProxyHostTls struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Cert    string `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key     string `json:"key,omitempty" yaml:"key,omitempty"`
}

func (o *ReverseProxyHostTls) Validate(diag *errors.Diagnostics) {
	if o.Cert == "" {
		diag.AddError("400", "missing reverse proxy host tls cert", "ReverseProxyHostTls")
		return
	}
	if o.Key == "" {
		diag.AddError("400", "missing reverse proxy host tls key", "ReverseProxyHostTls")
		return
	}
	return
}
