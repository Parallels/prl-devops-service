package models

type OrchestratorHostAuthentication struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	ApiKey   string `json:"api_key,omitempty"`
}

func (c *OrchestratorHostAuthentication) Diff(source OrchestratorHostAuthentication) bool {
	if c.Username != source.Username {
		return true
	}
	if c.Password != source.Password {
		return true
	}
	if c.ApiKey != source.ApiKey {
		return true
	}

	return false
}
