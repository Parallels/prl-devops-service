package models

// JobLink maps a host-side job back to the orchestrator job that spawned it.
type JobLink struct {
	HostJobID         string `json:"host_job_id" gorm:"primaryKey;column:host_job_id;type:varchar(64)"`
	OrchestratorJobID string `json:"orchestrator_job_id" gorm:"column:orchestrator_job_id;type:varchar(64)"`
	HostID            string `json:"host_id" gorm:"column:host_id;type:varchar(64)"`
	*DbRecord         `json:"db_record" gorm:"embedded"`
}
