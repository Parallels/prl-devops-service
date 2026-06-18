package models

type CatalogManager struct {
	ID                   string   `json:"id" yaml:"id" gorm:"primaryKey;type:varchar(255);column:id;not null"`
	Name                 string   `json:"name" yaml:"name" gorm:"column:name;not null;type:varchar(255);index"`
	URL                  string   `json:"url" yaml:"url" gorm:"column:url;not null;type:varchar(255);index"`
	Internal             bool     `json:"internal" yaml:"internal" gorm:"column:internal;not null;type:boolean;default:0"`
	Active               bool     `json:"active" yaml:"active" gorm:"column:active;not null;type:boolean;default:0"`
	AuthenticationMethod string   `json:"authentication_method" yaml:"authentication_method" gorm:"column:authentication_method;not null;type:varchar(255)"`
	Username             string   `json:"username,omitempty" yaml:"username,omitempty" gorm:"column:username;type:varchar(255)"`
	Password             string   `json:"password,omitempty" yaml:"password,omitempty" gorm:"column:password;type:varchar(255)"`
	ApiKey               string   `json:"api_key,omitempty" yaml:"api_key,omitempty" gorm:"column:api_key;type:varchar(255)"`
	Global               bool     `json:"global" yaml:"global" gorm:"column:global;not null;type:boolean;default:0"`
	RequiredClaims       []string `json:"required_claims,omitempty" yaml:"required_claims,omitempty" gorm:"column:required_claims;type:json;serializer:json"`
	OwnerID              string   `json:"owner_id" yaml:"owner_id" gorm:"column:owner_id;type:varchar(255)"`
	CreatedAt            string   `json:"created_at" yaml:"created_at" gorm:"column:created_at;type:timestamp"`
	UpdatedAt            string   `json:"updated_at" yaml:"updated_at" gorm:"column:updated_at;type:timestamp"`
}
