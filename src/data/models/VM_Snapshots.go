package models

type VMSnapshot struct {
	ID      string `json:"id,omitempty" gorm:"column:id;type:varchar(255);primaryKey;not null"`
	Name    string `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Date    string `json:"date" gorm:"column:date;type:timestamp;not null"`
	State   string `json:"state" gorm:"column:state;type:varchar(255);not null"`
	Current bool   `json:"current" gorm:"column:current;type:boolean;not null"`
	Parent  string `json:"parent" gorm:"column:parent;type:varchar(255);not null"`
}

type VMSnapshots struct {
	VMId       string       `json:"vm_id"`
	VMSnapshot []VMSnapshot `json:"vm_snapshots"`
}
