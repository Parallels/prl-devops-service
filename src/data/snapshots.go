package data

import (
	"errors"

	"github.com/Parallels/prl-devops-service/data/models"
)

func (j *JsonDatabase) SetListSnapshotsByVMId(vmID string, vmSnap models.VMSnapshot) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	for i, vmSnap := range j.data.VMSnapshots {
		if vmSnap.VMId == vmID {
			j.data.VMSnapshots[i] = vmSnap
			return nil
		}
	}
	j.data.VMSnapshots = append(j.data.VMSnapshots, vmSnap)
	return nil
}

func (j *JsonDatabase) GetListSnapshotsByVMId(vmID string) ([]models.Snapshot, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()
	for _, vmSnap := range j.data.VMSnapshots {
		if vmSnap.VMId == vmID {
			return vmSnap.Snapshots, nil
		}
	}
	return nil, errors.New("snapshots not found for VM ID: " + vmID)
}
