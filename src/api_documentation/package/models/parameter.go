package models

type Parameter struct {
	Name        string `yaml:"name"`
	Required    bool   `yaml:"required"`
	Type        string `yaml:"type,omitempty"`
	ValueType   string `yaml:"value_type,omitempty"`
	Description string `yaml:"description,omitempty"`
	Body        string `yaml:"body,omitempty"`
}
