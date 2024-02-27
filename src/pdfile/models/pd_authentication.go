package models

type PDFileAuthentication struct {
	Username string `json:"USERNAME,omitempty" yaml:"USERNAME,omitempty"`
	Password string `json:"PASSWORD,omitempty" yaml:"PASSWORD,omitempty"`
	ApiKey   string `json:"API_KEY,omitempty" yaml:"API_KEY,omitempty"`
}
