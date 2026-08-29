package models

import (
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/errors"
)

type ReverseProxyHostTls struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Cert    string `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key     string `json:"key,omitempty" yaml:"key,omitempty"`
}

func (o *ReverseProxyHostTls) Validate(diag *errors.Diagnostics) {
	if o.Cert == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing reverse proxy host tls cert", "ReverseProxyHostTls-Validate")
		return
	}
	if o.Key == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing reverse proxy host tls key", "ReverseProxyHostTls-Validate")
		return
	}
	return
}
