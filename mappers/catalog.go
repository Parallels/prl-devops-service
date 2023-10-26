package mappers

import (
	catalog_models "Parallels/pd-api-service/catalog/models"
	data_models "Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/models"
	"fmt"
)

func DtoCatalogManifestFromBase(m catalog_models.VirtualMachineManifest) data_models.CatalogVirtualMachineManifest {
	data := data_models.CatalogVirtualMachineManifest{
		ID:                 m.ID,
		Name:               m.Name,
		Path:               m.Path,
		MetadataPath:       m.MetadataPath,
		Type:               m.Type,
		Tags:               m.Tags,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		RequiredRoles:      m.RequiredRoles,
		RequiredClaims:     m.RequiredClaims,
		LastDownloadedAt:   m.LastDownloadedAt,
		LastDownloadedUser: m.LastDownloadedUser,
		Contents:           DtoCatalogManifestContentItemsFromBase(m.Contents),
		Size:               m.Size,
	}
	if m.Provider.Meta != nil {
		data.Provider = data_models.RemoteVirtualMachineProvider{
			Type: m.Provider.Type,
			Meta: m.Provider.Meta,
		}
	}

	return data
}

func DtoCatalogManifestToBase(m data_models.CatalogVirtualMachineManifest) catalog_models.VirtualMachineManifest {
	data := catalog_models.VirtualMachineManifest{
		ID:                 m.ID,
		Name:               m.Name,
		Path:               m.Path,
		MetadataPath:       m.MetadataPath,
		Type:               m.Type,
		Tags:               m.Tags,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		RequiredRoles:      m.RequiredRoles,
		RequiredClaims:     m.RequiredClaims,
		LastDownloadedAt:   m.LastDownloadedAt,
		LastDownloadedUser: m.LastDownloadedUser,
		Size:               m.Size,
		Contents:           DtoCatalogManifestContentItemsToBase(m.Contents),
	}

	if m.Provider.Meta != nil {
		data.Provider = catalog_models.CatalogManifestProvider{
			Type: m.Provider.Type,
			Meta: m.Provider.Meta,
		}
	}

	return data
}

func DtoCatalogManifestsFromBase(m []catalog_models.VirtualMachineManifest) []data_models.CatalogVirtualMachineManifest {
	var result []data_models.CatalogVirtualMachineManifest
	for _, item := range m {
		result = append(result, DtoCatalogManifestFromBase(item))
	}
	return result
}

func DtoCatalogManifestsToBase(m []data_models.CatalogVirtualMachineManifest) []catalog_models.VirtualMachineManifest {
	var result []catalog_models.VirtualMachineManifest
	for _, item := range m {
		result = append(result, DtoCatalogManifestToBase(item))
	}
	return result
}

func DtoCatalogManifestContentItemFromBase(m catalog_models.VirtualMachineManifestContentItem) data_models.RemoteVirtualMachineContentItem {
	return data_models.RemoteVirtualMachineContentItem{
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

func DtoCatalogManifestContentItemToBase(m data_models.RemoteVirtualMachineContentItem) catalog_models.VirtualMachineManifestContentItem {
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

func DtoCatalogManifestContentItemsFromBase(m []catalog_models.VirtualMachineManifestContentItem) []data_models.RemoteVirtualMachineContentItem {
	var result []data_models.RemoteVirtualMachineContentItem
	for _, item := range m {
		result = append(result, DtoCatalogManifestContentItemFromBase(item))
	}
	return result
}

func DtoCatalogManifestContentItemsToBase(m []data_models.RemoteVirtualMachineContentItem) []catalog_models.VirtualMachineManifestContentItem {
	var result []catalog_models.VirtualMachineManifestContentItem
	for _, item := range m {
		result = append(result, DtoCatalogManifestContentItemToBase(item))
	}
	return result
}

func DtoCatalogManifestFromApi(m models.CatalogVirtualMachineManifest) data_models.CatalogVirtualMachineManifest {
	data := data_models.CatalogVirtualMachineManifest{
		ID:                 m.ID,
		Name:               m.Name,
		Type:               m.Type,
		Tags:               m.Tags,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		LastDownloadedAt:   m.LastDownloadedAt,
		LastDownloadedUser: m.LastDownloadedUser,
	}

	if m.Provider.Meta != nil {
		for k, v := range m.Provider.Meta {
			data.Provider.Meta[k] = v
		}
		data.Provider.Type = m.Provider.Type
	}

	return data
}

func DtoCatalogManifestToApi(m data_models.CatalogVirtualMachineManifest) models.CatalogVirtualMachineManifest {
	data := models.CatalogVirtualMachineManifest{
		ID:                 m.ID,
		Name:               m.Name,
		Type:               m.Type,
		Tags:               m.Tags,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
		LastDownloadedAt:   m.LastDownloadedAt,
		LastDownloadedUser: m.LastDownloadedUser,
	}

	data.Size = fmt.Sprintf("%vGb", helpers.ConvertByteToGigabyte(m.Size))
	if m.Provider.Meta != nil {
		data.Provider = models.RemoteVirtualMachineProvider{
			Type: m.Provider.Type,
			Meta: m.Provider.Meta,
		}
	}
	return data
}

func DtoCatalogManifestsFromApi(m []models.CatalogVirtualMachineManifest) []data_models.CatalogVirtualMachineManifest {
	var result []data_models.CatalogVirtualMachineManifest
	for _, item := range m {
		result = append(result, DtoCatalogManifestFromApi(item))
	}
	return result
}

func DtoCatalogManifestsToApi(m []data_models.CatalogVirtualMachineManifest) []models.CatalogVirtualMachineManifest {
	var result []models.CatalogVirtualMachineManifest
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

	dto := DtoCatalogManifestFromBase(*m.Manifest)
	d := DtoCatalogManifestToApi(dto)
	data.Manifest = &d
	return data
}
