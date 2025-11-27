package data

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

var (
	ErrOrchestratorHostEmptyIdOrHost          = errors.NewWithCode("no host specified", 500)
	ErrOrchestratorHostEmptyName              = errors.NewWithCode("host name cannot be empty", 500)
	ErrOrchestratorHostNotFound               = errors.NewWithCode("host not found", 404)
	ErrOrchestratorReverseProxyHostNotFound   = errors.NewWithCode("reverse proxy host not found", 404)
	ErrOrchestratorHostVirtualMachineNotFound = errors.NewWithCode("host virtual machine not found", 404)
)

func (j *JsonDatabase) GetOrchestratorHosts(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	return j.getOrchestratorHostsLocked(ctx, filter)
}

// getOrchestratorHostsLocked returns the list of orchestrator hosts without acquiring a lock.
//
// IMPORTANT: The caller MUST hold j.dataMutex (either RLock or Lock) before calling this function.
// Failure to do so will lead to data races.
// Attempting to acquire the lock inside this function would cause deadlocks when called from functions that already hold a Write lock.
func (j *JsonDatabase) getOrchestratorHostsLocked(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error) {
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

func (j *JsonDatabase) GetActiveOrchestratorHosts(ctx basecontext.ApiContext, filter string) ([]models.OrchestratorHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	var activeHosts []models.OrchestratorHost
	for _, host := range j.data.OrchestratorHosts {
		if host.Enabled {
			activeHosts = append(activeHosts, host)
		}
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(activeHosts, dbFilter)
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

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	return j.getOrchestratorHostUnsafe(ctx, idOrHost)
}

func (j *JsonDatabase) getOrchestratorHostUnsafe(ctx basecontext.ApiContext, idOrHost string) (*models.OrchestratorHost, error) {
	hosts, err := j.getOrchestratorHostsLocked(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		dbHost := host.GetHost()
		ctx.LogDebugf("Processing Host: %s", dbHost)
		if strings.EqualFold(host.ID, idOrHost) || strings.EqualFold(host.GetHost(), idOrHost) {
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

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	host.ID = helpers.GenerateId()
	host.CreatedAt = helpers.GetUtcCurrentDateTime()
	host.UpdatedAt = helpers.GetUtcCurrentDateTime()
	host.Enabled = true

	if u, _ := j.getOrchestratorHostUnsafe(ctx, host.GetHost()); u != nil {
		return nil, errors.NewWithCodef(400, "host %s already exists with ID %s", host.GetHost(), host.ID)
	}

	j.data.OrchestratorHosts = append(j.data.OrchestratorHosts, host)

	return &host, nil
}

func (j *JsonDatabase) DeleteOrchestratorHost(ctx basecontext.ApiContext, idOrHost string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if idOrHost == "" {
		return ErrOrchestratorHostEmptyIdOrHost
	}

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	for i, host := range j.data.OrchestratorHosts {
		if strings.EqualFold(host.ID, idOrHost) || strings.EqualFold(host.Host, idOrHost) {
			j.data.OrchestratorHosts = append(j.data.OrchestratorHosts[:i], j.data.OrchestratorHosts[i+1:]...)

			return nil
		}
	}

	return ErrOrchestratorHostNotFound
}

func (j *JsonDatabase) DeleteOrchestratorVirtualMachine(ctx basecontext.ApiContext, idOrHost string, vmIdOrName string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if idOrHost == "" {
		return ErrOrchestratorHostEmptyIdOrHost
	}

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	for _, host := range j.data.OrchestratorHosts {
		if strings.EqualFold(host.ID, idOrHost) || strings.EqualFold(host.Host, idOrHost) {
			for j, vm := range host.VirtualMachines {
				if strings.EqualFold(vm.ID, vmIdOrName) || strings.EqualFold(vm.Name, vmIdOrName) {
					host.VirtualMachines = append(host.VirtualMachines[:j], host.VirtualMachines[j+1:]...)
				}
			}

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

	j.dataMutex.Lock()

	found := false
	var result *models.OrchestratorHost

	for i, dbHost := range j.data.OrchestratorHosts {
		if strings.EqualFold(dbHost.ID, host.ID) || strings.EqualFold(dbHost.Host, host.Host) {
			ctx.LogDebugf("[Database] Host %s already exists with ID %s", host.Host, dbHost.ID)
			if host.Diff(j.data.OrchestratorHosts[i]) {
				j.data.OrchestratorHosts[i].Enabled = host.Enabled
				j.data.OrchestratorHosts[i].UpdatedAt = helpers.GetUtcCurrentDateTime()
				j.data.OrchestratorHosts[i].Host = host.Host
				j.data.OrchestratorHosts[i].OsVersion = host.OsVersion
				j.data.OrchestratorHosts[i].OsName = host.OsName
				j.data.OrchestratorHosts[i].ExternalIpAddress = host.ExternalIpAddress
				j.data.OrchestratorHosts[i].DevOpsVersion = host.DevOpsVersion
				j.data.OrchestratorHosts[i].Architecture = host.Architecture
				j.data.OrchestratorHosts[i].CpuModel = host.CpuModel
				j.data.OrchestratorHosts[i].Port = host.Port
				j.data.OrchestratorHosts[i].Authentication = host.Authentication
				j.data.OrchestratorHosts[i].Resources = host.Resources
				j.data.OrchestratorHosts[i].RequiredClaims = host.RequiredClaims
				j.data.OrchestratorHosts[i].RequiredRoles = host.RequiredRoles
				j.data.OrchestratorHosts[i].Description = host.Description
				j.data.OrchestratorHosts[i].Tags = host.Tags
				j.data.OrchestratorHosts[i].PathPrefix = host.PathPrefix
				j.data.OrchestratorHosts[i].Schema = host.Schema
				j.data.OrchestratorHosts[i].State = host.State
				j.data.OrchestratorHosts[i].LastUnhealthy = host.LastUnhealthy
				j.data.OrchestratorHosts[i].LastUnhealthyErrorMessage = host.LastUnhealthyErrorMessage
				j.data.OrchestratorHosts[i].HealthCheck = host.HealthCheck
				j.data.OrchestratorHosts[i].VirtualMachines = host.VirtualMachines
				// Other Data
				j.data.OrchestratorHosts[i].ParallelsDesktopVersion = host.ParallelsDesktopVersion
				j.data.OrchestratorHosts[i].ParallelsDesktopLicensed = host.ParallelsDesktopLicensed
				// Reverse Proxy Hosts
				j.data.OrchestratorHosts[i].IsReverseProxyEnabled = host.IsReverseProxyEnabled
				j.data.OrchestratorHosts[i].ReverseProxy = host.ReverseProxy
				j.data.OrchestratorHosts[i].ReverseProxyHosts = host.ReverseProxyHosts

				found = true
				result = &j.data.OrchestratorHosts[i]
			} else {
				ctx.LogDebugf("[Database] No changes detected for host %s", host.Host)
				j.dataMutex.Unlock()
				return host, nil
			}
			break
		}
	}

	if !found {
		j.dataMutex.Unlock()
		ctx.LogDebugf("[Database] Host %s not found, cannot update it", host.Host)
		return nil, ErrOrchestratorHostNotFound
	}

	j.dataMutex.Unlock()

	_ = j.SaveNow(ctx)
	ctx.LogDebugf("[Database] Host %s updated and saved", host.Host)
	return result, nil
}

func (j *JsonDatabase) UpdateOrchestratorHostDetails(ctx basecontext.ApiContext, host *models.OrchestratorHost) (*models.OrchestratorHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if host.ID == "" {
		return nil, ErrOrchestratorHostEmptyIdOrHost
	}

	j.dataMutex.Lock()

	// Check for duplicates
	for _, dbHost := range j.data.OrchestratorHosts {
		if strings.EqualFold(dbHost.Host, host.Host) && dbHost.ID != host.ID {
			j.dataMutex.Unlock()
			return nil, errors.NewWithCodef(400, "host %s already exists with ID %s", host.Host, dbHost.ID)
		}
	}

	for i, dbHost := range j.data.OrchestratorHosts {
		if strings.EqualFold(dbHost.ID, host.ID) {
			if host.Diff(j.data.OrchestratorHosts[i]) {
				j.data.OrchestratorHosts[i].Enabled = host.Enabled
				j.data.OrchestratorHosts[i].UpdatedAt = helpers.GetUtcCurrentDateTime()
				j.data.OrchestratorHosts[i].Host = host.Host
				j.data.OrchestratorHosts[i].Architecture = host.Architecture
				j.data.OrchestratorHosts[i].CpuModel = host.CpuModel
				j.data.OrchestratorHosts[i].Port = host.Port
				j.data.OrchestratorHosts[i].Authentication = host.Authentication
				j.data.OrchestratorHosts[i].Resources = host.Resources
				j.data.OrchestratorHosts[i].RequiredClaims = host.RequiredClaims
				j.data.OrchestratorHosts[i].RequiredRoles = host.RequiredRoles
				j.data.OrchestratorHosts[i].Description = host.Description
				j.data.OrchestratorHosts[i].Tags = host.Tags
				j.data.OrchestratorHosts[i].PathPrefix = host.PathPrefix
				j.data.OrchestratorHosts[i].Schema = host.Schema
				j.data.OrchestratorHosts[i].State = host.State
				j.data.OrchestratorHosts[i].LastUnhealthy = host.LastUnhealthy
				j.data.OrchestratorHosts[i].LastUnhealthyErrorMessage = host.LastUnhealthyErrorMessage
				j.data.OrchestratorHosts[i].HealthCheck = host.HealthCheck
				j.data.OrchestratorHosts[i].VirtualMachines = host.VirtualMachines

				j.dataMutex.Unlock()
				return &j.data.OrchestratorHosts[i], nil
			} else {
				ctx.LogDebugf("[Database] No changes detected for host %s", host.Host)
				j.dataMutex.Unlock()
				return host, nil
			}
		}
	}

	j.dataMutex.Unlock()
	return nil, ErrOrchestratorHostNotFound
}

func (j *JsonDatabase) GetOrchestratorAvailableResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	result := make(map[string]models.HostResourceItem)

	for _, host := range j.data.OrchestratorHosts {
		if host.State == "healthy" && host.Enabled {
			if host.Resources != nil {
				if _, ok := result[host.Resources.CpuType]; !ok {
					result[host.Resources.CpuType] = models.HostResourceItem{}
				}
				item := result[host.Resources.CpuType]
				item.LogicalCpuCount += host.Resources.TotalAvailable.LogicalCpuCount
				item.PhysicalCpuCount += host.Resources.TotalAvailable.PhysicalCpuCount
				item.FreeDiskSize += host.Resources.TotalAvailable.FreeDiskSize
				item.MemorySize += host.Resources.TotalAvailable.MemorySize
				item.DiskSize += host.Resources.TotalAvailable.DiskSize
				item.TotalAppleVms += host.Resources.TotalAvailable.TotalAppleVms
				result[host.Resources.CpuType] = item
			}
		}
	}

	return result
}

func (j *JsonDatabase) GetOrchestratorTotalResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	result := make(map[string]models.HostResourceItem)

	for _, host := range j.data.OrchestratorHosts {
		if host.State == "healthy" && host.Enabled {
			if host.Resources != nil {
				if _, ok := result[host.Resources.CpuType]; !ok {
					result[host.Resources.CpuType] = models.HostResourceItem{}
				}
				item := result[host.Resources.CpuType]
				item.LogicalCpuCount += host.Resources.Total.LogicalCpuCount
				item.PhysicalCpuCount += host.Resources.Total.PhysicalCpuCount
				item.FreeDiskSize += host.Resources.Total.FreeDiskSize
				item.MemorySize += host.Resources.Total.MemorySize
				item.DiskSize += host.Resources.Total.DiskSize
				item.TotalAppleVms += host.Resources.Total.TotalAppleVms
				result[host.Resources.CpuType] = item
			}
		}
	}

	return result
}

func (j *JsonDatabase) GetOrchestratorInUseResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	result := make(map[string]models.HostResourceItem)

	for _, host := range j.data.OrchestratorHosts {
		if host.State == "healthy" && host.Enabled {
			if host.Resources != nil {
				if _, ok := result[host.Resources.CpuType]; !ok {
					result[host.Resources.CpuType] = models.HostResourceItem{}
				}
				item := result[host.Resources.CpuType]
				item.LogicalCpuCount += host.Resources.TotalInUse.LogicalCpuCount
				item.PhysicalCpuCount += host.Resources.TotalInUse.PhysicalCpuCount
				item.FreeDiskSize += host.Resources.TotalInUse.FreeDiskSize
				item.DiskSize += host.Resources.TotalInUse.DiskSize
				item.TotalAppleVms += host.Resources.TotalInUse.TotalAppleVms
				item.MemorySize += host.Resources.TotalInUse.MemorySize
				result[host.Resources.CpuType] = item
			}
		}
	}

	return result
}

func (j *JsonDatabase) GetOrchestratorReservedResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	result := make(map[string]models.HostResourceItem)

	for _, host := range j.data.OrchestratorHosts {
		if host.State == "healthy" && host.Enabled {
			if host.Resources != nil {
				if _, ok := result[host.Resources.CpuType]; !ok {
					result[host.Resources.CpuType] = models.HostResourceItem{}
				}
				item := result[host.Resources.CpuType]
				item.TotalAppleVms += host.Resources.TotalAppleVms
				item.LogicalCpuCount += host.Resources.TotalReserved.LogicalCpuCount
				item.PhysicalCpuCount += host.Resources.TotalReserved.PhysicalCpuCount
				item.FreeDiskSize += host.Resources.TotalReserved.FreeDiskSize
				item.MemorySize += host.Resources.TotalReserved.MemorySize
				item.DiskSize += host.Resources.TotalReserved.DiskSize
				item.TotalAppleVms += host.Resources.TotalReserved.TotalAppleVms
				result[host.Resources.CpuType] = item
			}
		}
	}

	return result
}

func (j *JsonDatabase) GetOrchestratorSystemReservedResources(ctx basecontext.ApiContext) map[string]models.HostResourceItem {
	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	result := make(map[string]models.HostResourceItem)

	for _, host := range j.data.OrchestratorHosts {
		if host.State == "healthy" && host.Enabled {
			if host.Resources != nil {
				if _, ok := result[host.Resources.CpuType]; !ok {
					result[host.Resources.CpuType] = models.HostResourceItem{}
				}
				item := result[host.Resources.CpuType]
				item.LogicalCpuCount += host.Resources.SystemReserved.LogicalCpuCount
				item.PhysicalCpuCount += host.Resources.SystemReserved.PhysicalCpuCount
				item.FreeDiskSize += host.Resources.SystemReserved.FreeDiskSize
				item.DiskSize += host.Resources.SystemReserved.DiskSize
				item.MemorySize += host.Resources.SystemReserved.MemorySize
				result[host.Resources.CpuType] = item
			}
		}
	}

	return result
}

func (j *JsonDatabase) GetOrchestratorHostResources(ctx basecontext.ApiContext, hostId string) (*models.HostResources, error) {
	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	host, err := j.getOrchestratorHostUnsafe(ctx, hostId)
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

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	var result []models.VirtualMachine

	hosts, err := j.getOrchestratorHostsLocked(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		if host.State == "healthy" && host.Enabled {
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

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	host, err := j.getOrchestratorHostUnsafe(ctx, hostId)
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

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	host, err := j.getOrchestratorHostUnsafe(ctx, hostId)
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

func (j *JsonDatabase) GetOrchestratorReverseProxyHosts(ctx basecontext.ApiContext, hostId string, filter string) ([]*models.ReverseProxyHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	host, err := j.getOrchestratorHostUnsafe(ctx, hostId)
	if err != nil {
		return nil, err
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(host.ReverseProxyHosts, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (j *JsonDatabase) GetOrchestratorReverseProxyHost(ctx basecontext.ApiContext, hostId string, rpIdOrName string) (*models.ReverseProxyHost, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	hosts, err := j.getOrchestratorHostsLocked(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, host := range hosts {
		dbHost := host.GetHost()
		ctx.LogDebugf("host: %s", dbHost)
		if strings.EqualFold(host.ID, hostId) || strings.EqualFold(host.GetHost(), hostId) {
			for _, rpHost := range host.ReverseProxyHosts {
				hostname := rpHost.GetHost()
				if strings.EqualFold(rpHost.ID, rpIdOrName) || strings.EqualFold(hostname, rpIdOrName) {
					return rpHost, nil
				}
			}
		}
	}

	return nil, ErrOrchestratorReverseProxyHostNotFound
}

func (j *JsonDatabase) GetOrchestratorReverseProxyConfig(ctx basecontext.ApiContext, hostId string) (*models.ReverseProxy, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	host, err := j.getOrchestratorHostUnsafe(ctx, hostId)
	if err != nil {
		return nil, err
	}

	return host.ReverseProxy, nil
}
