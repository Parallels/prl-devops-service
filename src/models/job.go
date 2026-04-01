package models

import "github.com/Parallels/prl-devops-service/constants"

type JobResponse struct {
	ID                 string             `json:"id"`
	Owner              string             `json:"owner"`
	OwnerName          string             `json:"owner_name,omitempty"`
	OwnerEmail         string             `json:"owner_email,omitempty"`
	State              constants.JobState `json:"state"`
	Message            string             `json:"message,omitempty"`
	Progress           int                `json:"progress"`
	JobType            string             `json:"job_type"`
	JobOperation       string             `json:"job_operation"`
	Steps              []JobStepResponse  `json:"steps,omitempty"`
	Result             string             `json:"result,omitempty"`
	ResultRecordId     string             `json:"result_record_id,omitempty"`
	ResultRecordName   string             `json:"result_record_name,omitempty"`
	ResultRecordType   string             `json:"result_record_type,omitempty"`
	ResultRecordLinkId string             `json:"result_record_link_id,omitempty"`
	Error              string             `json:"error,omitempty"`
	CreatedAt          string             `json:"created_at"`
	UpdatedAt          string             `json:"updated_at"`
}

type JobStepResponse struct {
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

type JobCreateRequest struct {
	Action       string `json:"action"`
	JobType      string `json:"job_type"`
	JobOperation string `json:"job_operation"`
	// Profile selects the scenario for /jobs/debug:
	//   "simple"       – linear 0→100% progress (default)
	//   "pull_remote"  – simulates a full multi-step remote pull (validate, download, decompress, register)
	//   "pull_cache"   – simulates a cached pull (validate, copy-from-cache, register)
	//   "skipped_steps"– simulates steps that are skipped / instant-complete
	Profile string `json:"profile"`
}
