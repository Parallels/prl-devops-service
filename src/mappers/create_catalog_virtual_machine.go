package mappers

import (
	catalog_models "github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/models"
)

func MapPullCatalogManifestRequestFromCreateCatalogVirtualMachineRequest(m models.CreateCatalogVirtualMachineRequest) catalog_models.PullCatalogManifestRequest {
	mapped := catalog_models.PullCatalogManifestRequest{
		CatalogId:        m.CatalogId,
		MachineName:      m.MachineName,
		Owner:            m.Owner,
		Version:          m.Version,
		Architecture:     m.Architecture,
		Connection:       m.Connection,
		ProviderMetadata: m.ProviderMetadata,
		StartAfterPull:   m.StartAfterPull,
		Path:             m.Path,
	}

	return mapped
}
