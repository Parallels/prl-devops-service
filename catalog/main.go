package catalog

import (
	"Parallels/pd-api-service/catalog/interfaces"
	"Parallels/pd-api-service/catalog/models"
	"Parallels/pd-api-service/catalog/providers/aws_s3_bucket"
	"Parallels/pd-api-service/catalog/providers/local"
	"Parallels/pd-api-service/common"
	"Parallels/pd-api-service/helpers"
	"Parallels/pd-api-service/mappers"
	"Parallels/pd-api-service/service_provider"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cjlapao/common-go/helper"
)

var logger = common.Logger

type CatalogManifestService struct {
	remoteServices []interfaces.RemoteStorageService
	context        context.Context
}

func NewManifestService(ctx context.Context) *CatalogManifestService {
	manifestService := &CatalogManifestService{
		context: ctx,
	}
	manifestService.remoteServices = make([]interfaces.RemoteStorageService, 0)
	manifestService.AddRemoteService(aws_s3_bucket.NewAwsS3RemoteService())
	manifestService.AddRemoteService(local.NewLocalProviderService())
	return manifestService
}

func (s *CatalogManifestService) AddRemoteService(service interfaces.RemoteStorageService) {
	exists := false
	for _, remoteService := range s.remoteServices {
		if remoteService.Name() == service.Name() {
			exists = true
			break
		}
	}

	if exists {
		return
	}

	s.remoteServices = append(s.remoteServices, service)
}

