package models

import "github.com/Parallels/prl-devops-service/errors"

type ReverseProxyHostTls struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Cert    string `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key     string `json:"key,omitempty" yaml:"key,omitempty"`
}

func (o *ReverseProxyHostTls) Validate() error {
	if o.Cert == "" {
		return errors.NewWithCode("missing reverse proxy host tls cert", 400)
	}
	if o.Key == "" {
		return errors.NewWithCode("missing reverse proxy host tls key", 400)
	}
	return nil
}
