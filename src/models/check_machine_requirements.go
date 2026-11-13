package models

import (
	"os"
	"strings"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
)

// MachineRequirements describes the minimum resources a virtual machine needs.
// It is accepted as explicit input to the pre-flight check and echoed back as the
// resolved set of requirements the check was evaluated against.
type MachineRequirements struct {
	Type   string `json:"type,omitempty"`   // pvm | macvm (resolved from os)
	Cpu    int    `json:"cpu,omitempty"`    // logical CPUs
	Memory int    `json:"memory,omitempty"` // MiB
	Disk   int    `json:"disk,omitempty"`   // MiB (VM disk setting)
	Size   int64  `json:"size,omitempty"`   // pack size in bytes; drives the host free-disk feasibility check
	Os     string `json:"os,omitempty"`     // request-only hint; "macos" selects the macvm decision path
}

// GetSpecsType resolves the VM type implied by the request: "macos" (any case)
// maps to "macvm", anything else defaults to "pvm".
func (m *MachineRequirements) GetSpecsType() string {
	if m == nil {
		return "pvm"
	}
	if strings.EqualFold(m.Type, "macvm") || strings.EqualFold(m.Os, "macos") {
		return "macvm"
	}
	return "pvm"
}

// CheckMachineRequirementsRequest is the body of POST /v1/orchestrator/machines/check.
// Exactly one of CatalogId or Requirements must be provided.
type CheckMachineRequirementsRequest struct {
	CatalogId        string               `json:"catalog_id,omitempty"`
	Version          string               `json:"version,omitempty"`
	Architecture     string               `json:"architecture,omitempty"`
	Connection       string               `json:"connection,omitempty"`
	CatalogManagerId string               `json:"catalog_manager_id,omitempty"`
	HostId           string               `json:"host_id,omitempty"`
	Owner            string               `json:"owner,omitempty"`
	Requirements     *MachineRequirements `json:"requirements,omitempty"`
	// TargetTags mirrors the create request's hard host-tag placement
	// constraint: only hosts carrying ALL listed tags (case-insensitive)
	// pass the host_tags check. Absent means no tag constraint.
	TargetTags []string `json:"target_tags,omitempty"`
}

func (r *CheckMachineRequirementsRequest) Validate() error {
	catalogId := strings.TrimSpace(r.CatalogId)
	if catalogId == "" && r.Requirements == nil {
		return errors.NewWithCode("either catalog_id or requirements must be provided", 400)
	}
	if catalogId != "" && r.Requirements != nil {
		return errors.NewWithCode("catalog_id and requirements cannot both be provided", 400)
	}

	connection := strings.TrimSpace(r.Connection)
	catalogManagerId := strings.TrimSpace(r.CatalogManagerId)
	if connection != "" && catalogManagerId != "" {
		return errors.NewWithCode("connection and catalog_manager_id cannot both be provided", 400)
	}

	if catalogId != "" {
		r.CatalogId = catalogId
		if r.Version == "" {
			r.Version = constants.LATEST_TAG
		}
		if strings.TrimSpace(r.Architecture) == "" {
			return errors.NewWithCode("architecture is required when catalog_id is provided", 400)
		}
	}

	if r.Requirements != nil {
		if r.Requirements.Cpu <= 0 {
			r.Requirements.Cpu = 2
		}
		if r.Requirements.Memory <= 0 {
			r.Requirements.Memory = 2048
		}
	}

	targetTags, err := NormalizeTagConstraints(r.TargetTags)
	if err != nil {
		return err
	}
	r.TargetTags = targetTags

	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}

	return nil
}

// CheckMachineRequirementsResponse is the pre-flight verdict. When Passed is true the
// response stays slim and no per-host detail is emitted; the hosts array is only
// populated when no host meets the requirements, listing exactly what each host is missing.
type CheckMachineRequirementsResponse struct {
	Passed       bool                            `json:"passed"`
	Requirements *MachineRequirements            `json:"requirements,omitempty"`
	Message      string                          `json:"message,omitempty"`
	Hosts        []MachineRequirementsHostResult `json:"hosts,omitempty"`
}

// MachineRequirementsHostResult carries the per-check breakdown for one evaluated host.
type MachineRequirementsHostResult struct {
	HostId string                   `json:"host_id,omitempty"`
	Host   string                   `json:"host,omitempty"`
	Passed bool                     `json:"passed"`
	Checks []RequirementCheckResult `json:"checks"`
}

// RequirementCheckResult is the outcome of a single decision-factor check on a host.
type RequirementCheckResult struct {
	Name      string   `json:"name"` // host_tags | host_state | architecture | cpu | memory | disk | vm_count | macos_inventory | macos_capacity
	Passed    bool     `json:"passed"`
	Required  *float64 `json:"required,omitempty"`
	Available *float64 `json:"available,omitempty"`
	Unit      string   `json:"unit,omitempty"` // logical_cpus | MiB | vms | apple_vms
	Message   string   `json:"message,omitempty"`
}

// NewRequirementCheckValue returns a pointer to v, for the Required/Available fields.
func NewRequirementCheckValue(v float64) *float64 {
	return &v
}
