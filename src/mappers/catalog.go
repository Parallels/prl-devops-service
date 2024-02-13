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
		VirtualMachineContents: CatalogManifestContentItemsToDto(m.VirtualMachineContents),
		PackContents:           CatalogManifestContentItemsToDto(m.PackContents),
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

	if m.Provider != nil {
		data.Provider = &data_models.CatalogManifestProvider{
			Type:     m.Provider.Type,
			Host:     m.Provider.Host,
			Port:     m.Provider.Port,
			Username: m.Provider.Username,
			Password: m.Provider.Password,
			ApiKey:   m.Provider.ApiKey,
			Meta:     m.Provider.Meta,
		}
	}
	if data.Provider.Meta == nil {
		data.Provider.Meta = make(map[string]string)
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
		Size:                   m.Size,
		VirtualMachineContents: DtoCatalogManifestContentItemsToBase(m.VirtualMachineContents),
		PackContents:           DtoCatalogManifestContentItemsToBase(m.PackContents),
		Tainted:                m.Tainted,
		TaintedBy:              m.TaintedBy,
		TaintedAt:              m.TaintedAt,
		UnTaintedBy:            m.UnTaintedBy,
		Revoked:                m.Revoked,
		RevokedAt:              m.RevokedAt,
		RevokedBy:              m.RevokedBy,
		DownloadCount:          m.DownloadCount,
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
	}
	if data.Provider.Meta == nil {
		data.Provider.Meta = make(map[string]string)
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
		Tainted:            m.Tainted,
		TaintedBy:          m.TaintedBy,
		TaintedAt:          m.TaintedAt,
		UnTaintedBy:        m.UnTaintedBy,
		Revoked:            m.Revoked,
		RevokedAt:          m.RevokedAt,
		RevokedBy:          m.RevokedBy,
		DownloadCount:      m.DownloadCount,
	}

	if m.Provider != nil {
		data.Provider = &data_models.CatalogManifestProvider{
			Type:     m.Provider.Type,
			Host:     m.Provider.Host,
			Port:     m.Provider.Port,
			Username: m.Provider.Username,
			Password: m.Provider.Password,
			ApiKey:   m.Provider.ApiKey,
			Meta:     m.Provider.Meta,
		}
	}
	if data.Provider.Meta == nil {
		data.Provider.Meta = make(map[string]string)
	}

	return data
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
		DownloadCount:      m.DownloadCount,
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
	if data.Provider.Meta == nil {
		data.Provider.Meta = make(map[string]string)
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
	if data.Provider.Meta == nil {
		data.Provider.Meta = make(map[string]string)
	}

	return data
}

func BaseImportCatalogManifestResponseToApi(m catalog_models.ImportCatalogManifestResponse) models.ImportCatalogManifestResponse {
	data := models.ImportCatalogManifestResponse{
		ID: m.ID,
	}

	return data
}
