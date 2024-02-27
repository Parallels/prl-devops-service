package models

type PullResponse struct {
	MachineId    string `json:"machine_id,omitempty" yaml:"machine_id,omitempty"`
	MachineName  string `json:"machine_name,omitempty" yaml:"machine_name,omitempty"`
	CatalogId    string `json:"catalog_id,omitempty" yaml:"catalog_id,omitempty"`
	Version      string `json:"version,omitempty" yaml:"version,omitempty"`
	Architecture string `json:"architecture,omitempty" yaml:"architecture,omitempty"`
	Type         string `json:"type,omitempty" yaml:"type,omitempty"`
}
