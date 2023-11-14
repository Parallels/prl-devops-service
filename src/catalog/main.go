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
	"github.com/Parallels/pd-api-service/helpers"

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

func (s *CatalogManifestService) GetProviders(ctx basecontext.ApiContext) []interfaces.RemoteStorageService {
	return s.remoteServices
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

func (s *CatalogManifestService) GenerateManifestContent(ctx basecontext.ApiContext, r *models.PushCatalogManifestRequest, manifest *models.VirtualMachineCatalogManifest) error {
	ctx.LogInfo("Generating manifest content for %v", r.CatalogId)
	if manifest == nil {
		manifest = models.NewVirtualMachineCatalogManifest()
	}

	manifest.CleanupRequest = cleanupservice.NewCleanupRequest()
	manifest.CreatedAt = helpers.GetUtcCurrentDateTime()
	manifest.UpdatedAt = helpers.GetUtcCurrentDateTime()

	manifest.Name = r.CatalogId
	manifest.Path = r.LocalPath
	manifest.ID = helpers.GenerateId()
	manifest.CatalogId = helpers.NormalizeString(r.CatalogId)
	manifest.Description = r.Description
	manifest.Version = helpers.NormalizeString(r.Version)
	manifest.Name = fmt.Sprintf("%v-%v", manifest.CatalogId, manifest.Version)
	manifestPackFileName := s.getPackFilename(manifest.Name)

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

	ctx.LogInfo("Getting manifest files for %v", r.CatalogId)
	files, err := s.getManifestFiles(r.LocalPath, "")
	if err != nil {
		return err
	}

	ctx.LogInfo("Compressing manifest files for %v", r.CatalogId)
	packFilePath, err := s.compressMachine(ctx, r.LocalPath, manifestPackFileName, "/tmp")
	if err != nil {
		return err
	}

	// Adding the zip file to the cleanup request
	manifest.CleanupRequest.AddLocalFileCleanupOperation(packFilePath, false)
	manifest.CompressedPath = packFilePath
	manifest.PackFile = "/tmp/" + manifestPackFileName

	fileInfo, err := os.Stat(packFilePath)
	if err != nil {
		return err
	}

	manifest.Size = fileInfo.Size()

	ctx.LogInfo("Getting manifest package checksum for %v", r.CatalogId)
	checksum, err := helpers.GetFileMD5Checksum(packFilePath)
	if err != nil {
		return err
	}
	manifest.CompressedChecksum = checksum

	manifest.VirtualMachineContents = files
	ctx.LogInfo("Finished generating manifest content for %v", r.CatalogId)
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
	return helpers.NormalizeString(name)
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

func (s *CatalogManifestService) compressMachine(ctx basecontext.ApiContext, path string, machineFileName string, destination string) (string, error) {
	tarFilename := machineFileName
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
