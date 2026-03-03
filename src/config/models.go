package config

import (
	"regexp"
	"strings"
)

type ConfigFile struct {
	Environment  map[string]string   `json:"environment,omitempty" yaml:"environment,omitempty"`
	ReverseProxy *ReverseProxyConfig `json:"reverse_proxy,omitempty" yaml:"reverse_proxy,omitempty"`
}

type ReverseProxyConfig struct {
	Host  string                    `json:"host,omitempty" yaml:"host,omitempty"`
	Port  string                    `json:"port,omitempty" yaml:"port,omitempty"`
	Ssl   *ReverseProxyConfigSsl    `json:"ssl,omitempty" yaml:"ssl,omitempty"`
	Cors  *ReverseProxyConfigCors   `json:"cors,omitempty" yaml:"cors,omitempty"`
	Hosts []*ReverseProxyConfigHost `json:"hosts,omitempty" yaml:"hosts,omitempty"`
}

func (r *ReverseProxyConfig) GetHost() string {
	if r.Port != "" {
		return strings.ToLower(r.Host + ":" + r.Port)
	}

	return strings.ToLower(r.Host)
}

type ReverseProxyConfigSsl struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Cert    string `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key     string `json:"key,omitempty" yaml:"key,omitempty"`
}

type ReverseProxyConfigCors struct {
	Enabled        bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty" yaml:"allowed_origins,omitempty"`
	AllowedMethods []string `json:"allowed_methods,omitempty" yaml:"allowed_methods,omitempty"`
	AllowedHeaders []string `json:"allowed_headers,omitempty" yaml:"allowed_headers,omitempty"`
}

type ReverseProxyConfigHost struct {
	Name       string                               `json:"name,omitempty" yaml:"name,omitempty"`
	Host       string                               `json:"host,omitempty" yaml:"host,omitempty"`
	Port       string                               `json:"port,omitempty" yaml:"port,omitempty"`
	HttpRoutes []*ReverseProxyConfigServerHttpRoute `json:"http_routes,omitempty" yaml:"http_routes,omitempty"`
	TcpRoute   *ReverseProxyConfigServerTcpRoute    `json:"tcp_route,omitempty" yaml:"tcp_route,omitempty"`
}

func (r *ReverseProxyConfigHost) GetHost() string {
	if r.Port != "" {
		return strings.ToLower(r.Host + ":" + r.Port)
	}

	return strings.ToLower(r.Host)
}

type ReverseProxyConfigServerHttpRoute struct {
	Path            string            `json:"path,omitempty" yaml:"path,omitempty"`
	TargetHost      string            `json:"target_host,omitempty" yaml:"target_host,omitempty"`
	TargetPort      string            `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	Scheme          string            `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	Pattern         string            `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	RegexpPattern   *regexp.Regexp    `json:"-" yaml:"-"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty" yaml:"request_headers,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty" yaml:"response_headers,omitempty"`
}

type ReverseProxyConfigServerTcpRoute struct {
	TargetPort string `json:"target_port,omitempty" yaml:"target_port,omitempty"`
	TargetHost string `json:"target_host,omitempty" yaml:"target_host,omitempty"`
}
