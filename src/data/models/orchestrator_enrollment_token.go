package models

type OrchestratorEnrollmentToken struct {
	ID        string    `json:"id"`
	Token     string    `json:"token"`
	HostName  string    `json:"host_name"`
	Used      bool      `json:"used"`
	ExpiresAt string    `json:"expires_at"`
	CreatedAt string    `json:"created_at"`
	*DbRecord `json:"db_record"`
}
