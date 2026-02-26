package models

import "github.com/Parallels/prl-devops-service/errors"

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

func (o *ReverseProxyHostCreateRequest) Validate() error {
	if o.Host == "" {
		return errors.NewWithCode("missing reverse proxy host", 400)
	}
	if o.Port == "" {
		return errors.NewWithCode("missing reverse proxy host port", 400)
	}

	if o.Tls != nil {
		if err := o.Tls.Validate(); err != nil {
			return err
		}
	}

	if o.Cors != nil {
		if err := o.Cors.Validate(); err != nil {
			return err
		}
	}

	if len(o.HttpRoutes) == 0 && o.TcpRoute == nil {
		return errors.NewWithCode("missing reverse proxy host routes", 400)
	}

	if len(o.HttpRoutes) > 0 && o.TcpRoute != nil {
		return errors.NewWithCode("reverse proxy host cannot have both http and tcp routes", 400)
	}

	if o.TcpRoute != nil {
		if o.Cors != nil {
			return errors.NewWithCode("reverse proxy host cannot have cors and tcp route", 400)
		}
		if o.Tls != nil {
			return errors.NewWithCode("reverse proxy host cannot have tls and tcp route", 400)
		}
		if err := o.TcpRoute.Validate(); err != nil {
			return err
		}
	}

	for _, route := range o.HttpRoutes {
		if err := route.Validate(); err != nil {
			return err
		}
	}

	return nil
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

func (o *ReverseProxyHostUpdateRequest) Validate() error {
	if o.Host == "" {
		return errors.NewWithCode("missing reverse proxy host", 400)
	}
	if o.Port == "" {
		return errors.NewWithCode("missing reverse proxy host port", 400)
	}

	if o.Tls != nil {
		if err := o.Tls.Validate(); err != nil {
			return err
		}
	}

	if o.Cors != nil {
		if err := o.Cors.Validate(); err != nil {
			return err
		}
	}
	return nil
}
