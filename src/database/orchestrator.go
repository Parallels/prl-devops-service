package database

// SQL implementation of the orchestrator HOSTS half (PBI #421 task 13);
// VM/snapshot/reverse-proxy-view methods stay delegated until task 14.
// The list pipeline is the D-4 trio (filter -> authz wrap -> Go-side
// UpdatedAt-desc order, load order rowid, no SQL ORDER BY) — same shape as
// catalogs. Legacy VirtualMachines/ReverseProxy fields are RELATIONS
// (#419), never preloaded (#419 D-7): list/single reads answer hosts with
// empty relation fields; VM payload on create/details-update lands in the
// virtual_machines table (the ONE task-13/task-14 overlap the task file
// mandates). Field-scoped discipline everywhere: host columns update via
// Updates maps so concurrent VM writes (task 14 paths) can never be
// clobbered by a whole-row Save.
// Quirks mirrored (spec-vs-code corrections in ledger entry):
//   - Create mints id, stamps CreatedAt/UpdatedAt with TWO separate clock
//     calls, force-enables, dup-probes the AUTHZ-VISIBLE list (an
//     invisible host is not a duplicate) and the 400 message names the
//     FRESH id (:139-141).
//   - Delete matches ID || RAW Host (not GetHost()), scans the RAW store
//     (NO authz :160) — unlike Get/Delete the write paths are authz-free.
//   - Update/Diff semantics: no-diff echoes the INPUT pointer; the
//     assignment sets of Update (:219-256) and Details (:304-324) are
//     DIFFERENT (Details writes VirtualMachines/Cache*, skips Os*/
//     ExternalIp/DevOps/PD*/RP fields; Update is the inverse). Neither
//     touches ID/CreatedAt/HasWebsocketEvents/IsLocal; HealthCheck is
//     gorm:"-:all".
//   - Details dup-precheck fires BEFORE existence and names the EXISTING
//     row's id (:297); its id-inequality guard is exact !=, not FoldEqual.
//   - Timestamp: caller-supplied UpdatedAt (no clock), State only when
//     non-empty, "healthy" clears both LastUnhealthy fields; no Diff.
//   - WebsocketStatus returns a CHANGED bool: same-value writes report
//     false WITHOUT writing.
//   - Aggregations have NO IsConnected guard and NO authz in legacy
//     (:399+): mirror answers an empty map when disconnected; the
//     reserved-family double-counts TotalAppleVms (host-level +
//     TotalReserved, :493+:499) and system-reserved omits TotalAppleVms
//     entirely (:521-525). GetOrchestratorHostResources ALSO has no guard
//     (:534) — disconnected answers the not-found sentinel.

import (
	"encoding/json"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"gorm.io/gorm"
)

// loadOrchestratorHosts: raw rows in insertion order. Relations (VMs,
// ReverseProxy*) are never loaded (D-7 #419).
func (g *GormDatabase) loadOrchestratorHosts() ([]models.OrchestratorHost, error) {
	var rows []models.OrchestratorHost
	if err := g.gormDB.Table("orchestrator_hosts").Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return rows, nil
}

// orchHostsPipeline mirrors getOrchestratorHostsLocked: filter -> authz ->
// UpdatedAt desc. The single-host lookups route through it, so authz and
// ordering apply there too.
func (g *GormDatabase) orchHostsPipeline(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error) {
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	rows, err := g.loadOrchestratorHosts()
	if err != nil {
		return nil, err
	}
	filtered, err := query.FilterByProperty(rows, dbFilter)
	if err != nil {
		return nil, err
	}
	authed := query.GetAuthorizedRecords(ctx, filtered...)
	return query.OrderByProperty(authed, &query.Order{Property: "UpdatedAt", Direction: query.OrderDirectionDesc})
}

func (g *GormDatabase) orchHostUnsafe(ctx basecontext.ApiContext, idOrHost string) (*models.OrchestratorHost, error) {
	hosts, err := g.orchHostsPipeline(ctx, "")
	if err != nil {
		return nil, err
	}
	for i := range hosts {
		if strings.EqualFold(hosts[i].ID, idOrHost) || strings.EqualFold(hosts[i].GetHost(), idOrHost) {
			g.orchAttachRP(&hosts[i])
			return &hosts[i], nil
		}
	}
	return nil, interfaces.ErrOrchestratorHostNotFound
}

