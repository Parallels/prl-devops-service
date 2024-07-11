package orchestrator

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *OrchestratorService) UpdateHost(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error) {
	dbService, err := serviceprovider.GetDatabaseService(ctx)
	if err != nil {
		return nil, err
	}

	dbHost, err := dbService.GetOrchestratorHost(ctx, host.ID)
	if err != nil {
		return nil, err
	}

	if dbHost.State == "healthy" {
		if hw, err := s.GetHostHardwareInfo(host); err == nil {
			if host.Architecture != hw.CpuType {
				if host.Resources == nil {
					host.Resources = &models.HostResources{}
				}

				dtoResources := mappers.MapHostResourcesFromSystemUsageResponse(*hw)
				host.Resources = &dtoResources
				host.Architecture = hw.CpuType
				host.CpuModel = hw.CpuBrand
			}
		}
	}

	host.Enabled = dbHost.Enabled
	host.CreatedAt = dbHost.CreatedAt
	host.UpdatedAt = dbHost.UpdatedAt
	if host.Authentication == nil {
		host.Authentication = dbHost.Authentication
	}
	if host.Resources == nil {
		host.Resources = dbHost.Resources
	}
	if host.Host == "" {
		host.Host = dbHost.Host
	}
	if host.Port == "" {
		host.Port = dbHost.Port
	}
	if host.PathPrefix == "" {
		host.PathPrefix = dbHost.PathPrefix
	}
	if len(dbHost.RequiredClaims) > 0 {
		for _, dbClaim := range dbHost.RequiredClaims {
			if dbClaim != "" {
				found := false
			hostClaim:
				for _, claim := range host.RequiredClaims {
					if strings.EqualFold(claim, dbClaim) {
						found = true
						break hostClaim
					}
				}
				if !found {
					host.RequiredClaims = append(host.RequiredClaims, dbClaim)
				}
			}
		}
	}
	if len(dbHost.RequiredRoles) > 0 {
		for _, dbRole := range dbHost.RequiredRoles {
			if dbRole != "" {
				found := false
			hostRole:
				for _, role := range host.RequiredRoles {
					if strings.EqualFold(role, dbRole) {
						found = true
						break hostRole
					}
				}
				if !found {
					host.RequiredRoles = append(host.RequiredRoles, dbRole)
				}
			}
		}
	}
	if len(dbHost.Tags) > 0 {
		for _, dbTag := range dbHost.Tags {
			if dbTag != "" {
				found := false
			hostTag:
				for _, tag := range host.Tags {
					if strings.EqualFold(tag, dbTag) {
						found = true
						break hostTag
					}
				}
				if !found {
					host.Tags = append(host.Tags, dbTag)
				}
			}
		}
	}
	if host.Description == "" {
		host.Description = dbHost.Description
	}
	if host.Schema == "" {
		host.Schema = dbHost.Schema
	}
	if len(dbHost.VirtualMachines) > 0 {
		host.VirtualMachines = dbHost.VirtualMachines
	}

	updatedHost, err := dbService.UpdateOrchestratorHostDetails(ctx, host)
	if err != nil {
		return nil, err
	}

	return updatedHost, nil
}
