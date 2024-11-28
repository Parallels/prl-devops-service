package models

import (
	"regexp"
	"strings"
)

type ReverseProxy struct {
	ID      string `json:"id,omitempty"`
	HostID  string `json:"host_id,omitempty"`
	Enabled bool   `json:"enabled"`
	Host    string `json:"host,omitempty"`
	Port    string `json:"port,omitempty"`
}

func (o *ReverseProxy) Diff(source ReverseProxy) bool {
	if o == nil {
		return true
	}
	if o.Enabled != source.Enabled {
		return true
	}
	if o.Host != source.Host {
		return true
	}
	if o.Port != source.Port {
		return true
	}
	if o.HostID != source.HostID {
		return true
	}

	return false
}

type ReverseProxyHost struct {
	ID         string                       `json:"id"`
	HostID     string                       `json:"host_id,omitempty"`
	Host       string                       `json:"host"`
	Port       string                       `json:"port"`
	Tls        *ReverseProxyHostTls         `json:"tls,omitempty"`
	Cors       *ReverseProxyHostCors        `json:"cors,omitempty"`
	HttpRoutes []*ReverseProxyHostHttpRoute `json:"http_routes,omitempty"`
	TcpRoute   *ReverseProxyHostTcpRoute    `json:"tcp_route,omitempty"`
}

func (o *ReverseProxyHost) GetHost() string {
	if o.Port != "" {
		return strings.ToLower(o.Host + ":" + o.Port)
	}

	return strings.ToLower(o.Host)
}

func (o *ReverseProxyHost) Diff(source ReverseProxyHost) bool {
	if o == nil {
		return true
	}
	if o.HostID != source.HostID {
		return true
	}

	if o.Host != source.Host {
		return true
	}
	if o.Port != source.Port {
		return true
	}

	if o.Tls == nil && source.Tls != nil {
		return true
	}

	if o.Tls != nil && source.Tls == nil {
		return true
	}

	if o.Tls != nil && source.Tls != nil {
		if o.Tls.Diff(*source.Tls) {
			return true
		}
	}

	if o.Cors == nil && source.Cors != nil {
		return true
	}

	if o.Cors != nil && source.Cors == nil {
		return true
	}

	if o.Cors != nil && source.Cors != nil {
		if o.Cors.Diff(*source.Cors) {
			return true
		}
	}

	if len(o.HttpRoutes) != len(source.HttpRoutes) {
		return true
	}
	for i, route := range o.HttpRoutes {
		if route.Diff(*source.HttpRoutes[i]) {
			return true
		}
	}

	if o.TcpRoute == nil && source.TcpRoute != nil {
		return true
	}
	if o.TcpRoute != nil && source.TcpRoute == nil {
		return true
	}
	if o.TcpRoute != nil && source.TcpRoute != nil {
		if o.TcpRoute.Diff(*source.TcpRoute) {
			return true
		}
	}

	return false
}

type ReverseProxyHostCors struct {
	Enabled        bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty" yaml:"allowed_origins,omitempty"`
	AllowedMethods []string `json:"allowed_methods,omitempty" yaml:"allowed_methods,omitempty"`
	AllowedHeaders []string `json:"allowed_headers,omitempty" yaml:"allowed_headers,omitempty"`
}

func (o *ReverseProxyHostCors) Diff(source ReverseProxyHostCors) bool {
	if o == nil {
		return true
	}

	if o.Enabled != source.Enabled {
		return true
	}
	if len(o.AllowedOrigins) != len(source.AllowedOrigins) {
		return true
	}
	for i, origin := range o.AllowedOrigins {
		if origin != source.AllowedOrigins[i] {
			return true
		}
	}
	if len(o.AllowedMethods) != len(source.AllowedMethods) {
		return true
	}
	for i, method := range o.AllowedMethods {
		if method != source.AllowedMethods[i] {
			return true
		}
	}
	if len(o.AllowedHeaders) != len(source.AllowedHeaders) {
		return true
	}
	for i, header := range o.AllowedHeaders {
		if header != source.AllowedHeaders[i] {
			return true
		}
	}
	return false
}

type ReverseProxyHostTls struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Cert    string `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key     string `json:"key,omitempty" yaml:"key,omitempty"`
}

func (o *ReverseProxyHostTls) Diff(source ReverseProxyHostTls) bool {
	if o == nil {
		return true
	}

	if o.Enabled != source.Enabled {
		return true
	}
	if o.Cert != source.Cert {
		return true
	}
	if o.Key != source.Key {
		return true
	}
	return false
}

type ReverseProxyHostHttpRoute struct {
	ID              string            `json:"id,omitempty" yaml:"id,omitempty"`
	Path            string            `json:"path,omitempty" yaml:"path,omitempty"`
	TargetVmId      string            `json:"target_vm_id,omitempty" yaml:"target_vm_id,omitempty"`
	TargetHost      string            `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetPort      string            `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	Schema          string            `json:"schema,omitempty" yaml:"scheme,omitempty"`
	Pattern         string            `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	RegexpPattern   *regexp.Regexp    `json:"-" yaml:"-"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty" yaml:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty" yaml:"response_headers,omitempty"`
}

func (o *ReverseProxyHostHttpRoute) Diff(source ReverseProxyHostHttpRoute) bool {
	if o == nil {
		return true
	}

	if o.Path != source.Path {
		return true
	}
	if o.TargetVmId != source.TargetVmId {
		return true
	}
	if o.TargetHost != source.TargetHost {
		return true
	}
	if o.TargetPort != source.TargetPort {
		return true
	}
	if o.Schema != source.Schema {
		return true
	}

	if o.Pattern != source.Pattern {
		return true
	}

	if o.RequestHeaders == nil && source.RequestHeaders != nil {
		return true
	}
	if o.RequestHeaders != nil && source.RequestHeaders == nil {
		return true
	}
	if len(o.RequestHeaders) != len(source.RequestHeaders) {
		return true
	}
	for key, value := range o.RequestHeaders {
		if source.RequestHeaders[key] != value {
			return true
		}
	}

	if o.ResponseHeaders == nil && source.ResponseHeaders != nil {
		return true
	}
	if o.ResponseHeaders != nil && source.ResponseHeaders == nil {
		return true
	}
	if len(o.ResponseHeaders) != len(source.ResponseHeaders) {
		return true
	}
	for key, value := range o.ResponseHeaders {
		if source.ResponseHeaders[key] != value {
			return true
		}
	}
	return false
}

func (r *ReverseProxyHostHttpRoute) GetRoute() string {
	if r.Path != "" {
		return r.Path
	}

	if r.Pattern != "" {
		return r.Pattern
	}

	return ""
}

type ReverseProxyHostTcpRoute struct {
	ID         string `json:"id,omitempty" yaml:"id,omitempty"`
	TargetPort string `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	TargetHost string `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetVmId string `json:"target_vm_id,omitempty" yaml:"target_vm_id,omitempty"`
}

func (o *ReverseProxyHostTcpRoute) Diff(source ReverseProxyHostTcpRoute) bool {
	if o == nil {
		return true
	}

	if o.TargetPort != source.TargetPort {
		return true
	}
	if o.TargetHost != source.TargetHost {
		return true
	}
	if o.TargetVmId != source.TargetVmId {
		return true
	}
	return false
}
