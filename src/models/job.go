package models

import "github.com/Parallels/prl-devops-service/constants"

type JobResponse struct {
	ID               string             `json:"id"`
	Owner            string             `json:"owner"`
	OwnerName        string             `json:"owner_name,omitempty"`
	OwnerEmail       string             `json:"owner_email,omitempty"`
	State            constants.JobState `json:"state"`
	Progress         int                `json:"progress"`
	JobType          string             `json:"job_type"`
	JobOperation     string             `json:"job_operation"`
	Action           string             `json:"action"`
	ActionMessage    string             `json:"action_message,omitempty"`
	ActionValue      int64              `json:"action_value,omitempty"`
	ActionPercentage int                `json:"action_percentage,omitempty"`
	ActionTotal      int64              `json:"action_total,omitempty"`
	ActionETA        string             `json:"action_eta,omitempty"`
	ActionValueUnit  string             `json:"action_value_unit,omitempty"`
	Result           string             `json:"result,omitempty"`
	Error            string             `json:"error,omitempty"`
	CreatedAt        string             `json:"created_at"`
	UpdatedAt        string             `json:"updated_at"`
}

type JobCreateRequest struct {
	Action       string `json:"action"`
	JobType      string `json:"job_type"`
	JobOperation string `json:"job_operation"`
}
