package catalog

import (
	"path/filepath"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/cleanupservice"
	"github.com/Parallels/prl-devops-service/catalog/models"
	db_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/mappers"
	"github.com/Parallels/prl-devops-service/serviceprovider"
)

func (s *CatalogManifestService) Delete(catalogId string, version string, architecture string) error {
	if s.ctx == nil {
		s.ctx = basecontext.NewRootBaseContext()
	}
	executed := false
	db := serviceprovider.Get().JsonDatabase
	if db == nil {
		return errors.New("no database connection")
	}
	if err := db.Connect(s.ctx); err != nil {
		return err
	}

	// Either we will be cleaning all of the catalog or just a specific version
	cleanItems := make([]models.VirtualMachineCatalogManifest, 0)

	if version == "" {
		dbManifest, err := db.GetCatalogManifestsByCatalogId(s.ctx, catalogId)
		if err != nil && err.Error() != "catalog manifest not found" {
			return err
		} else {
			for _, manifest := range dbManifest {
				cleanItems = append(cleanItems, mappers.DtoCatalogManifestToBase(manifest))
			}
		}
	} else if version != "" && architecture == "" {
		dbManifest, err := db.GetCatalogManifestsByCatalogIdAndVersion(s.ctx, catalogId, version)
		if err != nil && err.Error() != "catalog manifest not found" {
			return err
		} else {
			for _, manifest := range dbManifest {
				cleanItems = append(cleanItems, mappers.DtoCatalogManifestToBase(manifest))
			}
		}
	} else if version != "" && architecture != "" {
		dbManifest, err := db.GetCatalogManifestsByCatalogIdVersionAndArch(s.ctx, catalogId, version, architecture)
		if err != nil && err.Error() != "catalog manifest not found" {
			return err
		}
		cleanItems = append(cleanItems, mappers.DtoCatalogManifestToBase(*dbManifest))
	}

	connectionString := ""
	if len(cleanItems) == 0 {
		return errors.Newf("no catalog manifest found for id %s", catalogId)
	}

	cleanupService := cleanupservice.NewCleanupService()
	var foundCatalogIds []db_models.CatalogManifest = make([]db_models.CatalogManifest, 0)

	allManifestForCatalogId, _ := db.GetCatalogManifestsByCatalogId(s.ctx, catalogId)
	shouldCleanMainFolder := false

	if len(allManifestForCatalogId) > 0 {
		for idx, manifest := range allManifestForCatalogId {
			isFoundInCleanItems := false
			for _, cleanItem := range cleanItems {
				if cleanItem.ID == manifest.ID {
					isFoundInCleanItems = true
					break
				}
			}

			if !isFoundInCleanItems {
				foundCatalogIds = append(foundCatalogIds, allManifestForCatalogId[idx])
			}
		}

		if len(foundCatalogIds) == 0 {
			shouldCleanMainFolder = true
		}
	}

	for _, cleanItem := range cleanItems {
		for _, rs := range s.remoteServices {
			check, checkErr := rs.Check(s.ctx, cleanItem.Provider.String())
			if checkErr != nil {
				s.ns.NotifyErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
				return checkErr
			}

			if check {
				cleanupService.RemoteStorageService = rs
				executed = true
				metadataFilePath := filepath.Join(cleanItem.Path, cleanItem.MetadataFile)
				packFilePath := filepath.Join(cleanItem.Path, cleanItem.PackFile)
				cleanupService.AddRemoteFileCleanupOperation(metadataFilePath, false)
				cleanupService.AddRemoteFileCleanupOperation(packFilePath, false)
				if shouldCleanMainFolder {
					cleanupService.AddRemoteFileCleanupOperation(cleanItem.Path, true)
				}
			}
		}
	}

	if !executed {
		return errors.Newf("no remote service found for connection  %s", connectionString)
	}

	if cleanupErrors := cleanupService.Clean(s.ctx); len(cleanupErrors) > 0 {
		return errors.Newf("error cleaning up files: %v", cleanupErrors)
	}

	return nil
}
