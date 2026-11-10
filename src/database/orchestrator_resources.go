package database

// Host-field-scoped writers + resource aggregations (PBI #421 task 13
// split per the >200-line guard: orchestrator.go keeps the D-4 pipelines).

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	apimodels "github.com/Parallels/prl-devops-service/models"
	"gorm.io/gorm"
)

func (g *GormDatabase) UpdateOrchestratorHostDetails(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	if host.ID == "" {
		return nil, interfaces.ErrOrchestratorHostEmptyIdOrHost
	}
	rows, err := g.loadOrchestratorHosts()
	if err != nil {
		return nil, err
	}
	// Dup pre-check fires BEFORE existence (:294-299) and names the
	// EXISTING row's id (:297); the id guard is exact !=, not FoldEqual.
	for i := range rows {
		if strings.EqualFold(rows[i].Host, host.Host) && rows[i].ID != host.ID {
			return nil, errors.NewWithCodef(400, "host %s already exists with ID %s", host.Host, rows[i].ID)
		}
	}
	var found *models.OrchestratorHost
	for i := range rows {
		if strings.EqualFold(rows[i].ID, host.ID) {
			found = &rows[i]
			break
		}
	}
	if found == nil {
		return nil, interfaces.ErrOrchestratorHostNotFound
	}
	// RP relations participate in Diff (stored embedded objects) but are
	// NOT in the Details assignment set — read-attach for diff fidelity,
	// stored rows untouched.
	g.orchAttachRP(found)
	if !host.Diff(*found) {
		return host, nil // no-diff echoes the INPUT (:331)
	}
	set := map[string]interface{}{
		"enabled": host.Enabled, "updated_at": helpers.GetUtcCurrentDateTime(),
		"host": host.Host, "architecture": host.Architecture, "cpu_model": host.CpuModel,
		"port": host.Port, "description": host.Description, "path_prefix": host.PathPrefix,
		"schema": host.Schema, "state": host.State, "last_unhealthy": host.LastUnhealthy,
		"last_unhealthy_error_message": host.LastUnhealthyErrorMessage,
	}
	for col, v := range map[string]interface{}{
		"authentication": host.Authentication, "resources": host.Resources,
		"tags": host.Tags, "required_claims": host.RequiredClaims, "required_roles": host.RequiredRoles,
		"cache_config": host.CacheConfig, "cache_items": host.CacheItems,
	} {
		s, err := orchHostCol(v)
		if err != nil {
			return nil, mapGormError(err, nil, nil)
		}
		set[col] = s
	}
	// Details' set also REPLACES the VM relation rows (:322) — the task-14
	// overlap mandated by the task file: delete-then-insert under one tx.
	vms := host.VirtualMachines
	err = g.gormDB.Transaction(func(tx *gorm.DB) error {
		if err := mapGormError(tx.Table("orchestrator_hosts").Where("id = ?", found.ID).Updates(set).Error,
			interfaces.ErrOrchestratorHostNotFound, nil); err != nil {
			return err
		}
		if err := mapGormError(tx.Table("virtual_machines").Where("host_id = ?", found.ID).
			Delete(&models.VirtualMachine{}).Error, nil, nil); err != nil {
			return err
		}
		for i := range vms {
			vms[i].HostId = found.ID
			if err := mapGormError(tx.Table("virtual_machines").Create(&vms[i]).Error, nil, nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return g.orchHostByID(found.ID)
}

func (g *GormDatabase) UpdateOrchestratorHostTimestamp(ctx basecontext.ApiContext, hostID string, timestamp string, state string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if hostID == "" {
		return interfaces.ErrOrchestratorHostEmptyIdOrHost
	}
	rows, err := g.loadOrchestratorHosts()
	if err != nil {
		return err
	}
	for i := range rows {
		if strings.EqualFold(rows[i].ID, hostID) {
			set := map[string]interface{}{"updated_at": timestamp} // CALLER's clock, no Diff
			if state != "" {
				set["state"] = state
				if state == "healthy" {
					set["last_unhealthy"] = ""
					set["last_unhealthy_error_message"] = ""
				}
			}
			return mapGormError(g.gormDB.Table("orchestrator_hosts").Where("id = ?", rows[i].ID).
				Updates(set).Error, interfaces.ErrOrchestratorHostNotFound, nil)
		}
	}
	return interfaces.ErrOrchestratorHostNotFound
}

func (g *GormDatabase) UpdateOrchestratorHostWebsocketStatus(ctx basecontext.ApiContext, hostID string, hasWebsocketEvents bool) (bool, error) {
	if !g.IsConnected() {
		return false, interfaces.ErrDatabaseNotConnected
	}
	if hostID == "" {
		return false, interfaces.ErrOrchestratorHostEmptyIdOrHost
	}
	rows, err := g.loadOrchestratorHosts()
	if err != nil {
		return false, err
	}
	for i := range rows {
		if strings.EqualFold(rows[i].ID, hostID) {
			if rows[i].HasWebsocketEvents == hasWebsocketEvents {
				return false, nil // same value: NO write, changed=false (:387)
			}
			if err := mapGormError(g.gormDB.Table("orchestrator_hosts").Where("id = ?", rows[i].ID).
				Update("has_websocket_events", hasWebsocketEvents).Error, nil, nil); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, interfaces.ErrOrchestratorHostNotFound
}

// --- resource aggregations -----------------------------------------------
// Legacy scans the raw store under RLock with NO IsConnected guard and NO
// authz: disconnected/empty answers an empty map, never an error.
// pick selects the item to sum; preAdd mirrors per-method oddities
// (reserved adds the HOST-LEVEL TotalAppleVms first, :493).

func (g *GormDatabase) orchResourceAggregate(pick func(*models.HostResources) models.HostResourceItem, preAdd func(*models.HostResourceItem, *models.HostResources)) map[string]models.HostResourceItem {
	result := make(map[string]models.HostResourceItem)
	if !g.IsConnected() {
		return result
	}
	rows, err := g.loadOrchestratorHosts()
	if err != nil {
		return result
	}
	for i := range rows {
		h := &rows[i]
		if h.State != "healthy" || !h.Enabled || h.Resources == nil {
			continue
		}
		item, ok := result[h.Resources.CpuType]
		if !ok {
			item = models.HostResourceItem{}
		}
		if preAdd != nil {
			preAdd(&item, h.Resources)
		}
		sum := pick(h.Resources)
		item.LogicalCpuCount += sum.LogicalCpuCount
		item.PhysicalCpuCount += sum.PhysicalCpuCount
		item.FreeDiskSize += sum.FreeDiskSize
		item.MemorySize += sum.MemorySize
		item.DiskSize += sum.DiskSize
		item.TotalAppleVms += sum.TotalAppleVms
		result[h.Resources.CpuType] = item
	}
	return result
}

func (g *GormDatabase) GetOrchestratorAvailableResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	return g.orchResourceAggregate(func(r *models.HostResources) models.HostResourceItem { return r.TotalAvailable }, nil)
}

func (g *GormDatabase) GetOrchestratorTotalResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	return g.orchResourceAggregate(func(r *models.HostResources) models.HostResourceItem { return r.Total }, nil)
}

func (g *GormDatabase) GetOrchestratorInUseResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	return g.orchResourceAggregate(func(r *models.HostResources) models.HostResourceItem { return r.TotalInUse }, nil)
}

// GetOrchestratorReservedResources double-counts TotalAppleVms by design:
// legacy adds host-level Resources.TotalAppleVms THEN TotalReserved's
// (legacy :493 + :499 — mirrored, not fixed).
func (g *GormDatabase) GetOrchestratorReservedResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	return g.orchResourceAggregate(
		func(r *models.HostResources) models.HostResourceItem { return r.TotalReserved },
		func(item *models.HostResourceItem, r *models.HostResources) { item.TotalAppleVms += r.TotalAppleVms })
}

// System-reserved contributes NO TotalAppleVms (legacy :521-525 skips it);
// the shared accumulator adds sum.TotalAppleVms, so pick zeroes it.
func (g *GormDatabase) GetOrchestratorSystemReservedResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	return g.orchResourceAggregate(func(r *models.HostResources) models.HostResourceItem {
		item := r.SystemReserved
		item.TotalAppleVms = 0
		return item
	}, nil)
}

func (g *GormDatabase) GetOrchestratorHostResources(ctx basecontext.ApiContext, hostId string) (*models.HostResources, error) {
	// Legacy has NO IsConnected guard here (:534): disconnected, the empty
	// store surfaces the not-found sentinel — mirrored, not upgraded.
	if !g.IsConnected() {
		return nil, interfaces.ErrOrchestratorHostNotFound
	}
	host, err := g.orchHostUnsafe(ctx, hostId)
	if err != nil {
		return nil, err
	}
	if host.Resources == nil {
		return nil, interfaces.ErrOrchestratorHostNotFound // existing host, no resources (:542)
	}
	return host.Resources, nil
}

func (g *GormDatabase) UpdateOrchestratorHostResources(ctx basecontext.ApiContext, hostID string, resources *models.HostResources) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if hostID == "" {
		return interfaces.ErrOrchestratorHostEmptyIdOrHost
	}
	s, err := orchHostCol(resources) // nil -> "null" — the serializer's own encoding
	if err != nil {
		return mapGormError(err, nil, nil)
	}
	rows, err := g.loadOrchestratorHosts()
	if err != nil {
		return err
	}
	for i := range rows {
		if strings.EqualFold(rows[i].ID, hostID) {
			return mapGormError(g.gormDB.Table("orchestrator_hosts").Where("id = ?", rows[i].ID).
				Update("resources", s).Error, interfaces.ErrOrchestratorHostNotFound, nil) // UpdatedAt untouched (:817)
		}
	}
	return interfaces.ErrOrchestratorHostNotFound
}

func (g *GormDatabase) UpdateOrchestratorHostCacheConfig(ctx basecontext.ApiContext, hostID string, cacheConfig *apimodels.CatalogCacheConfig) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if hostID == "" {
		return interfaces.ErrOrchestratorHostEmptyIdOrHost
	}
	s, err := orchHostCol(cacheConfig)
	if err != nil {
		return mapGormError(err, nil, nil)
	}
	rows, err := g.loadOrchestratorHosts()
	if err != nil {
		return err
	}
	for i := range rows {
		if strings.EqualFold(rows[i].ID, hostID) {
			return mapGormError(g.gormDB.Table("orchestrator_hosts").Where("id = ?", rows[i].ID).
				Update("cache_config", s).Error, interfaces.ErrOrchestratorHostNotFound, nil)
		}
	}
	return interfaces.ErrOrchestratorHostNotFound
}
