package models

type Claim struct {
	BaseModel          // Embeds ID, CreatedAt, UpdatedAt, DeletedAt
	Name        string `json:"name" gorm:"column:name;type:varchar(64);unique;not null"`
	Description string `json:"description,omitempty" gorm:"column:description;type:text"`
	Internal    bool   `json:"internal" gorm:"column:internal;default:false;type:boolean"`
	Group       string `json:"group,omitempty" gorm:"column:group;type:varchar(32)"`
	Resource    string `json:"resource,omitempty" gorm:"column:resource;type:varchar(32)"`
	Action      string `json:"action,omitempty" gorm:"column:action;type:varchar(32)"`
	Users       []User `json:"-" gorm:"many2many:user_claims"`
}

func (Claim) TableName() string {
	return "claims"
}
