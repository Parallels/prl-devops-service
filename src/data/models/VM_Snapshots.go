package models

type VMSnapshot struct {
	ID            string `json:"id,omitempty" gorm:"primaryKey;column:id;type:varchar(64);not null"`
	Name          string `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Date          string `json:"date" gorm:"column:date;type:timestamp;not null"`
	State         string `json:"state" gorm:"column:state;type:varchar(32);not null"`
	Current       bool   `json:"current" gorm:"column:current;type:boolean;not null"`
	Parent        string `json:"parent" gorm:"column:parent;type:varchar(64);not null"`
	CreatedBy     string `json:"created_by,omitempty" gorm:"column:created_by;type:varchar(64)"`
	CreatedByUser *User  `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type VMSnapshots struct {
	VMId       string       `json:"vm_id"`
	VMSnapshot []VMSnapshot `json:"vm_snapshots"`
}
