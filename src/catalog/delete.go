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

func (s *CatalogManifestService) Delete(ctx basecontext.ApiContext, id string) error {
	executed := false
	db := serviceprovider.Get().JsonDatabase
	if db == nil {
		return errors.New("no database connection")
	}
	if err := db.Connect(ctx); err != nil {
		return err
	}

	connectionString := ""
	var manifest *models.VirtualMachineCatalogManifest
	dbManifest, err := db.GetCatalogManifest(ctx, id)
	if err != nil && err.Error() != "catalog manifest not found" {
		return err
	} else {
		if dbManifest != nil {
			m := mappers.DtoCatalogManifestToBase(*dbManifest)
			manifest = &m
		}
	}

	if manifest == nil || manifest.Provider == nil {
		return errors.Newf("no catalog manifest found for id %s", id)
	}

	cleanupService := cleanupservice.NewCleanupRequest()

	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(ctx, manifest.Provider.String())
		if checkErr != nil {
			ctx.LogError("Error checking remote service %v: %v", rs.Name(), checkErr)
			return checkErr
		}

		if check {
			cleanupService.RemoteStorageService = rs
			executed = true
			metadataFilePath := filepath.Join(manifest.Path, manifest.MetadataFile)
			packFilePath := filepath.Join(manifest.Path, manifest.PackFile)
			cleanupService.AddRemoteFileCleanupOperation(metadataFilePath, false)
			cleanupService.AddRemoteFileCleanupOperation(packFilePath, false)
			cleanupService.AddRemoteFileCleanupOperation(manifest.Path, true)
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
