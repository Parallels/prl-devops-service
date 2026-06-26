package entities


type UserRoles struct {
	UserID string `json:"user_id" gorm:"not null;type:text;index;default:'00000000-0000-0000-0000-000000000000'"`
	RoleID string `json:"role_id" gorm:"not null;type:text;index;default:'00000000-0000-0000-0000-000000000000'"`
}

func (UserRoles) TableName() string {
	return "user_roles"
}
