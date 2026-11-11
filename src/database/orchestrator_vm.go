package database

// Orchestrator VM rows + reverse-proxy VIEWS (PBI #421 task 14).
// VirtualMachines are child-table rows (#419 D-3, table virtual_machines,
// carrier host_id, global id primary key). Read/write matrix mirrored
// line-by-line from legacy (:548-815):
//   - GetOrchestratorVirtualMachines: AUTHZ host list, only healthy+
//     enabled hosts contribute, order = host UpdatedAt-desc then rowid;
//     property filter applies AFTER concatenation (:548-580).
//   - host-scoped reads go through the authz single-host lookup but have
//     NO healthy/enabled gate (:582).
//   - twins (Upsert/Remove/State/Uptime, :704-810): RAW scan (NO authz),
//     host match Fold ID only, VM match EXACT == (not Fold!); Remove and
//     State/Uptime are idempotent nil when the VM is absent; only a
//     missing HOST errors. Upsert replaces the row in place (exact id),
//     else inserts.
//   - DeleteOrchestratorVirtualMachine :170: RAW scan, host Fold ID||
//     RAW Host, VM Fold ID||NAME — FIRST match splices (SQL mirrors via
//     rowid-ordered probe then delete-by-pk).
//   - ReplaceOrchestratorHostVMs :871: whole-slice swap = delete-where-
//     host + insert-all in one tx (empty payload = delete-all).
//     Spec's "legacy stamps host ownership on swap" is FALSE — legacy
//     swaps verbatim; the SQL side still stamps the host_id carrier
//     (#419 D-3), pinned in units.
// RP views (:630-704) read the host's ReverseProxy/ReverseProxyHosts
// which in legacy are embedded objects written ONLY by
// UpdateOrchestratorHost (:253-254 — the fidelity gap fixed in
// orchestrator.go by this slice). SQL composes the same from the
// reverse_proxies / reverse_proxy_hosts tables via the host_id carrier.

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
	"gorm.io/gorm"
)

// orchHostRawByID scans the RAW store (authz-free, mirroring the twin
// methods' raw loops) and returns the stored row's CANONICAL id so later
// statements bind on the real primary key regardless of caller casing.
func (g *GormDatabase) orchHostRawByID(hostID string) (string, error) {
	rows, err := g.loadOrchestratorHosts()
	if err != nil {
		return "", err
	}
	for i := range rows {
		if strings.EqualFold(rows[i].ID, hostID) {
			return rows[i].ID, nil
		}
	}
	return "", interfaces.ErrOrchestratorHostNotFound
}

func (g *GormDatabase) vmRowsForHost(hostID string) ([]models.VirtualMachine, error) {
	var rows []models.VirtualMachine
	if err := g.gormDB.Table("virtual_machines").Where("host_id = ?", hostID).
		Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return rows, nil
}

func (g *GormDatabase) GetOrchestratorVirtualMachines(ctx basecontext.ApiContext, filter string) ([]models.VirtualMachine, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	// Legacy parses the filter AFTER building the list (:571 — after the
	// host concat); an unparseable filter still surfaces, just later. Same
	// observable outcome here.
	hosts, err := g.orchHostsPipeline(ctx, "")
	if err != nil {
		return nil, err
	}
	var all []models.VirtualMachine
	for _, h := range hosts {
		if h.State != "healthy" || !h.Enabled {
			continue
		}
		rows, err := g.vmRowsForHost(h.ID)
		if err != nil {
			return nil, err
		}
		all = append(all, rows...)
	}
	return query.FilterByProperty(all, dbFilter)
}

func (g *GormDatabase) GetOrchestratorHostVirtualMachines(ctx basecontext.ApiContext, hostId string, filter string) ([]models.VirtualMachine, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	host, err := g.orchHostUnsafe(ctx, hostId) // authz'd, NO health/enabled gate
	if err != nil {
		return nil, err
	}
	rows, err := g.vmRowsForHost(host.ID)
	if err != nil {
		return nil, err
	}
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	return query.FilterByProperty(rows, dbFilter)
}

func (g *GormDatabase) GetOrchestratorHostVirtualMachine(ctx basecontext.ApiContext, hostId string, machineId string) (*models.VirtualMachine, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	host, err := g.orchHostUnsafe(ctx, hostId)
	if err != nil {
		return nil, err
	}
	rows, err := g.vmRowsForHost(host.ID)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if strings.EqualFold(rows[i].ID, machineId) {
			return &rows[i], nil
		}
	}
	return nil, interfaces.ErrOrchestratorHostVirtualMachineNotFound
}