// --- reverse-proxy relation rows ------------------------------------------
// Legacy keeps ReverseProxy/ReverseProxyHosts EMBEDDED on the host object;
// #419 normalised them into reverse_proxies/reverse_proxy_hosts with the
// host_id carrier. Reads that return a single host attach them (legacy's
// shared object exposes them on the wire); LISTS do not (D-7 lineage) —
// recorded deviation. Empty result keeps nil (legacy embedded-nil shape).

func (g *GormDatabase) rpConfigRowForHost(hostID string) (*models.ReverseProxy, error) {
	var rows []models.ReverseProxy
	if err := g.gormDB.Table("reverse_proxies").Where("host_id = ?", hostID).
		Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func (g *GormDatabase) rpHostRowsForHost(hostID string) ([]*models.ReverseProxyHost, error) {
	var rows []models.ReverseProxyHost
	if err := g.gormDB.Table("reverse_proxy_hosts").Where("host_id = ?", hostID).
		Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	out := make([]*models.ReverseProxyHost, 0, len(rows))
	for i := range rows {
		out = append(out, &rows[i])
	}
	return out, nil
}

func (g *GormDatabase) orchAttachRP(row *models.OrchestratorHost) {
	cfg, err := g.rpConfigRowForHost(row.ID)
	if err == nil {
		row.ReverseProxy = cfg
	}
	hosts, err := g.rpHostRowsForHost(row.ID)
	if err == nil {
		row.ReverseProxyHosts = hosts
	}
}

func (g *GormDatabase) GetOrchestratorHosts(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	return g.orchHostsPipeline(ctx, filter)
}

func (g *GormDatabase) GetActiveOrchestratorHosts(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	rows, err := g.loadOrchestratorHosts()
	if err != nil {
		return nil, err
	}
	// Enabled pre-filter happens BEFORE the property filter (:68-79).
	active := make([]models.OrchestratorHost, 0, len(rows))
	for _, h := range rows {
		if h.Enabled {
			active = append(active, h)
		}
	}
	filtered, err := query.FilterByProperty(active, dbFilter)
	if err != nil {
		return nil, err
	}
	authed := query.GetAuthorizedRecords(ctx, filtered...)
	return query.OrderByProperty(authed, &query.Order{Property: "UpdatedAt", Direction: query.OrderDirectionDesc})
}

func (g *GormDatabase) GetOrchestratorHost(ctx basecontext.ApiContext, idOrHost string) (*models.OrchestratorHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	return g.orchHostUnsafe(ctx, idOrHost)
}

func (g *GormDatabase) CreateOrchestratorHost(ctx basecontext.ApiContext, host models.OrchestratorHost) (*models.OrchestratorHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	if host.Host == "" {
		return nil, interfaces.ErrOrchestratorHostEmptyName
	}
	host.ID = helpers.GenerateId()
	host.CreatedAt = helpers.GetUtcCurrentDateTime() // two separate calls (:135-136)
	host.UpdatedAt = helpers.GetUtcCurrentDateTime()
	host.Enabled = true
	if u, _ := g.orchHostUnsafe(ctx, host.GetHost()); u != nil {
		return nil, errors.NewWithCodef(400, "host %s already exists with ID %s", host.GetHost(), host.ID) // FRESH id (:140)
	}
	vms := host.VirtualMachines
	host.VirtualMachines = nil // relation: lands in its own rows, not the host record
	err := g.gormDB.Transaction(func(tx *gorm.DB) error {
		if err := mapGormError(tx.Table("orchestrator_hosts").Create(&host).Error,
			interfaces.ErrOrchestratorHostNotFound, nil); err != nil {
			return err
		}
		for i := range vms {
			vms[i].HostId = host.ID // normalisation carrier (#419 D-3)
			if err := mapGormError(tx.Table("virtual_machines").Create(&vms[i]).Error, nil, nil); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	host.VirtualMachines = vms // legacy returned the input with VM payload intact
	return &host, nil
}

func (g *GormDatabase) DeleteOrchestratorHost(ctx basecontext.ApiContext, idOrHost string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if idOrHost == "" {
		return interfaces.ErrOrchestratorHostEmptyIdOrHost
	}
	rows, err := g.loadOrchestratorHosts() // RAW scan — delete is authz-free (:160)
	if err != nil {
		return err
	}
	for i := range rows {
		// ID || RAW Host — NOT GetHost() (:161).
		if strings.EqualFold(rows[i].ID, idOrHost) || strings.EqualFold(rows[i].Host, idOrHost) {
			return g.gormDB.Transaction(func(tx *gorm.DB) error {
				if err := mapGormError(tx.Table("virtual_machines").Where("host_id = ?", rows[i].ID).
					Delete(&models.VirtualMachine{}).Error, nil, nil); err != nil {
					return err
				}
				return mapGormError(tx.Table("orchestrator_hosts").Where("id = ?", rows[i].ID).
					Delete(&models.OrchestratorHost{}).Error, interfaces.ErrOrchestratorHostNotFound, nil)
			})
		}
	}
	return interfaces.ErrOrchestratorHostNotFound
}

func orchHostCol(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	return string(b), err
}

func (g *GormDatabase) UpdateOrchestratorHost(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error) {
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
	// Diff must SEE the stored RP relations (legacy diffs against the
	// embedded objects, :253-254 writes them).
	g.orchAttachRP(found)
	if !host.Diff(*found) {
		return host, nil // no-diff echoes the INPUT (:263)
	}
	set := map[string]interface{}{
		"enabled": host.Enabled, "updated_at": helpers.GetUtcCurrentDateTime(),
		"host": host.Host, "os_version": host.OsVersion, "os_name": host.OsName,
		"external_ip_address": host.ExternalIpAddress, "dev_ops_version": host.DevOpsVersion,
		"architecture": host.Architecture, "cpu_model": host.CpuModel, "port": host.Port,
		"description": host.Description, "path_prefix": host.PathPrefix, "schema": host.Schema,
		"state": host.State, "last_unhealthy": host.LastUnhealthy,
		"last_unhealthy_error_message": host.LastUnhealthyErrorMessage,
		"parallels_desktop_version":    host.ParallelsDesktopVersion,
		"parallels_desktop_licensed":   host.ParallelsDesktopLicensed,
		"is_reverse_proxy_enabled":     host.IsReverseProxyEnabled,
		"is_log_streaming_enabled":     host.IsLogStreamingEnabled,
	}
	for col, v := range map[string]interface{}{
		"authentication": host.Authentication, "resources": host.Resources,
		"tags": host.Tags, "required_claims": host.RequiredClaims, "required_roles": host.RequiredRoles,
		"enabled_modules": host.EnabledModules, "cache_config": host.CacheConfig, "cache_items": host.CacheItems,
	} {
		s, err := orchHostCol(v)
		if err != nil {
			return nil, mapGormError(err, nil, nil)
		}
		set[col] = s
	}
	// VirtualMachines, HasWebsocketEvents, CreatedAt, ID, IsLocal are
	// EXCLUDED from the column set (:241-252 comments are legacy's own
	// warning against clobbering them). The RP relations ARE written
	// (:253-254) — as replaced rows below, nil input clearing them.
	err = g.gormDB.Transaction(func(tx *gorm.DB) error {
		if err := mapGormError(tx.Table("orchestrator_hosts").Where("id = ?", found.ID).
			Updates(set).Error, interfaces.ErrOrchestratorHostNotFound, nil); err != nil {
			return err
		}
		if err := mapGormError(tx.Table("reverse_proxies").Where("host_id = ?", found.ID).
			Delete(&models.ReverseProxy{}).Error, nil, nil); err != nil {
			return err
		}
		if host.ReverseProxy != nil {
			cfg := *host.ReverseProxy
			cfg.HostID = found.ID
			if err := mapGormError(tx.Table("reverse_proxies").Create(&cfg).Error, nil, nil); err != nil {
				return err
			}
		}
		if err := mapGormError(tx.Table("reverse_proxy_hosts").Where("host_id = ?", found.ID).
			Delete(&models.ReverseProxyHost{}).Error, nil, nil); err != nil {
			return err
		}
		for _, h := range host.ReverseProxyHosts {
			c := *h
			c.HostID = found.ID
			if err := mapGormError(tx.Table("reverse_proxy_hosts").Create(&c).Error, nil, nil); err != nil {
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

func (g *GormDatabase) orchHostByID(id string) (*models.OrchestratorHost, error) {
	rows, err := g.loadOrchestratorHosts()
	if err != nil {
		return nil, err
	}
	for i := range rows {
		if rows[i].ID == id {
			g.orchAttachRP(&rows[i])
			return &rows[i], nil
		}
	}
	return nil, interfaces.ErrOrchestratorHostNotFound
}
