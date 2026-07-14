package models

import "time"

type VMSnapshot struct {
	ID      string    `json:"id,omitempty" gorm:"primaryKey;column:id;type:varchar(64);not null"`
	Name    string    `json:"name" gorm:"column:name;type:varchar(255);not null"`
	Date    time.Time `json:"date" gorm:"column:date;type:timestamp;not null"`
	State   string    `json:"state" gorm:"column:state;type:varchar(32);not null"`
	Current bool      `json:"current" gorm:"column:current;type:boolean"`
	Parent  string    `json:"parent" gorm:"column:parent;type:varchar(64)"`
}

func (VMSnapshot) TableName() string {
	return "vm_snapshots"
}

type VMSnapshots struct {
	VMId       string       `json:"vm_id"`
	VMSnapshot []VMSnapshot `json:"vm_snapshots"`
}
