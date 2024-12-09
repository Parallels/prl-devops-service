package models

type Category struct {
	Name      string             `yaml:"name"`
	Path      string             `yaml:"path"`
	Endpoints []CategoryEndpoint `yaml:"endpoints,omitempty"`
}

type CategoryEndpoint struct {
	Anchor      string `yaml:"anchor,omitempty"`
	Method      string `yaml:"method,omitempty"`
	Path        string `yaml:"path,omitempty"`
	Description string `yaml:"description,omitempty"`
	Title       string `yaml:"title,omitempty"`
}

func (c *Category) AddEndpoint(endpoint Endpoint) {
	c.Endpoints = append(c.Endpoints, CategoryEndpoint{
		Anchor:      endpoint.Path + "-" + endpoint.Method,
		Method:      endpoint.Method,
		Path:        endpoint.Path,
		Description: endpoint.Description,
		Title:       endpoint.Title,
	})
}
