package data

import (
	"fmt"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/data/models"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
)

var (
	ErrOrchestratorHostEmptyIdOrHost          = errors.NewWithCode("no host specified", 500)
	ErrOrchestratorHostEmptyName              = errors.NewWithCode("host name cannot be empty", 500)
	ErrOrchestratorHostNotFound               = errors.NewWithCode("host not found", 404)
	ErrOrchestratorHostVirtualMachineNotFound = errors.NewWithCode("host virtual machine not found", 404)
)

func (j *JsonDatabase) GetOrchestratorHosts(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(j.data.OrchestratorHosts, dbFilter)
	if err != nil {
		return nil, err
	}

	result := GetAuthorizedRecords(ctx, filteredData...)

	orderedResult, err := OrderByProperty(result, &Order{Property: "UpdatedAt", Direction: OrderDirectionDesc})
	if err != nil {
		return nil, err
	}

	return orderedResult, nil
}

func (j *JsonDatabase) GetOrchestratorHost(ctx basecontext.ApiContext, idOrHost string) (*models.OrchestratorHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	hosts, err := j.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		hostname := fmt.Sprintf("%s%s", idOrHost, host.PathPrefix)
		t := host.GetHost()
		println(t)
		if strings.EqualFold(host.ID, idOrHost) || strings.EqualFold(host.Host, idOrHost) || strings.EqualFold(host.GetHost(), hostname) {
			return &host, nil
		}
	}

	return nil, ErrOrchestratorHostNotFound
}

func (j *JsonDatabase) CreateOrchestratorHost(ctx basecontext.ApiContext, host models.OrchestratorHost) (*models.OrchestratorHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if host.Host == "" {
		return nil, ErrOrchestratorHostEmptyName
	}

	host.ID = helpers.GenerateId()
	host.CreatedAt = helpers.GetUtcCurrentDateTime()
	host.UpdatedAt = helpers.GetUtcCurrentDateTime()

	if u, _ := j.GetOrchestratorHost(ctx, host.Host); u != nil {
		return nil, errors.NewWithCodef(400, "host %s already exists with ID %s", host.Host, host.ID)
	}

	j.data.OrchestratorHosts = append(j.data.OrchestratorHosts, host)
	j.Save(ctx)

	return &host, nil
}

func (j *JsonDatabase) DeleteOrchestratorHost(ctx basecontext.ApiContext, idOrHost string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if idOrHost == "" {
		return ErrOrchestratorHostEmptyIdOrHost
	}

	for i, host := range j.data.OrchestratorHosts {
		if strings.EqualFold(host.ID, idOrHost) || strings.EqualFold(host.Host, idOrHost) {
			j.data.OrchestratorHosts = append(j.data.OrchestratorHosts[:i], j.data.OrchestratorHosts[i+1:]...)

			j.Save(ctx)
			return nil
		}
	}

	return ErrOrchestratorHostNotFound
}

func (j *JsonDatabase) UpdateOrchestratorHost(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if host.ID == "" {
		return nil, ErrOrchestratorHostEmptyIdOrHost
	}

	for _, dbHost := range j.data.OrchestratorHosts {
		if strings.EqualFold(dbHost.ID, host.ID) || strings.EqualFold(dbHost.Host, host.Host) {
			index, err := GetRecordIndex(j.data.OrchestratorHosts, "id", host.ID)
			if err != nil {
				return nil, err
			}
			if host.Diff(j.data.OrchestratorHosts[index]) {

				j.data.OrchestratorHosts[index].UpdatedAt = helpers.GetUtcCurrentDateTime()
				j.data.OrchestratorHosts[index].Host = host.Host
				j.data.OrchestratorHosts[index].Port = host.Port
				j.data.OrchestratorHosts[index].Authentication = host.Authentication
				j.data.OrchestratorHosts[index].Resources = host.Resources
				j.data.OrchestratorHosts[index].RequiredClaims = host.RequiredClaims
				j.data.OrchestratorHosts[index].RequiredRoles = host.RequiredRoles
				j.data.OrchestratorHosts[index].Description = host.Description
				j.data.OrchestratorHosts[index].Tags = host.Tags
				j.data.OrchestratorHosts[index].PathPrefix = host.PathPrefix
				j.data.OrchestratorHosts[index].Schema = host.Schema
				j.data.OrchestratorHosts[index].State = host.State
				j.data.OrchestratorHosts[index].LastUnhealthy = host.LastUnhealthy
				j.data.OrchestratorHosts[index].LastUnhealthyErrorMessage = host.LastUnhealthyErrorMessage
				j.data.OrchestratorHosts[index].HealthCheck = host.HealthCheck
				j.data.OrchestratorHosts[index].VirtualMachines = host.VirtualMachines

				j.Save(ctx)

				return &j.data.OrchestratorHosts[index], nil
			} else {
				ctx.LogDebug("[Database] No changes detected for host %s", host.Host)
				return host, nil
			}
		}
	}

	return nil, ErrOrchestratorHostNotFound
}

