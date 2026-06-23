package models

import "github.com/Parallels/prl-devops-service/constants"

type Job struct {
	ID                 string             `json:"id" gorm:"primaryKey;column:id;type:varchar(64)"`
	Owner              string             `json:"owner" gorm:"column:owner;type:varchar(255)"`
	OwnerName          string             `json:"owner_name,omitempty" gorm:"column:owner_name;type:varchar(255)"`
	OwnerEmail         string             `json:"owner_email,omitempty" gorm:"column:owner_email;type:varchar(255)"`
	State              constants.JobState `json:"state" gorm:"column:state;type:varchar(32)"`
	Message            string             `json:"message,omitempty" gorm:"column:message;type:text"`
	Progress           int                `json:"progress" gorm:"column:progress;type:integer;default:0;not null"`
	JobType            string             `json:"job_type" gorm:"column:job_type;type:varchar(64)"`
	JobOperation       string             `json:"job_operation" gorm:"column:job_operation;type:varchar(64)"`
	IsOrchestratorJob  bool               `json:"is_orchestrator_job,omitempty" gorm:"column:is_orchestrator_job;type:boolean;default:false;not null"`
	Steps              []JobStep          `json:"steps,omitempty" gorm:"column:steps;type:json;serializer:json"`
	Result             string             `json:"result" gorm:"column:result;type:varchar(255)"`
	ResultRecordId     string             `json:"result_record_id,omitempty" gorm:"column:result_record_id;type:varchar(64)"`
	ResultRecordName   string             `json:"result_record_name,omitempty" gorm:"column:result_record_name;type:varchar(255)"`
	ResultRecordType   string             `json:"result_record_type,omitempty" gorm:"column:result_record_type;type:varchar(64)"`
	ResultRecordLinkId string             `json:"result_record_link_id,omitempty" gorm:"column:result_record_link_id;type:varchar(255)"`
	Error              string             `json:"error" gorm:"column:error;type:varchar(255)"`
	CreatedAt          string             `json:"created_at" gorm:"column:created_at;type:timestamp"`
	UpdatedAt          string             `json:"updated_at" gorm:"column:updated_at;type:timestamp"`
	*DbRecord          `json:"db_record" gorm:"embedded"`
}

type JobStep struct {
	Name              string             `json:"name"`
	DisplayName       string             `json:"display_name,omitempty"`
	Weight            float64            `json:"weight"`
	Parallel          bool               `json:"parallel"`
	HasPercentage     bool               `json:"has_percentage"`
	State             constants.JobState `json:"state"`
	CurrentPercentage float64            `json:"current_percentage"`
	Value             int64              `json:"value"`
	Total             int64              `json:"total"`
	ETA               string             `json:"eta"`
	Message           string             `json:"message"`
	Error             string             `json:"error"`
	Filename          string             `json:"filename"`
	Unit              string             `json:"unit"`
}
