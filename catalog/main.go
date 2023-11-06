package catalog

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog/cleanupservice"
	"github.com/Parallels/pd-api-service/catalog/interfaces"
	"github.com/Parallels/pd-api-service/catalog/models"
	"github.com/Parallels/pd-api-service/catalog/providers/aws_s3_bucket"
	"github.com/Parallels/pd-api-service/catalog/providers/azurestorageaccount"
	"github.com/Parallels/pd-api-service/catalog/providers/local"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/Parallels/pd-api-service/mappers"
	"github.com/Parallels/pd-api-service/serviceprovider"

	"github.com/cjlapao/common-go/helper"
)

type CatalogManifestService struct {
	remoteServices []interfaces.RemoteStorageService
}

func NewManifestService(ctx basecontext.ApiContext) *CatalogManifestService {
	manifestService := &CatalogManifestService{}
	manifestService.remoteServices = make([]interfaces.RemoteStorageService, 0)
	manifestService.AddRemoteService(aws_s3_bucket.NewAwsS3RemoteService())
	manifestService.AddRemoteService(local.NewLocalProviderService())
	manifestService.AddRemoteService(azurestorageaccount.NewAzureStorageAccountRemoteService())
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

func (s *CatalogManifestService) PushOld(ctx basecontext.ApiContext, r *models.PushCatalogManifestRequest) (*models.VirtualMachineCatalogManifest, error) {
	executed := false
	var manifest *models.VirtualMachineCatalogManifest
	var err error
	for _, rs := range s.remoteServices {
		check, checkErr := rs.Check(ctx, r.Connection)
		if checkErr != nil {
			ctx.LogError("Error checking remote service %v: %v", rs.Name(), checkErr)
			return nil, checkErr
		}

		if check {
			executed = true
			err = s.GenerateManifestContent(ctx, r, manifest)
			if err != nil {
				ctx.LogError("Error generating manifest for %v: %v", r.Name, err)
				return nil, err
			}
			manifest.Provider = &models.CatalogManifestProvider{
				Type: rs.Name(),
				Meta: rs.GetProviderMeta(ctx),
			}

			ctx.LogInfo("Pushing manifest %v to provider %s", manifest.Name, rs.Name())
			var catalogManifest *models.VirtualMachineCatalogManifest
			if err := rs.PullFile(ctx, rs.GetProviderRootPath(ctx), s.getMetaFilename(manifest.ID), "/tmp"); err == nil {
				ctx.LogInfo("Remote Manifest found, retrieving it")
				catalogManifest, err = s.readManifestFromFile(filepath.Join("/tmp", s.getMetaFilename(manifest.ID)))
				if err != nil {
					ctx.LogError("Error reading manifest from file %v: %v", filepath.Join("/tmp", s.getMetaFilename(manifest.ID)), err)
					if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
						ctx.LogError("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
						return nil, err
					}

					return nil, err
				}
				if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
					ctx.LogError("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}
				manifest.CreatedAt = catalogManifest.CreatedAt
				manifest.RequiredRoles = catalogManifest.RequiredRoles
				manifest.RequiredClaims = catalogManifest.RequiredClaims
			}

			db := serviceprovider.Get().JsonDatabase
			if catalogManifest == nil {
				exists, _ := db.GetCatalogManifest(ctx, manifest.ID)
				if exists != nil {
					ctx.LogError("Manifest %v already exists in db", manifest.ID)
					return nil, fmt.Errorf("manifest %v already exists in db", manifest.ID)
				}

				ctx.LogInfo("Remote Manifest not found, creating it")
				if err := rs.CreateFolder(ctx, manifest.ID, "/"); err != nil {
					return nil, err
				}
				ctx.LogInfo("Pushing manifest files %v", s.getMetaFilename(manifest.ID))
				// for _, file := range manifest.Contents {
				// 	destinationPath := filepath.Join(manifest.ID, file.Path)
				// 	sourcePath := filepath.Join(manifest.Path, file.Path)
				// 	if file.IsDir {
				// 		if err := rs.CreateFolder(ctx, destinationPath, file.Name); err != nil {
				// 			return nil, err
				// 		}
				// 	} else {
				// 		manifest.Size += file.Size
				// 		if err := rs.PushFile(ctx, sourcePath, destinationPath, file.Name); err != nil {
				// 			return nil, err
				// 		}
				// 	}
				// }
				ctx.LogInfo("Finished pushing %d files for manifest %v", len(manifest.VirtualMachineContents), s.getMetaFilename(manifest.ID))

				ctx.LogInfo("Pushing manifest %v", manifest.Name)
				manifest.Path = filepath.Join(rs.GetProviderRootPath(ctx), manifest.ID)
				manifest.MetadataFile = filepath.Join(rs.GetProviderRootPath(ctx), s.getMetaFilename(manifest.ID))
				manifestContent, err := json.MarshalIndent(manifest, "", "  ")
				if err != nil {
					ctx.LogError("Error marshalling manifest %v: %v", manifest, err)
					return nil, err
				}

				if err := helper.WriteToFile(string(manifestContent), "/tmp/"+s.getMetaFilename(manifest.ID)); err != nil {
					ctx.LogError("Error writing manifest to temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}

				if err := rs.PushFile(ctx, "/tmp", "/", s.getMetaFilename(manifest.ID)); err != nil {
					ctx.LogError("Error pushing manifest to remote storage %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
						ctx.LogError("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
						return nil, err
					}

					return nil, err
				}

				if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
					ctx.LogError("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}

				dto := mappers.CatalogManifestToDto(*manifest)

				if err := db.CreateCatalogManifest(ctx, dto); err != nil {
					ctx.LogError("Error adding manifest to database %v: %v", manifest.Name, err)
					return nil, err
				}
				ctx.LogInfo("Finished pushing manifest %v", manifest.Name)
				return manifest, nil
			} else {
				ctx.LogInfo("Remote Manifest found, checking for changes in files")
				for _, file := range manifest.VirtualMachineContents {
					destinationPath := filepath.Join(manifest.ID, file.Path)
					sourcePath := filepath.Join(manifest.Path, file.Path)
					if file.IsDir {
						exists, err := rs.FolderExists(ctx, destinationPath, file.Name)
						if err != nil {
							ctx.LogError("Error checking if folder exists %v: %v", destinationPath, err)
							return nil, err
						}
						if !exists {
							if err := rs.CreateFolder(ctx, destinationPath, file.Name); err != nil {
								ctx.LogError("Error creating folder %v: %v", destinationPath, err)
								return nil, err
							}
						}
					} else {
						manifest.Size += file.Size
						exists, err := rs.FileExists(ctx, destinationPath, file.Name)
						if err != nil {
							ctx.LogError("Error checking if file exists %v: %v", destinationPath, err)
							return nil, err
						}
						if !exists {
							if err := rs.PushFile(ctx, sourcePath, destinationPath, file.Name); err != nil {
								ctx.LogError("Error pushing file %v: %v", destinationPath, err)
								return nil, err
							}
						} else {
							checksum, err := rs.FileChecksum(ctx, destinationPath, file.Name)
							if err != nil {
								ctx.LogError("Error getting file checksum %v: %v", destinationPath, err)
								return nil, err
							}
							if checksum != file.Checksum {
								if err := rs.DeleteFile(ctx, destinationPath, file.Name); err != nil {
									ctx.LogError("Error deleting file %v: %v", destinationPath, err)
									return nil, err
								}
								if err := rs.PushFile(ctx, sourcePath, destinationPath, file.Name); err != nil {
									ctx.LogError("Error pushing file %v: %v", destinationPath, err)
									return nil, err
								}
							} else {
								ctx.LogInfo("File %v is up to date", filepath.Join(destinationPath, file.Name))
							}
						}
					}
				}

				ctx.LogInfo("Finished pushing %v files for manifest %v", len(manifest.VirtualMachineContents), s.getMetaFilename(manifest.ID))

				if err := rs.DeleteFile(ctx, "/", s.getMetaFilename(manifest.ID)); err != nil {
					ctx.LogError("Error deleting manifest %v: %v", s.getMetaFilename(manifest.ID), err)
					return nil, err
				}

				ctx.LogInfo("Pushing new manifest %v", manifest.Name)
				manifest.Path = filepath.Join(rs.GetProviderRootPath(ctx), manifest.ID)
				manifest.MetadataFile = filepath.Join(rs.GetProviderRootPath(ctx), s.getMetaFilename(manifest.ID))
				manifestContent, err := json.MarshalIndent(manifest, "", "  ")
				if err != nil {
					ctx.LogError("Error marshalling manifest %v: %v", manifest.Name, err)
					return nil, err
				}

				if err := helper.WriteToFile(string(manifestContent), "/tmp/"+s.getMetaFilename(manifest.ID)); err != nil {
					ctx.LogError("Error writing manifest to temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}

				if err := rs.PushFile(ctx, "/tmp", "/", s.getMetaFilename(manifest.ID)); err != nil {
					ctx.LogError("Error pushing manifest to remote storage %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
						ctx.LogError("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
						return nil, err
					}

					return nil, err
				}

				if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
					ctx.LogError("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}

				exists, _ := db.GetCatalogManifest(ctx, manifest.ID)
				if exists != nil {
					ctx.LogInfo("Updating manifest %v", manifest.Name)
					dto := mappers.CatalogManifestToDto(*manifest)
					if err := db.UpdateCatalogManifest(ctx, dto); err != nil {
						ctx.LogError("Error updating manifest %v: %v", manifest.Name, err)
						return nil, err
					}
				} else {
					ctx.LogInfo("Creating manifest %v", manifest.Name)
					dto := mappers.CatalogManifestToDto(*manifest)
					if err := db.CreateCatalogManifest(ctx, dto); err != nil {
						ctx.LogError("Error creating manifest %v: %v", manifest.Name, err)
						return nil, err
					}
				}

				ctx.LogInfo("Finished pushing manifest %v", manifest.Name)
				return manifest, nil
			}
		}
	}

	if !executed {
		return nil, fmt.Errorf("no remote service found for connection %v", r.Connection)
	}

	return manifest, nil
}

func (s *CatalogManifestService) PullOld(ctx basecontext.ApiContext, r *models.PullCatalogManifestRequest) (*models.PullCatalogManifestResponse, error) {
	executed := false
	db := serviceprovider.Get().JsonDatabase
	if db == nil {
		return nil, errors.New("no database connection")
	}
	if err := db.Connect(ctx); err != nil {
		return nil, err
	}

	connectionString := ""
	var manifest *models.VirtualMachineCatalogManifest
	dbManifest, err := db.GetCatalogManifest(ctx, r.ID)
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
		check, checkErr := rs.Check(ctx, connectionString)
		if checkErr != nil {
			ctx.LogError("Error checking remote service %v: %v", rs.Name(), checkErr)
			return nil, checkErr
		}

		if check {
			var catalogManifest *models.VirtualMachineCatalogManifest
			if err := rs.PullFile(ctx, filepath.Dir(manifest.MetadataFile), s.getMetaFilename(manifest.ID), "/tmp"); err == nil {
				ctx.LogInfo("Remote Manifest found, retrieving it")
				catalogManifest, err = s.readManifestFromFile(filepath.Join("/tmp", s.getMetaFilename(manifest.ID)))
				if err != nil {
					ctx.LogError("Error reading manifest from file %v: %v", filepath.Join("/tmp", s.getMetaFilename(manifest.ID)), err)
					if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
						ctx.LogError("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
						return nil, err
					}

					return nil, err
				}
				if err := helper.DeleteFile("/tmp/" + s.getMetaFilename(manifest.ID)); err != nil {
					ctx.LogError("Error deleting temporary file %v: %v", "/tmp/"+s.getMetaFilename(manifest.ID), err)
					return nil, err
				}
			}

			if catalogManifest == nil {
				return nil, errors.New("manifest not found in the provider")
			}

			ctx.LogInfo("Pulling manifest %v", catalogManifest.Name)
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

			for _, file := range catalogManifest.VirtualMachineContents {
				destinationPath := filepath.Join(localMachineFolder, file.Path)
				fullSrcPath := filepath.Join(catalogManifest.Path, file.Path)
				if file.IsDir {
					dirFolder := filepath.Join(destinationPath, file.Name)
					if err := helpers.CreateDirIfNotExist(dirFolder); err != nil {
						ctx.LogError("Error creating directory %v: %v", dirFolder, err)
						return nil, err
					}
					ctx.LogInfo("Created directory %v", fullSrcPath)
				} else {
					if err := rs.PullFile(ctx, fullSrcPath, file.Name, destinationPath); err != nil {
						ctx.LogError("Error pulling file %v: %v", filepath.Join(destinationPath, file.Name), err)
						return nil, err
					}
					ctx.LogInfo("Pulled file %v", filepath.Join(destinationPath, file.Name))
				}
			}

			ctx.LogInfo("Finished pulling %v files for manifest %v", len(catalogManifest.VirtualMachineContents), s.getMetaFilename(catalogManifest.ID))
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

func (s *CatalogManifestService) GenerateManifestContent(ctx basecontext.ApiContext, r *models.PushCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest) error {
	ctx.LogInfo("Generating manifest content for %v", r.Name)
	if manifest == nil {
		manifest = models.NewVirtualMachineCatalogManifest()
	}

	manifest.CleanupRequest = cleanupservice.NewCleanupRequest()
	manifest.CreatedAt = helpers.GetUtcCurrentDateTime()
	manifest.UpdatedAt = helpers.GetUtcCurrentDateTime()

	manifest.Name = r.Name
	manifest.Path = r.LocalPath
	if r.Uuid != "" {
		manifest.ID = s.getConformName(r.Uuid)
	} else {
		manifest.ID = s.getConformName(r.Name)
	}
	if r.RequiredRoles != nil {
		manifest.RequiredRoles = r.RequiredRoles
	}
	if r.RequiredClaims != nil {
		manifest.RequiredClaims = r.RequiredClaims
	}
	if r.Tags != nil {
		manifest.Tags = r.Tags
	}

	_, file := filepath.Split(r.LocalPath)
	ext := filepath.Ext(file)
	manifest.Type = ext[1:]

	isDir, err := helpers.IsDirectory(r.LocalPath)
	if err != nil {
		return err
	}
	if !isDir {
		return fmt.Errorf("the path %v is not a directory", r.LocalPath)
	}

	ctx.LogInfo("Getting manifest files for %v", r.Name)
	files, err := s.getManifestFiles(r.LocalPath, "")
	if err != nil {
		return err
	}

	ctx.LogInfo("Compressing manifest files for %v", r.Name)
	packFilePath, err := s.compressMachine(ctx, r.LocalPath, manifest.ID, "/tmp")
	if err != nil {
		return err
	}

	// Adding the zip file to the cleanup request
	manifest.CleanupRequest.AddLocalFileCleanupOperation(packFilePath, false)
	manifest.CompressedPath = packFilePath
	manifest.PackFile = "/tmp/" + manifest.ID + ".pdpack"

	fileInfo, err := os.Stat(packFilePath)
	if err != nil {
		return err
	}

	manifest.Size = fileInfo.Size()

	ctx.LogInfo("Getting manifest package checksum for %v", r.Name)
	checksum, err := helpers.GetFileMD5Checksum(packFilePath)
	if err != nil {
		return err
	}
	manifest.CompressedChecksum = checksum

	manifest.VirtualMachineContents = files
	ctx.LogInfo("Finished generating manifest content for %v", r.Name)
	return nil
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
		// fullPath := filepath.Join(path, file.Name())
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
		// checksum, err := helpers.GetFileMD5Checksum(fullPath)
		// if err != nil {
		// 	return nil, err
		// }
		// manifestFile.Checksum = checksum
		result = append(result, manifestFile)
	}

	return result, nil
}

func (s *CatalogManifestService) readManifestFromFile(path string) (*models.VirtualMachineCatalogManifest, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}

	manifestBytes, err := helper.ReadFromFile(path)
	if err != nil {
		return nil, err
	}

	manifest := &models.VirtualMachineCatalogManifest{}
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

func (s *CatalogManifestService) getPackFilename(name string) string {
	name = s.getConformName(name)
	if !strings.HasSuffix(name, ".pdpack") {
		name = name + ".pdpack"
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

func (s *CatalogManifestService) compressMachine(ctx basecontext.ApiContext, path string, machineName string, destination string) (string, error) {
	tarFilename := s.getPackFilename(machineName)
	tarFilePath := filepath.Join(destination, tarFilename)

	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		return "", err
	}
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	countFiles := 0
	if err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		countFiles += 1
		return nil
	}); err != nil {
		return "", err
	}

	compressed := 1
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		ctx.LogInfo("[%v/%v] Compressing file %v", compressed, countFiles, filePath)
		compressed += 1
		if err != nil {
			return err
		}

		if info.IsDir() {
			compressed -= 1
			return nil
		}

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		relPath := strings.TrimPrefix(filePath, path)
		hdr := &tar.Header{
			Name: relPath,
			Mode: int64(info.Mode()),
			Size: info.Size(),
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			return err
		}

		_, err = io.Copy(tarWriter, f)
		return err
	})

	if err != nil {
		return "", err
	}

	return tarFilePath, nil
}

func (s *CatalogManifestService) decompressMachine(ctx basecontext.ApiContext, filePath string, destination string) error {
	tarFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		filePath := filepath.Join(destination, header.Name)
		// Creating the basedir if it does not exist
		baseDir := filepath.Dir(filePath)
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			if err := os.MkdirAll(baseDir, 0755); err != nil {
				return err
			}
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				if err := os.MkdirAll(filePath, os.FileMode(header.Mode)); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
		}
	}

	return nil
}