func (j *JsonDatabase) GetOrchestratorAvailableResources(ctx basecontext.ApiContext) models.HostResourceItem {
	result := models.HostResourceItem{}

	for _, host := range j.data.OrchestratorHosts {
		if host.State == "healthy" {
			if host.Resources != nil {
				result.LogicalCpuCount += host.Resources.TotalAvailable.LogicalCpuCount
				result.PhysicalCpuCount += host.Resources.TotalAvailable.PhysicalCpuCount
				result.FreeDiskSize += host.Resources.TotalAvailable.FreeDiskSize
				result.MemorySize += host.Resources.TotalAvailable.MemorySize
			}
		}
	}

	return result
}

func (j *JsonDatabase) GetOrchestratorTotalResources(ctx basecontext.ApiContext) models.HostResourceItem {
	result := models.HostResourceItem{}

	for _, host := range j.data.OrchestratorHosts {
		if host.State == "healthy" {
			if host.Resources != nil {
				result.LogicalCpuCount += host.Resources.Total.LogicalCpuCount
				result.PhysicalCpuCount += host.Resources.Total.PhysicalCpuCount
				result.FreeDiskSize += host.Resources.Total.FreeDiskSize
				result.MemorySize += host.Resources.Total.MemorySize
			}
		}
	}

	return result
}

func (j *JsonDatabase) GetOrchestratorInUseResources(ctx basecontext.ApiContext) models.HostResourceItem {
	result := models.HostResourceItem{}

	for _, host := range j.data.OrchestratorHosts {
		if host.State == "healthy" {
			if host.Resources != nil {
				result.LogicalCpuCount += host.Resources.TotalInUse.LogicalCpuCount
				result.PhysicalCpuCount += host.Resources.TotalInUse.PhysicalCpuCount
				result.FreeDiskSize += host.Resources.TotalInUse.FreeDiskSize
				result.MemorySize += host.Resources.TotalInUse.MemorySize
			}
		}
	}

	return result
}

func (j *JsonDatabase) GetOrchestratorReservedResources(ctx basecontext.ApiContext) models.HostResourceItem {
	result := models.HostResourceItem{}

	for _, host := range j.data.OrchestratorHosts {
		if host.State == "healthy" {
			if host.Resources != nil {
				result.LogicalCpuCount += host.Resources.TotalReserved.LogicalCpuCount
				result.PhysicalCpuCount += host.Resources.TotalReserved.PhysicalCpuCount
				result.FreeDiskSize += host.Resources.TotalReserved.FreeDiskSize
				result.MemorySize += host.Resources.TotalReserved.MemorySize
			}
		}
	}

	return result
}

func (j *JsonDatabase) GetOrchestratorHostResources(ctx basecontext.ApiContext, hostId string) (*models.HostResources, error) {
	host, err := j.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		return nil, err
	}
	if host == nil || host.Resources == nil {
		return nil, ErrOrchestratorHostNotFound
	}

	return host.Resources, nil
}

func (j *JsonDatabase) GetOrchestratorVirtualMachines(ctx basecontext.ApiContext, filter string) ([]models.VirtualMachine, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	var result []models.VirtualMachine

	hosts, err := j.GetOrchestratorHosts(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		if host.State == "healthy" {
			result = append(result, host.VirtualMachines...)
		}
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(result, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (j *JsonDatabase) GetOrchestratorHostVirtualMachines(ctx basecontext.ApiContext, hostId string, filter string) ([]models.VirtualMachine, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	host, err := j.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		return nil, err
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(host.VirtualMachines, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (j *JsonDatabase) GetOrchestratorHostVirtualMachine(ctx basecontext.ApiContext, hostId string, machineId string) (*models.VirtualMachine, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	host, err := j.GetOrchestratorHost(ctx, hostId)
	if err != nil {
		return nil, err
	}

	for _, machine := range host.VirtualMachines {
		if strings.EqualFold(machine.ID, machineId) {
			return &machine, nil
		}
	}

	return nil, ErrOrchestratorHostVirtualMachineNotFound
}