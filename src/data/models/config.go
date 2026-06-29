package models

type Configuration struct {
	ID string `json:"id" yaml:"id" gorm:"primaryKey;type:varchar(64);column:id"`
}
