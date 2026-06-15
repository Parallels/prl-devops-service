package models

type OrchestratorEnrollmentToken struct {
	ID        string `json:"id" gorm:"primaryKey;column:id;type:varchar(255)"`
	Token     string `json:"token" gorm:"column:token;type:varchar(255)"`
	HostName  string `json:"host_name" gorm:"column:host_name;type:varchar(255)"`
	Used      bool   `json:"used" gorm:"column:used;type:boolean;default:false;not null"`
	ExpiresAt string `json:"expires_at" gorm:"column:expires_at;type:timestamp"`
	CreatedAt string `json:"created_at" gorm:"column:created_at;type:timestamp"`
	*DbRecord `json:"db_record"  gorm:"embedded"`
}
