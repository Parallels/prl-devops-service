package models

import (
	"github.com/Parallels/prl-devops-service/models"
)

type OrchestratorSnapshot struct {
	HostId string `json:"host_id"`
	// map of vmId to list of snapshots
	Snapshots map[string]models.ListSnapshotResponse `json:"snapshots"`
}