func (g *GormDatabase) DeleteOrchestratorVirtualMachine(ctx basecontext.ApiContext, idOrHost string, vmIdOrName string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if idOrHost == "" {
		return interfaces.ErrOrchestratorHostEmptyIdOrHost
	}
	rows, err := g.loadOrchestratorHosts() // raw scan, NO authz (:182)
	if err != nil {
		return err
	}
	for i := range rows {
		if strings.EqualFold(rows[i].ID, idOrHost) || strings.EqualFold(rows[i].Host, idOrHost) {
			vms, err := g.vmRowsForHost(rows[i].ID)
			if err != nil {
				return err
			}
			for k := range vms {
				// Fold over ID || NAME; FIRST match wins (legacy splice).
				if strings.EqualFold(vms[k].ID, vmIdOrName) || strings.EqualFold(vms[k].Name, vmIdOrName) {
					return mapGormError(g.gormDB.Table("virtual_machines").Where("id = ?", vms[k].ID).
						Delete(&models.VirtualMachine{}).Error, nil, nil)
				}
			}
			return interfaces.ErrOrchestratorHostVirtualMachineNotFound
		}
	}
	return interfaces.ErrOrchestratorHostNotFound
}

func (g *GormDatabase) UpsertOrchestratorHostVM(ctx basecontext.ApiContext, hostID string, vm models.VirtualMachine) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	// hostID=="" scans like legacy: never matches -> HostNotFound (no
	// empty-id sentinel on this method).
	canonical, err := g.orchHostRawByID(hostID)
	if err != nil {
		return err
	}
	vms, err := g.vmRowsForHost(canonical)
	if err != nil {
		return err
	}
	vm.HostId = canonical // carrier (#419 D-3)
	for k := range vms {
		if vms[k].ID == vm.ID { // EXACT == in legacy (:715)
			// Whole-row replace in place (rowid/order preserved, nils
			// included): Select("*") over the model schema so the JSON
			// serializers still own the encoded columns.
			res := g.gormDB.Model(&models.VirtualMachine{}).Where("id = ?", vm.ID).
				Select("*").Omit("id").Updates(&vm)
			return mapGormError(res.Error, nil, nil)
		}
	}
	return mapGormError(g.gormDB.Table("virtual_machines").Create(&vm).Error, nil, nil)
}

func (g *GormDatabase) RemoveOrchestratorHostVM(ctx basecontext.ApiContext, hostID string, vmID string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	canonical, err := g.orchHostRawByID(hostID)
	if err != nil {
		return err
	}
	vms, err := g.vmRowsForHost(canonical)
	if err != nil {
		return err
	}
	for k := range vms {
		if vms[k].ID == vmID { // EXACT; absent -> nil (idempotent)
			return mapGormError(g.gormDB.Table("virtual_machines").Where("id = ?", vmID).
				Delete(&models.VirtualMachine{}).Error, nil, nil)
		}
	}
	return nil
}

func (g *GormDatabase) UpdateOrchestratorHostVMState(ctx basecontext.ApiContext, hostID string, vmID string, state string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	canonical, err := g.orchHostRawByID(hostID)
	if err != nil {
		return err
	}
	_ = canonical
	res := g.gormDB.Table("virtual_machines").Where("host_id = ? AND id = ?", canonical, vmID).
		Update("state", state) // field-scoped; absent VM -> nil
	if res.Error != nil {
		return mapGormError(res.Error, nil, nil)
	}
	return nil
}

func (g *GormDatabase) UpdateOrchestratorHostVMUptime(ctx basecontext.ApiContext, hostID string, vmID string, uptime string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	canonical, err := g.orchHostRawByID(hostID)
	if err != nil {
		return err
	}
	res := g.gormDB.Table("virtual_machines").Where("host_id = ? AND id = ?", canonical, vmID).
		Update("uptime", uptime)
	if res.Error != nil {
		return mapGormError(res.Error, nil, nil)
	}
	return nil
}

func (g *GormDatabase) ReplaceOrchestratorHostVMs(ctx basecontext.ApiContext, hostID string, vms []models.VirtualMachine) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	canonical, err := g.orchHostRawByID(hostID)
	if err != nil {
		return err
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error {
		if err := mapGormError(tx.Table("virtual_machines").Where("host_id = ?", canonical).
			Delete(&models.VirtualMachine{}).Error, nil, nil); err != nil {
			return err
		}
		for i := range vms {
			vms[i].HostId = canonical
			if err := mapGormError(tx.Table("virtual_machines").Create(&vms[i]).Error, nil, nil); err != nil {
				return err
			}
		}
		return nil
	})
}
