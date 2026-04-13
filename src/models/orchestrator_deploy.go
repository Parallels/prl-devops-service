package models

import "github.com/Parallels/prl-devops-service/errors"

// DeployOrchestratorHostRequest is the body for the SSH-based agent deployment endpoints.
type DeployOrchestratorHostRequest struct {
	// SSH connection details for the target host
	SshHost            string `json:"ssh_host"`
	SshPort            string `json:"ssh_port,omitempty"` // defaults to "22"
	SshUser            string `json:"ssh_user"`
	SshPassword        string `json:"ssh_password,omitempty"`          // mutually exclusive with SshKey
	SshKey             string `json:"ssh_key,omitempty"`               // PEM-encoded private key
	SshInsecureHostKey bool   `json:"ssh_insecure_host_key,omitempty"` // skip known_hosts verification
	SudoPassword       string `json:"sudo_password,omitempty"`         // sudo password if different from ssh_password

	// Agent identity in the orchestrator
	HostName string   `json:"host_name"`
	Tags     []string `json:"tags,omitempty"`

	// Install options forwarded to the install script
	RootPassword   string `json:"root_password,omitempty"`
	EnabledModules string `json:"enabled_modules,omitempty"` // e.g. "api,host,catalog,cors"
	PdVersion      string `json:"pd_version,omitempty"`      // "latest" or explicit e.g. "26.2.2-57373"

	// AgentVersion pins the prldevops binary version installed on the remote host
	// (e.g. "v0.7.0-beta"). When empty the install script uses its own default (latest stable).
	AgentVersion string `json:"agent_version,omitempty"`
	// PreRelease instructs the install script to pick the latest pre-release tag
	// instead of the latest stable release. Ignored when AgentVersion is set.
	PreRelease bool `json:"pre_release,omitempty"`
	// AgentPort is the port the installed agent listens on. Defaults to 3080.
	AgentPort string `json:"agent_port,omitempty"`

	// EnrollmentTokenTTL overrides the default 15-minute TTL (minutes)
	EnrollmentTokenTTL int `json:"enrollment_token_ttl,omitempty"`
}

func (r *DeployOrchestratorHostRequest) Validate() error {
	if r.SshHost == "" {
		return errors.NewWithCode("ssh_host is required", 400)
	}
	if r.SshUser == "" {
		return errors.NewWithCode("ssh_user is required", 400)
	}
	if r.SshPassword == "" && r.SshKey == "" {
		return errors.NewWithCode("either ssh_password or ssh_key is required", 400)
	}
	if r.HostName == "" {
		return errors.NewWithCode("host_name is required", 400)
	}
	if r.SshPort == "" {
		r.SshPort = "22"
	}
	return nil
}

// DeployOrchestratorHostResponse is returned by the synchronous deploy endpoint.
type DeployOrchestratorHostResponse struct {
	HostID  string `json:"host_id"`
	Host    string `json:"host"`
	Message string `json:"message,omitempty"`
}

// CreateEnrollmentTokenRequest is the body for POST /orchestrator/enrollment-token.
type CreateEnrollmentTokenRequest struct {
	HostName   string `json:"host_name"`
	TTLMinutes int    `json:"ttl_minutes,omitempty"` // defaults to 15
}

func (r *CreateEnrollmentTokenRequest) Validate() error {
	if r.HostName == "" {
		return errors.NewWithCode("host_name is required", 400)
	}
	if r.TTLMinutes <= 0 {
		r.TTLMinutes = 15
	}
	return nil
}

// CreateEnrollmentTokenResponse is returned when an enrollment token is generated.
type CreateEnrollmentTokenResponse struct {
	Token     string `json:"token"`
	HostName  string `json:"host_name"`
	ExpiresAt string `json:"expires_at"`
}

// ValidateEnrollmentTokenResponse is returned by the public validate endpoint.
type ValidateEnrollmentTokenResponse struct {
	Valid     bool   `json:"valid"`
	HostName  string `json:"host_name,omitempty"`
	ExpiresAt string `json:"expires_at,omitempty"`
	Reason    string `json:"reason,omitempty"`
}
