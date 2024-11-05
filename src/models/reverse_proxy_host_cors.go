package models

type ReverseProxyHostCors struct {
	Enabled        bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	AllowedOrigins []string `json:"allowed_origins,omitempty" yaml:"allowed_origins,omitempty"`
	AllowedMethods []string `json:"allowed_methods,omitempty" yaml:"allowed_methods,omitempty"`
	AllowedHeaders []string `json:"allowed_headers,omitempty" yaml:"allowed_headers,omitempty"`
}

func (o *ReverseProxyHostCors) Validate() error {
	if o.Enabled {
		if len(o.AllowedOrigins) == 0 {
			o.AllowedOrigins = []string{"*"}
		}
		if len(o.AllowedMethods) == 0 {
			o.AllowedHeaders = []string{"X-Requested-With", "authorization", "Authorization", "content-type"}
		}
		if len(o.AllowedHeaders) == 0 {
			o.AllowedHeaders = []string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"}
		}
	}
	return nil
}
