package database

// SQL implementation of the reverse-proxy domain (PBI #421 task 12):
// config singleton + hosts + embedded serialised-JSON routes. NO authz
// anywhere (legacy has none — GetReverseProxyHosts filters but never wraps
// with GetAuthorizedRecords; confirmed at task start). Route mutations are
// read-modify-write of the http_routes/tcp_route columns under tx (4 sites:
// create/delete/update route + reorder; D-9 reverse_proxy.go:312-394).
// Quirks mirrored:
//   - GetReverseProxyConfig answers (nil, nil) when the singleton does not
//     exist (:34 — no sentinel); Enable/Disable 404 with the inline
//     "reverse proxy config not found".
//   - UpdateReverseProxy replaces ONLY when Diff fires (HostID/Enabled/
//     Host/Port — ID excluded), adopts the stored ID (or mints one), and
//     returns a pointer to the INPUT, not the stored row (:81).
//   - CreateReverseProxyHost mints a fresh id, dup-probes by host, and
//     reports the NEW id in the 400 message (:137 — the existing row's id
//     never appears).
//   - TLS/CORS configuration is a MIRRORED LEGACY BUG: the mutation lands
//     on a copy returned by GetReverseProxyHost and is never written back
//     (legacy :227/:246 persists unchanged bytes via its save queue). The ported path
//     returns the same mutated view but persists nothing, so both engines
//     agree that a later reload shows the old config. "Fixing" this in SQL
//     would diverge parity; flag for upstream.
//   - Route-update matching uses GetRoute() equality, not id (:324); an
//     unchanged TCP-route update falls through to the HOST-not-found
//     sentinel (:424); reorder out-of-range errors BEFORE the normalise
//     side effect, and a not-found route returns RouteNotFound after
//     legacy's in-memory renumber (not reproducible without abnormal
//     order values — normal data makes the renumber a no-op).
//   - UpdateReverseProxyHost is a partial write: only non-empty Name/Host/
//     Port and non-nil Tls/Cors land; no-diff returns the input untouched.

import (
	"encoding/json"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

// rpConfig loads the singleton row (first by rowid); nil,nil when absent.
func (g *GormDatabase) rpConfig() (*models.ReverseProxy, error) {
	var rows []models.ReverseProxy
	if err := g.gormDB.Table("reverse_proxies").Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

func rpConfigMissing() error {
	return errors.NewWithCode("reverse proxy config not found", 404) // inline, as legacy :43
}

func (g *GormDatabase) GetReverseProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	return g.rpConfig() // (nil, nil) when no config exists — legacy :34
}

func (g *GormDatabase) setProxyEnabled(ctx basecontext.ApiContext, enabled bool) (*models.ReverseProxy, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	cfg, err := g.rpConfig()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, rpConfigMissing()
	}
	if err := g.gormDB.Table("reverse_proxies").Where("id = ?", cfg.ID).
		Update("enabled", enabled).Error; err != nil { // field-scoped, NOT Save
		return nil, mapGormError(err, nil, nil)
	}
	cfg.Enabled = enabled
	return cfg, nil
}

func (g *GormDatabase) EnableProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error) {
	return g.setProxyEnabled(ctx, true)
}

func (g *GormDatabase) DisableProxyConfig(ctx basecontext.ApiContext) (*models.ReverseProxy, error) {
	return g.setProxyEnabled(ctx, false)
}

func (g *GormDatabase) UpdateReverseProxy(ctx basecontext.ApiContext, rp models.ReverseProxy) (*models.ReverseProxy, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	cfg, err := g.rpConfig()
	if err != nil {
		return nil, err
	}
	if !cfg.Diff(rp) {
		return &rp, nil // no-diff answers the input, nothing stored (:84)
	}
	rpCopy := rp
	if cfg != nil {
		rpCopy.ID = cfg.ID
	} else {
		rpCopy.ID = helpers.GenerateId()
		if err := mapGormError(g.gormDB.Table("reverse_proxies").Create(&rpCopy).Error, nil, nil); err != nil {
			return nil, err
		}
	}
	if cfg != nil {
		if err := g.gormDB.Table("reverse_proxies").Where("id = ?", cfg.ID).Updates(map[string]interface{}{
			"host_id": rpCopy.HostID, "enabled": rpCopy.Enabled, "host": rpCopy.Host, "port": rpCopy.Port,
		}).Error; err != nil {
			return nil, mapGormError(err, nil, nil)
		}
	}
	return &rp, nil // the INPUT pointer, pre-ID-adoption, as legacy (:81)
}

// rpHosts loads all host rows in insertion order (the raw scan every
// non-filtered host path shares).
func (g *GormDatabase) rpHosts() ([]models.ReverseProxyHost, error) {
	var rows []models.ReverseProxyHost
	if err := g.gormDB.Table("reverse_proxy_hosts").Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return rows, nil
}

func (g *GormDatabase) GetReverseProxyHosts(ctx basecontext.ApiContext, filter string) ([]models.ReverseProxyHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	rows, err := g.rpHosts()
	if err != nil {
		return nil, err
	}
	return query.FilterByProperty(rows, dbFilter) // NO authz wrap (legacy has none)
}

func (g *GormDatabase) GetReverseProxyHost(ctx basecontext.ApiContext, idOrName string) (*models.ReverseProxyHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	hosts, err := g.GetReverseProxyHosts(ctx, "")
	if err != nil {
		return nil, err
	}
	for i := range hosts {
		if strings.EqualFold(hosts[i].ID, idOrName) || strings.EqualFold(hosts[i].GetHost(), idOrName) {
			return &hosts[i], nil
		}
	}
	return nil, interfaces.ErrorReverseProxyHostNotFound
}

