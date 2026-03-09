package models

import "github.com/Parallels/prl-devops-service/constants"

type Job struct {
	ID               string             `json:"id"`
	Owner            string             `json:"owner"`
	OwnerName        string             `json:"owner_name,omitempty"`
	OwnerEmail       string             `json:"owner_email,omitempty"`
	State            constants.JobState `json:"state"`
	Message          string             `json:"message,omitempty"`
	Progress         int                `json:"progress"`
	JobType          string             `json:"job_type"`
	JobOperation     string             `json:"job_operation"`
	Steps            []JobStep          `json:"steps,omitempty"`
	Result           string             `json:"result"`
	ResultRecordId   string             `json:"result_record_id,omitempty"`
	ResultRecordType string             `json:"result_record_type,omitempty"`
	Error            string             `json:"error"`
	CreatedAt        string             `json:"created_at"`
	UpdatedAt        string             `json:"updated_at"`
	*DbRecord        `json:"db_record"`
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
