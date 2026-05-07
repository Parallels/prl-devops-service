package models

import (
	"net/http"
	"strconv"

	"github.com/Parallels/prl-devops-service/errors"
)

type ReverseProxyHost struct {
	ID         string                       `json:"id"`
	Name       string                       `json:"name,omitempty" yaml:"name,omitempty"`
	Host       string                       `json:"host"`
	Port       string                       `json:"port"`
	Tls        *ReverseProxyHostTls         `json:"tls,omitempty"`
	Cors       *ReverseProxyHostCors        `json:"cors,omitempty"`
	HttpRoutes []*ReverseProxyHostHttpRoute `json:"http_routes,omitempty"`
	TcpRoute   *ReverseProxyHostTcpRoute    `json:"tcp_route,omitempty"`
}

type ReverseProxyRouteVmDetails struct {
	Name                  string `json:"name,omitempty"`
	State                 string `json:"state,omitempty"`
	OS                    string `json:"os,omitempty"`
	Uptime                string `json:"uptime,omitempty"`
	GuestToolsState       string `json:"guest_tools_state,omitempty"`
	GuestToolsVersion     string `json:"guest_tools_version,omitempty"`
	InternalIpAddress     string `json:"internal_ip_address,omitempty"`
	HostExternalIpAddress string `json:"host_external_ip_address,omitempty"`
}

func (o *ReverseProxyHost) GetHost() string {
	if o.Port != "" {
		return o.Host + ":" + o.Port
	}
	return o.Host
}

type ReverseProxyHostCreateRequest struct {
	Name       string                       `json:"name,omitempty" yaml:"name,omitempty"`
	Host       string                       `json:"host"`
	Port       string                       `json:"port"`
	Tls        *ReverseProxyHostTls         `json:"tls,omitempty"`
	Cors       *ReverseProxyHostCors        `json:"cors,omitempty"`
	HttpRoutes []*ReverseProxyHostHttpRoute `json:"http_routes,omitempty"`
	TcpRoute   *ReverseProxyHostTcpRoute    `json:"tcp_route,omitempty"`
}

func (o *ReverseProxyHostCreateRequest) GetHost() string {
	if o.Port != "" {
		return o.Host + ":" + o.Port
	}
	return o.Host
}

func (o *ReverseProxyHostCreateRequest) Validate(diag *errors.Diagnostics) {
	if o.Host == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing reverse proxy host", "")
		return
	}
	if o.Port == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing reverse proxy host port", "")
		return
	}

	if o.Tls != nil {
		o.Tls.Validate(diag)
		if diag.HasErrors() {
			return
		}
	}

	if o.Cors != nil {
		err := o.Cors.Validate()
		if err != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), err.Error(), "")
			return
		}
	}

	if len(o.HttpRoutes) == 0 && o.TcpRoute == nil {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing reverse proxy host routes", "")
		return
	}

	if len(o.HttpRoutes) > 0 && o.TcpRoute != nil {
		diag.AddError("400", "reverse proxy host cannot have both http and tcp routes", "")
		return
	}

	if o.TcpRoute != nil {
		if o.Cors != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy host cannot have cors and tcp route", "")
			return
		}
		if o.Tls != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), "reverse proxy host cannot have tls and tcp route", "")
			return
		}

		o.TcpRoute.Validate(diag)
		if diag.HasErrors() {
			return
		}
	}

	for _, route := range o.HttpRoutes {
		route.Validate(diag)
		if diag.HasErrors() {
			return
		}
	}

	return
}

type ReverseProxyHostUpdateRequest struct {
	Name string                `json:"name,omitempty" yaml:"name,omitempty"`
	Host string                `json:"host"`
	Port string                `json:"port"`
	Tls  *ReverseProxyHostTls  `json:"tls,omitempty"`
	Cors *ReverseProxyHostCors `json:"cors,omitempty"`
}

func (o *ReverseProxyHostUpdateRequest) GetHost() string {
	if o.Port != "" {
		return o.Host + ":" + o.Port
	}
	return o.Host
}

func (o *ReverseProxyHostUpdateRequest) Validate(diag *errors.Diagnostics) {
	if o.Host == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing reverse proxy host", "")
		return
	}
	if o.Port == "" {
		diag.AddError(strconv.Itoa(http.StatusBadRequest), "missing reverse proxy host port", "")
		return
	}

	if o.Tls != nil {
		o.Tls.Validate(diag)
		if diag.HasErrors() {
			return
		}
	}

	if o.Cors != nil {
		err := o.Cors.Validate()
		if err != nil {
			diag.AddError(strconv.Itoa(http.StatusBadRequest), err.Error(), "")
			return
		}
	}

	return
}
