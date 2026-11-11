package database

// Reverse-proxy VIEWS over the orchestrator host (PBI #421 task 14 split
// per the size guard; rationale header lives in orchestrator_vm.go).
// Read-only projections composing the task-12 tables through the host_id
// carrier — never the legacy embedded objects. Views answer nil when the
// host exists but has no config/hosts rows (legacy embedded-nil parity).

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/query"
)

func (g *GormDatabase) GetOrchestratorReverseProxyHosts(ctx basecontext.ApiContext, hostId string, filter string) ([]*models.ReverseProxyHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	host, err := g.orchHostUnsafe(ctx, hostId)
	if err != nil {
		return nil, err
	}
	hosts, err := g.rpHostRowsForHost(host.ID)
	if err != nil {
		return nil, err
	}
	dbFilter, err := query.ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	return query.FilterByProperty(hosts, dbFilter)
}

func (g *GormDatabase) GetOrchestratorReverseProxyHost(ctx basecontext.ApiContext, hostId string, rpIdOrName string) (*models.ReverseProxyHost, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	// Legacy iterates the AUTHZ'd list, matches host by ID||GetHost, then
	// the FIRST host's list by ID||GetHost (:657-682).
	hosts, err := g.orchHostsPipeline(ctx, "")
	if err != nil {
		return nil, err
	}
	for i := range hosts {
		if strings.EqualFold(hosts[i].ID, hostId) || strings.EqualFold(hosts[i].GetHost(), hostId) {
			rpHosts, err := g.rpHostRowsForHost(hosts[i].ID)
			if err != nil {
				return nil, err
			}
			for k := range rpHosts {
				if strings.EqualFold(rpHosts[k].ID, rpIdOrName) || strings.EqualFold(rpHosts[k].GetHost(), rpIdOrName) {
					return rpHosts[k], nil
				}
			}
		}
	}
	return nil, interfaces.ErrOrchestratorReverseProxyHostNotFound
}

func (g *GormDatabase) GetOrchestratorReverseProxyConfig(ctx basecontext.ApiContext, hostId string) (*models.ReverseProxy, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	host, err := g.orchHostUnsafe(ctx, hostId)
	if err != nil {
		return nil, err
	}
	cfg, err := g.rpConfigRowForHost(host.ID)
	if err != nil {
		return nil, err
	}
	return cfg, nil // nil,nil when absent — legacy embedded-nil shape
}