func (g *GormDatabase) CreateReverseProxyHost(ctx basecontext.ApiContext, rpHost models.ReverseProxyHost) (*models.ReverseProxyHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	if rpHost.Host == "" {
		return nil, interfaces.ErrReverseProxyHostEmptyNameOrId
	}
	rpHost.ID = helpers.GenerateId()
	if u, _ := g.GetReverseProxyHost(ctx, rpHost.GetHost()); u != nil {
		// The message names the FRESH id, not the existing row's (:137).
		return nil, errors.NewWithCodef(400, "reverse proxy host %s already exists with ID %s", rpHost.GetHost(), rpHost.ID)
	}
	if rpHost.TcpRoute != nil {
		rpHost.TcpRoute.ID = helpers.GenerateId()
	}
	if rpHost.HttpRoutes != nil {
		for _, route := range rpHost.HttpRoutes {
			route.ID = helpers.GenerateId()
		}
	}
	if err := mapGormError(g.gormDB.Table("reverse_proxy_hosts").Create(&rpHost).Error,
		interfaces.ErrorReverseProxyHostNotFound, nil); err != nil {
		return nil, err
	}
	return &rpHost, nil
}

func (g *GormDatabase) DeleteReverseProxyHost(ctx basecontext.ApiContext, idOrName string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	if idOrName == "" {
		return interfaces.ErrReverseProxyHostEmptyNameOrId
	}
	hosts, err := g.rpHosts()
	if err != nil {
		return err
	}
	for i := range hosts {
		if strings.EqualFold(hosts[i].ID, idOrName) || strings.EqualFold(hosts[i].GetHost(), idOrName) {
			return mapGormError(g.gormDB.Table("reverse_proxy_hosts").Where("id = ?", hosts[i].ID).
				Delete(&models.ReverseProxyHost{}).Error, interfaces.ErrorReverseProxyHostNotFound, nil)
		}
	}
	return interfaces.ErrorReverseProxyHostNotFound
}

func (g *GormDatabase) UpdateReverseProxyHost(ctx basecontext.ApiContext, rpHost *models.ReverseProxyHost) (*models.ReverseProxyHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	if rpHost.ID == "" {
		return nil, interfaces.ErrReverseProxyHostEmptyNameOrId
	}
	hosts, err := g.rpHosts()
	if err != nil {
		return nil, err
	}
	var found *models.ReverseProxyHost
	for i := range hosts {
		if strings.EqualFold(hosts[i].ID, rpHost.ID) ||
			strings.EqualFold(hosts[i].GetHost(), rpHost.ID) ||
			strings.EqualFold(hosts[i].GetHost(), rpHost.GetHost()) {
			found = &hosts[i]
			break
		}
	}
	if found == nil {
		return nil, interfaces.ErrorReverseProxyHostNotFound
	}
	if !found.Diff(*rpHost) {
		return rpHost, nil // no-diff answers the input (:209)
	}
	set := map[string]interface{}{}
	if rpHost.Name != "" {
		set["name"] = rpHost.Name
	}
	if rpHost.Host != "" {
		set["host"] = rpHost.Host
	}
	if rpHost.Port != "" {
		set["port"] = rpHost.Port
	}
	// Typed-nil trap avoided: a nil *Tls stored in an interface{} is NOT a
	// nil interface, so the pointer must be checked before marshalling
	// (the naive map loop wrote literal "null" into the column — caught by
	// the parity suite, which compares reloaded content).
	if rpHost.Tls != nil {
		b, err := json.Marshal(rpHost.Tls)
		if err != nil {
			return nil, mapGormError(err, nil, nil)
		}
		set["tls"] = string(b)
	}
	if rpHost.Cors != nil {
		b, err := json.Marshal(rpHost.Cors)
		if err != nil {
			return nil, mapGormError(err, nil, nil)
		}
		set["cors"] = string(b)
	}
	if err := g.gormDB.Table("reverse_proxy_hosts").Where("id = ?", found.ID).Updates(set).Error; err != nil {
		return nil, mapGormError(err, interfaces.ErrorReverseProxyHostNotFound, nil)
	}
	updated, err := g.rpHostByID(found.ID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (g *GormDatabase) rpHostByID(id string) (*models.ReverseProxyHost, error) {
	hosts, err := g.rpHosts()
	if err != nil {
		return nil, err
	}
	for i := range hosts {
		if hosts[i].ID == id {
			return &hosts[i], nil
		}
	}
	return nil, interfaces.ErrorReverseProxyHostNotFound
}

// ConfigureTLS/CORS mirror the legacy lost-write (:227/:246 mutate a COPY
// of the list result; the save path then persists the UNCHANGED row). The view
// returned here carries the new config; the stored row does not. Pinned by
// parity + unit reload checks — do not "fix" without a legacy change.
func (g *GormDatabase) ConfigureReverseProxyHostTls(ctx basecontext.ApiContext, rpHostId string, tlsConfig models.ReverseProxyHostTls) (*models.ReverseProxyHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rpHost, err := g.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}
	if rpHost.Tls.Diff(tlsConfig) {
		rpHost.Tls = &tlsConfig
	}
	return rpHost, nil
}

func (g *GormDatabase) ConfigureReverseProxyHostCors(ctx basecontext.ApiContext, rpHostId string, corsConfig models.ReverseProxyHostCors) (*models.ReverseProxyHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rpHost, err := g.GetReverseProxyHost(ctx, rpHostId)
	if err != nil {
		return nil, err
	}
	if rpHost.Cors.Diff(corsConfig) {
		rpHost.Cors = &corsConfig
	}
	return rpHost, nil
}
