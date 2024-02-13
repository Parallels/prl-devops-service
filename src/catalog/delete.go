package catalog

import (
	"path/filepath"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/cleanupservice"
	"github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/serviceprovider"
)

func (s *CatalogManifestService) Delete(ctx basecontext.ApiContext, catalogId string, version string, architecture string) error {
	executed := false
	db := serviceprovider.Get().JsonDatabase
	if db == nil {
		return errors.New("no database connection")
	}
	if err := db.Connect(ctx); err != nil {
		return err
	}

	// Either we will be cleaning all of the catalog or just a specific version
	cleanItems := make([]models.VirtualMachineCatalogManifest, 0)

	if version == "" {
		dbManifest, err := db.GetCatalogManifestsByCatalogId(ctx, catalogId)
		if err != nil && err.Error() != "catalog manifest not found" {
			return err
		} else {
			for _, manifest := range dbManifest {
				cleanItems = append(cleanItems, mappers.DtoCatalogManifestToBase(manifest))
			}
		}
	} else if version != "" && architecture == "" {
		dbManifest, err := db.GetCatalogManifestsByCatalogIdAndVersion(ctx, catalogId, version)
		if err != nil && err.Error() != "catalog manifest not found" {
			return err
		} else {
			for _, manifest := range dbManifest {
				cleanItems = append(cleanItems, mappers.DtoCatalogManifestToBase(manifest))
			}
		}
	} else if version != "" && architecture != "" {
		dbManifest, err := db.GetCatalogManifestsByCatalogIdVersionAndArch(ctx, catalogId, version, architecture)
		if err != nil && err.Error() != "catalog manifest not found" {
			return err
		}
		cleanItems = append(cleanItems, mappers.DtoCatalogManifestToBase(*dbManifest))
	}

	connectionString := ""
	if len(cleanItems) == 0 {
		return errors.Newf("no catalog manifest found for id %s", catalogId)
	}

	cleanupService := cleanupservice.NewCleanupRequest()

	for _, cleanItem := range cleanItems {
		for _, rs := range s.remoteServices {
			check, checkErr := rs.Check(ctx, cleanItem.Provider.String())
			if checkErr != nil {
				ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
				return checkErr
			}

			if check {
				cleanupService.RemoteStorageService = rs
				executed = true
				metadataFilePath := filepath.Join(cleanItem.Path, cleanItem.MetadataFile)
				packFilePath := filepath.Join(cleanItem.Path, cleanItem.PackFile)
				cleanupService.AddRemoteFileCleanupOperation(metadataFilePath, false)
				cleanupService.AddRemoteFileCleanupOperation(packFilePath, false)
				cleanupService.AddRemoteFileCleanupOperation(cleanItem.Path, true)
			}
		}
	}

	if !executed {
		return errors.Newf("no remote service found for connection  %s", connectionString)
	}

	if cleanupErrors := cleanupService.Clean(ctx); len(cleanupErrors) > 0 {
		return errors.Newf("error cleaning up files: %v", cleanupErrors)
	}

	return nil
}
