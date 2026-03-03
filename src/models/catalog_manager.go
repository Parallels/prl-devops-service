package models

type CatalogManager struct {
	ID                   string   `json:"id" yaml:"id"`
	Name                 string   `json:"name" yaml:"name"`
	URL                  string   `json:"url" yaml:"url"`
	Internal             bool     `json:"internal" yaml:"internal"`
	Active               bool     `json:"active" yaml:"active"`
	AuthenticationMethod string   `json:"authentication_method" yaml:"authentication_method"`
	Username             string   `json:"username,omitempty" yaml:"username,omitempty"`
	Password             string   `json:"password,omitempty" yaml:"password,omitempty"`
	ApiKey               string   `json:"api_key,omitempty" yaml:"api_key,omitempty"`
	Global               bool     `json:"global" yaml:"global"`
	RequiredClaims       []string `json:"required_claims,omitempty" yaml:"required_claims,omitempty"`
	OwnerID              string   `json:"owner_id" yaml:"owner_id"`
	CreatedAt            string   `json:"created_at" yaml:"created_at"`
	UpdatedAt            string   `json:"updated_at" yaml:"updated_at"`
}

type CatalogManagerRequest struct {
	Name                 string   `json:"name" yaml:"name"`
	URL                  string   `json:"url" yaml:"url"`
	Internal             bool     `json:"internal" yaml:"internal"`
	Active               bool     `json:"active" yaml:"active"`
	AuthenticationMethod string   `json:"authentication_method" yaml:"authentication_method"`
	Username             string   `json:"username,omitempty" yaml:"username,omitempty"`
	Password             string   `json:"password,omitempty" yaml:"password,omitempty"`
	ApiKey               string   `json:"api_key,omitempty" yaml:"api_key,omitempty"`
	Global               bool     `json:"global" yaml:"global"`
	RequiredClaims       []string `json:"required_claims,omitempty" yaml:"required_claims,omitempty"`
}
