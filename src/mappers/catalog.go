package mappers

import (
	catalog_models "github.com/Parallels/prl-devops-service/catalog/models"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

func CatalogManifestToDto(m catalog_models.VirtualMachineCatalogManifest) data_models.CatalogManifest {
	data := data_models.CatalogManifest{
		ID:                     m.ID,
		CatalogId:              m.CatalogId,
		Description:            m.Description,
		Name:                   m.Name,
		Architecture:           m.Architecture,
		Version:                m.Version,
		Path:                   m.Path,
		MetadataFile:           m.MetadataFile,
		PackFile:               m.PackFile,
		Type:                   m.Type,
		Tags:                   m.Tags,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
		RequiredRoles:          m.RequiredRoles,
		RequiredClaims:         m.RequiredClaims,
		LastDownloadedAt:       m.LastDownloadedAt,
		LastDownloadedUser:     m.LastDownloadedUser,
		IsCompressed:           m.IsCompressed,
		PackRelativePath:       m.PackRelativePath,
		VirtualMachineContents: CatalogManifestContentItemsToDto(m.VirtualMachineContents),
		PackContents:           CatalogManifestContentItemsToDto(m.PackContents),
		PackSize:               m.PackSize,
		Size:                   m.Size,
		Tainted:                m.Tainted,
		TaintedBy:              m.TaintedBy,
		TaintedAt:              m.TaintedAt,
		UnTaintedBy:            m.UnTaintedBy,
		Revoked:                m.Revoked,
		RevokedAt:              m.RevokedAt,
		RevokedBy:              m.RevokedBy,
		DownloadCount:          m.DownloadCount,
	}

	if m.MinimumSpecRequirements != nil {
		data.MinimumSpecRequirements = &data_models.MinimumSpecRequirement{
			Cpu:    m.MinimumSpecRequirements.Cpu,
			Memory: m.MinimumSpecRequirements.Memory,
			Disk:   m.MinimumSpecRequirements.Disk,
		}
	}

	if m.Provider != nil {
		provider := CatalogManifestProviderToDto(*m.Provider)
		data.Provider = &provider
	}

	if m.Tags == nil {
		data.Tags = make([]string, 0)
	}
	if m.RequiredRoles == nil {
		data.RequiredRoles = make([]string, 0)
	}
	if m.RequiredClaims == nil {
		data.RequiredClaims = make([]string, 0)
	}

	return data
}

func CatalogManifestProviderToDto(m catalog_models.CatalogManifestProvider) data_models.CatalogManifestProvider {
	provider := data_models.CatalogManifestProvider{
		Type:     m.Type,
		Host:     m.Host,
		Port:     m.Port,
		Username: m.Username,
		Password: m.Password,
		ApiKey:   m.ApiKey,
		Meta:     m.Meta,
	}

	if provider.Meta == nil {
		provider.Meta = make(map[string]string)
	}

	return provider
}

func DtoCatalogManifestProviderToBase(m data_models.CatalogManifestProvider) catalog_models.CatalogManifestProvider {
	provider := catalog_models.CatalogManifestProvider{
		Type:     m.Type,
		Host:     m.Host,
		Port:     m.Port,
		Username: m.Username,
		Password: m.Password,
		ApiKey:   m.ApiKey,
		Meta:     m.Meta,
	}

	if provider.Meta == nil {
		provider.Meta = make(map[string]string)
	}

	return provider
}

func DtoCatalogManifestToBase(m data_models.CatalogManifest) catalog_models.VirtualMachineCatalogManifest {
	data := catalog_models.VirtualMachineCatalogManifest{
		ID:                     m.ID,
		CatalogId:              m.CatalogId,
		Version:                m.Version,
		Name:                   m.Name,
		Description:            m.Description,
		Architecture:           m.Architecture,
		Path:                   m.Path,
		MetadataFile:           m.MetadataFile,
		PackFile:               m.PackFile,
		Type:                   m.Type,
		Tags:                   m.Tags,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
		RequiredRoles:          m.RequiredRoles,
		RequiredClaims:         m.RequiredClaims,
		LastDownloadedAt:       m.LastDownloadedAt,
		LastDownloadedUser:     m.LastDownloadedUser,
		IsCompressed:           m.IsCompressed,
		PackRelativePath:       m.PackRelativePath,
		Size:                   m.Size,
		VirtualMachineContents: DtoCatalogManifestContentItemsToBase(m.VirtualMachineContents),
		PackContents:           DtoCatalogManifestContentItemsToBase(m.PackContents),
		PackSize:               m.PackSize,
		Tainted:                m.Tainted,
		TaintedBy:              m.TaintedBy,
		TaintedAt:              m.TaintedAt,
		UnTaintedBy:            m.UnTaintedBy,
		Revoked:                m.Revoked,
		RevokedAt:              m.RevokedAt,
		RevokedBy:              m.RevokedBy,
		DownloadCount:          m.DownloadCount,
	}

	if m.MinimumSpecRequirements != nil {
		data.MinimumSpecRequirements = &catalog_models.MinimumSpecRequirement{
			Cpu:    m.MinimumSpecRequirements.Cpu,
			Memory: m.MinimumSpecRequirements.Memory,
			Disk:   m.MinimumSpecRequirements.Disk,
		}
	}

	if m.Provider != nil {
		provider := DtoCatalogManifestProviderToBase(*m.Provider)
		data.Provider = &provider
	}

	if m.Tags == nil {
		data.Tags = make([]string, 0)
	}

	if m.RequiredRoles == nil {
		data.RequiredRoles = make([]string, 0)
	}

	if m.RequiredClaims == nil {
		data.RequiredClaims = make([]string, 0)
	}

	return data
}

func CatalogManifestsToDto(m []catalog_models.VirtualMachineCatalogManifest) []data_models.CatalogManifest {
	var result []data_models.CatalogManifest
	for _, item := range m {
		result = append(result, CatalogManifestToDto(item))
	}
	return result
}

func DtoCatalogManifestsToBase(m []data_models.CatalogManifest) []catalog_models.VirtualMachineCatalogManifest {
	var result []catalog_models.VirtualMachineCatalogManifest
	for _, item := range m {
		result = append(result, DtoCatalogManifestToBase(item))
	}
	return result
}

func CatalogManifestContentItemToDto(m catalog_models.VirtualMachineManifestContentItem) data_models.CatalogManifestContentItem {
	return data_models.CatalogManifestContentItem{
		IsDir:     m.IsDir,
		Name:      m.Name,
		Path:      m.Path,
		Checksum:  m.Checksum,
		Size:      m.Size,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
	}
}

func DtoCatalogManifestContentItemToBase(m data_models.CatalogManifestContentItem) catalog_models.VirtualMachineManifestContentItem {
	return catalog_models.VirtualMachineManifestContentItem{
		IsDir:     m.IsDir,
		Name:      m.Name,
		Path:      m.Path,
		Checksum:  m.Checksum,
		Size:      m.Size,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
	}
}

func CatalogManifestContentItemsToDto(m []catalog_models.VirtualMachineManifestContentItem) []data_models.CatalogManifestContentItem {
	var result []data_models.CatalogManifestContentItem
	for _, item := range m {
		result = append(result, CatalogManifestContentItemToDto(item))
	}
	return result
}

func CatalogManifestMinimumSpecsToDto(m catalog_models.MinimumSpecRequirement) data_models.MinimumSpecRequirement {
	return data_models.MinimumSpecRequirement{
		Cpu:    m.Cpu,
		Memory: m.Memory,
		Disk:   m.Disk,
	}
}

func DtoCatalogManifestContentItemsToBase(m []data_models.CatalogManifestContentItem) []catalog_models.VirtualMachineManifestContentItem {
	var result []catalog_models.VirtualMachineManifestContentItem
	for _, item := range m {
		result = append(result, DtoCatalogManifestContentItemToBase(item))
	}
	return result
}

func ApiCatalogManifestToDto(m models.CatalogManifest) data_models.CatalogManifest {
	data := data_models.CatalogManifest{
		ID:                 m.ID,
		Name:               m.Name,
		CatalogId:          m.CatalogId,
		Description:        m.Description,
		Architecture:       m.Architecture,
		Version:            m.Version,
		Type:               m.Type,
		Tags:               m.Tags,
		Path:               m.Path,
		PackFile:           m.PackFilename,
		MetadataFile:       m.MetadataFilename,
		RequiredRoles:      m.RequiredRoles,
		RequiredClaims:     m.RequiredClaims,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		LastDownloadedAt:   m.LastDownloadedAt,
		LastDownloadedUser: m.LastDownloadedUser,
		IsCompressed:       m.IsCompressed,
		PackRelativePath:   m.PackRelativePath,
		Tainted:            m.Tainted,
		TaintedBy:          m.TaintedBy,
		TaintedAt:          m.TaintedAt,
		UnTaintedBy:        m.UnTaintedBy,
		Revoked:            m.Revoked,
		RevokedAt:          m.RevokedAt,
		RevokedBy:          m.RevokedBy,
		PackSize:           m.PackSize,
		DownloadCount:      m.DownloadCount,
	}

	if m.Provider != nil {
		provider := ApiCatalogManifestProviderToDto(*m.Provider)
		data.Provider = &provider
	}

	if data.Tags == nil {
		data.Tags = make([]string, 0)
	}

	if data.RequiredRoles == nil {
		data.RequiredRoles = make([]string, 0)
	}

	if data.RequiredClaims == nil {
		data.RequiredClaims = make([]string, 0)
	}

	return data
}

func ApiCatalogManifestProviderToDto(m models.RemoteVirtualMachineProvider) data_models.CatalogManifestProvider {
	provider := data_models.CatalogManifestProvider{
		Type:     m.Type,
		Host:     m.Host,
		Port:     m.Port,
		Username: m.Username,
		Password: m.Password,
		ApiKey:   m.ApiKey,
		Meta:     m.Meta,
	}

	if provider.Meta == nil {
		provider.Meta = make(map[string]string)
	}

	return provider
}

func DtoCatalogManifestToApi(m data_models.CatalogManifest) models.CatalogManifest {
	data := models.CatalogManifest{
		ID:                 m.ID,
		CatalogId:          m.CatalogId,
		Description:        m.Description,
		Version:            m.Version,
		Architecture:       m.Architecture,
		Name:               m.Name,
		Type:               m.Type,
		Tags:               m.Tags,
		Path:               m.Path,
		PackFilename:       m.PackFile,
		MetadataFilename:   m.MetadataFile,
		RequiredRoles:      m.RequiredRoles,
		RequiredClaims:     m.RequiredClaims,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		LastDownloadedAt:   m.LastDownloadedAt,
		LastDownloadedUser: m.LastDownloadedUser,
		Tainted:            m.Tainted,
		TaintedBy:          m.TaintedBy,
		TaintedAt:          m.TaintedAt,
		UnTaintedBy:        m.UnTaintedBy,
		Revoked:            m.Revoked,
		RevokedAt:          m.RevokedAt,
		RevokedBy:          m.RevokedBy,
		PackSize:           m.PackSize,
		Size:               m.Size,
		DownloadCount:      m.DownloadCount,
		IsCompressed:       m.IsCompressed,
	}

	if data.Tags == nil {
		data.Tags = make([]string, 0)
	}

	if m.Provider != nil {
		data.Provider = &models.RemoteVirtualMachineProvider{
			Type:     m.Provider.Type,
			Host:     m.Provider.Host,
			Port:     m.Provider.Port,
			Username: m.Provider.Username,
			Password: m.Provider.Password,
			ApiKey:   m.Provider.ApiKey,
			Meta:     m.Provider.Meta,
		}
	}
	if m.MinimumSpecRequirements != nil {
		data.MinimumSpecRequirements = &models.MinimumSpecRequirement{
			Cpu:    m.MinimumSpecRequirements.Cpu,
			Memory: m.MinimumSpecRequirements.Memory,
			Disk:   m.MinimumSpecRequirements.Disk,
		}
	}

	if m.PackContents != nil {
		data.PackContents = make([]models.CatalogManifestPackItem, 0)
		for _, item := range m.PackContents {
			data.PackContents = append(data.PackContents, models.CatalogManifestPackItem{
				IsDir: item.IsDir,
				Name:  item.Name,
				Path:  item.Path,
			})
		}
	}

	return data
}

func DtoCatalogManifestsToApi(m []data_models.CatalogManifest) []models.CatalogManifest {
	var result []models.CatalogManifest
	for _, item := range m {
		result = append(result, DtoCatalogManifestToApi(item))
	}
	return result
}

func BasePullCatalogManifestResponseToApi(m catalog_models.PullCatalogManifestResponse) models.PullCatalogManifestResponse {
	data := models.PullCatalogManifestResponse{
		ID:          m.ID,
		MachineID:   m.MachineID,
		LocalPath:   m.LocalPath,
		MachineName: m.MachineName,
	}

	dto := CatalogManifestToDto(*m.Manifest)
	d := DtoCatalogManifestToApi(dto)
	data.Manifest = &d
	return data
}

func ApiCatalogManifestToCatalogManifest(m models.CatalogManifest) catalog_models.VirtualMachineCatalogManifest {
	data := catalog_models.VirtualMachineCatalogManifest{
		ID:                 m.ID,
		CatalogId:          m.CatalogId,
		Version:            m.Version,
		Name:               m.Name,
		Description:        m.Description,
		Architecture:       m.Architecture,
		Path:               m.Path,
		PackFile:           m.PackFilename,
		MetadataFile:       m.MetadataFilename,
		Type:               m.Type,
		Tags:               m.Tags,
		RequiredRoles:      m.RequiredRoles,
		RequiredClaims:     m.RequiredClaims,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		LastDownloadedAt:   m.LastDownloadedAt,
		LastDownloadedUser: m.LastDownloadedUser,
		Tainted:            m.Tainted,
		TaintedBy:          m.TaintedBy,
		TaintedAt:          m.TaintedAt,
		UnTaintedBy:        m.UnTaintedBy,
		Revoked:            m.Revoked,
		RevokedAt:          m.RevokedAt,
		RevokedBy:          m.RevokedBy,
		DownloadCount:      m.DownloadCount,
		PackSize:           m.PackSize,
		IsCompressed:       m.IsCompressed,
	}

	if m.Provider != nil {
		data.Provider = &catalog_models.CatalogManifestProvider{
			Type:     m.Provider.Type,
			Host:     m.Provider.Host,
			Port:     m.Provider.Port,
			Username: m.Provider.Username,
			Password: m.Provider.Password,
			ApiKey:   m.Provider.ApiKey,
			Meta:     m.Provider.Meta,
		}

		if data.Provider.Meta == nil {
			data.Provider.Meta = make(map[string]string)
		}
	}

	if m.PackContents != nil {
		data.PackContents = make([]catalog_models.VirtualMachineManifestContentItem, 0)
		for _, item := range m.PackContents {
			data.PackContents = append(data.PackContents, catalog_models.VirtualMachineManifestContentItem{
				IsDir: item.IsDir,
				Name:  item.Name,
				Path:  item.Path,
			})
		}
	}

	return data
}

func BaseImportCatalogManifestResponseToApi(m catalog_models.ImportCatalogManifestResponse) models.ImportCatalogManifestResponse {
	data := models.ImportCatalogManifestResponse{
		ID: m.ID,
	}

	return data
}

func BaseImportVmResponseToApi(m catalog_models.ImportVmResponse) models.ImportVmResponse {
	data := models.ImportVmResponse{
		ID: m.ID,
	}

	return data
}

func BaseVirtualMachineCatalogManifestListToApi(m catalog_models.CachedManifests) models.VirtualMachineCatalogManifestList {
	data := models.VirtualMachineCatalogManifestList{
		TotalSize: m.TotalSize,
		Manifests: BaseVirtualMachineCatalogManifestsToApi(m.Manifests),
	}

	return data
}

func BaseVirtualMachineCatalogManifestToApi(m catalog_models.VirtualMachineCatalogManifest) models.CatalogManifest {
	data := models.CatalogManifest{
		ID:                      m.ID,
		CatalogId:               m.CatalogId,
		Version:                 m.Version,
		Name:                    m.Name,
		Description:             m.Description,
		Architecture:            m.Architecture,
		Type:                    m.Type,
		Tags:                    m.Tags,
		Size:                    m.Size,
		Path:                    m.Path,
		PackFilename:            m.PackFile,
		MetadataFilename:        m.MetadataFile,
		CreatedAt:               m.CreatedAt,
		Provider:                BaseRemoteVirtualMachineProviderToApi(m.Provider),
		UpdatedAt:               m.UpdatedAt,
		RequiredClaims:          m.RequiredClaims,
		RequiredRoles:           m.RequiredRoles,
		LastDownloadedAt:        m.LastDownloadedAt,
		LastDownloadedUser:      m.LastDownloadedUser,
		IsCompressed:            m.IsCompressed,
		PackRelativePath:        m.PackRelativePath,
		DownloadCount:           m.DownloadCount,
		PackContents:            BaseCatalogManifestContentItemsToApi(m.PackContents),
		PackSize:                m.PackSize,
		Tainted:                 m.Tainted,
		TaintedBy:               m.TaintedBy,
		TaintedAt:               m.TaintedAt,
		UnTaintedBy:             m.UnTaintedBy,
		Revoked:                 m.Revoked,
		RevokedAt:               m.RevokedAt,
		RevokedBy:               m.RevokedBy,
		MinimumSpecRequirements: BaseMinimumSpecRequirementToApi(m.MinimumSpecRequirements),
		CacheDate:               m.CachedDate,
		CacheLocalFullPath:      m.CacheLocalFullPath,
		CacheMetadataName:       m.CacheMetadataName,
		CacheFileName:           m.CacheFileName,
		CacheType:               m.CacheType,
		CacheSize:               m.CacheSize,
	}

	return data
}

func BaseVirtualMachineCatalogManifestsToApi(m []catalog_models.VirtualMachineCatalogManifest) []models.CatalogManifest {
	var result []models.CatalogManifest
	if len(m) == 0 {
		result = make([]models.CatalogManifest, 0)
		return result
	}

	for _, item := range m {
		result = append(result, BaseVirtualMachineCatalogManifestToApi(item))
	}
	return result
}

func BaseCatalogManifestContentItemToApi(m catalog_models.VirtualMachineManifestContentItem) models.CatalogManifestPackItem {
	result := models.CatalogManifestPackItem{
		IsDir:     m.IsDir,
		Name:      m.Name,
		Path:      m.Path,
		Checksum:  m.Checksum,
		Size:      m.Size,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
	}

	return result
}

func BaseCatalogManifestContentItemsToApi(m []catalog_models.VirtualMachineManifestContentItem) []models.CatalogManifestPackItem {
	var result []models.CatalogManifestPackItem
	if len(m) == 0 {
		result = make([]models.CatalogManifestPackItem, 0)
		return result
	}

	for _, item := range m {
		result = append(result, BaseCatalogManifestContentItemToApi(item))
	}
	return result
}

func BaseMinimumSpecRequirementToApi(m *catalog_models.MinimumSpecRequirement) *models.MinimumSpecRequirement {
	if m == nil {
		return nil
	}
	result := &models.MinimumSpecRequirement{
		Cpu:    m.Cpu,
		Memory: m.Memory,
		Disk:   m.Disk,
	}

	return result
}

func BaseRemoteVirtualMachineProviderToApi(m *catalog_models.CatalogManifestProvider) *models.RemoteVirtualMachineProvider {
	if m == nil {
		return nil
	}
	result := &models.RemoteVirtualMachineProvider{
		Type:     m.Type,
		Host:     m.Host,
		Port:     m.Port,
		Username: m.Username,
		Password: m.Password,
		ApiKey:   m.ApiKey,
		Meta:     m.Meta,
	}

	return result
}
