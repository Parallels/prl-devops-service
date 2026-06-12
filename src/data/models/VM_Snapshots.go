package models

type VMSnapshot struct {
	ID      string `json:"id,omitempty" gorm:"column:id;type:varchar(255)"`
	Name    string `json:"name" gorm:"column:name;type:varchar(255)"`
	Date    string `json:"date" gorm:"column:date;type:varchar(255)"`
	State   string `json:"state" gorm:"column:state;type:varchar(255)"`
	Current bool   `json:"current" gorm:"column:current;type:boolean"`
	Parent  string `json:"parent" gorm:"column:parent;type:varchar(255)"`
}

type VMSnapshots struct {
	VMId       string       `json:"vm_id"`
	VMSnapshot []VMSnapshot `json:"vm_snapshots"`
}