func (s *CatalogManifestService) Push(r *models.PushCatalogManifestRequest) (*models.VirtualMachineManifest, error) {
	executed := false
	manifest, err := s.GenerateManifest(r)
	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(r.Connection)
		if checkErr != nil {
			logger.Error("Error checking remote service %v: %v", rs.Name(), checkErr)
			return nil, checkErr
		}

		if check {
			executed = true
			logger.Info("Generating vm manifest for %v", r.Name)
			if err != nil {
				logger.Error("Error generating manifest for %v: %v", r.Name, err)
				return nil, err
			}
			manifest.Provider = models.CatalogManifestProvider{
				Type: rs.Name(),
				Meta: rs.GetProviderMeta(),
			}

			logger.Info("Pushing manifest %v", manifest.Name)
			var catalogManifest *models.VirtualMachineManifest
			if err := rs.PullFile(rs.GetProviderRootPath(), s.getMetaFilename(manifest.ID), "/tmp"); err == nil {
				logger.Info("Remote Manifest found, retrieving it")
				catalogManifest, err = s.readManifestFromFile(filepath.Join("/tmp", s.getMetaFilename(manifest.ID)))
				if err != nil {
					logger.Error("Error reading manifest from file %v: %v", filepath.Join("/tmp", s.getMetaFilename(manifest.ID)), err)
					if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
						logger.Error("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
						return nil, err
					}

					return nil, err
				}
				if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
					logger.Error("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}
				manifest.CreatedAt = catalogManifest.CreatedAt
				manifest.RequiredRoles = catalogManifest.RequiredRoles
				manifest.RequiredClaims = catalogManifest.RequiredClaims
			}

			db := service_provider.Get().JsonDatabase
			if catalogManifest == nil {
				exists, _ := db.GetCatalogManifest(s.context, manifest.ID)
				if exists != nil {
					logger.Error("Manifest %v already exists in db", manifest.ID)
					return nil, fmt.Errorf("manifest %v already exists in db", manifest.ID)
				}

				logger.Info("Remote Manifest not found, creating it")
				if err := rs.CreateFolder(manifest.ID, "/"); err != nil {
					return nil, err
				}
				logger.Info("Pushing manifest files %v", s.getMetaFilename(manifest.ID))
				for _, file := range manifest.Contents {
					destinationPath := filepath.Join(manifest.ID, file.Path)
					sourcePath := filepath.Join(manifest.Path, file.Path)
					if file.IsDir {
						if err := rs.CreateFolder(destinationPath, file.Name); err != nil {
							return nil, err
						}
					} else {
						manifest.Size += file.Size
						if err := rs.PushFile(sourcePath, destinationPath, file.Name); err != nil {
							return nil, err
						}
					}
				}
				logger.Info("Finished pushing %d files for manifest %v", len(manifest.Contents), s.getMetaFilename(manifest.ID))

				logger.Info("Pushing manifest %v", manifest.Name)
				manifest.Path = filepath.Join(rs.GetProviderRootPath(), manifest.ID)
				manifest.MetadataPath = filepath.Join(rs.GetProviderRootPath(), s.getMetaFilename(manifest.ID))
				manifestContent, err := json.MarshalIndent(manifest, "", "  ")
				if err != nil {
					logger.Error("Error marshalling manifest %v: %v", manifest, err)
					return nil, err
				}

				if err := helper.WriteToFile(string(manifestContent), "/tmp/"+s.getMetaFilename(manifest.ID)); err != nil {
					logger.Error("Error writing manifest to temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}

				if err := rs.PushFile("/tmp", "/", s.getMetaFilename(manifest.ID)); err != nil {
					logger.Error("Error pushing manifest to remote storage %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
						logger.Error("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
						return nil, err
					}

					return nil, err
				}

				if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
					logger.Error("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}

				dto := mappers.DtoCatalogManifestFromBase(*manifest)

				if err := db.CreateCatalogManifest(s.context, &dto); err != nil {
					logger.Error("Error adding manifest to database %v: %v", manifest.Name, err)
					return nil, err
				}
				logger.Info("Finished pushing manifest %v", manifest.Name)
				return manifest, nil
			} else {
				logger.Info("Remote Manifest found, checking for changes in files")
				for _, file := range manifest.Contents {
					destinationPath := filepath.Join(manifest.ID, file.Path)
					sourcePath := filepath.Join(manifest.Path, file.Path)
					if file.IsDir {
						exists, err := rs.FolderExists(destinationPath, file.Name)
						if err != nil {
							logger.Error("Error checking if folder exists %v: %v", destinationPath, err)
							return nil, err
						}
						if !exists {
							if err := rs.CreateFolder(destinationPath, file.Name); err != nil {
								logger.Error("Error creating folder %v: %v", destinationPath, err)
								return nil, err
							}
						}
					} else {
						manifest.Size += file.Size
						exists, err := rs.FileExists(destinationPath, file.Name)
						if err != nil {
							logger.Error("Error checking if file exists %v: %v", destinationPath, err)
							return nil, err
						}
						if !exists {
							if err := rs.PushFile(sourcePath, destinationPath, file.Name); err != nil {
								logger.Error("Error pushing file %v: %v", destinationPath, err)
								return nil, err
							}
						} else {
							checksum, err := rs.FileChecksum(destinationPath, file.Name)
							if err != nil {
								logger.Error("Error getting file checksum %v: %v", destinationPath, err)
								return nil, err
							}
							if checksum != file.Checksum {
								if err := rs.DeleteFile(destinationPath, file.Name); err != nil {
									logger.Error("Error deleting file %v: %v", destinationPath, err)
									return nil, err
								}
								if err := rs.PushFile(sourcePath, destinationPath, file.Name); err != nil {
									logger.Error("Error pushing file %v: %v", destinationPath, err)
									return nil, err
								}
							} else {
								logger.Info("File %v is up to date", filepath.Join(destinationPath, file.Name))
							}
						}
					}
				}

				logger.Info("Finished pushing %v files for manifest %v", len(manifest.Contents), s.getMetaFilename(manifest.ID))

				if err := rs.DeleteFile("/", s.getMetaFilename(manifest.ID)); err != nil {
					logger.Error("Error deleting manifest %v: %v", s.getMetaFilename(manifest.ID), err)
					return nil, err
				}

				logger.Info("Pushing new manifest %v", manifest.Name)
				manifest.Path = filepath.Join(rs.GetProviderRootPath(), manifest.ID)
				manifest.MetadataPath = filepath.Join(rs.GetProviderRootPath(), s.getMetaFilename(manifest.ID))
				manifestContent, err := json.MarshalIndent(manifest, "", "  ")
				if err != nil {
					logger.Error("Error marshalling manifest %v: %v", manifest.Name, err)
					return nil, err
				}

				if err := helper.WriteToFile(string(manifestContent), "/tmp/"+s.getMetaFilename(manifest.ID)); err != nil {
					logger.Error("Error writing manifest to temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}

				if err := rs.PushFile("/tmp", "/", s.getMetaFilename(manifest.ID)); err != nil {
					logger.Error("Error pushing manifest to remote storage %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
						logger.Error("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
						return nil, err
					}

					return nil, err
				}

				if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
					logger.Error("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}

				exists, _ := db.GetCatalogManifest(s.context, manifest.ID)
				if exists != nil {
					logger.Info("Updating manifest %v", manifest.Name)
					dto := mappers.DtoCatalogManifestFromBase(*manifest)
					if err := db.UpdateCatalogManifest(s.context, dto); err != nil {
						logger.Error("Error updating manifest %v: %v", manifest.Name, err)
						return nil, err
					}
				} else {
					logger.Info("Creating manifest %v", manifest.Name)
					dto := mappers.DtoCatalogManifestFromBase(*manifest)
					if err := db.CreateCatalogManifest(s.context, &dto); err != nil {
						logger.Error("Error creating manifest %v: %v", manifest.Name, err)
						return nil, err
					}
				}

				logger.Info("Finished pushing manifest %v", manifest.Name)
				return manifest, nil
			}
		}
	}

	if !executed {
		return nil, fmt.Errorf("no remote service found for connection %v", r.Connection)
	}

	return manifest, nil
}

func (s *CatalogManifestService) Pull(r *models.PullCatalogManifestRequest) (*models.PullCatalogManifestResponse, error) {
	executed := false
	db := service_provider.Get().JsonDatabase
	if db == nil {
		return nil, errors.New("no database connection")
	}
	if err := db.Connect(); err != nil {
		return nil, err
	}

	connectionString := ""
	var manifest *models.VirtualMachineManifest
	dbManifest, err := db.GetCatalogManifest(s.context, r.ID)
	if err != nil && err.Error() != "catalog manifest not found" {
		return nil, err
	} else {
		if dbManifest != nil {
			m := mappers.DtoCatalogManifestToBase(*dbManifest)
			manifest = &m
		}
	}

	if manifest != nil {
		connectionString = s.parseProviderMetadata(manifest.Provider.String(), r.ProviderMetadata)
	} else if r.Connection != "" {
		connectionString = s.parseProviderMetadata(r.Connection, r.ProviderMetadata)
	}

	if connectionString == "" {
		return nil, errors.New("no connection string")
	}

	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(connectionString)
		if checkErr != nil {
			logger.Error("Error checking remote service %v: %v", rs.Name(), checkErr)
			return nil, checkErr
		}

		if check {
			var catalogManifest *models.VirtualMachineManifest
			if err := rs.PullFile(filepath.Dir(manifest.MetadataPath), s.getMetaFilename(manifest.ID), "/tmp"); err == nil {
				logger.Info("Remote Manifest found, retrieving it")
				catalogManifest, err = s.readManifestFromFile(filepath.Join("/tmp", s.getMetaFilename(manifest.ID)))
				if err != nil {
					logger.Error("Error reading manifest from file %v: %v", filepath.Join("/tmp", s.getMetaFilename(manifest.ID)), err)
					if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
						logger.Error("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
						return nil, err
					}

					return nil, err
				}
				if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
					logger.Error("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}
			}

			if catalogManifest == nil {
				return nil, errors.New("manifest not found in the provider")
			}

			logger.Info("Pulling manifest %v", catalogManifest.Name)
			machineName := ""
			if r.MachineName != "" {
				machineName = r.MachineName
			} else {
				machineName = catalogManifest.Name
			}

			localMachineFolder := fmt.Sprintf("%s.%s", filepath.Join(r.Path, machineName), catalogManifest.Type)
			if err := helpers.CreateDirIfNotExist(localMachineFolder); err != nil {
				return nil, err
			}

			for _, file := range catalogManifest.Contents {
				destinationPath := filepath.Join(localMachineFolder, file.Path)
				fullSrcPath := filepath.Join(catalogManifest.Path, file.Path)
				if file.IsDir {
					dirFolder := filepath.Join(destinationPath, file.Name)
					if err := helpers.CreateDirIfNotExist(dirFolder); err != nil {
						logger.Error("Error creating directory %v: %v", dirFolder, err)
						return nil, err
					}
					logger.Info("Created directory %v", fullSrcPath)
				} else {
					if err := rs.PullFile(fullSrcPath, file.Name, destinationPath); err != nil {
						logger.Error("Error pulling file %v: %v", filepath.Join(destinationPath, file.Name), err)
						return nil, err
					}
					logger.Info("Pulled file %v", filepath.Join(destinationPath, file.Name))
				}
			}

			logger.Info("Finished pulling %v files for manifest %v", len(catalogManifest.Contents), s.getMetaFilename(catalogManifest.ID))
			resultData := models.PullCatalogManifestResponse{
				ID:          catalogManifest.ID,
				LocalPath:   localMachineFolder,
				MachineName: machineName,
				Manifest:    catalogManifest,
			}

			return &resultData, nil
		}
	}

	if !executed {
		return nil, errors.New("no remote service found for connection " + connectionString)
	}

	return nil, errors.New("unknown error")
}

func (s *CatalogManifestService) GenerateManifest(r *models.PushCatalogManifestRequest) (*models.VirtualMachineManifest, error) {
	result := models.VirtualMachineManifest{}
	result.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	result.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	result.Name = r.Name
	result.Path = r.LocalPath
	if r.Uuid != "" {
		result.ID = s.getConformName(r.Uuid)
	} else {
		result.ID = s.getConformName(r.Name)
	}
	if r.RequiredRoles != nil {
		result.RequiredRoles = r.RequiredRoles
	}
	if r.RequiredClaims != nil {
		result.RequiredClaims = r.RequiredClaims
	}
	if r.Tags != nil {
		result.Tags = r.Tags
	}

	_, file := filepath.Split(r.LocalPath)
	ext := filepath.Ext(file)
	result.Type = ext[1:]

	isDir, err := helpers.IsDirectory(r.LocalPath)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, fmt.Errorf("the path %v is not a directory", r.LocalPath)
	}

	files, err := s.getManifestFiles(r.LocalPath, "")
	if err != nil {
		return nil, err
	}

	result.Contents = files
	return &result, nil
}

func (s *CatalogManifestService) getManifestFiles(path string, relativePath string) ([]models.VirtualMachineManifestContentItem, error) {
	if relativePath == "" {
		relativePath = "/"
	}

	result := make([]models.VirtualMachineManifestContentItem, 0)
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fullPath := filepath.Join(path, file.Name())
		if file.IsDir() {
			result = append(result, models.VirtualMachineManifestContentItem{
				IsDir: true,
				Name:  file.Name(),
				Path:  relativePath,
			})
			files, err := s.getManifestFiles(filepath.Join(path, file.Name()), filepath.Join(relativePath, file.Name()))
			if err != nil {
				return nil, err
			}
			result = append(result, files...)
			continue
		}

		manifestFile := models.VirtualMachineManifestContentItem{
			Path: relativePath,
		}
		fileInfo, err := file.Info()
		if err != nil {
			return nil, err
		}
		manifestFile.Name = file.Name()
		manifestFile.Size = fileInfo.Size()
		manifestFile.CreatedAt = fileInfo.ModTime().Format(time.RFC3339Nano)
		manifestFile.UpdatedAt = fileInfo.ModTime().Format(time.RFC3339Nano)
		checksum, err := helpers.GetFileChecksum(fullPath)
		if err != nil {
			return nil, err
		}
		manifestFile.Checksum = checksum
		result = append(result, manifestFile)
	}

	return result, nil
}

func (s *CatalogManifestService) readManifestFromFile(path string) (*models.VirtualMachineManifest, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}

	manifestBytes, err := helper.ReadFromFile(path)
	if err != nil {
		return nil, err
	}

	manifest := &models.VirtualMachineManifest{}
	err = json.Unmarshal(manifestBytes, manifest)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func (s *CatalogManifestService) getConformName(name string) string {
	replaceChars := []string{" ", "_", "-", ".", "$", "@", "\"", "'", "{", "}", "[", "]", "+", "!", "#", "%", "^", "&", "*", "(", ")", "=", ",", "<", ">", "?", "/", "\\", "|", "~", "`", ":", ";"}
	for _, replaceChar := range replaceChars {
		name = strings.ReplaceAll(name, replaceChar, "_")
	}

	return name
}

func (s *CatalogManifestService) getMetaFilename(name string) string {
	name = s.getConformName(name)
	if !strings.HasSuffix(name, ".meta") {
		name = name + ".meta"
	}

	return name
}

func (s *CatalogManifestService) parseProviderMetadata(connection string, meta map[string]string) string {
	newConnection := make(map[string]string)
	parts := strings.Split(connection, ";")
	for _, part := range parts {
		part := strings.TrimSpace(part)
		kv := strings.Split(part, "=")
		if len(kv) == 2 {
			newConnection[kv[0]] = kv[1]
		}
	}

	for k, v := range meta {
		newConnection[k] = v
	}

	result := ""
	for k, v := range newConnection {
		result = result + fmt.Sprintf("%s=%s;", k, v)
	}
	result = strings.TrimSuffix(result, ";")

	return result
}
