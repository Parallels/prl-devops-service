package models

type ReverseProxyConfig struct {
	ID      string `json:"-"`
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Host    string `json:"host,omitempty" yaml:"host,omitempty"`
	Port    string `json:"port,omitempty" yaml:"port,omitempty"`
}
