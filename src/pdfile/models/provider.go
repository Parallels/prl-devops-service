package models

type PDFileProvider struct {
	Name       string            `json:"NAME,omitempty" yaml:"NAME,omitempty"`
	Attributes map[string]string `json:"ATTRIBUTES,omitempty" yaml:"ATTRIBUTES,omitempty"`
}
