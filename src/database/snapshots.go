package database

// VM-snapshot rows on GORM (PBI #421 task 14, second half). #419 D-3
// normalised BOTH legacy stores into the one `vm_snapshots` table:
//   - global j.data.VMSnapshots      -> rows with vm_id=VMId, host_id=''
//   - j.data.HostsVMSnapshots records -> rows with host_id+vm_id carriers
//     (HostID/VMID are json:"-" carriers, so wire JSON is unchanged).
// Mirrored quirks: host record lookup is EXACT hostId (no Fold, :896);
// the upsert replaces only the touched VMId key (:925); GetHostVMSnapshots
// answers ErrOrchestratorHostsVMSnapshotsNotFound when the host has no
// rows; GetListVMSnapshotsByVMId answers an EMPTY non-nil slice (:38).
// Deviations (documented in ledger):
//   - a snapshot list saved empty leaves no rows, so the touched VMId key
//     vanishes instead of persisting as null;
//   - empty snapshot IDs get helpers.GenerateId on write (SQL primary key),
//     legacy keeps "" on the wire (json omitempty hides ids only when
//     empty, so seeded fixtures use explicit ids in parity);
//   - a host-snapshot record with hostId=="" is indistinguishable from a
//     global-store row in the normalised table (pathological: real host
//     ids are uuids; no legacy flow produces it).

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"gorm.io/gorm"
)

func (g *GormDatabase) GetHostVMSnapshots(ctx basecontext.ApiContext, hostId string) (*models.HostsVMSnapshotsRecord, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	var rows []models.VMSnapshot
	if err := g.gormDB.Table("vm_snapshots").Where("host_id = ?", hostId).
		Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	if len(rows) == 0 {
		return nil, interfaces.ErrOrchestratorHostsVMSnapshotsNotFound
	}
	rec := &models.HostsVMSnapshotsRecord{HostId: hostId, VMSnapshots: map[string][]models.VMSnapshot{}}
	for _, r := range rows {
		rec.VMSnapshots[r.VMID] = append(rec.VMSnapshots[r.VMID], r)
	}
	return rec, nil
}

func (g *GormDatabase) SetHostVMSnapshots(ctx basecontext.ApiContext, hostId string, snapshots models.VMSnapshots) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		// Upsert semantics (:907-933): replace the VMId key; other keys of
		// an existing record survive; a fresh record is created otherwise.
		if err := mapGormError(tx.Table("vm_snapshots").
			Where("host_id = ? AND vm_id = ?", hostId, snapshots.VMId).
			Delete(&models.VMSnapshot{}).Error, nil, nil); err != nil {
			return err
		}
		for i := range snapshots.VMSnapshot {
			snapshots.VMSnapshot[i].HostID = hostId
			snapshots.VMSnapshot[i].VMID = snapshots.VMId
			if snapshots.VMSnapshot[i].ID == "" {
				snapshots.VMSnapshot[i].ID = helpers.GenerateId()
			}
			if err := mapGormError(tx.Table("vm_snapshots").Create(&snapshots.VMSnapshot[i]).Error, nil, nil); err != nil {
				return err
			}
		}
		return nil
	})
}

// SetListVMSnapshotsByVMId replaces the GLOBAL store record for vmID
// (legacy signature takes no ctx). host_id=” keeps global rows separate
// from host-record rows in the shared table.
func (g *GormDatabase) SetListVMSnapshotsByVMId(vmID string, newVmSnap models.VMSnapshots) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		if err := mapGormError(tx.Table("vm_snapshots").Where("vm_id = ? AND host_id = ''", vmID).
			Delete(&models.VMSnapshot{}).Error, nil, nil); err != nil {
			return err
		}
		list := newVmSnap.VMSnapshot
		for i := range list {
			list[i].VMID = vmID
			list[i].HostID = ""
			if list[i].ID == "" {
				list[i].ID = helpers.GenerateId()
			}
			if err := mapGormError(tx.Table("vm_snapshots").Create(&list[i]).Error, nil, nil); err != nil {
				return err
			}
		}
		return nil
	})
}

func (g *GormDatabase) GetListVMSnapshotsByVMId(vmID string) ([]models.VMSnapshot, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	var rows []models.VMSnapshot
	if err := g.gormDB.Table("vm_snapshots").Where("vm_id = ? AND host_id = ''", vmID).
		Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	if len(rows) == 0 {
		return make([]models.VMSnapshot, 0), nil // legacy's non-nil empty (:38)
	}
	return rows, nil
}
